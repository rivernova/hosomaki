// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"strings"
	"testing"
)

// unit tests for the crons collector

func TestIsVariableAssignment_ShellVar(t *testing.T) {
	if !isVariableAssignment("SHELL=/bin/sh") {
		t.Error("isVariableAssignment('SHELL=/bin/sh') should be true")
	}
}

func TestIsVariableAssignment_Mailto(t *testing.T) {
	if !isVariableAssignment("MAILTO=root") {
		t.Error("isVariableAssignment('MAILTO=root') should be true")
	}
}

func TestIsVariableAssignment_EmptyValue(t *testing.T) {
	if !isVariableAssignment("MAILTO=") {
		t.Error("isVariableAssignment('MAILTO=') should be true (empty value is valid)")
	}
}

func TestIsVariableAssignment_CronJobWithArgEquals(t *testing.T) {
	if isVariableAssignment("* * * * * cmd --flag=value") {
		t.Error("isVariableAssignment('* * * * * cmd --flag=value') should be false")
	}
}

func TestIsVariableAssignment_NoEquals(t *testing.T) {
	if isVariableAssignment("* * * * * /bin/cmd") {
		t.Error("isVariableAssignment with no '=' should be false")
	}
}

func TestParseSystemCrontabLines_BasicJob(t *testing.T) {
	lines := []string{
		"# comment",
		"SHELL=/bin/sh",
		"0 2 * * * root /usr/sbin/logrotate /etc/logrotate.conf",
	}
	jobs := parseSystemCrontabLines("/etc/crontab", lines)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	j := jobs[0]
	if j.Schedule != "0 2 * * *" {
		t.Errorf("Schedule = %q, want %q", j.Schedule, "0 2 * * *")
	}
	if j.User != "root" {
		t.Errorf("User = %q, want %q", j.User, "root")
	}
	if !strings.Contains(j.Command, "logrotate") {
		t.Errorf("Command = %q, want it to contain 'logrotate'", j.Command)
	}
	if j.Source != "/etc/crontab" {
		t.Errorf("Source = %q, want %q", j.Source, "/etc/crontab")
	}
}

func TestParseSystemCrontabLines_AtRebootWithUser(t *testing.T) {
	lines := []string{"@reboot root /usr/bin/start-daemon --quiet"}
	jobs := parseSystemCrontabLines("/etc/cron.d/daemon", lines)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job for @reboot system line, got %d", len(jobs))
	}
	j := jobs[0]
	if j.Schedule != "@reboot" {
		t.Errorf("Schedule = %q, want %q", j.Schedule, "@reboot")
	}
	if j.User != "root" {
		t.Errorf("User = %q, want %q", j.User, "root")
	}
	if !strings.Contains(j.Command, "start-daemon") {
		t.Errorf("Command = %q, want it to contain 'start-daemon'", j.Command)
	}
}

func TestParseSystemCrontabLines_AtRebootTooFewFields(t *testing.T) {
	lines := []string{"@reboot root"}
	jobs := parseSystemCrontabLines("/etc/crontab", lines)
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs for '@reboot root' (no command), got %d", len(jobs))
	}
}

func TestParseSystemCrontabLines_SkipsComments(t *testing.T) {
	lines := []string{
		"# this is a comment",
		"## double comment",
		"5 4 * * * root /bin/cmd",
	}
	jobs := parseSystemCrontabLines("/etc/crontab", lines)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d: %+v", len(jobs), jobs)
	}
}

func TestParseSystemCrontabLines_SkipsVariableAssignments(t *testing.T) {
	lines := []string{
		"MAILTO=root",
		"PATH=/usr/bin:/bin",
		"0 1 * * * root /bin/true",
	}
	jobs := parseSystemCrontabLines("/etc/crontab", lines)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job (no variable lines), got %d", len(jobs))
	}
}

func TestParseSystemCrontabLines_TooFewFields(t *testing.T) {
	lines := []string{"* * * *"} // only 4 fields; need ≥7
	jobs := parseSystemCrontabLines("/etc/crontab", lines)
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs for under-field line, got %d", len(jobs))
	}
}

func TestParseSystemCrontabLines_Empty(t *testing.T) {
	jobs := parseSystemCrontabLines("/etc/crontab", nil)
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs for nil input, got %d", len(jobs))
	}
}

func TestParseUserCrontabLines_BasicJob(t *testing.T) {
	lines := []string{
		"# user crontab",
		"*/5 * * * * /home/alice/backup.sh",
	}
	jobs := parseUserCrontabLines("user:alice", lines)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	j := jobs[0]
	if j.Schedule != "*/5 * * * *" {
		t.Errorf("Schedule = %q, want %q", j.Schedule, "*/5 * * * *")
	}
	if j.User != "" {
		t.Errorf("User = %q, want empty (user crontab format)", j.User)
	}
	if j.Source != "user:alice" {
		t.Errorf("Source = %q, want %q", j.Source, "user:alice")
	}
	if !strings.Contains(j.Command, "backup.sh") {
		t.Errorf("Command = %q, want it to contain 'backup.sh'", j.Command)
	}
}

func TestParseUserCrontabLines_AtRebootShorthand(t *testing.T) {
	lines := []string{"@reboot /usr/bin/myserver --daemon"}
	jobs := parseUserCrontabLines("user:root", lines)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job for @reboot line, got %d", len(jobs))
	}
	if jobs[0].Schedule != "@reboot" {
		t.Errorf("Schedule = %q, want %q", jobs[0].Schedule, "@reboot")
	}
	if !strings.Contains(jobs[0].Command, "myserver") {
		t.Errorf("Command = %q, want it to contain 'myserver'", jobs[0].Command)
	}
}

func TestParseUserCrontabLines_AtHourlyShorthand(t *testing.T) {
	lines := []string{"@hourly /usr/local/bin/check"}
	jobs := parseUserCrontabLines("user:bob", lines)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job for @hourly line, got %d", len(jobs))
	}
	if jobs[0].Schedule != "@hourly" {
		t.Errorf("Schedule = %q, want %q", jobs[0].Schedule, "@hourly")
	}
}

func TestParseUserCrontabLines_AtShorthandNoCommand(t *testing.T) {
	lines := []string{"@reboot"}
	jobs := parseUserCrontabLines("user:root", lines)
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs for '@reboot' with no command, got %d", len(jobs))
	}
}

func TestParseUserCrontabLines_TooFewFields(t *testing.T) {
	lines := []string{"* * * * "} // only 4 fields; need ≥6
	jobs := parseUserCrontabLines("user:bob", lines)
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs for under-field line, got %d", len(jobs))
	}
}

func TestParseUserCrontabLines_SkipsVariableAssignments(t *testing.T) {
	lines := []string{
		"MAILTO=''",
		"0 3 * * * /bin/backup",
	}
	jobs := parseUserCrontabLines("user:root", lines)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
}

func TestParseUserCrontabLines_CommandWithEqualsSign(t *testing.T) {
	lines := []string{"0 * * * * /usr/bin/curl --data key=value https://example.com"}
	jobs := parseUserCrontabLines("user:root", lines)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job for command with '=' in arg, got %d", len(jobs))
	}
	if !strings.Contains(jobs[0].Command, "key=value") {
		t.Errorf("Command = %q should contain 'key=value'", jobs[0].Command)
	}
}

func TestFormatCronsForPrompt_Empty(t *testing.T) {
	out := FormatCronsForPrompt(nil)
	if !strings.Contains(out, "no cron jobs found") {
		t.Errorf("FormatCronsForPrompt(nil) = %q, want 'no cron jobs found'", out)
	}
}

func TestFormatCronsForPrompt_AllFieldsPresent(t *testing.T) {
	jobs := []CronJob{
		{
			Source:   "/etc/crontab",
			Schedule: "0 2 * * *",
			User:     "root",
			Command:  "/usr/sbin/logrotate /etc/logrotate.conf",
		},
	}
	out := FormatCronsForPrompt(jobs)
	for _, want := range []string{"source:", "schedule:", "user:", "command:"} {
		if !strings.Contains(out, want) {
			t.Errorf("FormatCronsForPrompt() missing field %q, got:\n%s", want, out)
		}
	}
}

func TestFormatCronsForPrompt_UserCrontabOmitsUserLine(t *testing.T) {
	jobs := []CronJob{
		{
			Source:   "user:alice",
			Schedule: "*/5 * * * *",
			User:     "",
			Command:  "/home/alice/backup.sh",
		},
	}
	out := FormatCronsForPrompt(jobs)
	if strings.Contains(out, "\nuser:") {
		t.Error("FormatCronsForPrompt() should omit the 'user:' line when User is empty")
	}
}

func TestFormatCronsForPrompt_MultipleJobs(t *testing.T) {
	jobs := []CronJob{
		{Source: "user:root", Schedule: "@reboot", Command: "/usr/bin/start"},
		{Source: "/etc/cron.d/apt", Schedule: "0 4 * * *", User: "root", Command: "/usr/lib/apt/apt.systemd.daily"},
	}
	out := FormatCronsForPrompt(jobs)
	if !strings.Contains(out, "@reboot") {
		t.Error("FormatCronsForPrompt() missing @reboot job")
	}
	if !strings.Contains(out, "apt.systemd.daily") {
		t.Error("FormatCronsForPrompt() missing apt daily job")
	}
}
