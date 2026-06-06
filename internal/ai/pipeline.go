// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"context"
	"fmt"
	"io"
)

// validation and repair pipeline for JSON outputs

const MaxRepairAttempts = 3

type Pipeline[T any] struct {
	provider    Provider
	validator   StructValidator[T]
	repairer    Repairer
	schema      Schema
	debugWriter io.Writer
}

func NewPipeline[T any](provider Provider, schema Schema, validator StructValidator[T]) Pipeline[T] {
	return Pipeline[T]{
		provider:  provider,
		validator: validator,
		repairer:  Repairer{},
		schema:    schema,
	}
}

func (p Pipeline[T]) WithDebug(w io.Writer) Pipeline[T] {
	p.debugWriter = w
	return p
}

type RunOptions struct {
	OnFirstToken  func()
	OnRepairStart func(attempt int)
}

func (p Pipeline[T]) Run(ctx context.Context, generationPrompt string, opts RunOptions) (T, error) {
	raw, err := p.provider.GenerateJSON(ctx, generationPrompt, opts.OnFirstToken)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("pipeline: generate: %w", err)
	}

	vr := p.validator.Validate(raw)
	p.debugf("[pipeline] initial response (%d bytes):\n%s\n[pipeline] validation: %s\n",
		len(raw), raw, validationSummary(vr))

	if vr.Valid() {
		return p.mustDecode(raw)
	}

	current := raw
	lastVR := vr

	for attempt := range MaxRepairAttempts {
		if err := ctx.Err(); err != nil {
			var zero T
			return zero, fmt.Errorf("pipeline: cancelled before repair attempt %d: %w", attempt+1, err)
		}
		if opts.OnRepairStart != nil {
			opts.OnRepairStart(attempt + 1)
		}

		repairPrompt := p.selectRepairPrompt(lastVR, generationPrompt, current)
		p.debugf("[pipeline] repair attempt %d (%s) — prompt:\n%s\n",
			attempt+1, repairKind(lastVR), repairPrompt)

		repaired, err := p.provider.GenerateJSON(ctx, repairPrompt, nil)
		if err != nil {
			var zero T
			return zero, fmt.Errorf("pipeline: repair attempt %d: %w", attempt+1, err)
		}

		lastVR = p.validator.Validate(repaired)
		p.debugf("[pipeline] repair %d response (%d bytes):\n%s\n[pipeline] validation: %s\n",
			attempt+1, len(repaired), repaired, validationSummary(lastVR))

		if lastVR.Valid() {
			return p.mustDecode(repaired)
		}

		current = repaired
	}

	var zero T
	return zero, fmt.Errorf(
		"pipeline: output still invalid after %d repair attempt(s): %w",
		MaxRepairAttempts,
		lastVR,
	)
}

func (p Pipeline[T]) selectRepairPrompt(vr *ValidationResult, generationPrompt, current string) string {
	if vr.StructurallyValid() {
		return p.repairer.BuildSemanticRepairPrompt(p.schema, generationPrompt, vr.SemanticErrors())
	}
	if isEssentiallyEmpty(current) {
		return p.repairer.BuildStructuralRepairPromptWithContext(
			p.schema, current, vr.StructuralErrors(), generationPrompt,
		)
	}
	return p.repairer.BuildStructuralRepairPrompt(p.schema, current, vr.StructuralErrors())
}

func isEssentiallyEmpty(raw string) bool {
	trimmed := stripWhitespace(raw)
	return trimmed == "" || trimmed == "{}" || trimmed == "null" || trimmed == "[]"
}

func stripWhitespace(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		b = append(b, c)
	}
	return string(b)
}

func (p Pipeline[T]) mustDecode(raw string) (T, error) {
	value, err := p.validator.Decode(raw)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("pipeline: decode after validation: %w (this is a validator bug)", err)
	}
	return value, nil
}

func (p Pipeline[T]) debugf(format string, args ...any) {
	if p.debugWriter != nil {
		fmt.Fprintf(p.debugWriter, format, args...)
	}
}

func validationSummary(vr *ValidationResult) string {
	if vr.Valid() {
		return "OK"
	}
	return fmt.Sprintf("INVALID (%d error(s)): %v", len(vr.Errors()), vr.Errors())
}

func repairKind(vr *ValidationResult) string {
	if vr.StructurallyValid() {
		return "semantic"
	}
	return "structural"
}
