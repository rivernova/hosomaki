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

// unit testing stream pipeline

type streamStubProvider struct {
	responses []string
	callIdx   int
}

func (s *streamStubProvider) Generate(_ context.Context, _ string) (string, error) {
	return s.next(), nil
}

func (s *streamStubProvider) GenerateStream(_ context.Context, _ string, onFirstToken func(), w io.Writer) (string, error) {
	resp := s.next()
	if onFirstToken != nil && resp != "" {
		onFirstToken()
	}
	if w != nil {
		if _, err := io.WriteString(w, resp); err != nil {
			return "", err
		}
	}
	return resp, nil
}

func (s *streamStubProvider) GenerateJSON(_ context.Context, _ string, onFirstToken func()) (string, error) {
	if onFirstToken != nil {
		onFirstToken()
	}
	return s.next(), nil
}

func (s *streamStubProvider) next() string {
	idx := s.callIdx
	s.callIdx++
	if idx >= len(s.responses) {
		return s.responses[len(s.responses)-1]
	}
	return s.responses[idx]
}

type streamErrProvider struct{ err error }

func (e *streamErrProvider) Generate(_ context.Context, _ string) (string, error) {
	return "", e.err
}
func (e *streamErrProvider) GenerateStream(_ context.Context, _ string, _ func(), _ io.Writer) (string, error) {
	return "", e.err
}
func (e *streamErrProvider) GenerateJSON(_ context.Context, _ string, _ func()) (string, error) {
	return "", e.err
}

type streamFuncProvider struct {
	stream func(prompt string, w io.Writer) (string, error)
	repair func(prompt string) (string, error)
	calls  int
}

func (f *streamFuncProvider) Generate(_ context.Context, _ string) (string, error) { return "", nil }
func (f *streamFuncProvider) GenerateStream(_ context.Context, prompt string, onFirstToken func(), w io.Writer) (string, error) {
	f.calls++
	resp, err := f.stream(prompt, w)
	if err != nil {
		return "", err
	}
	if onFirstToken != nil && resp != "" {
		onFirstToken()
	}
	return resp, nil
}
func (f *streamFuncProvider) GenerateJSON(_ context.Context, prompt string, _ func()) (string, error) {
	f.calls++
	return f.repair(prompt)
}

type streamSchema struct {
	Items []streamItem `json:"items"`
}

type streamItem struct {
	Name string `json:"name"`
}

const streamSchemaStr = `{"items":[{"name":"string"}]}`

func newStreamPipe(p Provider) StreamPipeline[streamSchema] {
	return NewStreamPipeline(p, NewSchema(streamSchemaStr), StructValidator[streamSchema]{})
}

func TestStreamPipeline_EmitsItemsAsTheyArrive(t *testing.T) {
	resp := `{"items":[{"name":"alpha"},{"name":"beta"},{"name":"gamma"}]}`
	p := &streamStubProvider{responses: []string{resp}}

	var keys, raws []string
	_, err := newStreamPipe(p).Run(context.Background(), "prompt", StreamOptions{
		OnItem: func(key, raw string) {
			keys = append(keys, key)
			raws = append(raws, raw)
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 OnItem calls, got %d", len(keys))
	}
	for _, k := range keys {
		if k != "items" {
			t.Fatalf("expected key 'items', got %q", k)
		}
	}
	if !strings.Contains(raws[0], "alpha") || !strings.Contains(raws[1], "beta") || !strings.Contains(raws[2], "gamma") {
		t.Fatalf("unexpected item raws: %v", raws)
	}
}

func TestStreamPipeline_DecodesFullResult(t *testing.T) {
	resp := `{"items":[{"name":"x"},{"name":"y"}]}`
	p := &streamStubProvider{responses: []string{resp}}

	result, err := newStreamPipe(p).Run(context.Background(), "prompt", StreamOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Name != "x" || result.Items[1].Name != "y" {
		t.Fatalf("unexpected item names: %v", result.Items)
	}
}

func TestStreamPipeline_ValidOnFirstGeneration(t *testing.T) {
	p := &streamStubProvider{responses: []string{`{"items":[]}`}}
	result, err := newStreamPipe(p).Run(context.Background(), "prompt", StreamOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Items == nil {
		t.Fatal("expected non-nil items slice")
	}
}

func TestStreamPipeline_ErrorOnStreamGeneration(t *testing.T) {
	provErr := errors.New("stream failure")
	_, err := newStreamPipe(&streamErrProvider{err: provErr}).Run(context.Background(), "prompt", StreamOptions{})
	if !errors.Is(err, provErr) {
		t.Fatalf("expected wrapped provErr, got: %v", err)
	}
}

func TestStreamPipeline_RepairOnInvalidResponse(t *testing.T) {
	const originalTask = "STREAM_TASK_MARKER"

	firstStreamed := false
	p := &streamFuncProvider{
		stream: func(_ string, w io.Writer) (string, error) {
			firstStreamed = true
			resp := `{"items":[{"junk":true}]}`
			if w != nil {
				_, err := io.WriteString(w, resp)
				if err != nil {
					return "", err
				}
			}
			return resp, nil
		},
		repair: func(prompt string) (string, error) {
			if !strings.Contains(prompt, originalTask) {
				return "", errors.New("repair prompt missing original task")
			}
			return `{"items":[{"name":"repaired"}]}`, nil
		},
	}

	result, err := newStreamPipe(p).Run(context.Background(), originalTask, StreamOptions{})
	if err != nil {
		t.Fatalf("unexpected error after repair: %v", err)
	}
	if !firstStreamed {
		t.Fatal("expected streaming call to have been made")
	}
	if len(result.Items) == 0 || result.Items[0].Name != "repaired" {
		t.Fatalf("unexpected result after repair: %+v", result)
	}
}

func TestStreamPipeline_ReturnsCorrectResultAfterRepair(t *testing.T) {
	p := &streamFuncProvider{
		stream: func(_ string, w io.Writer) (string, error) {
			resp := `{"items":[{"junk":true}]}`
			if w != nil {
				_, _ = io.WriteString(w, resp)
			}
			return resp, nil
		},
		repair: func(_ string) (string, error) {
			return `{"items":[{"name":"fixed"}]}`, nil
		},
	}

	result, err := newStreamPipe(p).Run(context.Background(), "task", StreamOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) == 0 || result.Items[0].Name != "fixed" {
		t.Fatalf("expected repaired result, got: %+v", result)
	}
}

func TestStreamPipeline_OnFirstTokenFired(t *testing.T) {
	p := &streamStubProvider{responses: []string{`{"items":[]}`}}
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
	p := &streamFuncProvider{
		stream: func(_ string, w io.Writer) (string, error) {
			resp := `{"items":[{"junk":true}]}`
			if w != nil {
				_, _ = io.WriteString(w, resp)
			}
			return resp, nil
		},
		repair: func(_ string) (string, error) {
			return `{"items":[{"name":"ok"}]}`, nil
		},
	}

	var attempts []int
	_, _ = newStreamPipe(p).Run(context.Background(), "task", StreamOptions{
		OnRepairStart: func(n int) { attempts = append(attempts, n) },
	})
	if len(attempts) == 0 {
		t.Fatal("expected at least one OnRepairStart call")
	}
}

func TestStreamPipeline_ContextCancellationStopsRepair(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	p := &streamFuncProvider{
		stream: func(_ string, w io.Writer) (string, error) {
			calls++
			cancel()
			resp := `{"items":[{"junk":true}]}`
			if w != nil {
				_, _ = io.WriteString(w, resp)
			}
			return resp, nil
		},
		repair: func(_ string) (string, error) {
			t.Fatal("repair called after context cancellation")
			return "", nil
		},
	}

	_, err := newStreamPipe(p).Run(ctx, "prompt", StreamOptions{})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected wrapped context.Canceled, got: %v", err)
	}
}

func TestStreamPipeline_ExhaustsAllRepairAttempts(t *testing.T) {
	p := &streamFuncProvider{
		stream: func(_ string, w io.Writer) (string, error) {
			resp := `{"items":[{"junk":true}]}`
			if w != nil {
				_, _ = io.WriteString(w, resp)
			}
			return resp, nil
		},
		repair: func(_ string) (string, error) {
			return `{"items":[{"junk":true}]}`, nil
		},
	}

	_, err := newStreamPipe(p).Run(context.Background(), "task", StreamOptions{})
	if err == nil {
		t.Fatal("expected error after exhausting repair attempts")
	}
	want := 1 + MaxRepairAttempts
	if p.calls != want {
		t.Fatalf("expected %d provider calls, got %d", want, p.calls)
	}
}
