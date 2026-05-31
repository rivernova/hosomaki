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

// composition. assembles primitives from ui into the sections each command prints
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

type DoctorCounts struct {
	Anomalies int
	Actions   int
}

type StatusCounts struct {
	FailedServices   int
	WarnServices     int
	PatternsDetected int
}

type ExplainCounts struct {
	Patterns int
	Causes   int
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

func StatusSummary(d SnapshotData, c StatusCounts) string {
	var b strings.Builder
	b.WriteString(SummaryLine(plural(c.FailedServices, "service failing", "services failing")))
	b.WriteString(SummaryLine(plural(c.WarnServices, "service with warnings", "services with warnings")))
	b.WriteString(SummaryLine(plural(c.PatternsDetected, "pattern detected by AI", "patterns detected by AI")))
	return SectionSummary(b.String())
}

func StatusSummaryBrief(d SnapshotData, c StatusCounts) string {
	var b strings.Builder
	b.WriteString(SummaryLine(plural(c.FailedServices, "service failing", "services failing")))
	b.WriteString(SummaryLine(plural(c.PatternsDetected, "pattern detected", "patterns detected")))
	return SectionSummary(b.String())
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

func DoctorSummary(c DoctorCounts) string {
	var b strings.Builder
	b.WriteString(SummaryLine(plural(c.Anomalies, "anomaly detected", "anomalies detected")))
	b.WriteString(SummaryLine(plural(c.Actions, "action suggested", "actions suggested")))
	return SectionSummary(b.String())
}

func DoctorSummaryBrief(c DoctorCounts) string {
	var b strings.Builder
	b.WriteString(SummaryLine(plural(c.Anomalies, "anomaly detected", "anomalies detected")))
	b.WriteString(SummaryLine(plural(c.Actions, "action suggested", "actions suggested")))
	return SectionSummary(b.String())
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

func ExplainSummary(c ExplainCounts) string {
	var b strings.Builder
	b.WriteString(SummaryLine(plural(c.Patterns, "pattern detected", "patterns detected")))
	b.WriteString(SummaryLine(plural(c.Causes, "probable cause", "probable causes")))
	return SectionSummary(b.String())
}

func ParseDoctorAI(text string) DoctorCounts {
	actionStarters := []string{
		"to investigate", "to fix", "to resolve", "to address",
		"you should", "you can ", "you may", "you could",
	}
	summaryStarters := []string{
		"overall", "in summary", "in conclusion", "these issues", "the system is experiencing",
	}

	anomalies := 0
	actions := 0

	for _, p := range splitParagraphs(text) {
		lower := strings.ToLower(p)
		if startsWithAny(lower, summaryStarters) {
			continue
		}
		if startsWithAny(lower, actionStarters) {
			actions += len(regexp.MustCompile("`[^`]+`").FindAllString(p, -1))
			continue
		}
		anomalies++
		actions += len(regexp.MustCompile("`[^`]+`").FindAllString(p, -1))
	}

	return DoctorCounts{Anomalies: anomalies, Actions: actions}
}

func ParseStatusAI(text string, d SnapshotData) StatusCounts {
	failed := len(nonEmptyLines(d.FailedServices))
	warnings := 0
	if strings.TrimSpace(d.RecentErrors) != "" {
		warnings = 1
	}
	patterns := len(splitParagraphs(text))
	return StatusCounts{
		FailedServices:   failed,
		WarnServices:     warnings,
		PatternsDetected: patterns,
	}
}

func ParseExplainAI(text string) ExplainCounts {
	causeKeywords := []string{"because", "caused by", "due to", "result of", "triggered by", "likely"}
	sentences := splitSentences(text)
	causes := 0
	for _, s := range sentences {
		lower := strings.ToLower(s)
		for _, kw := range causeKeywords {
			if strings.Contains(lower, kw) {
				causes++
				break
			}
		}
	}
	return ExplainCounts{
		Patterns: len(sentences),
		Causes:   causes,
	}
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

func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, pluralForm)
}

func splitParagraphs(text string) []string {
	var out []string
	for _, p := range strings.Split(text, "\n\n") {
		if strings.TrimSpace(p) != "" {
			out = append(out, strings.TrimSpace(p))
		}
	}
	return out
}

func splitSentences(text string) []string {
	var out []string
	for _, s := range regexp.MustCompile(`(?:[.!?])\s+`).Split(strings.TrimSpace(text), -1) {
		if strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	if len(out) == 0 && strings.TrimSpace(text) != "" {
		out = append(out, strings.TrimSpace(text))
	}
	return out
}

func startsWithAny(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
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
