// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/prompt"
)

// unit tests for the crons command

func TestCronsCmdRegistered(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "crons" {
			found = true
			break
		}
	}
	if !found {
		t.Error("crons command is not registered on the root command")
	}
}

func TestCronsCmdHasDebugFlag(t *testing.T) {
	cmd := newCronsCmd()
	f := cmd.Flags().Lookup("debug")
	if f == nil {
		t.Fatal("crons command is missing the --debug flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--debug default = %q, want %q", f.DefValue, "false")
	}
}

func TestCronsCmdRejectsArgs(t *testing.T) {
	cmd := newCronsCmd()
	err := cmd.Args(cmd, []string{"unexpected-arg"})
	if err == nil {
		t.Error("crons command should reject positional arguments")
	}
}

func TestCronsCmdShortDescription(t *testing.T) {
	cmd := newCronsCmd()
	if cmd.Short == "" {
		t.Error("crons command must have a non-empty Short description")
	}
	if !strings.Contains(strings.ToLower(cmd.Short), "cron") {
		t.Errorf("crons Short description should mention 'cron', got: %q", cmd.Short)
	}
}

func TestCronsCmdLongDescriptionMentionsNeverModifies(t *testing.T) {
	cmd := newCronsCmd()
	if !strings.Contains(cmd.Long, "never modifies") {
		t.Error("crons Long help must contain 'never modifies'")
	}
}

func TestCronsCmdLongMentionsV1Scope(t *testing.T) {
	cmd := newCronsCmd()
	if !strings.Contains(cmd.Long, "v1") || !strings.Contains(cmd.Long, "classic crontab") {
		t.Error("crons Long help must clarify v1 scope (classic crontab files)")
	}
}

func TestValidateCronsResult_ValidResult(t *testing.T) {
	r := prompt.CronsResult{
		Summary: "Found 2 cron jobs.",
		Jobs: []prompt.CronJobEntry{
			{
				Source:     "/etc/crontab",
				Schedule:   "daily at 2 AM",
				Command:    "/usr/sbin/logrotate /etc/logrotate.conf",
				WhatItDoes: "Rotates system log files to prevent disk exhaustion.",
				LastRun:    "unknown",
				Status:     "ok",
				Detail:     "",
			},
		},
	}
	errs := validateCronsResult(r)
	if len(errs) != 0 {
		t.Errorf("validateCronsResult() = %v, want no errors", errs)
	}
}

func TestValidateCronsResult_EmptySummary(t *testing.T) {
	r := prompt.CronsResult{Summary: "", Jobs: nil}
	errs := validateCronsResult(r)
	if len(errs) == 0 {
		t.Error("validateCronsResult() should reject empty summary")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "summary") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("validateCronsResult() errors should mention 'summary', got: %v", errs)
	}
}

func TestValidateCronsResult_InvalidStatus(t *testing.T) {
	r := prompt.CronsResult{
		Summary: "test",
		Jobs: []prompt.CronJobEntry{
			{Source: "/etc/crontab", Command: "/bin/true", WhatItDoes: "Does nothing.", LastRun: "unknown", Status: "maybe"},
		},
	}
	errs := validateCronsResult(r)
	if len(errs) == 0 {
		t.Error("validateCronsResult() should reject invalid status value")
	}
}

func TestValidateCronsResult_AllStatusValuesAccepted(t *testing.T) {
	for _, status := range []string{"ok", "warning", "failed"} {
		r := prompt.CronsResult{
			Summary: "test",
			Jobs: []prompt.CronJobEntry{
				{Source: "/etc/crontab", Command: "/bin/true", WhatItDoes: "x", LastRun: "unknown", Status: status},
			},
		}
		errs := validateCronsResult(r)
		if len(errs) != 0 {
			t.Errorf("validateCronsResult() rejected valid status %q: %v", status, errs)
		}
	}
}

func TestValidateCronsResult_EmptySource(t *testing.T) {
	r := prompt.CronsResult{
		Summary: "test",
		Jobs: []prompt.CronJobEntry{
			{Source: "", Command: "/bin/true", WhatItDoes: "x", LastRun: "unknown", Status: "ok"},
		},
	}
	errs := validateCronsResult(r)
	if len(errs) == 0 {
		t.Error("validateCronsResult() should reject empty source")
	}
}

func TestValidateCronsResult_EmptyCommand(t *testing.T) {
	r := prompt.CronsResult{
		Summary: "test",
		Jobs: []prompt.CronJobEntry{
			{Source: "/etc/crontab", Command: "", WhatItDoes: "x", LastRun: "unknown", Status: "ok"},
		},
	}
	errs := validateCronsResult(r)
	if len(errs) == 0 {
		t.Error("validateCronsResult() should reject empty command")
	}
}

func TestValidateCronsResult_EmptyWhatItDoes(t *testing.T) {
	r := prompt.CronsResult{
		Summary: "test",
		Jobs: []prompt.CronJobEntry{
			{Source: "/etc/crontab", Command: "/bin/true", WhatItDoes: "", LastRun: "unknown", Status: "ok"},
		},
	}
	errs := validateCronsResult(r)
	if len(errs) == 0 {
		t.Error("validateCronsResult() should reject empty what_it_does")
	}
}

func TestValidateCronsResult_EmptyLastRun(t *testing.T) {
	r := prompt.CronsResult{
		Summary: "test",
		Jobs: []prompt.CronJobEntry{
			{Source: "/etc/crontab", Command: "/bin/true", WhatItDoes: "x", LastRun: "", Status: "ok"},
		},
	}
	errs := validateCronsResult(r)
	if len(errs) == 0 {
		t.Error("validateCronsResult() should reject empty last_run")
	}
}
