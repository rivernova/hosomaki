// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"testing"
)

// unit tests for why log collection

func TestWhyLogs_NonexistentServiceReturnsError(t *testing.T) {
	_, err := WhyLogs("hosomaki-nonexistent-why-xyzzy", LogOptions{})
	if err == nil {
		t.Fatal("expected error for nonexistent service, got nil")
	}
}

func TestWhyLogs_DefaultLinesWhenZero(t *testing.T) {
	if got := lines(0, defaultServiceLines); got != defaultServiceLines {
		t.Fatalf("lines(0, %d) = %d, want %d",
			defaultServiceLines, got, defaultServiceLines)
	}
}

func TestWhyLogs_ExplicitLinesRespected(t *testing.T) {
	if got := lines(10, defaultServiceLines); got != 10 {
		t.Fatalf("lines(10, %d) = %d, want 10", defaultServiceLines, got)
	}
}

func TestWhyLogs_NegativeLinesUsesDefault(t *testing.T) {
	if got := lines(-1, defaultServiceLines); got != defaultServiceLines {
		t.Fatalf("lines(-1, %d) = %d, want %d",
			defaultServiceLines, got, defaultServiceLines)
	}
}

func TestWhyLogs_KnownService(t *testing.T) {
	logs, err := WhyLogs("systemd-journald", LogOptions{})
	if err != nil {
		t.Skipf("skipping: journal unavailable or empty: %v", err)
	}
	if logs == "" {
		t.Fatal("expected non-empty log output for systemd-journald")
	}
}
