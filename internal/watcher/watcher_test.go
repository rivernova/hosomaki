// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package watcher

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// unit tests for watcher

func noopSanitise(s string) string { return s }

func classifySanitise(s string) string {
	lower := strings.ToLower(s)
	switch {
	case strings.Contains(lower, "error") || strings.Contains(lower, "fatal"):
		return "<ERROR> " + s
	case strings.Contains(lower, "warn"):
		return "<WARN> " + s
	default:
		return "<INFO> " + s
	}
}

func validConfig(onFlush FlushFunc) Config {
	return Config{
		Service:   "nginx.service",
		SeedLines: 10,
		Buffer:    DefaultBufferConfig(),
		Sanitise:  noopSanitise,
		OnFlush:   onFlush,
	}
}

func TestNew_RejectsEmptyService(t *testing.T) {
	cfg := validConfig(func(_ context.Context, _ string) error { return nil })
	cfg.Service = ""
	_, err := New(cfg)
	if err == nil {
		t.Fatal("New() must reject empty service name")
	}
	if !strings.Contains(err.Error(), "service name") {
		t.Errorf("error should mention 'service name', got %q", err.Error())
	}
}

func TestNew_RejectsWhitespaceOnlyService(t *testing.T) {
	cfg := validConfig(func(_ context.Context, _ string) error { return nil })
	cfg.Service = "   "
	_, err := New(cfg)
	if err == nil {
		t.Fatal("New() must reject whitespace-only service name")
	}
}

func TestNew_RejectsNilSanitise(t *testing.T) {
	cfg := validConfig(func(_ context.Context, _ string) error { return nil })
	cfg.Sanitise = nil
	_, err := New(cfg)
	if err == nil {
		t.Fatal("New() must reject nil Sanitise function")
	}
	if !strings.Contains(err.Error(), "Sanitise") {
		t.Errorf("error should mention 'Sanitise', got %q", err.Error())
	}
}

func TestNew_RejectsNilOnFlush(t *testing.T) {
	cfg := validConfig(nil)
	_, err := New(cfg)
	if err == nil {
		t.Fatal("New() must reject nil OnFlush function")
	}
	if !strings.Contains(err.Error(), "OnFlush") {
		t.Errorf("error should mention 'OnFlush', got %q", err.Error())
	}
}

func TestNew_AcceptsValidConfig(t *testing.T) {
	cfg := validConfig(func(_ context.Context, _ string) error { return nil })
	w, err := New(cfg)
	if err != nil {
		t.Fatalf("New() with valid config returned error: %v", err)
	}
	if w == nil {
		t.Fatal("New() returned nil watcher")
	}
}

func TestNew_SubstitutesDefaultBufferWhenZero(t *testing.T) {
	cfg := validConfig(func(_ context.Context, _ string) error { return nil })
	cfg.Buffer = BufferConfig{} // zero values
	w, err := New(cfg)
	if err != nil {
		t.Fatalf("New() with zero BufferConfig returned error: %v", err)
	}
	if w.cfg.Buffer.SilenceWindow != DefaultBufferConfig().SilenceWindow {
		t.Errorf("SilenceWindow not substituted: got %v", w.cfg.Buffer.SilenceWindow)
	}
	if w.cfg.Buffer.MaxLines != DefaultBufferConfig().MaxLines {
		t.Errorf("MaxLines not substituted: got %d", w.cfg.Buffer.MaxLines)
	}
}

func TestNew_NegativeSeedLinesBecomesZero(t *testing.T) {
	cfg := validConfig(func(_ context.Context, _ string) error { return nil })
	cfg.SeedLines = -5
	w, err := New(cfg)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if w.cfg.SeedLines != 0 {
		t.Errorf("SeedLines = %d, want 0", w.cfg.SeedLines)
	}
}

func TestIngestLine_CallsOnLine(t *testing.T) {
	cfg := validConfig(func(_ context.Context, _ string) error { return nil })
	var received []string
	cfg.OnLine = func(raw string) { received = append(received, raw) }
	w, _ := New(cfg)

	w.ingestLine("raw journal line")
	if len(received) != 1 || received[0] != "raw journal line" {
		t.Errorf("OnLine not called correctly: %v", received)
	}
}

func TestIngestLine_CallsSanitise(t *testing.T) {
	var sanitised []string
	cfg := validConfig(func(_ context.Context, _ string) error { return nil })
	cfg.Sanitise = func(s string) string {
		sanitised = append(sanitised, s)
		return "<INFO> " + s
	}
	w, _ := New(cfg)
	w.ingestLine("test line")
	if len(sanitised) != 1 || sanitised[0] != "test line" {
		t.Errorf("Sanitise not called with raw line: %v", sanitised)
	}
}

func TestIngestLine_AddsSanitisedLineToBuffer(t *testing.T) {
	cfg := validConfig(func(_ context.Context, _ string) error { return nil })
	cfg.Sanitise = classifySanitise
	w, _ := New(cfg)
	w.ingestLine("fatal error occurred")
	if w.buf.len() != 1 {
		t.Errorf("buffer len = %d, want 1 after ingest", w.buf.len())
	}
	if !w.buf.hasAlert {
		t.Error("buffer hasAlert should be true after ingesting an error line")
	}
}

func TestIngestLine_SkipsEmptyAfterSanitise(t *testing.T) {
	cfg := validConfig(func(_ context.Context, _ string) error { return nil })
	cfg.Sanitise = func(_ string) string { return "   " } // returns only whitespace
	w, _ := New(cfg)
	w.ingestLine("some line")
	if w.buf.len() != 0 {
		t.Error("lines that sanitise to whitespace must not be buffered")
	}
}

func TestFlush_CallsOnFlushWhenActionable(t *testing.T) {
	called := false
	var received string
	cfg := validConfig(func(_ context.Context, text string) error {
		called = true
		received = text
		return nil
	})
	w, _ := New(cfg)
	w.buf.add("<ERROR> something broke")
	_ = w.flush(context.Background(), flushReasonSilence)
	if !called {
		t.Fatal("OnFlush must be called when buffer has actionable lines")
	}
	if !strings.Contains(received, "<ERROR>") {
		t.Errorf("OnFlush received unexpected text: %q", received)
	}
}

func TestFlush_DoesNotCallOnFlushWhenNotActionable(t *testing.T) {
	called := false
	cfg := validConfig(func(_ context.Context, _ string) error {
		called = true
		return nil
	})
	w, _ := New(cfg)
	w.buf.add("<INFO> harmless")
	_ = w.flush(context.Background(), flushReasonSilence)
	if called {
		t.Fatal("OnFlush must not be called when buffer has no actionable lines")
	}
}

func TestFlush_DoesNotCallOnFlushWhenEmpty(t *testing.T) {
	called := false
	cfg := validConfig(func(_ context.Context, _ string) error {
		called = true
		return nil
	})
	w, _ := New(cfg)
	_ = w.flush(context.Background(), flushReasonSilence)
	if called {
		t.Fatal("OnFlush must not be called on an empty buffer")
	}
}

func TestFlush_DrainsClearBuffer(t *testing.T) {
	cfg := validConfig(func(_ context.Context, _ string) error { return nil })
	w, _ := New(cfg)
	w.buf.add("<ERROR> bang")
	_ = w.flush(context.Background(), flushReasonSilence)
	if w.buf.len() != 0 {
		t.Error("buffer must be empty after flush")
	}
}

func TestFlush_PropagatesOnFlushError(t *testing.T) {
	flushErr := fmt.Errorf("pipeline failed")
	cfg := validConfig(func(_ context.Context, _ string) error { return flushErr })
	w, _ := New(cfg)
	w.buf.add("<ERROR> bang")
	err := w.flush(context.Background(), flushReasonSilence)
	if err == nil {
		t.Fatal("flush must propagate OnFlush errors")
	}
}

func TestRun_FlushTriggeredByMaxLines(t *testing.T) {
	var mu sync.Mutex
	var batches []string

	cfg := validConfig(func(_ context.Context, text string) error {
		mu.Lock()
		batches = append(batches, text)
		mu.Unlock()
		return nil
	})
	cfg.Buffer = BufferConfig{MaxLines: 3, SilenceWindow: 10 * time.Second}
	cfg.Sanitise = classifySanitise

	w, _ := New(cfg)

	w.ingestLine("error: connection refused")
	w.ingestLine("info: retrying")
	w.ingestLine("error: gave up")

	reason, ok := w.buf.shouldFlush()
	if !ok || reason != flushReasonFull {
		t.Fatalf("expected full-buffer flush, got ok=%v reason=%v", ok, reason)
	}

	_ = w.flush(context.Background(), reason)

	mu.Lock()
	defer mu.Unlock()
	if len(batches) != 1 {
		t.Fatalf("expected 1 flushed batch, got %d", len(batches))
	}
	if !strings.Contains(batches[0], "<ERROR>") {
		t.Errorf("flushed batch should contain classified error lines, got: %q", batches[0])
	}
}

func TestRun_SilenceWindowFlush(t *testing.T) {
	var mu sync.Mutex
	var batches []string

	cfg := validConfig(func(_ context.Context, text string) error {
		mu.Lock()
		batches = append(batches, text)
		mu.Unlock()
		return nil
	})
	cfg.Buffer = BufferConfig{MaxLines: 100, SilenceWindow: 10 * time.Millisecond}
	cfg.Sanitise = classifySanitise

	w, _ := New(cfg)
	w.ingestLine("error: disk full")

	// wait for silence window to elapse.
	time.Sleep(20 * time.Millisecond)

	reason, ok := w.buf.shouldFlush()
	if !ok || reason != flushReasonSilence {
		t.Fatalf("expected silence flush, got ok=%v reason=%v", ok, reason)
	}
	_ = w.flush(context.Background(), reason)

	mu.Lock()
	defer mu.Unlock()
	if len(batches) != 1 {
		t.Fatalf("expected 1 flushed batch, got %d", len(batches))
	}
}
