// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"

	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt logic for the timers command

type TimerEntry struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	LastRun  string `json:"last_run"`
	NextRun  string `json:"next_run"`
	Status   string `json:"status"`
	Detail   string `json:"detail"`
}

type TimersResult struct {
	Summary string       `json:"summary"`
	Timers  []TimerEntry `json:"timers"`
}

type TimersInput struct {
	Environment collector.Environment
	Timers      string
}

func Timers(in TimersInput) string {
	return fmt.Sprintf(`You are a Linux systems expert reviewing all systemd timers on a live system.

%s
TASK
Analyse the systemd timer list below. For each timer, explain what it does and
when it last ran. Flag timers that appear to have failed, are overdue, or have
never successfully run when they should have.

DEFINITIONS
- A timer is "failed" when its associated service last exited with a non-"success"
  Result (e.g. "exit-code", "signal", "timeout", "watchdog").
- A timer is at "warning" level when it has never run but is expected to
  (i.e. has a valid next_run), or when its last run is unexpectedly old.
- A timer that ran successfully and is scheduled normally is "ok".

OUTPUT
Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.
The JSON must use exactly these field names. Do not rename, abbreviate, or add fields.

SCHEMA
%s

FIELD RULES
- "summary": one to two sentences. Describe the overall timer health. State how many
  timers are present and whether any are failing or overdue. Maximum 40 words.
- "timers": one entry per timer in the input. Do not invent timers not present.
- "name": the timer unit name verbatim from the input.
- "schedule": a plain-English description of the schedule (e.g. "daily at midnight",
  "every 15 minutes"). Do not copy the cron expression or systemd OnCalendar value;
  translate it into natural language.
- "last_run": copy the last_run value verbatim from the input. If the input says
  "never", use exactly the string "never".
- "next_run": copy the next_run value verbatim from the input. If the input says
  "never", use exactly the string "never".
- "status": exactly one of "ok", "warning", or "failed". Apply the definitions above.
- "detail": 1–3 plain-text sentences about what this timer does, why its status was
  assigned, and any action the operator should take for non-ok timers. For "ok"
  timers with nothing notable, an empty string is acceptable.

No markdown. No bullet points. No headers.

Systemd timer list:
%s`, EnvironmentSection(in.Environment), SchemaTimers, in.Timers)
}
