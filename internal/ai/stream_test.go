// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

// unit testing for the streaming response of the LLM

type streamSchema struct {
	Items []streamItem `json:"items"`
}

type streamItem struct {
	Name string `json:"name"`
}

const (
	streamSchemaStr  = `{"items":[{"name":"string"}]}`
	elementSchemaStr = `{"name":"string"}`
)

func newStreamPipe(p Provider) StreamPipeline[streamSchema] {
	return NewStreamPipeline(p, NewSchema(streamSchemaStr), StructValidator[streamSchema]{})
}

type scalarSchema struct {
	Summary  string       `json:"summary"`
	Findings []scalarItem `json:"findings"`
}

type scalarItem struct {
	Severity string `json:"severity"`
	Title    string `json:"title"`
}

const (
	scalarSchemaStr  = `{"summary":"string","findings":[{"severity":"string","title":"string"}]}`
	findingSchemaStr = `{"severity":"string","title":"string"}`
)

func scalarDocCheck(s scalarSchema) []string {
	var errs []string
	if strings.TrimSpace(s.Summary) == "" {
		errs = append(errs, "summary must not be empty")
	}
	for i, f := range s.Findings {
		for _, e := range scalarFindingCheck(f) {
			errs = append(errs, fmt.Sprintf("findings[%d].%s", i, e))
		}
	}
	return errs
}

func scalarFindingCheck(f scalarItem) []string {
	if f.Severity != "warning" && f.Severity != "info" {
		return []string{"severity must be 'warning' or 'info'"}
	}
	return nil
}

func newScalarPipe(p Provider) StreamPipeline[scalarSchema] {
	return NewStreamPipeline(
		p,
		NewSchema(scalarSchemaStr),
		StructValidator[scalarSchema]{SemanticCheck: scalarDocCheck},
	).WithElementCheck("findings", ElementCheck(scalarFindingCheck))
}

type scriptProvider struct {
	stream      string
	repair      func(prompt string) (string, error)
	streamErr   error
	streamCalls int
	repairCalls int
}

func (s *scriptProvider) Generate(_ context.Context, _ string) (string, error) { return "", nil }

func (s *scriptProvider) GenerateStream(_ context.Context, _ string, onFirstToken func(), w io.Writer) (string, error) {
	s.streamCalls++
	if s.streamErr != nil {
		return "", s.streamErr
	}
	if onFirstToken != nil && s.stream != "" {
		onFirstToken()
	}
	if w != nil {
		if _, err := io.WriteString(w, s.stream); err != nil {
			return "", err
		}
	}
	return s.stream, nil
}

func (s *scriptProvider) GenerateJSON(_ context.Context, prompt string, _ func()) (string, error) {
	s.repairCalls++
	if s.repair == nil {
		return "", errors.New("unexpected repair call")
	}
	return s.repair(prompt)
}

func collect(raws *[]string) func(key, raw string) {
	return func(_, raw string) { *raws = append(*raws, raw) }
}

func hasDuplicate(raws []string) bool {
	seen := make(map[string]struct{}, len(raws))
	for _, r := range raws {
		if _, ok := seen[r]; ok {
			return true
		}
		seen[r] = struct{}{}
	}
	return false
}

func TestStreamPipeline_EmitsValidItemsAsTheyArrive(t *testing.T) {
	p := &scriptProvider{stream: `{"items":[{"name":"alpha"},{"name":"beta"},{"name":"gamma"}]}`}

	var raws []string
	result, err := newStreamPipe(p).Run(context.Background(), "prompt", StreamOptions{OnItem: collect(&raws)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(raws) != 3 {
		t.Fatalf("expected 3 emitted items, got %d: %v", len(raws), raws)
	}
	if hasDuplicate(raws) {
		t.Fatalf("items emitted more than once: %v", raws)
	}
	if len(result.Items) != 3 {
		t.Fatalf("expected 3 items in result, got %d", len(result.Items))
	}
}

func TestStreamPipeline_DecodesFullResult(t *testing.T) {
	p := &scriptProvider{stream: `{"items":[{"name":"x"},{"name":"y"}]}`}
	result, err := newStreamPipe(p).Run(context.Background(), "prompt", StreamOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 || result.Items[0].Name != "x" || result.Items[1].Name != "y" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestStreamPipeline_ValidEmptyArray(t *testing.T) {
	p := &scriptProvider{stream: `{"items":[]}`}
	result, err := newStreamPipe(p).Run(context.Background(), "prompt", StreamOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Items == nil {
		t.Fatal("expected non-nil items slice")
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected empty slice, got %d", len(result.Items))
	}
}

func TestStreamPipeline_ErrorOnStreamGeneration(t *testing.T) {
	provErr := errors.New("stream failure")
	p := &scriptProvider{streamErr: provErr}
	_, err := newStreamPipe(p).Run(context.Background(), "prompt", StreamOptions{})
	if !errors.Is(err, provErr) {
		t.Fatalf("expected wrapped provErr, got: %v", err)
	}
}

func TestStreamPipeline_HoldsBackInvalidUntilRepaired(t *testing.T) {
	const task = "ORIGINAL_TASK_MARKER"
	p := &scriptProvider{
		stream: `{"items":[{"junk":true}]}`,
		repair: func(prompt string) (string, error) {
			if !strings.Contains(prompt, elementSchemaStr) {
				return "", errors.New("element repair prompt missing element schema")
			}
			if strings.Contains(prompt, task) {
				return "", errors.New("element repair prompt must stay item-scoped, not inject the whole task")
			}
			return `{"name":"repaired"}`, nil
		},
	}

	var raws []string
	result, err := newStreamPipe(p).Run(context.Background(), task, StreamOptions{OnItem: collect(&raws)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(raws) != 1 {
		t.Fatalf("expected exactly 1 emission (held back until verified), got %d: %v", len(raws), raws)
	}
	if !strings.Contains(raws[0], "repaired") {
		t.Fatalf("emitted item was not the repaired one: %v", raws)
	}
	if len(result.Items) != 1 || result.Items[0].Name != "repaired" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestStreamPipeline_NoDuplicationMixedValidAndInvalid(t *testing.T) {
	p := &scriptProvider{
		stream: `{"items":[{"name":"a"},{"junk":true},{"name":"c"}]}`,
		repair: func(_ string) (string, error) { return `{"name":"b"}`, nil },
	}

	var raws []string
	result, err := newStreamPipe(p).Run(context.Background(), "task", StreamOptions{OnItem: collect(&raws)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(raws) != 3 {
		t.Fatalf("expected 3 total emissions, got %d: %v", len(raws), raws)
	}
	if hasDuplicate(raws) {
		t.Fatalf("duplicate emission detected: %v", raws)
	}
	names := []string{result.Items[0].Name, result.Items[1].Name, result.Items[2].Name}
	if names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Fatalf("result order not preserved: %v", names)
	}
}

func TestStreamPipeline_UnrepairableElementDroppedGracefully(t *testing.T) {
	p := &scriptProvider{
		stream: `{"items":[{"name":"a"},{"junk":true}]}`,
		repair: func(_ string) (string, error) { return `{"still":"broken"}`, nil },
	}

	var raws []string
	result, err := newStreamPipe(p).Run(context.Background(), "task", StreamOptions{OnItem: collect(&raws)})
	if !errors.Is(err, ErrIncomplete) {
		t.Fatalf("a dropped element must signal ErrIncomplete (non-fatal), got: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Name != "a" {
		t.Fatalf("expected only the valid item to survive, got: %+v", result)
	}
	if len(raws) != 1 || raws[0] != `{"name":"a"}` {
		t.Fatalf("expected exactly the valid item emitted, got: %v", raws)
	}
}

func TestStreamPipeline_ExhaustsElementRepairThenDrops(t *testing.T) {
	p := &scriptProvider{
		stream: `{"items":[{"junk":true}]}`,
		repair: func(_ string) (string, error) { return `{"oops":1}`, nil },
	}
	result, err := newStreamPipe(p).Run(context.Background(), "task", StreamOptions{})
	if !errors.Is(err, ErrIncomplete) {
		t.Fatalf("an unrepairable element must signal ErrIncomplete (non-fatal), got: %v", err)
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected no surviving items, got: %+v", result)
	}
	if p.repairCalls != MaxRepairAttempts {
		t.Fatalf("expected %d element repair attempts, got %d", MaxRepairAttempts, p.repairCalls)
	}
	if p.streamCalls != 1 {
		t.Fatalf("expected 1 stream call, got %d", p.streamCalls)
	}
}

func TestStreamPipeline_WrongTypedFieldFallsBackToWholeRepair(t *testing.T) {
	const task = "WHOLE_TASK_MARKER"
	p := &scriptProvider{
		stream: `{"items":"oops"}`,
		repair: func(prompt string) (string, error) {
			if !strings.Contains(prompt, task) {
				return "", errors.New("whole-document repair must include original task")
			}
			return `{"items":[{"name":"recovered"}]}`, nil
		},
	}

	var raws []string
	result, err := newStreamPipe(p).Run(context.Background(), task, StreamOptions{OnItem: collect(&raws)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Name != "recovered" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(raws) != 1 || hasDuplicate(raws) {
		t.Fatalf("expected exactly one clean emission, got: %v", raws)
	}
}

func TestStreamPipeline_EmptyResponseTriggersContextualRepair(t *testing.T) {
	const task = "EMPTY_TASK_MARKER"
	p := &scriptProvider{
		stream: "{}",
		repair: func(prompt string) (string, error) {
			if !strings.Contains(prompt, task) {
				return "", errors.New("contextual repair prompt must include the original task")
			}
			return `{"items":[{"name":"recovered"}]}`, nil
		},
	}
	result, err := newStreamPipe(p).Run(context.Background(), task, StreamOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Name != "recovered" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestStreamPipeline_OnFirstTokenFired(t *testing.T) {
	p := &scriptProvider{stream: `{"items":[]}`}
	fired := false
	_, err := newStreamPipe(p).Run(context.Background(), "prompt", StreamOptions{
		OnFirstToken: func() { fired = true },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fired {
		t.Fatal("OnFirstToken was not fired")
	}
}

func TestStreamPipeline_OnRepairStartFires(t *testing.T) {
	p := &scriptProvider{
		stream: `{"items":[{"junk":true}]}`,
		repair: func(_ string) (string, error) { return `{"name":"ok"}`, nil },
	}
	var attempts []int
	_, err := newStreamPipe(p).Run(context.Background(), "task", StreamOptions{
		OnRepairStart: func(n int) { attempts = append(attempts, n) },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attempts) == 0 {
		t.Fatal("expected at least one OnRepairStart call")
	}
}

func TestStreamPipeline_ContextCancellationStopsRepair(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := &scriptProvider{
		stream: `{"items":[{"junk":true}]}`,
		repair: func(_ string) (string, error) {
			t.Fatal("repair must not be called after cancellation")
			return "", nil
		},
	}
	cancel()

	_, err := newStreamPipe(p).Run(ctx, "prompt", StreamOptions{})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected wrapped context.Canceled, got: %v", err)
	}
	if p.repairCalls != 0 {
		t.Fatalf("expected no repair calls, got %d", p.repairCalls)
	}
}

func TestStreamPipeline_EmitsScalarFieldLive(t *testing.T) {
	p := &scriptProvider{stream: `{"summary":"all clear","findings":[]}`}

	var keys, raws []string
	result, err := newScalarPipe(p).Run(context.Background(), "prompt", StreamOptions{
		OnItem: func(key, raw string) { keys = append(keys, key); raws = append(raws, raw) },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 1 || keys[0] != "summary" {
		t.Fatalf("expected exactly one summary emission, got keys=%v", keys)
	}
	if raws[0] != `"all clear"` {
		t.Fatalf("scalar emitted in wrong shape: %q", raws[0])
	}
	if result.Summary != "all clear" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestStreamPipeline_ScalarNotDuplicatedWithFindings(t *testing.T) {
	p := &scriptProvider{stream: `{"summary":"two issues","findings":[{"severity":"warning","title":"a"},{"severity":"info","title":"b"}]}`}

	var raws []string
	result, err := newScalarPipe(p).Run(context.Background(), "prompt", StreamOptions{OnItem: collect(&raws)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(raws) != 3 {
		t.Fatalf("expected 3 emissions (summary + 2 findings), got %d: %v", len(raws), raws)
	}
	if hasDuplicate(raws) {
		t.Fatalf("duplicate emission detected: %v", raws)
	}
	if len(result.Findings) != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestStreamPipeline_SemanticallyInvalidElementHeldThenRepaired(t *testing.T) {
	const task = "ELEMENT_SEMANTIC_TASK"
	p := &scriptProvider{
		stream: `{"summary":"one issue","findings":[{"severity":"bogus","title":"a"}]}`,
		repair: func(prompt string) (string, error) {
			if !strings.Contains(prompt, findingSchemaStr) {
				return "", errors.New("element repair prompt missing element schema")
			}
			if !strings.Contains(prompt, "severity must be 'warning' or 'info'") {
				return "", errors.New("element repair prompt missing the semantic error")
			}
			if !strings.Contains(prompt, `"severity":"bogus"`) {
				return "", errors.New("element repair prompt missing the invalid item")
			}
			if strings.Contains(prompt, task) {
				return "", errors.New("element repair prompt must stay item-scoped, not inject the whole task")
			}
			return `{"severity":"warning","title":"a"}`, nil
		},
	}

	var raws []string
	result, err := newScalarPipe(p).Run(context.Background(), task, StreamOptions{OnItem: collect(&raws)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.repairCalls == 0 {
		t.Fatal("expected a per-element repair")
	}
	if len(raws) != 2 {
		t.Fatalf("expected summary + one finding emitted, got %d: %v", len(raws), raws)
	}
	if hasDuplicate(raws) {
		t.Fatalf("duplicate emission detected: %v", raws)
	}
	if !strings.Contains(raws[1], `"warning"`) || strings.Contains(strings.Join(raws, " "), "bogus") {
		t.Fatalf("displayed element must be the verified one, never the invalid one: %v", raws)
	}
	if len(result.Findings) != 1 || result.Findings[0].Severity != "warning" {
		t.Fatalf("returned value must reflect the repair: %+v", result)
	}
}

func TestStreamPipeline_SemanticallyUnrepairableElementDropped(t *testing.T) {
	p := &scriptProvider{
		stream: `{"summary":"x","findings":[{"severity":"warning","title":"a"},{"severity":"bogus","title":"b"}]}`,
		repair: func(_ string) (string, error) {
			return `{"severity":"still-bogus","title":"b"}`, nil
		},
	}

	var raws []string
	result, err := newScalarPipe(p).Run(context.Background(), "task", StreamOptions{OnItem: collect(&raws)})
	if !errors.Is(err, ErrIncomplete) {
		t.Fatalf("a dropped finding must signal ErrIncomplete (non-fatal), got: %v", err)
	}
	if p.repairCalls != MaxRepairAttempts {
		t.Fatalf("expected %d element repair attempts, got %d", MaxRepairAttempts, p.repairCalls)
	}
	if len(result.Findings) != 1 || result.Findings[0].Severity != "warning" {
		t.Fatalf("only the valid finding may survive: %+v", result)
	}
	if strings.Contains(strings.Join(raws, " "), "bogus") {
		t.Fatalf("an unverified element must never be emitted: %v", raws)
	}
}

func TestStreamPipeline_EmptyStreamEmitsScalarAfterRepair(t *testing.T) {
	p := &scriptProvider{
		stream: "{}",
		repair: func(_ string) (string, error) {
			return `{"summary":"recovered","findings":[{"severity":"info","title":"x"}]}`, nil
		},
	}

	var keys []string
	result, err := newScalarPipe(p).Run(context.Background(), "task", StreamOptions{
		OnItem: func(key, _ string) { keys = append(keys, key) },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected summary + finding emitted after whole repair, got keys=%v", keys)
	}
	if result.Summary != "recovered" || len(result.Findings) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestStreamPipeline_ProseWrappedStreamNotDuplicated(t *testing.T) {
	p := &scriptProvider{
		stream: "Here is the analysis you requested:\n" +
			`{"summary":"one job","findings":[{"severity":"warning","title":"a"}]}` +
			"\n\nLet me know if you need anything else.",
		repair: func(_ string) (string, error) {
			return "", errors.New("must not repair: the streamed JSON is valid once extracted")
		},
	}

	var raws []string
	result, err := newScalarPipe(p).Run(context.Background(), "task", StreamOptions{OnItem: collect(&raws)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.repairCalls != 0 {
		t.Fatalf("expected no repair, got %d", p.repairCalls)
	}
	if len(raws) != 2 {
		t.Fatalf("expected summary + one finding emitted exactly once, got %d: %v", len(raws), raws)
	}
	if hasDuplicate(raws) {
		t.Fatalf("prose-wrapped stream must not duplicate: %v", raws)
	}
	if result.Summary != "one job" || len(result.Findings) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestStreamPipeline_FencedStreamNotDuplicated(t *testing.T) {
	p := &scriptProvider{
		stream: "```json\n" +
			`{"summary":"two","findings":[{"severity":"warning","title":"a"},{"severity":"info","title":"b"}]}` +
			"\n```",
		repair: func(_ string) (string, error) {
			return "", errors.New("must not repair fenced-but-valid JSON")
		},
	}

	var raws []string
	result, err := newScalarPipe(p).Run(context.Background(), "task", StreamOptions{OnItem: collect(&raws)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.repairCalls != 0 {
		t.Fatalf("expected no repair, got %d", p.repairCalls)
	}
	if len(raws) != 3 {
		t.Fatalf("expected summary + two findings, got %d: %v", len(raws), raws)
	}
	if hasDuplicate(raws) {
		t.Fatalf("fenced stream must not duplicate: %v", raws)
	}
	if len(result.Findings) != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestExtractJSONObject(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{"clean", `{"a":1}`, `{"a":1}`, true},
		{"leading prose", `sure: {"a":1}`, `{"a":1}`, true},
		{"trailing prose", `{"a":1} done`, `{"a":1}`, true},
		{"fenced", "```json\n{\"a\":1}\n```", `{"a":1}`, true},
		{"nested", `{"a":{"b":2},"c":3}`, `{"a":{"b":2},"c":3}`, true},
		{"brace in string", `{"a":"}{"}`, `{"a":"}{"}`, true},
		{"escaped quote in string", `{"a":"x\"}"}`, `{"a":"x\"}"}`, true},
		{"no object", `no json here`, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := extractJSONObject(tt.in)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("extractJSONObject(%q) = (%q,%v), want (%q,%v)", tt.in, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestStreamPipeline_DocRepairExhaustionDegradesNotCrash(t *testing.T) {
	p := &scriptProvider{
		stream: `{"summary":"","findings":[{"severity":"warning","title":"a"}]}`,
		repair: func(_ string) (string, error) {
			return `{"summary":"","findings":[{"severity":"warning","title":"a"}]}`, nil
		},
	}

	result, err := newScalarPipe(p).Run(context.Background(), "task", StreamOptions{})
	if !errors.Is(err, ErrIncomplete) {
		t.Fatalf("doc repair exhaustion must degrade to ErrIncomplete, got: %v", err)
	}
	if p.repairCalls != MaxRepairAttempts {
		t.Fatalf("expected %d document repair attempts, got %d", MaxRepairAttempts, p.repairCalls)
	}
	if len(result.Findings) != 1 || result.Findings[0].Severity != "warning" {
		t.Fatalf("best-effort result must keep the verified finding: %+v", result)
	}
}

func TestStreamPipeline_EmptyUnrepairableDegradesNotCrash(t *testing.T) {
	p := &scriptProvider{
		stream: "{}",
		repair: func(_ string) (string, error) { return "{}", nil },
	}
	result, err := newScalarPipe(p).Run(context.Background(), "task", StreamOptions{})
	if !errors.Is(err, ErrIncomplete) {
		t.Fatalf("unrepairable empty document must degrade to ErrIncomplete, got: %v", err)
	}
	if result.Summary != "" || len(result.Findings) != 0 {
		t.Fatalf("expected empty best-effort result, got: %+v", result)
	}
}

func TestStreamPipeline_ProviderErrorIsFatalNotIncomplete(t *testing.T) {
	p := &scriptProvider{streamErr: errors.New("ollama unreachable")}
	_, err := newScalarPipe(p).Run(context.Background(), "task", StreamOptions{})
	if err == nil || errors.Is(err, ErrIncomplete) {
		t.Fatalf("a provider/transport failure must stay fatal, got: %v", err)
	}
}

func TestRepairer_ElementPromptIsItemScoped(t *testing.T) {
	r := Repairer{}
	prompt := r.BuildElementRepairPrompt(
		NewSchema(`{"name":"string"}`),
		`{"name":42}`,
		[]string{`field "name": expected string`},
	)
	for _, want := range []string{
		"corrected item",
		"Preserve every value that is already correct",
		`{"name":"string"}`,
		`{"name":42}`,
		`field "name": expected string`,
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("element repair prompt missing %q, got:\n%s", want, prompt)
		}
	}
}

func newScalarPipeWithEnum(p Provider) StreamPipeline[scalarSchema] {
	return newScalarPipe(p).WithEnum("severity", "warning", "info")
}

func TestStreamPipeline_EnumNearMissFixedWithoutRepair(t *testing.T) {
	p := &scriptProvider{
		stream: `{"summary":"s","findings":[{"severity":"Warning","title":"a"},{"severity":"INFO","title":"b"}]}`,
		repair: func(_ string) (string, error) {
			return "", errors.New("must not repair: enum case is deterministically fixable")
		},
	}

	var raws []string
	result, err := newScalarPipeWithEnum(p).Run(context.Background(), "task", StreamOptions{OnItem: collect(&raws)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.repairCalls != 0 {
		t.Fatalf("enum near-misses must be fixed deterministically, got %d repairs", p.repairCalls)
	}
	if len(result.Findings) != 2 || result.Findings[0].Severity != "warning" || result.Findings[1].Severity != "info" {
		t.Fatalf("enums must be canonicalized in the result: %+v", result)
	}
	for _, r := range raws {
		if strings.Contains(r, "Warning") || strings.Contains(r, "INFO") {
			t.Fatalf("emitted item must carry the canonical enum, got: %v", raws)
		}
	}
}

func TestStreamPipeline_ElementRepairProviderErrorDropsAndContinues(t *testing.T) {
	p := &scriptProvider{
		stream: `{"summary":"s","findings":[{"severity":"warning","title":"a"},{"severity":"bogus","title":"b"},{"severity":"info","title":"c"}]}`,
		repair: func(_ string) (string, error) {
			return "", errors.New("ollama hiccup")
		},
	}

	var raws []string
	result, err := newScalarPipeWithEnum(p).Run(context.Background(), "task", StreamOptions{OnItem: collect(&raws)})
	if !errors.Is(err, ErrIncomplete) {
		t.Fatalf("a provider error on one element must degrade, not crash: %v", err)
	}
	if len(result.Findings) != 2 {
		t.Fatalf("the two valid findings must survive the one failed element: %+v", result)
	}
	if result.Findings[0].Title != "a" || result.Findings[1].Title != "c" {
		t.Fatalf("survivors must be the valid elements in order: %+v", result)
	}
	if strings.Contains(strings.Join(raws, " "), "bogus") {
		t.Fatalf("the unrepairable element must never be emitted: %v", raws)
	}
}

func TestStreamPipeline_TruncatedDocumentResultMatchesDisplay(t *testing.T) {
	p := &scriptProvider{
		stream: `{"summary":"two findings","findings":[{"severity":"warning","title":"a"},{"severity":"info","title":"b"}]`,
		repair: func(_ string) (string, error) {
			return "", errors.New("must not whole-document repair: the elements streamed cleanly")
		},
	}

	var raws []string
	result, err := newScalarPipeWithEnum(p).Run(context.Background(), "task", StreamOptions{OnItem: collect(&raws)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.repairCalls != 0 {
		t.Fatalf("a truncated tail must not trigger whole-document repair, got %d", p.repairCalls)
	}
	if len(result.Findings) != 2 {
		t.Fatalf("result must match the two streamed findings, got: %+v", result)
	}
	if len(raws) != 3 {
		t.Fatalf("summary + two findings must have been displayed, got %d: %v", len(raws), raws)
	}
}

type twoArrayDoc struct {
	Issues  []scalarItem `json:"issues"`
	Actions []scalarItem `json:"actions"`
}

func TestStreamPipeline_TwoArraysTruncatedSummaryMatches(t *testing.T) {
	p := &scriptProvider{
		stream: `{"issues":[{"severity":"warning","title":"a"},{"severity":"info","title":"b"}],"actions":[{"severity":"warning","title":"c"},{"severity":"info","title":"d"}]`,
		repair: func(_ string) (string, error) {
			return "", errors.New("must not repair")
		},
	}
	pipe := NewStreamPipeline(p, NewSchema(`{"issues":[{"severity":"string","title":"string"}],"actions":[{"severity":"string","title":"string"}]}`),
		StructValidator[twoArrayDoc]{})

	var raws []string
	result, err := pipe.Run(context.Background(), "task", StreamOptions{OnItem: collect(&raws)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Issues) != 2 || len(result.Actions) != 2 {
		t.Fatalf("result counts must match the 2 issues + 2 actions displayed, got issues=%d actions=%d",
			len(result.Issues), len(result.Actions))
	}
	if len(raws) != 4 {
		t.Fatalf("all four elements must have been displayed, got %d: %v", len(raws), raws)
	}
}
