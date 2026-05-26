// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"
	"time"
)

// this file contains logic for constructing the prompt for the "doctor" command

type DoctorInput struct {
	CollectedAt    time.Time
	Uptime         string
	Memory         string
	Disk           string
	FailedServices string
	RecentErrors   string
	TopProcesses   string
}

func Doctor(d DoctorInput, brief bool) string {
	var style string
	if brief {
		style = `Write exactly one sentence per detected issue.
For each issue state: what is wrong, and the single most important action to take.
If nothing is wrong, write one sentence confirming the system is healthy.`
	} else {
		style = `Write a structured plain-prose diagnosis.

For each issue you detect, write a short paragraph that covers three things in order:
1. What is wrong and what the likely cause is.
2. The concrete action or actions the user should take to fix or investigate it (for example: a command to run, a file to inspect, a configuration value to change).
3. If any suggested action is potentially disruptive or irreversible, say so explicitly.

Separate each issue paragraph with a blank line.
If multiple issues are related or share a root cause, group them in the same paragraph.
After all issues, write a final short paragraph summarising the overall health of the system.
If nothing is wrong, write a single short paragraph confirming the system is healthy and why you think so.`
	}

	return fmt.Sprintf(`You are a Linux system expert performing a full diagnostic of a live system.

RULES — follow every one without exception:
- Plain prose only. No markdown. No bullet points. No numbered lists. No headers. No bold. No italics.
- Unlike a status report, you MUST suggest concrete next steps for every problem you find.
- Suggested actions must be specific: name the exact command to run, the file to edit, or the configuration key to change.
- If an action could cause data loss, downtime, or is otherwise risky, explicitly state that it is potentially disruptive before describing it.
- Do not open with a preamble. Do not close with an offer to help further.
- %s

System snapshot:
%s`, style, formatDoctorSnapshot(d))
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
