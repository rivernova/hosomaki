// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

// unit testing for the Pipeline type and its repair logic

type stubProvider struct {
	responses []string
	callIdx   int
}

func (s *stubProvider) Generate(_ context.Context, _ string) (string, error) {
	return s.next(), nil
}

func (s *stubProvider) GenerateStream(_ context.Context, _ string, _ func(), _ io.Writer) (string, error) {
	return s.next(), nil
}

func (s *stubProvider) GenerateJSON(_ context.Context, _ string, onFirstToken func()) (string, error) {
	if onFirstToken != nil {
		onFirstToken()
	}
	return s.next(), nil
}

func (s *stubProvider) next() string {
	idx := s.callIdx
	s.callIdx++
	if idx >= len(s.responses) {
		return s.responses[len(s.responses)-1]
	}
	return s.responses[idx]
}

type errProvider struct{ err error }

func (e *errProvider) Generate(_ context.Context, _ string) (string, error) { return "", e.err }
func (e *errProvider) GenerateStream(_ context.Context, _ string, _ func(), _ io.Writer) (string, error) {
	return "", e.err
}
func (e *errProvider) GenerateJSON(_ context.Context, _ string, _ func()) (string, error) {
	return "", e.err
}

type conditionalErrProvider struct {
	firstResponse string
	repairErr     error
	calls         int
}

func (c *conditionalErrProvider) Generate(_ context.Context, _ string) (string, error) {
	return "", nil
}
func (c *conditionalErrProvider) GenerateStream(_ context.Context, _ string, _ func(), _ io.Writer) (string, error) {
	return "", nil
}
func (c *conditionalErrProvider) GenerateJSON(_ context.Context, _ string, onFirstToken func()) (string, error) {
	if onFirstToken != nil {
		onFirstToken()
	}
	c.calls++
	if c.calls == 1 {
		return c.firstResponse, nil
	}
	return "", c.repairErr
}

type funcProvider struct {
	fn func(prompt string) (string, error)
}

func (f *funcProvider) Generate(_ context.Context, _ string) (string, error) { return "", nil }
func (f *funcProvider) GenerateStream(_ context.Context, _ string, _ func(), _ io.Writer) (string, error) {
	return "", nil
}
func (f *funcProvider) GenerateJSON(_ context.Context, prompt string, _ func()) (string, error) {
	return f.fn(prompt)
}

type pipelineItem struct {
	Label  string `json:"label"`
	Active bool   `json:"active"`
}

type pipelineSchema struct {
	Name  string         `json:"name"`
	Count int            `json:"count"`
	Items []pipelineItem `json:"items"`
}

const pipelineSchemaStr = `{"name":"string","count":0,"items":[]}`

func newPipeline(p Provider) Pipeline[pipelineSchema] {
	return NewPipeline(p, NewSchema(pipelineSchemaStr), StructValidator[pipelineSchema]{})
}

func newPipelineWithCheck(p Provider, check func(pipelineSchema) []string) Pipeline[pipelineSchema] {
	v := StructValidator[pipelineSchema]{SemanticCheck: check}
	return NewPipeline(p, NewSchema(pipelineSchemaStr), v)
}

func TestPipeline_ValidOnFirstGeneration(t *testing.T) {
	p := &stubProvider{responses: []string{`{"name":"ok","count":1,"items":[]}`}}
	v, err := newPipeline(p).Run(context.Background(), "prompt", RunOptions{})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if v.Name != "ok" {
		t.Fatalf("Name = %q, want %q", v.Name, "ok")
	}
	if p.callIdx != 1 {
		t.Fatalf("expected 1 provider call, got %d", p.callIdx)
	}
}

func TestPipeline_StructuralRepairOnFirstAttempt(t *testing.T) {
	p := &stubProvider{responses: []string{
		`{"count":1,"items":[]}`,
		`{"name":"fixed","count":1,"items":[]}`,
	}}
	v, err := newPipeline(p).Run(context.Background(), "prompt", RunOptions{})
	if err != nil {
		t.Fatalf("expected success after structural repair, got: %v", err)
	}
	if v.Name != "fixed" {
		t.Fatalf("Name = %q, want %q", v.Name, "fixed")
	}
	if p.callIdx != 2 {
		t.Fatalf("expected 2 provider calls, got %d", p.callIdx)
	}
}

func TestPipeline_EmptyResponseTriggersContextualRepair(t *testing.T) {
	const originalTask = "ANALYSE_THIS_LOG_X12345"
	calls := 0
	p := &funcProvider{
		fn: func(prompt string) (string, error) {
			calls++
			if calls == 1 {
				return "{}", nil
			}
			if !strings.Contains(prompt, originalTask) {
				return "", errors.New("repair prompt missing original task")
			}
			return `{"name":"recovered","count":1,"items":[]}`, nil
		},
	}
	v, err := newPipeline(p).Run(context.Background(), originalTask, RunOptions{})
	if err != nil {
		t.Fatalf("expected success after contextual repair, got: %v", err)
	}
	if v.Name != "recovered" {
		t.Fatalf("unexpected name: %q", v.Name)
	}
}

func TestPipeline_SemanticRepairUsesOriginalPrompt(t *testing.T) {
	const originalPrompt = "analyse this log and return items"
	calls := 0
	rp := &funcProvider{
		fn: func(prompt string) (string, error) {
			calls++
			if calls == 1 {
				return `{"name":"ok","count":0,"items":[]}`, nil
			}
			if !strings.Contains(prompt, originalPrompt) {
				return "", errors.New("repair prompt missing original task")
			}
			return `{"name":"ok","count":0,"items":[{"label":"real","active":true}]}`, nil
		},
	}
	check := func(s pipelineSchema) []string {
		if len(s.Items) == 0 {
			return []string{"items must not be empty"}
		}
		return nil
	}
	result, err := newPipelineWithCheck(rp, check).Run(context.Background(), originalPrompt, RunOptions{})
	if err != nil {
		t.Fatalf("expected success after semantic repair, got: %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatal("expected non-empty Items after semantic repair")
	}
}

func TestPipeline_ExhaustsAllRepairAttempts(t *testing.T) {
	p := &stubProvider{responses: []string{`{"count":1,"items":[]}`}}
	_, err := newPipeline(p).Run(context.Background(), "prompt", RunOptions{})
	if err == nil {
		t.Fatal("expected error after exhausting repair attempts")
	}
	if p.callIdx != 1+MaxRepairAttempts {
		t.Fatalf("expected %d provider calls, got %d", 1+MaxRepairAttempts, p.callIdx)
	}
}

func TestPipeline_ErrorOnGeneration(t *testing.T) {
	provErr := errors.New("network unavailable")
	_, err := newPipeline(&errProvider{err: provErr}).Run(context.Background(), "prompt", RunOptions{})
	if !errors.Is(err, provErr) {
		t.Fatalf("expected wrapped provErr, got: %v", err)
	}
}

func TestPipeline_ErrorOnRepairCall(t *testing.T) {
	provErr := errors.New("connection reset")
	p := &conditionalErrProvider{
		firstResponse: `{"count":1,"items":[]}`,
		repairErr:     provErr,
	}
	_, err := newPipeline(p).Run(context.Background(), "prompt", RunOptions{})
	if !errors.Is(err, provErr) {
		t.Fatalf("expected wrapped provErr, got: %v", err)
	}
}

func TestPipeline_OnFirstTokenFiredOnInitialGeneration(t *testing.T) {
	p := &stubProvider{responses: []string{`{"name":"ok","count":1,"items":[]}`}}
	fired := false
	_, err := newPipeline(p).Run(context.Background(), "prompt", RunOptions{
		OnFirstToken: func() { fired = true },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fired {
		t.Fatal("OnFirstToken was not fired during initial generation")
	}
}

func TestPipeline_OnFirstTokenNotFiredOnRepair(t *testing.T) {
	p := &stubProvider{responses: []string{
		`{"count":1,"items":[]}`,
		`{"name":"fixed","count":1,"items":[]}`,
	}}
	fireCount := 0
	_, err := newPipeline(p).Run(context.Background(), "prompt", RunOptions{
		OnFirstToken: func() { fireCount++ },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fireCount != 1 {
		t.Fatalf("OnFirstToken fired %d times, want exactly 1", fireCount)
	}
}

func TestPipeline_OnRepairStartFiresPerAttempt(t *testing.T) {
	p := &stubProvider{responses: []string{`{"count":1,"items":[]}`}}
	attempts := []int{}
	_, _ = newPipeline(p).Run(context.Background(), "prompt", RunOptions{
		OnRepairStart: func(n int) { attempts = append(attempts, n) },
	})
	if len(attempts) != MaxRepairAttempts {
		t.Fatalf("expected %d OnRepairStart fires, got %d", MaxRepairAttempts, len(attempts))
	}
	for i, n := range attempts {
		if n != i+1 {
			t.Fatalf("attempt #%d reported as %d", i+1, n)
		}
	}
}

func TestPipeline_ContextCancellationStopsRepair(t *testing.T) {
	calls := 0
	ctx, cancel := context.WithCancel(context.Background())
	p := &funcProvider{
		fn: func(_ string) (string, error) {
			calls++
			if calls == 1 {
				cancel()
				return `{"count":1,"items":[]}`, nil
			}
			t.Fatal("provider called after context cancellation")
			return "", nil
		},
	}
	_, err := newPipeline(p).Run(ctx, "prompt", RunOptions{})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected wrapped context.Canceled, got: %v", err)
	}
}

func TestPipeline_ErrorMessageContainsValidationErrors(t *testing.T) {
	p := &stubProvider{responses: []string{`{"count":1,"items":[]}`}}
	_, err := newPipeline(p).Run(context.Background(), "prompt", RunOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Fatalf("error must mention the failing field 'name', got: %q", err.Error())
	}
}

func TestNewSchema_PanicsOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty schema, got none")
		}
	}()
	NewSchema("")
}

func TestNewSchema_StringRoundTrip(t *testing.T) {
	s := NewSchema(`{"foo":"string"}`)
	if s.String() != `{"foo":"string"}` {
		t.Fatalf("String() = %q", s.String())
	}
}

func TestRepairer_StructuralPromptContainsSchema(t *testing.T) {
	r := Repairer{}
	schema := NewSchema(`{"key":"string"}`)
	prompt := r.BuildStructuralRepairPrompt(schema, `{"key":42}`, []string{`field "key": expected string`})
	if !strings.Contains(prompt, `{"key":"string"}`) {
		t.Fatalf("structural repair prompt must contain the schema, got:\n%s", prompt)
	}
}

func TestRepairer_StructuralPromptContainsInvalidJSON(t *testing.T) {
	r := Repairer{}
	schema := NewSchema(`{"key":"string"}`)
	invalid := `{"key":42}`
	prompt := r.BuildStructuralRepairPrompt(schema, invalid, []string{`field "key": expected string`})
	if !strings.Contains(prompt, invalid) {
		t.Fatalf("structural repair prompt must contain the invalid JSON, got:\n%s", prompt)
	}
}

func TestRepairer_StructuralPromptContainsAllErrors(t *testing.T) {
	r := Repairer{}
	schema := NewSchema(`{"a":"string","b":"string"}`)
	errs := []string{`field "a" is missing`, `field "b" is missing`}
	prompt := r.BuildStructuralRepairPrompt(schema, `{}`, errs)
	for _, e := range errs {
		if !strings.Contains(prompt, e) {
			t.Fatalf("structural repair prompt missing error %q, got:\n%s", e, prompt)
		}
	}
}

func TestRepairer_StructuralWithContextContainsOriginalTask(t *testing.T) {
	r := Repairer{}
	schema := NewSchema(`{"issues":[]}`)
	prompt := r.BuildStructuralRepairPromptWithContext(
		schema, `{}`, []string{`field "issues" is missing`}, "original-task-marker",
	)
	if !strings.Contains(prompt, "original-task-marker") {
		t.Fatalf("contextual structural prompt missing original task, got:\n%s", prompt)
	}
}

func TestRepairer_SemanticPromptContainsOriginalPrompt(t *testing.T) {
	r := Repairer{}
	schema := NewSchema(`{"issues":[]}`)
	const originalTask = "analyse log and find errors"
	prompt := r.BuildSemanticRepairPrompt(schema, originalTask, []string{"issues must not be empty"})
	if !strings.Contains(prompt, originalTask) {
		t.Fatalf("semantic repair prompt must contain the original task, got:\n%s", prompt)
	}
}

func TestRepairer_SemanticPromptContainsSchema(t *testing.T) {
	r := Repairer{}
	schema := NewSchema(`{"issues":[{"what":"string","why":"string"}]}`)
	prompt := r.BuildSemanticRepairPrompt(schema, "original task", []string{"issues must not be empty"})
	if !strings.Contains(prompt, schema.String()) {
		t.Fatalf("semantic repair prompt must contain the schema, got:\n%s", prompt)
	}
}

func TestRepairer_SemanticPromptContainsSemanticErrors(t *testing.T) {
	r := Repairer{}
	schema := NewSchema(`{"issues":[]}`)
	errs := []string{"issues must not be empty", "at least one issue required"}
	prompt := r.BuildSemanticRepairPrompt(schema, "task", errs)
	for _, e := range errs {
		if !strings.Contains(prompt, e) {
			t.Fatalf("semantic repair prompt missing error %q, got:\n%s", e, prompt)
		}
	}
}

func TestIsEssentiallyEmpty(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"", true},
		{"{}", true},
		{" { } ", true},
		{"null", true},
		{"[]", true},
		{`{"x":1}`, false},
		{`{"x":null}`, false},
	}
	for _, tc := range tests {
		if got := isEssentiallyEmpty(tc.in); got != tc.want {
			t.Errorf("isEssentiallyEmpty(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
