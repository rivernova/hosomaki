// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivernova/hosomaki/internal/collector"
)

// template for prompt for doctor command
type DoctorInput struct {
	CollectedAt    time.Time
	Environment    collector.Environment
	Uptime         string
	Memory         string
	Disk           string
	FailedServices string
	RecentErrors   string
	TopProcesses   string
}

type DoctorIssue struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
}

type DoctorAction struct {
	Description string `json:"description"`
	Disruptive  bool   `json:"disruptive"`
}

type DoctorResult struct {
	Issues  []DoctorIssue  `json:"issues"`
	Actions []DoctorAction `json:"actions"`
}

type DoctorBriefResult = DoctorResult

func Doctor(d DoctorInput, brief bool) string {
	var volumeInstr string
	if brief {
		volumeInstr = `Include only the most critical issues and their matching actions. Maximum 3 issues and 3 actions. If the system is healthy, return empty arrays.`
	} else {
		volumeInstr = `Include every distinct issue you find. There is no maximum. If the system is healthy, return empty arrays.`
	}

	return fmt.Sprintf(`You are a Linux system expert performing a full diagnostic of a live system.

%s
TASK
Analyse the system snapshot below and return a JSON object — and nothing else.
No prose, no explanation, no markdown fences, no commentary. Pure JSON only.

JSON SCHEMA (you must follow this exactly):
{
  "issues": [
    {
      "severity": "<string: must be exactly 'failed' or 'warning'>",
      "summary":  "<string: one short line naming the problem, e.g. 'nginx.service has failed'>"
    }
  ],
  "actions": [
    {
      "description": "<string: one actionable sentence; name the exact command to run or file to inspect>",
      "disruptive":  <boolean: true if this action could cause downtime, data loss, or is otherwise risky>
    }
  ]
}

RULES
- Return valid JSON and nothing else. Do not wrap it in backticks or add any text outside the JSON.
- issues[i] and actions[i] must correspond: issue 0 is addressed by action 0, and so on.
- If there are no issues return {"issues":[],"actions":[]}.
- severity must be exactly the string "failed" (service is down / critical error) or "warning" (degraded / non-critical).
- Every command in an action must be correct for the host environment described above.
- If an action could cause downtime or data loss, set disruptive to true.
- %s

System snapshot:
%s`, EnvironmentSection(d.Environment), volumeInstr, formatDoctorSnapshot(d))
}

func formatDoctorSnapshot(d DoctorInput) string {
	var b strings.Builder

	section := func(title, content string) {
		content = strings.TrimSpace(content)
		if content == "" {
			content = "(no data)"
		}
		fmt.Fprintf(&b, "=== %s ===\n%s\n\n", title, content)
	}

	section("Collected at", d.CollectedAt.Format("2006-01-02 15:04:05"))
	section("Uptime", d.Uptime)
	section("Memory", d.Memory)
	section("Disk", d.Disk)
	section("Failed services", d.FailedServices)
	section("Recent errors (journalctl)", d.RecentErrors)
	section("Top processes by CPU", limitLines(d.TopProcesses, maxTopProcessLines))

	return b.String()
}
