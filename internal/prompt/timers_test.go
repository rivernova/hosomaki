// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit tests for the timers prompt builder

func makeTimersInput(timers string) TimersInput {
	return TimersInput{
		Environment: collector.Environment{
			DistroID:   "ubuntu",
			InitSystem: "systemd",
		},
		Timers: timers,
	}
}

func TestTimers_ContainsSchema(t *testing.T) {
	p := Timers(makeTimersInput("unit: logrotate.timer"))
	if !strings.Contains(p, SchemaTimers) {
		t.Error("Timers() prompt must contain the schema constant")
	}
}

func TestTimers_ContainsEnvironmentSection(t *testing.T) {
	p := Timers(makeTimersInput("unit: logrotate.timer"))
	if !strings.Contains(p, "Host environment") {
		t.Error("Timers() prompt must contain the environment section")
	}
}

func TestTimers_ContainsTimerData(t *testing.T) {
	data := "unit:      logrotate.timer\nactivates: logrotate.service"
	p := Timers(makeTimersInput(data))
	if !strings.Contains(p, "logrotate.timer") {
		t.Errorf("Timers() prompt must embed the timer list, missing 'logrotate.timer'")
	}
}

func TestTimers_InstructsPureJSON(t *testing.T) {
	p := Timers(makeTimersInput("unit: test.timer"))
	if !strings.Contains(p, "Return ONLY a JSON object") {
		t.Error("Timers() prompt must instruct the model to return only JSON")
	}
}

func TestTimers_InstructsNoMarkdown(t *testing.T) {
	p := Timers(makeTimersInput("unit: test.timer"))
	if !strings.Contains(p, "No markdown") {
		t.Error("Timers() prompt must instruct the model to produce no markdown")
	}
}

func TestTimers_DefinesNeverSemantics(t *testing.T) {
	p := Timers(makeTimersInput("unit: test.timer"))
	if !strings.Contains(p, `"never"`) {
		t.Error("Timers() prompt must define how to handle 'never' last_run values")
	}
}

func TestTimers_DefinesAllStatusValues(t *testing.T) {
	p := Timers(makeTimersInput("unit: test.timer"))
	for _, status := range []string{"ok", "warning", "failed"} {
		if !strings.Contains(p, `"`+status+`"`) {
			t.Errorf("Timers() prompt must define status value %q", status)
		}
	}
}

func TestTimers_RequiresVerbatimLastRunNextRun(t *testing.T) {
	p := Timers(makeTimersInput("unit: test.timer"))
	if !strings.Contains(p, "verbatim") {
		t.Error("Timers() prompt must instruct the model to copy last_run/next_run verbatim")
	}
}
