// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"
)

// unit tests for status prompt generation, focused on ensuring the JSON sentinel is present in all prompts and prompt styles

func TestLimitLinesTruncatesLongInput(t *testing.T) {
	lines := []string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
		"line 5",
		"line 6",
		"line 7",
		"line 8",
		"line 9",
		"line 10",
		"line 11",
	}

	got := limitLines(strings.Join(lines, "\n"), maxTopProcessLines)
	gotLines := strings.Split(got, "\n")

	if len(gotLines) != maxTopProcessLines {
		t.Fatalf("limitLines() returned %d lines, want %d", len(gotLines), maxTopProcessLines)
	}
	if gotLines[len(gotLines)-1] != "line 10" {
		t.Fatalf("limitLines() last line = %q, want line 10", gotLines[len(gotLines)-1])
	}
}

func TestLimitLinesKeepsShortInput(t *testing.T) {
	input := "line 1\nline 2"

	got := limitLines(input, maxTopProcessLines)
	if got != input {
		t.Fatalf("limitLines() = %q, want %q", got, input)
	}
}

func TestLimitLinesKeepsEmptyInput(t *testing.T) {
	if got := limitLines("", maxTopProcessLines); got != "" {
		t.Fatalf("limitLines() = %q, want empty string", got)
	}
}

func TestFormatSnapshotLimitsTopProcesses(t *testing.T) {
	lines := []string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
		"line 5",
		"line 6",
		"line 7",
		"line 8",
		"line 9",
		"line 10",
		"line 11",
	}

	got := formatSnapshot(StatusInput{TopProcesses: strings.Join(lines, "\n")})

	if !strings.Contains(got, "=== Top processes by CPU ===\nline 1") {
		t.Fatalf("formatSnapshot() did not include top process section:\n%s", got)
	}
	if !strings.Contains(got, "line 10") {
		t.Fatalf("formatSnapshot() = %q, want line 10 to be included", got)
	}
	if strings.Contains(got, "line 11") {
		t.Fatalf("formatSnapshot() = %q, want line 11 to be omitted", got)
	}
}
