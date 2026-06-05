// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"context"
	"fmt"
	"io"
)

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

// Run executes the full pipeline:
//
//  1. Call provider.GenerateJSON with the generation prompt.  onFirstToken
//     fires on the first streamed token so the caller's spinner can react.
//  2. Validate the result.  Return immediately if valid.
//  3. If invalid, choose the repair strategy based on failure kind and call
//     the provider again.  Re-validate.  Repeat up to MaxRepairAttempts times.
//  4. If still invalid after all attempts, return a descriptive error.
func (p Pipeline[T]) Run(ctx context.Context, generationPrompt string, onFirstToken func()) (T, error) {
	raw, err := p.provider.GenerateJSON(ctx, generationPrompt, onFirstToken)
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
	for attempt := range MaxRepairAttempts {
		repairPrompt := p.selectRepairPrompt(vr, generationPrompt, current)
		p.debugf("[pipeline] repair attempt %d (%s) — prompt:\n%s\n",
			attempt+1, repairKind(vr), repairPrompt)

		repaired, err := p.provider.GenerateJSON(ctx, repairPrompt, nil)
		if err != nil {
			var zero T
			return zero, fmt.Errorf("pipeline: repair attempt %d: %w", attempt+1, err)
		}

		vr = p.validator.Validate(repaired)
		p.debugf("[pipeline] repair %d response (%d bytes):\n%s\n[pipeline] validation: %s\n",
			attempt+1, len(repaired), repaired, validationSummary(vr))

		if vr.Valid() {
			return p.mustDecode(repaired)
		}

		current = repaired
	}

	var zero T
	return zero, fmt.Errorf(
		"pipeline: output still invalid after %d repair attempt(s): %w",
		MaxRepairAttempts,
		p.validator.Validate(current),
	)
}

func (p Pipeline[T]) selectRepairPrompt(vr *ValidationResult, generationPrompt, current string) string {
	if vr.StructurallyValid() {
		return p.repairer.BuildSemanticRepairPrompt(p.schema, generationPrompt, vr.SemanticErrors())
	}
	return p.repairer.BuildStructuralRepairPrompt(p.schema, current, vr.StructuralErrors())
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
