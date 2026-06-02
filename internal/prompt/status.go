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

// template for prompt for status command

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
	Summary  string `json:"summary"`
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
Analyse the system snapshot below and return a JSON object — and nothing else.
No prose, no explanation, no markdown fences, no commentary. Pure JSON only.

JSON SCHEMA:
{
  "summary": "<string: exactly one sentence, maximum 30 words, stating overall health and the most critical issue if any>"
}

RULES
- Return valid JSON and nothing else.
- summary must be a single sentence of at most 30 words.

System snapshot:
%s`, EnvironmentSection(s.Environment), formatSnapshot(s))
	}

	return fmt.Sprintf(`You are a Linux system expert reading a live system snapshot.

%s
TASK
Analyse the system snapshot below and return a JSON object — and nothing else.
No prose, no explanation, no markdown fences, no commentary. Pure JSON only.

JSON SCHEMA:
{
  "overview":  "<string: 2–4 sentence prose paragraph covering uptime, memory, and disk health; must not mention any problems or anomalies>",
  "anomalies": [
    {
      "severity": "<string: must be exactly 'failed' or 'warning'>",
      "summary":  "<string: one short line describing the anomaly>"
    }
  ]
}

RULES
- Return valid JSON and nothing else. Do not wrap it in backticks or add any text outside the JSON.
- overview must describe uptime, memory, and disk only. Never mention problems in overview.
- anomalies lists every issue found: failed services, high memory/disk usage, concerning error patterns.
- If no anomalies exist, return an empty array: "anomalies": [].
- severity must be exactly "failed" (service is down / critical) or "warning" (degraded / non-critical).

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
