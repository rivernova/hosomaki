// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type SnapshotData struct {
	CollectedAt    time.Time
	Uptime         string
	Memory         string
	Disk           string
	FailedServices string
	RecentErrors   string
}

type ExplainContext struct {
	Source string
	Cmd    string
	Lines  int
}

func StatusHeader() string {
	return Title("status")
}

func StatusHeaderBrief() string {
	return Title("status (brief)")
}

func StatusSystemSection(d SnapshotData) string {
	return Section("system status", systemKV(d))
}

func StatusSystemSectionBrief(d SnapshotData) string {
	return SectionCompact("system", systemCompact(d))
}

func StatusInsightsSection(d SnapshotData) string {
	return Section("local insights", insightBullets(d))
}

func StatusInsightsSectionBrief(d SnapshotData) string {
	return SectionCompact("insights", insightBullets(d))
}

func StatusAIHeader() string {
	return sectionHeader("ai analysis")
}

func StatusAIHeaderBrief() string {
	return compactHeader("ai")
}

func StatusSummary(d SnapshotData) string {
	return Section("summary", summaryBullets(d))
}

func StatusSummaryBrief(d SnapshotData) string {
	return SectionCompact("summary", summaryCompact(d))
}

func DoctorHeader() string {
	return Title("doctor")
}

func DoctorHeaderBrief() string {
	return Title("doctor (brief)")
}

func DoctorSystemSection(d SnapshotData) string {
	return Section("system analysis", systemKV(d))
}

func DoctorSystemSectionBrief(d SnapshotData) string {
	return SectionCompact("system", systemCompact(d))
}

func DoctorInsightsSection(d SnapshotData) string {
	return Section("local insights", insightBullets(d))
}

func DoctorInsightsSectionBrief(d SnapshotData) string {
	return SectionCompact("insights", insightBullets(d))
}

func DoctorAIHeader() string {
	return sectionHeader("ai analysis")
}

func DoctorAIHeaderBrief() string {
	return compactHeader("ai")
}

func DoctorSummary(d SnapshotData) string {
	return Section("summary", summaryBullets(d))
}

func DoctorSummaryBrief(d SnapshotData) string {
	return SectionCompact("summary", summaryCompact(d))
}

func ExplainHeader() string {
	return Title("explain")
}

func ExplainContextSection(c ExplainContext) string {
	return Section("context", explainContextKV(c))
}

func ExplainExplanationSection(c ExplainContext) string {
	return Section("explanation", explainBullets(c))
}

func ExplainAIHeader() string {
	return sectionHeader("ai analysis")
}

func ExplainSummary(c ExplainContext) string {
	return Section("summary", explainSummaryBullets(c))
}

func systemKV(d SnapshotData) string {
	var b strings.Builder
	b.WriteString(KeyValue("uptime", formatUptime(d.Uptime)))
	for _, line := range formatMemory(d.Memory) {
		b.WriteString(line)
	}
	for _, line := range formatDisk(d.Disk) {
		b.WriteString(line)
	}
	return b.String()
}

func systemCompact(d SnapshotData) string {
	var parts []string
	if u := formatUptime(d.Uptime); u != "" {
		parts = append(parts, u)
	}
	if lines := formatMemory(d.Memory); len(lines) > 0 {
		parts = append(parts, strings.TrimRight(lines[0], "\n"))
	}
	if lines := formatDisk(d.Disk); len(lines) > 0 {
		parts = append(parts, strings.TrimRight(lines[0], "\n"))
	}
	if len(parts) == 0 {
		return "(no data)\n"
	}
	return strings.Join(parts, " · ") + "\n"
}

func formatUptime(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "(none)"
	}
	s = strings.TrimPrefix(s, "up ")
	s = regexp.MustCompile(`(\d+)\s+days?`).ReplaceAllString(s, "${1}d")
	s = regexp.MustCompile(`(\d+)\s+hours?`).ReplaceAllString(s, "${1}h")
	s = regexp.MustCompile(`(\d+)\s+minutes?`).ReplaceAllString(s, "${1}m")
	s = strings.ReplaceAll(s, ",", "")
	return strings.Join(strings.Fields(s), " ")
}

func formatMemory(raw string) []string {
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 7 && fields[0] == "Mem:" {
			used := cleanUnit(fields[2])
			total := cleanUnit(fields[1])
			avail := cleanUnit(fields[6])
			out = append(out, KeyValue("memory", fmt.Sprintf("%s / %s  (%s free)", used, total, avail)))
		}
		if len(fields) >= 3 && fields[0] == "Swap:" {
			if fields[2] == "0B" || fields[2] == "0" {
				out = append(out, KeyValue("swap", "inactive"))
			} else {
				out = append(out, KeyValue("swap", fmt.Sprintf("%s / %s", cleanUnit(fields[2]), cleanUnit(fields[1]))))
			}
		}
	}
	if len(out) == 0 {
		out = append(out, KeyValue("memory", "(none)"))
	}
	return out
}

func formatDisk(raw string) []string {
	var out []string
	seen := map[string]bool{}
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 || !strings.HasPrefix(fields[0], "/dev/") {
			continue
		}
		dev := fields[0]
		if seen[dev] {
			continue
		}
		seen[dev] = true
		used := cleanUnit(fields[2])
		size := cleanUnit(fields[1])
		pct := fields[4]
		mount := fields[5]
		out = append(out, KeyValue("disk "+mount, fmt.Sprintf("%s / %s  (%s)", used, size, pct)))
	}
	if len(out) == 0 {
		out = append(out, KeyValue("disk", "(none)"))
	}
	return out
}

func cleanUnit(s string) string {
	s = regexp.MustCompile(`(\d)Gi\b`).ReplaceAllString(s, "${1}G")
	s = regexp.MustCompile(`(\d)Mi\b`).ReplaceAllString(s, "${1}M")
	s = regexp.MustCompile(`(\d)Ki\b`).ReplaceAllString(s, "${1}K")
	s = regexp.MustCompile(`(\d+)\.0([GMK])`).ReplaceAllString(s, "${1}${2}")
	return s
}

func insightBullets(d SnapshotData) string {
	var b strings.Builder
	if strings.TrimSpace(d.FailedServices) == "" {
		b.WriteString(BulletOK("no failed services"))
	} else {
		for _, line := range nonEmptyLines(d.FailedServices) {
			b.WriteString(BulletFail(line))
		}
	}
	if strings.TrimSpace(d.RecentErrors) == "" {
		b.WriteString(BulletOK("no recent errors in journal"))
	} else {
		b.WriteString(BulletWarn("recent errors detected in journal"))
	}
	return b.String()
}

func summaryBullets(d SnapshotData) string {
	var b strings.Builder
	failedCount := len(nonEmptyLines(d.FailedServices))
	if failedCount == 0 {
		b.WriteString(BulletSummary("no failed services"))
	} else {
		b.WriteString(BulletSummary(fmt.Sprintf("%d failed service(s) detected", failedCount)))
	}
	if strings.TrimSpace(d.RecentErrors) == "" {
		b.WriteString(BulletSummary("no recent errors"))
	} else {
		b.WriteString(BulletSummary("recent errors present in journal"))
	}
	return b.String()
}

func summaryCompact(d SnapshotData) string {
	failedCount := len(nonEmptyLines(d.FailedServices))
	hasErrors := strings.TrimSpace(d.RecentErrors) != ""
	switch {
	case failedCount > 0 && hasErrors:
		return fmt.Sprintf("%d failed · errors present\n", failedCount)
	case failedCount > 0:
		return fmt.Sprintf("%d failed service(s)\n", failedCount)
	case hasErrors:
		return "errors present in journal\n"
	default:
		return "system healthy\n"
	}
}

func explainContextKV(c ExplainContext) string {
	var b strings.Builder
	b.WriteString(KeyValue("source", orNone(c.Source)))
	if c.Cmd != "" {
		b.WriteString(KeyValue("command", c.Cmd))
	}
	if c.Lines > 0 {
		b.WriteString(KeyValue("lines", fmt.Sprintf("%d", c.Lines)))
	}
	return b.String()
}

func explainBullets(c ExplainContext) string {
	var b strings.Builder
	if c.Cmd != "" {
		b.WriteString(BulletOK(fmt.Sprintf("originating command: %s", c.Cmd)))
	}
	if c.Source != "" {
		b.WriteString(BulletOK(fmt.Sprintf("input source: %s", c.Source)))
	}
	if c.Lines > 0 {
		b.WriteString(BulletWarn(fmt.Sprintf("reading last %d lines", c.Lines)))
	}
	return b.String()
}

func explainSummaryBullets(c ExplainContext) string {
	var b strings.Builder
	b.WriteString(BulletSummary(fmt.Sprintf("source: %s", orNone(c.Source))))
	if c.Cmd != "" {
		b.WriteString(BulletSummary(fmt.Sprintf("command: %s", c.Cmd)))
	}
	return b.String()
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(none)"
	}
	return strings.TrimSpace(s)
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return out
}
