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

// template for the status command prompt

const maxTopProcessLines = 10

type StatusInput struct {
	CollectedAt    time.Time
	Environment    collector.Environment
	Uptime         string
	Memory         string
	Disk           string
	FailedServices string
	RecentErrors   string
	TopProcesses   string
}

type StatusAnomaly struct {
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

type StatusResult struct {
	Overview  string          `json:"overview"`
	Anomalies []StatusAnomaly `json:"anomalies"`
}

type StatusBriefResult struct {
	Summary string `json:"summary"`
}

func Status(s StatusInput, brief bool) string {
	if brief {
		return fmt.Sprintf(`You are a Linux system expert reading a live system snapshot.

%s
TASK
Analyse the system snapshot below. Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.

The JSON must use exactly these field names:

SCHEMA
{"summary": "string"}

FIELD RULES
- "summary": exactly one sentence, maximum 30 words. State overall health and the single most critical issue if any.

System snapshot:
%s`, EnvironmentSection(s.Environment), formatSnapshot(s))
	}

	return fmt.Sprintf(`You are a Linux system expert reading a live system snapshot.

%s
TASK
Analyse the system snapshot below. Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.

The JSON must use exactly these field names. Do not rename, abbreviate, or add fields.

SCHEMA
{
  "overview": "string",
  "anomalies": [
    {
      "severity": "string",
      "title": "string",
      "detail": "string"
    }
  ]
}

FIELD RULES
- "overview": 3–5 sentences of prose covering uptime, memory, and disk. Do not mention any problems or anomalies here.
- "anomalies": every issue, warning, or concern found in the snapshot.
- "severity": the string "failed" for a downed or broken component, "warning" for degraded or concerning.
- "title": a concise plain-text label for the anomaly, e.g. "postgresql.service has failed".
- "detail": 2–4 sentences. Describe exactly what was observed (reference specific values or log lines),
  explain why it is a problem, and state what impact it has or could have on the system.
- If no anomalies exist return an empty array.

System snapshot:
%s`, EnvironmentSection(s.Environment), formatSnapshot(s))
}

func formatSnapshot(s StatusInput) string {
	var b strings.Builder

	section := func(title, content string) {
		content = strings.TrimSpace(content)
		if content == "" {
			content = "(no data)"
		}
		fmt.Fprintf(&b, "=== %s ===\n%s\n\n", title, content)
	}

	section("Collected at", s.CollectedAt.Format("2006-01-02 15:04:05"))
	section("Uptime", s.Uptime)
	section("Memory", s.Memory)
	section("Disk", s.Disk)
	section("Failed services", s.FailedServices)
	section("Recent errors (journalctl)", s.RecentErrors)
	section("Top processes by CPU", limitLines(s.TopProcesses, maxTopProcessLines))

	return b.String()
}

func limitLines(s string, n int) string {
	if s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n")
}
