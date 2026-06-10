// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package watcher

import (
	"strings"
	"time"
)

// manages the accumulation of incoming  lines and the decision of when to flush them to the pipeline

type BufferConfig struct {
	SilenceWindow time.Duration
	MaxLines      int
}

func DefaultBufferConfig() BufferConfig {
	return BufferConfig{
		SilenceWindow: 3 * time.Second,
		MaxLines:      50,
	}
}

type flushReason int

const (
	flushReasonSilence  flushReason = iota // silence window elapsed
	flushReasonFull                        // buffer reached MaxLines
	flushReasonShutdown                    // context cancelled. Drain
)

type lineBuffer struct {
	cfg      BufferConfig
	lines    []string
	hasAlert bool
	lastLine time.Time
}

func newLineBuffer(cfg BufferConfig) *lineBuffer {
	return &lineBuffer{
		cfg:   cfg,
		lines: make([]string, 0, cfg.MaxLines),
	}
}

func (b *lineBuffer) add(classified string) {
	b.lines = append(b.lines, classified)
	b.lastLine = time.Now()
	if isActionable(classified) {
		b.hasAlert = true
	}
}

func (b *lineBuffer) shouldFlush() (flushReason, bool) {
	if len(b.lines) == 0 {
		return 0, false
	}
	if len(b.lines) >= b.cfg.MaxLines {
		return flushReasonFull, true
	}
	if b.hasAlert && time.Since(b.lastLine) >= b.cfg.SilenceWindow {
		return flushReasonSilence, true
	}
	return 0, false
}

func (b *lineBuffer) drain() (text string, actionable bool) {
	actionable = b.hasAlert
	if len(b.lines) > 0 {
		text = strings.Join(b.lines, "\n")
	}
	b.lines = b.lines[:0]
	b.hasAlert = false
	b.lastLine = time.Time{}
	return text, actionable
}

func (b *lineBuffer) len() int { return len(b.lines) }

func isActionable(classified string) bool {
	return strings.HasPrefix(classified, "<ERROR>") ||
		strings.HasPrefix(classified, "<WARN>")
}
