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

// unit test for the pipeline

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

type recordingProvider struct {
	response string
	prompts  []string
}

func (r *recordingProvider) Generate(_ context.Context, _ string) (string, error) { return "", nil }
func (r *recordingProvider) GenerateStream(_ context.Context, _ string, _ func(), _ io.Writer) (string, error) {
	return "", nil
}
func (r *recordingProvider) GenerateJSON(_ context.Context, prompt string, _ func()) (string, error) {
	r.prompts = append(r.prompts, prompt)
	return r.response, nil
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
	v, err := newPipeline(p).Run(context.Background(), "prompt", nil)
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
	v, err := newPipeline(p).Run(context.Background(), "prompt", nil)
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

func TestPipeline_SemanticRepairUsesOriginalPrompt(t *testing.T) {
	check := func(s pipelineSchema) []string {
		if len(s.Items) == 0 {
			return []string{"items must not be empty"}
		}
		return nil
	}
	v := StructValidator[pipelineSchema]{SemanticCheck: check}

	const originalPrompt = "analyse this log and return items"
	callCount := 0
	repairProvider := &funcProvider{
		fn: func(prompt string) (string, error) {
			callCount++
			if callCount == 1 {
				return `{"name":"ok","count":0,"items":[]}`, nil
			}
			if !strings.Contains(prompt, originalPrompt) {
				return "", errors.New("repair prompt did not contain original task")
			}
			return `{"name":"ok","count":0,"items":[{"label":"real","active":true}]}`, nil
		},
	}

	result, err := NewPipeline(repairProvider, NewSchema(pipelineSchemaStr), v).
		Run(context.Background(), originalPrompt, nil)
	if err != nil {
		t.Fatalf("expected success after semantic repair, got: %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatal("expected non-empty Items after semantic repair")
	}
}

func TestPipeline_SemanticCheckTriggersRepair(t *testing.T) {
	p := &stubProvider{responses: []string{
		`{"name":"ok","count":0,"items":[]}`,
		`{"name":"ok","count":0,"items":[{"label":"a","active":true}]}`,
	}}
	v, err := newPipelineWithCheck(p, func(s pipelineSchema) []string {
		if len(s.Items) == 0 {
			return []string{"items must not be empty"}
		}
		return nil
	}).Run(context.Background(), "prompt", nil)
	if err != nil {
		t.Fatalf("expected success after semantic repair, got: %v", err)
	}
	if len(v.Items) == 0 {
		t.Fatal("expected non-empty Items after semantic repair")
	}
	if p.callIdx != 2 {
		t.Fatalf("expected 2 provider calls, got %d", p.callIdx)
	}
}

func TestPipeline_ExhaustsAllRepairAttempts(t *testing.T) {
	p := &stubProvider{responses: []string{`{"count":1,"items":[]}`}}
	_, err := newPipeline(p).Run(context.Background(), "prompt", nil)
	if err == nil {
		t.Fatal("expected error after exhausting repair attempts, got nil")
	}
	if p.callIdx != 1+MaxRepairAttempts {
		t.Fatalf("expected %d provider calls, got %d", 1+MaxRepairAttempts, p.callIdx)
	}
}

func TestPipeline_ErrorOnGeneration(t *testing.T) {
	provErr := errors.New("network unavailable")
	_, err := newPipeline(&errProvider{err: provErr}).Run(context.Background(), "prompt", nil)
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
	_, err := newPipeline(p).Run(context.Background(), "prompt", nil)
	if !errors.Is(err, provErr) {
		t.Fatalf("expected wrapped provErr, got: %v", err)
	}
}

func TestPipeline_OnFirstTokenFiredOnInitialGeneration(t *testing.T) {
	p := &stubProvider{responses: []string{`{"name":"ok","count":1,"items":[]}`}}
	fired := false
	_, err := newPipeline(p).Run(context.Background(), "prompt", func() { fired = true })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fired {
		t.Fatal("onFirstToken was not fired during initial generation")
	}
}

func TestPipeline_OnFirstTokenNotFiredOnRepair(t *testing.T) {
	p := &stubProvider{responses: []string{
		`{"count":1,"items":[]}`,
		`{"name":"fixed","count":1,"items":[]}`,
	}}
	fireCount := 0
	_, err := newPipeline(p).Run(context.Background(), "prompt", func() { fireCount++ })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fireCount != 1 {
		t.Fatalf("onFirstToken fired %d times, want exactly 1", fireCount)
	}
}

func TestPipeline_ErrorMessageContainsValidationErrors(t *testing.T) {
	p := &stubProvider{responses: []string{`{"count":1,"items":[]}`}}
	_, err := newPipeline(p).Run(context.Background(), "prompt", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Fatalf("error message should mention the failing field 'name', got: %q", err.Error())
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

func TestRepairer_SemanticPromptContainsOriginalPrompt(t *testing.T) {
	r := Repairer{}
	schema := NewSchema(`{"issues":[]}`)
	const originalTask = "analyse /var/log/dnf5.log and find errors"
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
