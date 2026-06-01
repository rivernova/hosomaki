// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"bytes"
	"strings"
	"testing"
)

// unit tests for sentinel writer

func writeFull(t *testing.T, sw *SentinelWriter, s string) {
	t.Helper()
	if _, err := sw.Write([]byte(s)); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
}

func writeChunked(t *testing.T, sw *SentinelWriter, s string, size int) {
	t.Helper()
	for i := 0; i < len(s); i += size {
		end := i + size
		if end > len(s) {
			end = len(s)
		}
		if _, err := sw.Write([]byte(s[i:end])); err != nil {
			t.Fatalf("Write() chunk error: %v", err)
		}
	}
}

func TestSentinelWriterProsePassesThrough(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "hello world\n")
	sw.Flush()
	if !strings.Contains(out.String(), "hello world") {
		t.Errorf("expected prose to pass through, got %q", out.String())
	}
}

func TestSentinelWriterSentinelNotWrittenToTerminal(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "prose\n---JSON---\n{}\n---END---\n")
	sw.Flush()
	if strings.Contains(out.String(), "---JSON---") {
		t.Errorf("sentinel should not appear in terminal output, got %q", out.String())
	}
	if strings.Contains(out.String(), "---END---") {
		t.Errorf("sentinel end should not appear in terminal output, got %q", out.String())
	}
}

func TestSentinelWriterProseBeforeSentinelPreserved(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "analysis text\n---JSON---\n{\"anomalies\":2}\n---END---\n")
	sw.Flush()
	if !strings.Contains(out.String(), "analysis text") {
		t.Errorf("prose before sentinel should be preserved, got %q", out.String())
	}
}

func TestSentinelWriterSentinelSplitAcrossWrites(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeChunked(t, sw, "prose\n---JSON---\n{\"anomalies\":3,\"actions\":2}\n---END---\n", 4)
	sw.Flush()
	if strings.Contains(out.String(), "---JSON---") {
		t.Errorf("sentinel split across writes should still be intercepted, got %q", out.String())
	}
	got := sw.ExtractJSON()
	if got == "" {
		t.Error("ExtractJSON() should return JSON even when sentinel was split across writes")
	}
}

func TestSentinelWriterContentAfterSentinelBufferedSilently(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "prose\n---JSON---\n{\"anomalies\":1}\n---END---\nextra after end\n")
	sw.Flush()
	if strings.Contains(out.String(), "extra after end") {
		t.Errorf("content after sentinel should be buffered, not written to terminal")
	}
}

func TestSentinelWriterFlushWritesRemainingProse(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "some prose without sentinel")
	sw.Flush()
	if !strings.Contains(out.String(), "some prose without sentinel") {
		t.Errorf("Flush() should write remaining buffered prose, got %q", out.String())
	}
}

func TestSentinelWriterFlushNoOpWhenCutting(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "prose\n---JSON---\n{}\n---END---\n")
	out.Reset()
	sw.Flush()
	if out.Len() > 0 {
		t.Errorf("Flush() after cutting should write nothing, got %q", out.String())
	}
}

func TestSentinelWriterExtractJSONReturnsJSON(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "prose\n---JSON---\n{\"anomalies\":3,\"actions\":5}\n---END---\n")
	sw.Flush()
	got := sw.ExtractJSON()
	want := `{"anomalies":3,"actions":5}`
	if got != want {
		t.Errorf("ExtractJSON() = %q, want %q", got, want)
	}
}

func TestSentinelWriterExtractJSONEmptyWhenNoSentinel(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "prose with no sentinel block\n")
	sw.Flush()
	if got := sw.ExtractJSON(); got != "" {
		t.Errorf("ExtractJSON() without sentinel = %q, want empty string", got)
	}
}

func TestSentinelWriterExtractJSONMissingEndMarker(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "prose\n---JSON---\n{\"anomalies\":1}\n")
	sw.Flush()
	got := sw.ExtractJSON()
	if got == "" {
		t.Error("ExtractJSON() without ---END--- should still return whatever JSON was buffered")
	}
}

func TestParseDoctorCountsValidJSON(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "prose\n---JSON---\n{\"anomalies\":4,\"actions\":6}\n---END---\n")
	sw.Flush()
	c := ParseDoctorCounts(sw)
	if c.Anomalies != 4 {
		t.Errorf("Anomalies = %d, want 4", c.Anomalies)
	}
	if c.Actions != 6 {
		t.Errorf("Actions = %d, want 6", c.Actions)
	}
}

func TestParseDoctorCountsInvalidJSONReturnsZero(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "prose with no JSON block\n")
	sw.Flush()
	c := ParseDoctorCounts(sw)
	if c.Anomalies != 0 || c.Actions != 0 {
		t.Errorf("ParseDoctorCounts() with no JSON = {%d,%d}, want {0,0}", c.Anomalies, c.Actions)
	}
}

func TestParseStatusCountsValidJSON(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "prose\n---JSON---\n{\"failed_services\":1,\"warn_services\":2,\"patterns_detected\":3}\n---END---\n")
	sw.Flush()
	c := ParseStatusCounts(sw)
	if c.FailedServices != 1 {
		t.Errorf("FailedServices = %d, want 1", c.FailedServices)
	}
	if c.WarnServices != 2 {
		t.Errorf("WarnServices = %d, want 2", c.WarnServices)
	}
	if c.PatternsDetected != 3 {
		t.Errorf("PatternsDetected = %d, want 3", c.PatternsDetected)
	}
}

func TestParseExplainCountsValidJSON(t *testing.T) {
	var out bytes.Buffer
	sw := NewSentinelWriter(&out)
	writeFull(t, sw, "prose\n---JSON---\n{\"patterns\":2,\"causes\":1}\n---END---\n")
	sw.Flush()
	c := ParseExplainCounts(sw)
	if c.Patterns != 2 {
		t.Errorf("Patterns = %d, want 2", c.Patterns)
	}
	if c.Causes != 1 {
		t.Errorf("Causes = %d, want 1", c.Causes)
	}
}
