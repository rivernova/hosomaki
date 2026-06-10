// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package watcher

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// unit tests for buffer management

func testCfg() BufferConfig {
	return BufferConfig{
		SilenceWindow: 100 * time.Millisecond,
		MaxLines:      5,
	}
}

func TestIsActionable_Error(t *testing.T) {
	if !isActionable("<ERROR> connection refused") {
		t.Error("<ERROR> lines must be actionable")
	}
}

func TestIsActionable_Warn(t *testing.T) {
	if !isActionable("<WARN> disk usage at 90%") {
		t.Error("<WARN> lines must be actionable")
	}
}

func TestIsActionable_Info(t *testing.T) {
	if isActionable("<INFO> service started") {
		t.Error("<INFO> lines must not be actionable")
	}
}

func TestIsActionable_Debug(t *testing.T) {
	if isActionable("<DEBUG> entering handler") {
		t.Error("<DEBUG> lines must not be actionable")
	}
}

func TestIsActionable_Transaction(t *testing.T) {
	if isActionable("<TRANSACTION> installed curl") {
		t.Error("<TRANSACTION> lines must not be actionable")
	}
}

func TestIsActionable_Scriptlet(t *testing.T) {
	if isActionable("<SCRIPTLET> running postinst") {
		t.Error("<SCRIPTLET> lines must not be actionable")
	}
}

func TestIsActionable_EmptyString(t *testing.T) {
	if isActionable("") {
		t.Error("empty string must not be actionable")
	}
}

func TestIsActionable_ErrorMidLine(t *testing.T) {
	if isActionable("<INFO> contains <ERROR> word mid-line") {
		t.Error("actionability must be determined by line prefix, not substring")
	}
}

func TestLineBuffer_Add_SetsHasAlertOnError(t *testing.T) {
	b := newLineBuffer(testCfg())
	b.add("<ERROR> something broke")
	if !b.hasAlert {
		t.Error("hasAlert should be true after adding an <ERROR> line")
	}
}

func TestLineBuffer_Add_SetsHasAlertOnWarn(t *testing.T) {
	b := newLineBuffer(testCfg())
	b.add("<WARN> high memory usage")
	if !b.hasAlert {
		t.Error("hasAlert should be true after adding a <WARN> line")
	}
}

func TestLineBuffer_Add_DoesNotSetHasAlertOnInfo(t *testing.T) {
	b := newLineBuffer(testCfg())
	b.add("<INFO> all is well")
	if b.hasAlert {
		t.Error("hasAlert must not be set by <INFO> lines")
	}
}

func TestLineBuffer_Add_UpdatesLastLine(t *testing.T) {
	b := newLineBuffer(testCfg())
	before := time.Now()
	b.add("<INFO> hello")
	if b.lastLine.Before(before) {
		t.Error("lastLine should be updated when a line is added")
	}
}

func TestLineBuffer_Len_Empty(t *testing.T) {
	b := newLineBuffer(testCfg())
	if b.len() != 0 {
		t.Errorf("len() = %d, want 0 on empty buffer", b.len())
	}
}

func TestLineBuffer_Len_AfterAdds(t *testing.T) {
	b := newLineBuffer(testCfg())
	b.add("<INFO> one")
	b.add("<INFO> two")
	b.add("<ERROR> three")
	if b.len() != 3 {
		t.Errorf("len() = %d, want 3", b.len())
	}
}

func TestShouldFlush_EmptyBufferNeverFlushes(t *testing.T) {
	b := newLineBuffer(testCfg())
	_, ok := b.shouldFlush()
	if ok {
		t.Error("empty buffer must never trigger a flush")
	}
}

func TestShouldFlush_FullBuffer(t *testing.T) {
	b := newLineBuffer(testCfg())
	for i := range 5 {
		b.add(fmt.Sprintf("<INFO> line %d", i))
	}
	reason, ok := b.shouldFlush()
	if !ok {
		t.Error("full buffer must trigger a flush")
	}
	if reason != flushReasonFull {
		t.Errorf("reason = %v, want flushReasonFull", reason)
	}
}

func TestShouldFlush_SilenceWindowNotElapsed(t *testing.T) {
	b := newLineBuffer(BufferConfig{
		SilenceWindow: 10 * time.Second,
		MaxLines:      50,
	})
	b.add("<ERROR> something bad")
	_, ok := b.shouldFlush()
	if ok {
		t.Error("silence window not elapsed — must not flush yet")
	}
}

func TestShouldFlush_SilenceWindowElapsedWithAlert(t *testing.T) {
	b := newLineBuffer(BufferConfig{
		SilenceWindow: 1 * time.Millisecond,
		MaxLines:      50,
	})
	b.add("<ERROR> something bad")
	time.Sleep(5 * time.Millisecond)
	reason, ok := b.shouldFlush()
	if !ok {
		t.Error("silence window elapsed with alert — must flush")
	}
	if reason != flushReasonSilence {
		t.Errorf("reason = %v, want flushReasonSilence", reason)
	}
}

func TestShouldFlush_SilenceWindowElapsedWithoutAlert(t *testing.T) {
	b := newLineBuffer(BufferConfig{
		SilenceWindow: 1 * time.Millisecond,
		MaxLines:      50,
	})
	b.add("<INFO> harmless")
	time.Sleep(5 * time.Millisecond)
	_, ok := b.shouldFlush()
	if ok {
		t.Error("silence window elapsed but no alert — must not flush")
	}
}

func TestDrain_ReturnsJoinedLines(t *testing.T) {
	b := newLineBuffer(testCfg())
	b.add("<ERROR> first")
	b.add("<INFO> second")
	text, _ := b.drain()
	if !strings.Contains(text, "<ERROR> first") {
		t.Error("drain must include all buffered lines")
	}
	if !strings.Contains(text, "<INFO> second") {
		t.Error("drain must include all buffered lines")
	}
}

func TestDrain_ReturnsActionableTrueWhenAlertPresent(t *testing.T) {
	b := newLineBuffer(testCfg())
	b.add("<ERROR> something bad")
	_, actionable := b.drain()
	if !actionable {
		t.Error("drain must return actionable=true when an alert line was buffered")
	}
}

func TestDrain_ReturnsActionableFalseWhenNoAlert(t *testing.T) {
	b := newLineBuffer(testCfg())
	b.add("<INFO> all good")
	_, actionable := b.drain()
	if actionable {
		t.Error("drain must return actionable=false when only <INFO> lines were buffered")
	}
}

func TestDrain_ResetsBuffer(t *testing.T) {
	b := newLineBuffer(testCfg())
	b.add("<ERROR> bang")
	b.drain()
	if b.len() != 0 {
		t.Error("drain must reset the line count to zero")
	}
	if b.hasAlert {
		t.Error("drain must reset hasAlert to false")
	}
	if !b.lastLine.IsZero() {
		t.Error("drain must reset lastLine to zero")
	}
}

func TestDrain_EmptyBufferReturnsEmptyString(t *testing.T) {
	b := newLineBuffer(testCfg())
	text, actionable := b.drain()
	if text != "" {
		t.Errorf("drain of empty buffer must return empty string, got %q", text)
	}
	if actionable {
		t.Error("drain of empty buffer must return actionable=false")
	}
}

func TestDrain_CanBeCalledRepeatedly(t *testing.T) {
	b := newLineBuffer(testCfg())
	b.add("<ERROR> first batch")
	b.drain()
	b.add("<WARN> second batch")
	text, actionable := b.drain()
	if !actionable {
		t.Error("second drain must still report actionable=true")
	}
	if !strings.Contains(text, "second batch") {
		t.Error("second drain must contain lines from the second batch only")
	}
}

func TestDefaultBufferConfig_SilenceWindow(t *testing.T) {
	cfg := DefaultBufferConfig()
	if cfg.SilenceWindow != 3*time.Second {
		t.Errorf("default SilenceWindow = %v, want 3s", cfg.SilenceWindow)
	}
}

func TestDefaultBufferConfig_MaxLines(t *testing.T) {
	cfg := DefaultBufferConfig()
	if cfg.MaxLines != 50 {
		t.Errorf("default MaxLines = %d, want 50", cfg.MaxLines)
	}
}
