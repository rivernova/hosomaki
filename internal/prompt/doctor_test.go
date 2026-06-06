// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"
	"time"
)

// unit testing for doctor prompt generation logic

func TestDoctorContainsSnapshot(t *testing.T) {
	input := DoctorInput{
		CollectedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Uptime:         "up 3 days, 4 hours",
		Memory:         "total 16G used 12G free 4G",
		Disk:           "/dev/sda1 95% /",
		FailedServices: "nginx.service",
		RecentErrors:   "kernel: OOM killer activated",
		TopProcesses:   "PID USER CMD\n1234 root nginx",
	}

	p := Doctor(input, false)

	checks := []struct {
		name    string
		contain string
	}{
		{"uptime", "up 3 days"},
		{"memory", "16G"},
		{"disk", "95%"},
		{"failed service", "nginx.service"},
		{"recent error", "OOM killer"},
		{"top process", "nginx"},
		{"collected at", "2024-01-15"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(p, c.contain) {
				t.Errorf("Doctor() prompt missing %q", c.contain)
			}
		})
	}
}

func TestDoctorBriefStyle(t *testing.T) {
	input := DoctorInput{CollectedAt: time.Now()}

	brief := Doctor(input, true)
	full := Doctor(input, false)

	if !strings.Contains(brief, "one sentence") {
		t.Error("Doctor() brief prompt should contain one-sentence instruction")
	}

	if strings.Contains(brief, "structured plain-prose") {
		t.Error("Doctor() brief prompt should not contain full-mode instructions")
	}

	if !strings.Contains(full, "concrete action") {
		t.Error("Doctor() full prompt should instruct model to provide concrete actions")
	}
	if strings.Contains(full, "one sentence") {
		t.Error("Doctor() full prompt should not contain brief-mode instructions")
	}
}

func TestDoctorActionableInstructions(t *testing.T) {
	input := DoctorInput{CollectedAt: time.Now()}
	p := Doctor(input, false)

	if !strings.Contains(p, "suggest concrete next steps") {
		t.Error("Doctor() prompt must instruct model to suggest concrete next steps")
	}

	if !strings.Contains(p, "potentially disruptive") {
		t.Error("Doctor() prompt must ask model to flag potentially disruptive actions")
	}
}

func TestDoctorNoMarkdownInstruction(t *testing.T) {
	input := DoctorInput{CollectedAt: time.Now()}
	p := Doctor(input, false)

	for _, forbidden := range []string{"No markdown", "No bullet points", "No numbered lists", "No headers"} {
		if !strings.Contains(p, forbidden) {
			t.Errorf("Doctor() prompt should contain formatting rule %q", forbidden)
		}
	}
}

func TestDoctorSnapshotSectionsPresent(t *testing.T) {
	input := DoctorInput{
		CollectedAt: time.Now(),
		Uptime:      "up 1 hour",
	}
	p := Doctor(input, false)

	for _, section := range []string{
		"=== Collected at ===",
		"=== Uptime ===",
		"=== Memory ===",
		"=== Disk ===",
		"=== Failed services ===",
		"=== Recent errors (journalctl) ===",
		"=== Top processes by CPU ===",
	} {
		if !strings.Contains(p, section) {
			t.Errorf("Doctor() snapshot missing section header %q", section)
		}
	}
}

func TestDoctorEmptyFieldsFallback(t *testing.T) {
	input := DoctorInput{CollectedAt: time.Now()}
	p := Doctor(input, false)

	if !strings.Contains(p, "(no data)") {
		t.Error("Doctor() should show (no data) for empty snapshot fields")
	}
}

func TestDoctorTopProcessesLimited(t *testing.T) {
	lines := make([]string, 15)
	for i := range lines {
		lines[i] = "row"
	}
	input := DoctorInput{
		CollectedAt:  time.Now(),
		TopProcesses: strings.Join(lines, "\n"),
	}
	p := Doctor(input, false)

	count := strings.Count(p, "row")
	if count > maxTopProcessLines {
		t.Errorf("Doctor() top-processes section has %d rows, want at most %d", count, maxTopProcessLines)
	}
}
