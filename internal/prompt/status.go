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

func Status(s StatusInput, brief bool) string {
	var style string
	if brief {
		style = `OUTPUT FORMAT — follow this exactly, no exceptions:
- Output exactly ONE sentence. Hard limit: 30 words.
- State overall health and name the single most critical issue if any exists.
- NOTHING else. No introduction. No elaboration. No second sentence. Stop immediately after the period.`
	} else {
		style = "Write a paragraph of five to eight sentences summarising the overall health of this system. Cover uptime, memory, disk, failed services, and recent errors. Call out any anomalies or points of concern."
	}

	return fmt.Sprintf(`You are a Linux system expert reading a live system snapshot.

%sRULES — follow every one without exception:
- Plain prose only. No markdown. No bullet points. No numbered lists. No headers. No bold. No italics.
- Do not suggest fixes, commands to run, or remediation steps of any kind.
- Do not open with a preamble. Do not close with an offer to help further.
- %s

After your analysis, on a new line write exactly:
---JSON---
{"failed_services": <count of failed services>, "warn_services": <count of services with warnings or errors>, "patterns_detected": <count of distinct issues or anomalies you observed>}
---END---

System snapshot:
%s`, EnvironmentSection(s.Environment), style, formatSnapshot(s))
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
