// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit tests for the crons prompt builder

func makeCronsInput(jobs string) CronsInput {
	return CronsInput{
		Environment: collector.Environment{
			DistroID:   "ubuntu",
			InitSystem: "systemd",
		},
		Jobs: jobs,
	}
}

func TestCrons_ContainsSchema(t *testing.T) {
	p := Crons(makeCronsInput("source: /etc/crontab"))
	if !strings.Contains(p, SchemaCrons) {
		t.Error("Crons() prompt must contain the schema constant")
	}
}

func TestCrons_ContainsEnvironmentSection(t *testing.T) {
	p := Crons(makeCronsInput("source: /etc/crontab"))
	if !strings.Contains(p, "Host environment") {
		t.Error("Crons() prompt must contain the environment section")
	}
}

func TestCrons_ContainsJobData(t *testing.T) {
	data := "source:   /etc/crontab\nschedule: 0 2 * * *\ncommand:  /usr/sbin/logrotate"
	p := Crons(makeCronsInput(data))
	if !strings.Contains(p, "logrotate") {
		t.Errorf("Crons() prompt must embed the job list verbatim, missing 'logrotate'")
	}
}

func TestCrons_InstructsPureJSON(t *testing.T) {
	p := Crons(makeCronsInput("source: test"))
	if !strings.Contains(p, "Return ONLY a JSON object") {
		t.Error("Crons() prompt must instruct the model to return only JSON")
	}
}

func TestCrons_InstructsNoMarkdown(t *testing.T) {
	p := Crons(makeCronsInput("source: test"))
	if !strings.Contains(p, "No markdown") {
		t.Error("Crons() prompt must instruct the model to produce no markdown")
	}
}

func TestCrons_DefinesLastRunUnknown(t *testing.T) {
	p := Crons(makeCronsInput("source: test"))
	if !strings.Contains(p, "unknown") {
		t.Error("Crons() prompt must instruct the model to use 'unknown' when last_run is unavailable")
	}
}

func TestCrons_DefinesAllStatusValues(t *testing.T) {
	p := Crons(makeCronsInput("source: test"))
	for _, status := range []string{"ok", "warning", "failed"} {
		if !strings.Contains(p, `"`+status+`"`) {
			t.Errorf("Crons() prompt must define status value %q", status)
		}
	}
}

func TestCrons_WarnsSuspiciousPatterns(t *testing.T) {
	p := Crons(makeCronsInput("source: test"))
	if !strings.Contains(p, "curl") || !strings.Contains(p, "wget") {
		t.Error("Crons() prompt must mention curl/wget piped to shell as a warning trigger")
	}
}

func TestCrons_InstructsVerbatimSourceAndCommand(t *testing.T) {
	p := Crons(makeCronsInput("source: test"))
	if !strings.Contains(p, "verbatim") {
		t.Error("Crons() prompt must instruct the model to copy source and command verbatim")
	}
}
