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
	result := run("echo", "hello world")
	if result != "hello world" {
		t.Errorf("run() = %v, want %v", result, "hello world")
	}
}

func TestRunFailure(t *testing.T) {
	result := run("nonexistent_command_xyz", "arg1")
	if !strings.Contains(result, "command failed") {
		t.Errorf("run() with nonexistent command should return error message, got: %v", result)
	}
}

func TestRunShellSuccess(t *testing.T) {
	result := runShell("echo hello_from_shell")
	if result != "hello_from_shell" {
		t.Errorf("runShell() = %v, want %v", result, "hello_from_shell")
	}
}

func TestRunShellFailure(t *testing.T) {
	result := runShell("nonexistent_command_xyz arg1")
	if !strings.Contains(result, "command failed") {
		t.Errorf("runShell() with nonexistent command should return error message, got: %v", result)
	}
}

func TestRunEmptyOutput(t *testing.T) {
	result := run("echo", "")
	if !strings.Contains(result, "command failed") {
	}
}

func TestSnapshotReturnsData(t *testing.T) {
	snapshot, err := Snapshot()
	if err != nil {
		t.Errorf("Snapshot() returned error: %v", err)
		return
	}
	if snapshot.CollectedAt.IsZero() {
		t.Error("Snapshot() CollectedAt should not be zero")
	}
	if snapshot.Uptime == "" {
		t.Error("Snapshot() Uptime should not be empty")
	}
}

func TestSnapshotCollectedAt(t *testing.T) {
	snapshot, err := Snapshot()
	if err != nil {
		t.Errorf("Snapshot() returned error: %v", err)
		return
	}
	if snapshot.CollectedAt.After(time.Now()) {
		t.Error("Snapshot() CollectedAt should not be in the future")
	}
}
