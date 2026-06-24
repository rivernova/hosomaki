// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/rivernova/hosomaki/internal/stream"
)

// stream pipeline for the LLM

var ErrIncomplete = errors.New("llm response incomplete")

type StreamOptions struct {
	OnFirstToken  func()
	OnItem        func(key, raw string)
	OnRepairStart func(attempt int)
}

type ElementChecker func(raw json.RawMessage) []string

func ElementCheck[E any](check func(E) []string) ElementChecker {
	return func(raw json.RawMessage) []string {
		var value E
		if err := json.Unmarshal(raw, &value); err != nil {
			return []string{fmt.Sprintf("element decode: %v", err)}
		}
		return check(value)
	}
}

type StreamPipeline[T any] struct {
	provider      Provider
	validator     StructValidator[T]
	repairer      Repairer
	schema        Schema
	elementChecks map[string]ElementChecker
	enums         map[string][]string
	debugWriter   io.Writer
}

func NewStreamPipeline[T any](provider Provider, schema Schema, validator StructValidator[T]) StreamPipeline[T] {
	return StreamPipeline[T]{
		provider:  provider,
		validator: validator,
		repairer:  Repairer{},
		schema:    schema,
	}
}

func (p StreamPipeline[T]) WithElementCheck(field string, check ElementChecker) StreamPipeline[T] {
	next := make(map[string]ElementChecker, len(p.elementChecks)+1)
	for k, v := range p.elementChecks {
		next[k] = v
	}
	next[field] = check
	p.elementChecks = next
	return p
}

func (p StreamPipeline[T]) WithEnum(field string, allowed ...string) StreamPipeline[T] {
	next := make(map[string][]string, len(p.enums)+1)
	for k, v := range p.enums {
		next[k] = v
	}
	next[field] = allowed
	p.enums = next
	return p
}

func (p StreamPipeline[T]) WithDebug(w io.Writer) StreamPipeline[T] {
	p.debugWriter = w
	return p
}

func (p StreamPipeline[T]) Run(ctx context.Context, generationPrompt string, opts StreamOptions) (T, error) {
	var zero T

	elemTypes := arrayElemTypes(reflect.TypeFor[T]())
	known := knownFieldNames(reflect.TypeFor[T]())

	emit := func(key, raw string) {
		if opts.OnItem != nil {
			opts.OnItem(key, raw)
		}
	}

	collectedElems := make(map[string][]string)
	collectedScalars := make(map[string]string)

	collect := func(key, raw string) {
		if et, isArray := elemTypes[key]; isArray {
			if !looksLikeObject(raw) {
				return
			}
			collectedElems[key] = append(collectedElems[key], raw)
			if opts.OnItem != nil {
				if norm, vr := p.prepareElement(key, et, raw); vr.Valid() {
					emit(key, norm)
				}
			}
			return
		}
		if _, ok := known[key]; ok {
			collectedScalars[key] = raw
			if opts.OnItem != nil {
				emit(key, raw)
			}
		}
	}

	scanner := stream.NewArrayItemScanner(collect)

	raw, err := p.provider.GenerateStream(ctx, generationPrompt, opts.OnFirstToken, scanner)
	if err != nil {
		return zero, fmt.Errorf("stream pipeline: generate: %w", err)
	}
	p.debugf("[stream] full response (%d bytes):\n%s\n", len(raw), raw)

	if len(collectedElems) == 0 && len(collectedScalars) == 0 {
		p.debugf("[stream] nothing streamed cleanly — whole-document repair\n")
		return p.recoverWholeDocument(ctx, generationPrompt, raw, elemTypes, known, emit, opts)
	}

	assembled, dropped, err := p.assembleCollected(ctx, collectedElems, collectedScalars, elemTypes, emit, opts)
	if err != nil {
		return zero, err
	}

	result, err := p.finalize(ctx, generationPrompt, assembled, opts)
	if err != nil {
		if errors.Is(err, ErrIncomplete) {
			return result, ErrIncomplete
		}
		return zero, err
	}
	if dropped > 0 {
		return result, ErrIncomplete
	}
	return result, nil
}

func (p StreamPipeline[T]) assembleCollected(
	ctx context.Context,
	collectedElems map[string][]string,
	collectedScalars map[string]string,
	elemTypes map[string]reflect.Type,
	emit func(key, raw string),
	opts StreamOptions,
) (map[string]json.RawMessage, int, error) {
	assembled := make(map[string]json.RawMessage, len(elemTypes)+len(collectedScalars))
	dropped := 0

	for name := range elemTypes {
		assembled[name] = json.RawMessage("[]")
	}
	for name, raw := range collectedScalars {
		assembled[name] = json.RawMessage(raw)
	}

	for name, elems := range collectedElems {
		et := elemTypes[name]
		verified := make([]json.RawMessage, 0, len(elems))
		for _, elem := range elems {
			norm, vr := p.prepareElement(name, et, elem)
			if vr.Valid() {
				verified = append(verified, json.RawMessage(norm))
				continue
			}

			if err := ctx.Err(); err != nil {
				return nil, dropped, fmt.Errorf("stream pipeline: cancelled before element repair: %w", err)
			}

			repaired, ok, err := p.repairElement(ctx, name, norm, vr, et, opts)
			if err != nil {
				return nil, dropped, err
			}
			if !ok {
				p.debugf("[stream] dropping unrepairable element in %q\n", name)
				dropped++
				continue
			}
			verified = append(verified, json.RawMessage(repaired))
			emit(name, repaired)
		}

		arr, err := json.Marshal(verified)
		if err != nil {
			return nil, dropped, fmt.Errorf("stream pipeline: marshal %q: %w", name, err)
		}
		assembled[name] = arr
	}

	return assembled, dropped, nil
}

func (p StreamPipeline[T]) recoverWholeDocument(
	ctx context.Context,
	generationPrompt, raw string,
	elemTypes map[string]reflect.Type,
	known map[string]struct{},
	emit func(key, raw string),
	opts StreamOptions,
) (T, error) {
	var zero T

	clean := raw
	if obj, ok := extractJSONObject(raw); ok {
		clean = obj
	}

	repaired, err := p.repairDocument(ctx, generationPrompt, clean, opts)
	if err != nil {
		if errors.Is(err, ErrIncomplete) {
			return p.bestEffortDecode(clean), ErrIncomplete
		}
		return zero, err
	}

	top, _ := parseObject(repaired)
	for name := range known {
		if _, isArray := elemTypes[name]; isArray {
			continue
		}
		if val, present := top[name]; present {
			emit(name, string(val))
		}
	}
	for name, et := range elemTypes {
		rawArr, present := top[name]
		if !present {
			continue
		}
		var elems []json.RawMessage
		if err := json.Unmarshal(rawArr, &elems); err != nil {
			continue
		}
		for _, elem := range elems {
			if norm, vr := p.prepareElement(name, et, string(elem)); vr.Valid() {
				emit(name, norm)
			}
		}
	}

	return p.decode(repaired)
}

func (p StreamPipeline[T]) verifyElement(field string, et reflect.Type, raw string) *ValidationResult {
	vr := validateElement(et, raw)
	if !vr.StructurallyValid() {
		return vr
	}
	if check := p.elementChecks[field]; check != nil {
		vr.semanticErrors = append(vr.semanticErrors, check(json.RawMessage(raw))...)
	}
	return vr
}

func (p StreamPipeline[T]) prepareElement(field string, et reflect.Type, raw string) (string, *ValidationResult) {
	norm := normalizeJSON(raw, et, p.enums)
	return norm, p.verifyElement(field, et, norm)
}

func (p StreamPipeline[T]) repairElement(
	ctx context.Context,
	key, invalid string,
	vr *ValidationResult,
	et reflect.Type,
	opts StreamOptions,
) (string, bool, error) {
	elemSchema := NewSchema(extractElementSchema(p.schema.String(), key))
	current := invalid
	lastVR := vr

	for attempt := range MaxRepairAttempts {
		if err := ctx.Err(); err != nil {
			return "", false, fmt.Errorf("stream pipeline: cancelled before element repair attempt %d: %w", attempt+1, err)
		}
		if opts.OnRepairStart != nil {
			opts.OnRepairStart(attempt + 1)
		}

		prompt := p.repairer.BuildElementRepairPrompt(elemSchema, current, lastVR.Errors())
		p.debugf("[stream] repair element %q attempt %d\n", key, attempt+1)

		repaired, err := p.provider.GenerateJSON(ctx, prompt, nil)
		if err != nil {
			if ctx.Err() != nil {
				return "", false, fmt.Errorf("stream pipeline: cancelled during element repair: %w", ctx.Err())
			}
			p.debugf("[stream] element %q repair attempt %d failed (%v) — moving on\n", key, attempt+1, err)
			return "", false, nil
		}

		repaired = unwrapElement(repaired, key)
		repaired = normalizeJSON(repaired, et, p.enums)
		lastVR = p.verifyElement(key, et, repaired)
		if lastVR.Valid() {
			return repaired, true, nil
		}
		current = repaired
	}

	return "", false, nil
}

func (p StreamPipeline[T]) repairDocument(ctx context.Context, generationPrompt, raw string, opts StreamOptions) (string, error) {
	docType := reflect.TypeFor[T]()
	current := normalizeJSON(raw, docType, p.enums)
	lastVR := p.validator.Validate(current)
	if lastVR.Valid() {
		return current, nil
	}

	for attempt := range MaxRepairAttempts {
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("stream pipeline: cancelled before repair attempt %d: %w", attempt+1, err)
		}
		if opts.OnRepairStart != nil {
			opts.OnRepairStart(attempt + 1)
		}

		prompt := p.selectDocumentRepairPrompt(lastVR, generationPrompt, current)
		p.debugf("[stream] whole-document repair attempt %d (%s)\n", attempt+1, repairKind(lastVR))

		repaired, err := p.provider.GenerateJSON(ctx, prompt, nil)
		if err != nil {
			if ctx.Err() != nil {
				return "", fmt.Errorf("stream pipeline: cancelled during repair: %w", ctx.Err())
			}
			return "", fmt.Errorf("%w (repair call failed: %v)", ErrIncomplete, err)
		}

		current = normalizeJSON(repaired, docType, p.enums)
		lastVR = p.validator.Validate(current)
		p.debugf("[stream] whole-document repair %d: %s\n", attempt+1, validationSummary(lastVR))
		if lastVR.Valid() {
			return current, nil
		}
	}

	return "", fmt.Errorf(
		"%w (still invalid after %d repair attempt(s): %v)",
		ErrIncomplete, MaxRepairAttempts, lastVR,
	)
}

func (p StreamPipeline[T]) bestEffortDecode(raw string) T {
	value, err := p.validator.Decode(raw)
	if err != nil {
		var zero T
		return zero
	}
	return value
}

func (p StreamPipeline[T]) finalize(
	ctx context.Context,
	generationPrompt string,
	assembled map[string]json.RawMessage,
	opts StreamOptions,
) (T, error) {
	var zero T

	b, err := json.Marshal(assembled)
	if err != nil {
		return zero, fmt.Errorf("stream pipeline: marshal assembled result: %w", err)
	}

	repaired, err := p.repairDocument(ctx, generationPrompt, string(b), opts)
	if err != nil {
		if errors.Is(err, ErrIncomplete) {
			return p.bestEffortDecode(string(b)), ErrIncomplete
		}
		return zero, err
	}
	return p.decode(repaired)
}

func (p StreamPipeline[T]) selectDocumentRepairPrompt(vr *ValidationResult, generationPrompt, current string) string {
	if vr.StructurallyValid() {
		return p.repairer.BuildSemanticRepairPrompt(p.schema, generationPrompt, vr.SemanticErrors())
	}
	return p.repairer.BuildStructuralRepairPromptWithContext(
		p.schema, current, vr.StructuralErrors(), generationPrompt,
	)
}

func (p StreamPipeline[T]) decode(raw string) (T, error) {
	var zero T
	value, err := p.validator.Decode(raw)
	if err != nil {
		return zero, fmt.Errorf("stream pipeline: decode after validation: %w (this is a validator bug)", err)
	}
	return value, nil
}

func (p StreamPipeline[T]) debugf(format string, args ...any) {
	if p.debugWriter != nil {
		_, err := fmt.Fprintf(p.debugWriter, format, args...)
		if err != nil {
			return
		}
	}
}

func validateElement(et reflect.Type, raw string) *ValidationResult {
	result := &ValidationResult{}
	var elemMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &elemMap); err != nil {
		result.structuralErrors = append(result.structuralErrors,
			fmt.Sprintf("element is not a JSON object: %v", err))
		return result
	}
	validateStruct(et, elemMap, "", result)
	return result
}

func arrayElemTypes(t reflect.Type) map[string]reflect.Type {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	out := make(map[string]reflect.Type)
	if t.Kind() != reflect.Struct {
		return out
	}
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() || field.Type.Kind() != reflect.Slice {
			continue
		}
		name, _ := parseTag(field.Tag.Get("json"), field.Name)
		if name == "-" {
			continue
		}
		elem := field.Type.Elem()
		if elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Struct {
			continue
		}
		out[name] = elem
	}
	return out
}

func knownFieldNames(t reflect.Type) map[string]struct{} {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	out := make(map[string]struct{})
	if t.Kind() != reflect.Struct {
		return out
	}
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name, _ := parseTag(field.Tag.Get("json"), field.Name)
		if name == "-" {
			continue
		}
		out[name] = struct{}{}
	}
	return out
}

func parseObject(raw string) (map[string]json.RawMessage, bool) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, false
	}
	return m, true
}

func extractJSONObject(raw string) (string, bool) {
	start := strings.IndexByte(raw, '{')
	if start < 0 {
		return "", false
	}
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(raw); i++ {
		c := raw[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return raw[start : i+1], true
			}
		}
	}
	return "", false
}

func looksLikeObject(raw string) bool {
	return strings.HasPrefix(strings.TrimSpace(raw), "{")
}

func unwrapElement(raw, key string) string {
	m, ok := parseObject(raw)
	if !ok {
		return raw
	}
	val, present := m[key]
	if !present || len(m) != 1 {
		return raw
	}
	var elems []json.RawMessage
	if err := json.Unmarshal(val, &elems); err != nil || len(elems) != 1 {
		return raw
	}
	return string(elems[0])
}

func extractElementSchema(schema, key string) string {
	keyIdx := strings.Index(schema, `"`+key+`"`)
	if keyIdx < 0 {
		return schema
	}
	open := strings.IndexByte(schema[keyIdx:], '[')
	if open < 0 {
		return schema
	}
	rest := schema[keyIdx+open:]
	brace := strings.IndexByte(rest, '{')
	if brace < 0 {
		return schema
	}
	depth := 0
	for i := brace; i < len(rest); i++ {
		switch rest[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return rest[brace : i+1]
			}
		}
	}
	return schema
}
