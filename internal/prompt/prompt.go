// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// this file contains the shared prompt templates for all commands

type DoctorInput struct {
	Snapshot *collector.SystemSnapshot
	Language string
	Brief    bool
}
type StatusInput struct {
	Snapshot *collector.SystemSnapshot
	Language string
	Brief    bool
}

const (
	maxRecentErrorLinesBrief = 20
	maxTopProcessLinesBrief  = 5

	maxRecentErrorLinesFull = 50
	maxTopProcessLinesFull  = 10
)

const prohibitions = `
ABSOLUTE PROHIBITIONS — any response that contains any of the following is wrong and must never be produced:
Any text, sentence, or paragraph before <analysis> or after </analysis>.
Any prose introduction such as "Here is the analysis", "Here's a breakdown", "The log contains", "I found the following", or any similar preamble.
Any numbered list (1. 2. 3.) or lettered list (a. b. c.) anywhere in the response.
Any bullet list (- * •) anywhere in the response.
Any section heading, summary paragraph, recommendation list, or conclusion paragraph outside a <component>.
Any text of the form "By addressing these issues…", "I recommend…", "You should…", or "To fix this…" outside a <component>.
Any XML tag not defined in the schema above.
Any truncated field. Every field MUST be complete. Never end a field mid-sentence.
Any ellipsis ("…", "...", "[...]") used to shorten a field. Shortening is forbidden.
Any field that ends with "and more", "etc.", "and so on", or any similar placeholder.
If you are tempted to write any of the above, stop immediately. Write only the <analysis> XML instead.`

const summaryRule = `
MANDATORY SUMMARY COMPONENT — the LAST <component> in every response MUST have <source>summary</source>.
The summary component MUST synthesise the key findings across all preceding components.
The summary component MUST follow exactly the same schema as all other components.
The summary component MUST NOT be truncated or shortened.
The summary component is REQUIRED even when there is only one other component.
If the system is completely healthy with no issues, the single component IS the summary: use <source>summary</source> with a brief healthy-state description.`

func assemblePrompt(instructions, dataLabel, data string) string {
	var b strings.Builder
	b.WriteString(strings.TrimRight(instructions, "\n"))
	b.WriteString(prohibitions)
	b.WriteString(summaryRule)
	b.WriteString("\n\n")
	if dataLabel != "" {
		b.WriteString(dataLabel)
		b.WriteString("\n\n")
	}
	b.WriteString(strings.TrimSpace(data))
	return b.String()
}

func languageLine(lang string) string {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return ""
	}
	return fmt.Sprintf(
		"CRITICAL: Write every human-readable text block inside the tags in this language: %s. "+
			"Keep all command lines, unit names, and identifiers verbatim. Do not use markdown wrappers.\n",
		lang,
	)
}

func formatSnapshot(s *collector.SystemSnapshot) string {
	return buildSnapshot(s, maxRecentErrorLinesBrief, maxTopProcessLinesBrief)
}

func formatSnapshotFull(s *collector.SystemSnapshot) string {
	return buildSnapshot(s, maxRecentErrorLinesFull, maxTopProcessLinesFull)
}

func buildSnapshot(s *collector.SystemSnapshot, maxErrors, maxProcs int) string {
	var b strings.Builder

	sec := func(name, body string) {
		b.WriteString("=== ")
		b.WriteString(name)
		b.WriteString(" ===\n")
		body = strings.TrimSpace(body)
		if body == "" {
			b.WriteString("(no data)\n\n")
			return
		}
		b.WriteString(body)
		b.WriteString("\n\n")
	}

	sec("HOST ENVIRONMENT", environmentBody(s.Environment))
	sec("UPTIME", s.Uptime)
	sec("MEMORY", s.Memory)
	sec("DISK", s.Disk)
	sec("FAILED SERVICES", s.FailedServices)
	sec("RECENT ERRORS", headLines(s.RecentErrors, maxErrors))
	sec("TOP PROCESSES", headLines(s.TopProcesses, maxProcs))

	if len(s.Errors) > 0 {
		sec("COLLECTION NOTES", strings.Join(s.Errors, "\n"))
	}

	return strings.TrimRight(b.String(), "\n")
}

func headLines(text string, n int) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	if len(lines) <= n {
		return text
	}
	return strings.Join(lines[:n], "\n")
}
