// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// this file contains shared prompt utilities and constants.

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
Any section heading, summary paragraph, recommendation list, or conclusion paragraph.
Any text of the form "By addressing these issues…", "I recommend…", "You should…", or "To fix this…".
Any XML tag not defined in the schema above.
If you are tempted to write any of the above, stop immediately. Write only the <analysis> XML instead.`

func languageLine(lang string) string {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return ""
	}
	return fmt.Sprintf(
		"CRITICAL: Write every human-readable text block inside the tags in this language: %s. "+
			"Keep all command lines, unit names and identifiers verbatim. Do not use markdown wrappers.\n",
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
	sec("RECENT ERRORS", limitLines(s.RecentErrors, maxErrors))
	sec("TOP PROCESSES", limitLines(s.TopProcesses, maxProcs))

	if len(s.Errors) > 0 {
		sec("COLLECTION NOTES", strings.Join(s.Errors, "\n"))
	}

	return strings.TrimRight(b.String(), "\n")
}

func limitLines(text string, n int) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	if len(lines) <= n {
		return text
	}
	kept := append(lines[:n:n], fmt.Sprintf("… (%d more lines omitted)", len(lines)-n))
	return strings.Join(kept, "\n")
}
