// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"

	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt logic for the crons command

type CronJobEntry struct {
	Source     string `json:"source"`
	Schedule   string `json:"schedule"`
	Command    string `json:"command"`
	WhatItDoes string `json:"what_it_does"`
	LastRun    string `json:"last_run"`
	Status     string `json:"status"`
	Detail     string `json:"detail"`
}

type CronsResult struct {
	Summary string         `json:"summary"`
	Jobs    []CronJobEntry `json:"jobs"`
}

type CronsInput struct {
	Environment collector.Environment
	Jobs        string
}

func Crons(in CronsInput) string {
	return fmt.Sprintf(`You are a Linux systems expert reviewing all cron jobs on a live system.

%s
TASK
Analyse the cron job list below. For each job, explain what it does in plain English
and assess its status. Flag jobs whose command looks broken, suspicious, or whose
schedule is unusual or potentially dangerous.

IMPORTANT: You do not have live journal data for these jobs. For "last_run", respond
with "unknown" unless the input explicitly provides this information. Do not invent
execution timestamps.

DEFINITIONS
- "failed": the command path does not exist, the schedule is invalid, or there is
  strong evidence the job is broken (e.g. obvious syntax error in the command).
- "warning": the schedule is unusually frequent (more often than every minute would be
  pathological), the command writes to sensitive locations, uses curl/wget piped to a
  shell, or runs as root with world-writable paths. Also flag jobs whose schedule
  means they may be overdue based on the schedule alone.
- "ok": the job looks correct, the command is a known system utility, and the schedule
  is reasonable.

OUTPUT
Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.
The JSON must use exactly these field names. Do not rename, abbreviate, or add fields.

SCHEMA
%s

FIELD RULES
- "summary": one to two sentences. State how many jobs were found, where they came
  from, and whether anything is concerning. Maximum 40 words.
- "jobs": one entry per cron job in the input. Do not invent jobs not present.
- "source": copy the source value verbatim from the input.
- "schedule": a plain-English description of the schedule (e.g. "every day at 2 AM",
  "at system reboot"). Do not copy the raw cron expression; translate it.
- "command": copy the command verbatim from the input.
- "what_it_does": one sentence describing what the command does in plain English.
- "last_run": "unknown" unless the input explicitly provides a timestamp.
- "status": exactly one of "ok", "warning", or "failed".
- "detail": 1–3 plain-text sentences about what is notable, why the status was
  assigned, and any action to take for non-ok jobs. Empty string is acceptable for
  "ok" jobs with nothing to remark on.

No markdown. No bullet points. No headers.

Cron job list:
%s`, EnvironmentSection(in.Environment), SchemaCrons, in.Jobs)
}
