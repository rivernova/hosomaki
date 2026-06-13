// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/prompt"
)

// renders the results

func RenderDoctorBrief(result prompt.DoctorBriefResult) string {
	summary := strings.TrimSpace(result.Summary)
	if summary == "" {
		summary = "(no summary)"
	}
	return Section("summary", summary)
}

func RenderDoctorSummary(result prompt.DoctorResult) string {
	disruptive := 0
	for _, act := range result.Actions {
		if act.Disruptive {
			disruptive++
		}
	}
	var b strings.Builder
	b.WriteString(SummaryLine(plural(len(result.Issues), "issue found", "issues found")))
	b.WriteString(SummaryLine(plural(len(result.Actions), "action suggested", "actions suggested")))
	if disruptive > 0 {
		b.WriteString(SummaryLine(plural(disruptive, "action flagged as disruptive", "actions flagged as disruptive")))
	}
	return SectionSummary(b.String())
}

func RenderStatusBrief(result prompt.StatusBriefResult) string {
	summary := strings.TrimSpace(result.Summary)
	if summary == "" {
		summary = "(no summary)"
	}
	return Section("summary", summary)
}

func RenderStatusSummary(result prompt.StatusResult) string {
	critical, warnings := 0, 0
	for _, a := range result.Anomalies {
		if a.Severity == "failed" {
			critical++
		} else {
			warnings++
		}
	}
	var b strings.Builder
	b.WriteString(SummaryLine(plural(critical, "critical issue", "critical issues")))
	b.WriteString(SummaryLine(plural(warnings, "warning", "warnings")))
	return SectionSummary(b.String())
}

func RenderAuditResultSummary(result prompt.AuditResult) string {
	critical, warnings, info := 0, 0, 0
	for _, f := range result.Findings {
		switch f.Severity {
		case "critical":
			critical++
		case "warning":
			warnings++
		default:
			info++
		}
	}
	var b strings.Builder
	b.WriteString(SummaryLine(plural(critical, "critical finding", "critical findings")))
	b.WriteString(SummaryLine(plural(warnings, "warning", "warnings")))
	if info > 0 {
		b.WriteString(SummaryLine(plural(info, "informational finding", "informational findings")))
	}
	return SectionSummary(b.String())
}

func indentProse(text string) string {
	const indent = "     "
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var b strings.Builder
	b.Grow(len(text) + len(indent)*len(lines))
	for _, line := range lines {
		b.WriteString(indent)
		b.WriteString(strings.TrimSpace(line))
		b.WriteByte('\n')
	}
	return b.String()
}

func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, pluralForm)
}

func RenderWhySummary(result prompt.WhyResult) string {
	var b strings.Builder
	b.WriteString(SummaryLine(plural(len(result.Chain),
		"step in failure chain", "steps in failure chain")))
	b.WriteString(SummaryLine(plural(len(result.NextSteps),
		"remediation step suggested", "remediation steps suggested")))
	return SectionSummary(b.String())
}

func RenderPortsResultSummary(result prompt.PortsResult) string {
	warnings, info := 0, 0
	for _, f := range result.Findings {
		switch f.Severity {
		case "warning":
			warnings++
		default:
			info++
		}
	}
	var b strings.Builder
	b.WriteString(SummaryLine(plural(warnings, "warning", "warnings")))
	if info > 0 {
		b.WriteString(SummaryLine(plural(info, "informational finding", "informational findings")))
	}
	return SectionSummary(b.String())
}
