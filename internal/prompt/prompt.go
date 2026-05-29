// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// this file contains shared prompt utilities

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
	maxRecentErrorLines = 20
	maxTopProcessLines  = 8
)

func languageLine(lang string) string {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return ""
	}
	return fmt.Sprintf(
		"Write every human-readable text value in this language: %s. "+
			"Keep command lines, unit names and identifiers verbatim.\n",
		lang,
	)
}

func formatSnapshot(s *collector.SystemSnapshot) string {
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
	sec("RECENT ERRORS", limitLines(s.RecentErrors, maxRecentErrorLines))
	sec("TOP PROCESSES", limitLines(s.TopProcesses, maxTopProcessLines))

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
