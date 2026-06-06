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

// prompt logic for the doctor command

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
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

type DoctorAction struct {
	Description string `json:"description"`
	Disruptive  bool   `json:"disruptive"`
}

type DoctorResult struct {
	Issues  []DoctorIssue  `json:"issues"`
	Actions []DoctorAction `json:"actions"`
}

type DoctorBriefResult struct {
	Summary string `json:"summary"`
}

func Doctor(d DoctorInput, brief bool) string {
	if brief {
		return fmt.Sprintf(`You are a Linux system expert reading a live system snapshot.

%s
TASK
Analyse the system snapshot below. Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.

The JSON must use exactly these field names:

SCHEMA
%s

FIELD RULES
- "summary": exactly ONE sentence, maximum 35 words. State overall health and the single most critical issue found (if any), and name the one immediate action to take. If the system is healthy, say so.

OUTPUT FORMAT
No markdown. No bullet points. No numbered lists. No headers. The value of "summary" is plain prose.

System snapshot:
%s`, EnvironmentSection(d.Environment), SchemaDoctorBrief, formatDoctorSnapshot(d))
	}

	return fmt.Sprintf(`You are a Linux system expert performing a full diagnostic of a live system.

%s
TASK
Analyse the system snapshot below. Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.

The JSON must use exactly these field names. Do not rename, abbreviate, or add fields.

SCHEMA
%s

FIELD RULES
- "severity": the string "failed" for a downed or broken component, "warning" for degraded or concerning.
- "title": plain text label, no punctuation at the end.
- "detail": a thorough diagnostic paragraph (3–6 sentences). Explain what is wrong,
  what the evidence in the snapshot shows, what the likely root cause is, and what
  impact this has or could have on the system. Reference specific values, service
  names, and log excerpts from the snapshot where relevant.
- "description": a full paragraph (2–4 sentences) describing the concrete action to
  take. Name the exact command to run or file to inspect. Explain what to look for and
  what a successful outcome looks like. If multiple steps are needed, describe them
  in sequence. If the action is potentially disruptive or irreversible, say so first.
- "disruptive": boolean true only if the action risks downtime or data loss.
- issues[i] and actions[i] correspond 1-to-1: issue 0 is fixed by action 0.
- All commands in "description" must be correct for the host environment above.
- To resolve each issue, suggest concrete next steps.
- If the system is healthy return {"issues":[],"actions":[]}.

OUTPUT FORMAT
No markdown. No bullet points. No numbered lists. No headers. All string values are plain prose.

System snapshot:
%s`, EnvironmentSection(d.Environment), SchemaDoctorFull, formatDoctorSnapshot(d))
}

func formatDoctorSnapshot(d DoctorInput) string {
	var b strings.Builder

	section := func(title, content string) {
		content = strings.TrimSpace(content)
		if content == "" {
			content = "(no data)"
		}
		_, err := fmt.Fprintf(&b, "=== %s ===\n%s\n\n", title, content)
		if err != nil {
			return
		}
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
