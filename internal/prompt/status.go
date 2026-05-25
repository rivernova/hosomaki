// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"
	"time"
)

// this file contains logic for constructing the prompt sent to the AI provider for status summaries

type StatusInput struct {
	CollectedAt    time.Time
	Uptime         string
	Memory         string
	Disk           string
	FailedServices string
	RecentErrors   string
	TopProcesses   string
}

func Status(s StatusInput, brief bool) string {
	style := "Write a clear, concise paragraph (5–8 sentences) summarising system health. Highlight any anomalies or points of attention."
	if brief {
		style = "Summarise system health in a single sentence. Mention the most critical issue if any."
	}

	return fmt.Sprintf(`You are a Linux system expert. Here is a snapshot of the current system state.

%s

Rules: plain text only, no markdown, no bullet points.

System snapshot:
%s`, style, formatSnapshot(s))
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
	section("Top processes by CPU", s.TopProcesses)

	return b.String()
}
