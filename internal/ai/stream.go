// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/rivernova/hosomaki/internal/stream"
)

// stream pipeline
type StreamOptions struct {
	OnFirstToken  func()
	OnItem        func(key, raw string)
	OnRepairStart func(attempt int)
}

type StreamPipeline[T any] struct {
	provider    Provider
	validator   StructValidator[T]
	repairer    Repairer
	schema      Schema
	debugWriter io.Writer
}

func NewStreamPipeline[T any](provider Provider, schema Schema, validator StructValidator[T]) StreamPipeline[T] {
	return StreamPipeline[T]{
		provider:  provider,
		validator: validator,
		repairer:  Repairer{},
		schema:    schema,
	}
}

func (p StreamPipeline[T]) WithDebug(w io.Writer) StreamPipeline[T] {
	p.debugWriter = w
	return p
}

func (p StreamPipeline[T]) Run(ctx context.Context, generationPrompt string, opts StreamOptions) (T, error) {
	scanner := stream.NewArrayItemScanner(func(key, raw string) {
		if opts.OnItem != nil {
			opts.OnItem(key, raw)
		}
	})

	raw, err := p.provider.GenerateStream(ctx, generationPrompt, opts.OnFirstToken, scanner)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("stream pipeline: generate: %w", err)
	}

	p.debugf("[stream] full response (%d bytes):\n%s\n", len(raw), raw)

	vr := p.validator.Validate(raw)
	p.debugf("[stream] validation: %s\n", validationSummary(vr))

	if vr.Valid() {
		return p.mustDecode(raw)
	}

	current := raw
	lastVR := vr

	for attempt := range MaxRepairAttempts {
		if err := ctx.Err(); err != nil {
			var zero T
			return zero, fmt.Errorf("stream pipeline: cancelled before repair attempt %d: %w", attempt+1, err)
		}
		if opts.OnRepairStart != nil {
			opts.OnRepairStart(attempt + 1)
		}

		repairPrompt := p.selectRepairPrompt(lastVR, generationPrompt, current)
		p.debugf("[stream] repair attempt %d (%s)\n", attempt+1, repairKind(lastVR))

		repaired, err := p.provider.GenerateJSON(ctx, repairPrompt, nil)
		if err != nil {
			var zero T
			return zero, fmt.Errorf("stream pipeline: repair attempt %d: %w", attempt+1, err)
		}

		lastVR = p.validator.Validate(repaired)
		p.debugf("[stream] repair %d (%d bytes): %s\n", attempt+1, len(repaired), validationSummary(lastVR))

		if lastVR.Valid() {
			return p.mustDecode(repaired)
		}

		current = repaired
	}

	var zero T
	return zero, fmt.Errorf(
		"stream pipeline: output still invalid after %d repair attempt(s): %w",
		MaxRepairAttempts,
		lastVR,
	)
}

func (p StreamPipeline[T]) selectRepairPrompt(vr *ValidationResult, generationPrompt, current string) string {
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

func (p StreamPipeline[T]) mustDecode(raw string) (T, error) {
	var value T
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		var zero T
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
