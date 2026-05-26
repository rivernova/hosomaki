// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"strings"
	"testing"
	"time"
)

func TestRunSuccess(t *testing.T) {
	result, errMsg := run("echo", "hello world")
	if errMsg != "" {
		t.Fatalf("run() returned unexpected error: %v", errMsg)
	}
	if result != "hello world" {
		t.Errorf("run() = %v, want %v", result, "hello world")
	}
}

func TestRunFailure(t *testing.T) {
	result, errMsg := run("nonexistent_command_xyz", "arg1")
	if result != "" {
		t.Errorf("run() with nonexistent command should return empty string, got: %v", result)
	}
	if errMsg == "" {
		t.Error("run() with nonexistent command should return an error message")
	}
}

func TestRunShellSuccess(t *testing.T) {
	result, errMsg := runShell("echo hello_from_shell")
	if errMsg != "" {
		t.Fatalf("runShell() returned unexpected error: %v", errMsg)
	}
	if result != "hello_from_shell" {
		t.Errorf("runShell() = %v, want %v", result, "hello_from_shell")
	}
}

func TestRunShellFailure(t *testing.T) {
	result, errMsg := runShell("nonexistent_command_xyz arg1")
	if result != "" {
		t.Errorf("runShell() with nonexistent command should return empty string, got: %v", result)
	}
	if errMsg == "" {
		t.Error("runShell() with nonexistent command should return an error message")
	}
}

func TestRunEmptyOutput(t *testing.T) {
	result, errMsg := run("echo", "")
	if errMsg != "" {
		t.Fatalf("run() returned unexpected error: %v", errMsg)
	}
	// echo with an empty string argument still exits 0 and returns an empty line
	if result != "" {
		t.Errorf("run() with empty arg = %v, want empty string", result)
	}
}

func TestSnapshotReturnsData(t *testing.T) {
	snap, err := Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() returned error: %v", err)
	}
	if snap.CollectedAt.IsZero() {
		t.Error("Snapshot() CollectedAt should not be zero")
	}
	if snap.Uptime == "" {
		t.Error("Snapshot() Uptime should not be empty")
	}
}

func TestSnapshotCollectedAt(t *testing.T) {
	snap, err := Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() returned error: %v", err)
	}
	if snap.CollectedAt.After(time.Now()) {
		t.Error("Snapshot() CollectedAt should not be in the future")
	}
}

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
