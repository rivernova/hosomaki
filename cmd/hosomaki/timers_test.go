// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/prompt"
)

// unit tests for the timers command

func TestTimersCmdRegistered(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "timers" {
			found = true
			break
		}
	}
	if !found {
		t.Error("timers command is not registered on the root command")
	}
}

func TestTimersCmdHasDebugFlag(t *testing.T) {
	cmd := newTimersCmd()
	f := cmd.Flags().Lookup("debug")
	if f == nil {
		t.Fatal("timers command is missing the --debug flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--debug default = %q, want %q", f.DefValue, "false")
	}
}

func TestTimersCmdRejectsArgs(t *testing.T) {
	cmd := newTimersCmd()
	err := cmd.Args(cmd, []string{"unexpected-arg"})
	if err == nil {
		t.Error("timers command should reject positional arguments")
	}
}

func TestTimersCmdShortDescription(t *testing.T) {
	cmd := newTimersCmd()
	if cmd.Short == "" {
		t.Error("timers command must have a non-empty Short description")
	}
	if !strings.Contains(strings.ToLower(cmd.Short), "timer") {
		t.Errorf("timers Short description should mention 'timer', got: %q", cmd.Short)
	}
}

func TestTimersCmdLongDescriptionMentionsNeverModifies(t *testing.T) {
	cmd := newTimersCmd()
	if !strings.Contains(cmd.Long, "never modifies") {
		t.Error("timers Long help must contain 'never modifies'")
	}
}

func TestTimersCmdLongMentionsNever(t *testing.T) {
	cmd := newTimersCmd()
	if !strings.Contains(cmd.Long, "never") {
		t.Error("timers Long help must mention the 'never' last_run behaviour")
	}
}

func TestValidateTimersResult_ValidResult(t *testing.T) {
	r := prompt.TimersResult{
		Summary: "All timers are healthy.",
		Timers: []prompt.TimerEntry{
			{
				Name:     "logrotate.timer",
				Schedule: "daily at midnight",
				LastRun:  "2024-06-10 00:00:01 UTC",
				NextRun:  "2024-06-11 00:00:00 UTC",
				Status:   "ok",
				Detail:   "",
			},
		},
	}
	errs := validateTimersResult(r)
	if len(errs) != 0 {
		t.Errorf("validateTimersResult() = %v, want no errors", errs)
	}
}

func TestValidateTimersResult_ValidNeverValues(t *testing.T) {
	r := prompt.TimersResult{
		Summary: "One timer has never run.",
		Timers: []prompt.TimerEntry{
			{Name: "test.timer", Schedule: "weekly", LastRun: "never", NextRun: "never", Status: "warning"},
		},
	}
	errs := validateTimersResult(r)
	if len(errs) != 0 {
		t.Errorf("validateTimersResult() = %v, want no errors for 'never' values", errs)
	}
}

func TestValidateTimersResult_EmptySummary(t *testing.T) {
	r := prompt.TimersResult{Summary: "", Timers: nil}
	errs := validateTimersResult(r)
	if len(errs) == 0 {
		t.Error("validateTimersResult() should reject empty summary")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "summary") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("validateTimersResult() errors should mention 'summary', got: %v", errs)
	}
}

func TestValidateTimersResult_InvalidStatus(t *testing.T) {
	r := prompt.TimersResult{
		Summary: "test",
		Timers: []prompt.TimerEntry{
			{Name: "test.timer", Schedule: "daily", LastRun: "never", NextRun: "never", Status: "unknown"},
		},
	}
	errs := validateTimersResult(r)
	if len(errs) == 0 {
		t.Error("validateTimersResult() should reject invalid status value")
	}
}

func TestValidateTimersResult_EmptyLastRun(t *testing.T) {
	r := prompt.TimersResult{
		Summary: "test",
		Timers: []prompt.TimerEntry{
			{Name: "test.timer", Schedule: "daily", LastRun: "", NextRun: "never", Status: "ok"},
		},
	}
	errs := validateTimersResult(r)
	if len(errs) == 0 {
		t.Error("validateTimersResult() should reject empty last_run")
	}
}

func TestValidateTimersResult_EmptyNextRun(t *testing.T) {
	r := prompt.TimersResult{
		Summary: "test",
		Timers: []prompt.TimerEntry{
			{Name: "test.timer", Schedule: "daily", LastRun: "never", NextRun: "", Status: "ok"},
		},
	}
	errs := validateTimersResult(r)
	if len(errs) == 0 {
		t.Error("validateTimersResult() should reject empty next_run")
	}
}

func TestValidateTimersResult_EmptyName(t *testing.T) {
	r := prompt.TimersResult{
		Summary: "test",
		Timers: []prompt.TimerEntry{
			{Name: "", Schedule: "daily", LastRun: "never", NextRun: "never", Status: "ok"},
		},
	}
	errs := validateTimersResult(r)
	if len(errs) == 0 {
		t.Error("validateTimersResult() should reject empty name")
	}
}

func TestValidateTimersResult_AllStatusValuesAccepted(t *testing.T) {
	for _, status := range []string{"ok", "warning", "failed"} {
		r := prompt.TimersResult{
			Summary: "test",
			Timers: []prompt.TimerEntry{
				{Name: "t.timer", Schedule: "x", LastRun: "never", NextRun: "never", Status: status},
			},
		}
		errs := validateTimersResult(r)
		if len(errs) != 0 {
			t.Errorf("validateTimersResult() rejected valid status %q: %v", status, errs)
		}
	}
}
