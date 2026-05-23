// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package normalizer

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// ForStatus converts a SystemSnapshot into a structured text payload
// suitable for sending to the AI as context.
func ForStatus(s *collector.SystemSnapshot) string {
	var b strings.Builder

	section := func(title, content string) {
		if strings.TrimSpace(content) == "" {
			content = "(no data)"
		}
		fmt.Fprintf(&b, "=== %s ===\n%s\n\n", title, strings.TrimSpace(content))
	}

	section("Collected at", s.CollectedAt.Format("2006-01-02 15:04:05"))
	section("Uptime", s.Uptime)
	section("Memory", s.Memory)
	section("Disk", s.Disk)
	section("Failed services", s.FailedServices)
	section("Recent errors (journalctl)", s.RecentErrors)
	section("Top processes by CPU", limitLines(s.TopProcesses, 10))

	return b.String()
}

// limitLines returns at most n lines from a multi-line string.
func limitLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n")
}
