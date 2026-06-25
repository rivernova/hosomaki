// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/prompt"
)

// live ux

func DoctorIssuesHeader() string  { return sectionHeader("issues") }
func DoctorActionsHeader() string { return sectionHeader("suggested actions") }
func RenderDoctorIssueLive(iss prompt.DoctorIssue, _ int) string {
	title := strings.TrimSpace(iss.Title)
	detail := strings.TrimSpace(iss.Detail)
	if title == "" && detail == "" {
		return ""
	}
	label := title
	if label == "" {
		label = detail
		detail = ""
	}
	var bullet string
	if iss.Severity == "failed" {
		bullet = BulletTitleFail(label)
	} else {
		bullet = BulletTitleWarn(label)
	}
	if detail == "" {
		return bullet
	}
	return bullet + indentProse(detail)
}

func RenderDoctorActionLive(act prompt.DoctorAction, _ int) string {
	desc := strings.TrimSpace(act.Description)
	if desc == "" {
		return ""
	}
	if act.Disruptive {
		return BulletFail(fmt.Sprintf("[disruptive] %s", desc))
	}
	return BulletOK(desc)
}

func StatusOverviewHeader() string  { return sectionHeader("system overview") }
func StatusAnomaliesHeader() string { return sectionHeader("anomalies") }

func RenderStatusOverviewLive(overview string) string {
	t := strings.TrimSpace(overview)
	if t == "" {
		return ""
	}
	return t + "\n"
}

func RenderStatusAnomalyLive(a prompt.StatusAnomaly, _ int) string {
	title := strings.TrimSpace(a.Title)
	detail := strings.TrimSpace(a.Detail)
	if title == "" && detail == "" {
		return ""
	}
	label := title
	if label == "" {
		label = detail
		detail = ""
	}
	var bullet string
	if a.Severity == "failed" {
		bullet = BulletTitleFail(label)
	} else {
		bullet = BulletTitleWarn(label)
	}
	if detail == "" {
		return bullet
	}
	return bullet + indentProse(detail)
}

func RenderExplainEntryLive(entry prompt.ExplainEntry, index int, multi bool) string {
	what := strings.TrimSpace(entry.What)
	why := strings.TrimSpace(entry.Why)
	if what == "" {
		what = "(no information)"
	}
	if why == "" {
		why = "(no information)"
	}
	if multi {
		return Section(fmt.Sprintf("issue %d — what is happening", index), what) +
			Section(fmt.Sprintf("issue %d — why it is happening", index), why)
	}
	return Section("what is happening", what) +
		Section("why it is happening", why)
}

func AuditFindingsHeader() string { return sectionHeader("analysis") }

func RenderAuditSummaryLive(summary string) string {
	t := strings.TrimSpace(summary)
	if t == "" {
		return ""
	}
	return t + "\n"
}

func RenderAuditFindingLive(f prompt.AuditFinding, _ int) string {
	title := strings.TrimSpace(f.Title)
	detail := strings.TrimSpace(f.Detail)
	if title == "" && detail == "" {
		return ""
	}
	label := title
	if label == "" {
		label = detail
		detail = ""
	}

	var bullet string
	switch f.Severity {
	case "critical":
		bullet = BulletTitleFail(label)
	case "warning":
		bullet = BulletTitleWarn(label)
	default: // "info" and any unexpected value
		bullet = BulletOK(label)
	}

	if detail == "" {
		return bullet
	}
	return bullet + indentProse(detail)
}

func WhySummaryHeader() string   { return sectionHeader("failure summary") }
func WhyChainHeader() string     { return sectionHeader("failure chain") }
func WhyNextStepsHeader() string { return sectionHeader("next steps") }

func RenderWhySummaryLive(summary string) string {
	t := strings.TrimSpace(summary)
	if t == "" {
		return ""
	}
	return t + "\n"
}

func RenderWhyStepLive(step prompt.WhyStep, index int) string {
	event := strings.TrimSpace(step.Event)
	detail := strings.TrimSpace(step.Detail)
	if event == "" && detail == "" {
		return ""
	}
	label := event
	if label == "" {
		label = detail
		detail = ""
	}
	bullet := BulletWarn(fmt.Sprintf("step %d — %s", index, label))
	if detail == "" {
		return bullet
	}
	return bullet + indentProse(detail)
}

func RenderWhyNextStepLive(step string, index int) string {
	text := strings.TrimSpace(step)
	if text == "" {
		return ""
	}
	return BulletOK(fmt.Sprintf("%d. %s", index, text))
}

func RenderPortsSummaryLive(summary string) string {
	t := strings.TrimSpace(summary)
	if t == "" {
		return ""
	}
	return t + "\n"
}

func RenderPortsFindingLive(f prompt.PortsFinding, _ int) string {
	title := strings.TrimSpace(f.Title)
	detail := strings.TrimSpace(f.Detail)
	if title == "" && detail == "" {
		return ""
	}
	label := title
	if label == "" {
		label = detail
		detail = ""
	}
	var bullet string
	switch f.Severity {
	case "warning":
		bullet = BulletTitleWarn(label)
	default:
		bullet = BulletOK(label)
	}
	if detail == "" {
		return bullet
	}
	return bullet + indentProse(detail)
}

func TimersFindingsHeader() string { return sectionHeader("analysis") }

func RenderTimersSummaryLive(summary string) string {
	t := strings.TrimSpace(summary)
	if t == "" {
		return ""
	}
	return t + "\n"
}

func RenderTimerLive(entry prompt.TimerEntry, _ int) string {
	name := strings.TrimSpace(entry.Name)
	if name == "" {
		return ""
	}

	schedule := strings.TrimSpace(entry.Schedule)
	label := name
	if schedule != "" {
		label = fmt.Sprintf("%s  —  %s", name, schedule)
	}

	var bullet string
	switch entry.Status {
	case "failed":
		bullet = BulletTitleFail(label)
	case "warning":
		bullet = BulletTitleWarn(label)
	default:
		bullet = BulletOK(label)
	}

	var meta strings.Builder
	if v := strings.TrimSpace(entry.LastRun); v != "" {
		meta.WriteString(KeyValue("last run", v))
	}
	if v := strings.TrimSpace(entry.NextRun); v != "" {
		meta.WriteString(KeyValue("next run", v))
	}
	if detail := strings.TrimSpace(entry.Detail); detail != "" {
		meta.WriteString(indentProse(detail))
	}

	if meta.Len() == 0 {
		return bullet
	}
	return bullet + meta.String()
}

func CronsFindingsHeader() string { return sectionHeader("analysis") }

func RenderCronsSummaryLive(summary string) string {
	t := strings.TrimSpace(summary)
	if t == "" {
		return ""
	}
	return t + "\n"
}

func RenderCronJobLive(entry prompt.CronJobEntry, _ int) string {
	source := strings.TrimSpace(entry.Source)
	if source == "" {
		return ""
	}

	schedule := strings.TrimSpace(entry.Schedule)
	whatItDoes := strings.TrimSpace(entry.WhatItDoes)

	var label string
	switch {
	case schedule != "" && whatItDoes != "":
		label = fmt.Sprintf("%s  —  %s  —  %s", source, schedule, whatItDoes)
	case schedule != "":
		label = fmt.Sprintf("%s  —  %s", source, schedule)
	default:
		label = source
	}

	var bullet string
	switch entry.Status {
	case "failed":
		bullet = BulletTitleFail(label)
	case "warning":
		bullet = BulletTitleWarn(label)
	default:
		bullet = BulletOK(label)
	}

	detail := strings.TrimSpace(entry.Detail)
	if detail == "" {
		return bullet
	}
	return bullet + indentProse(detail)
}

func MountsFindingsHeader() string { return sectionHeader("analysis") }

func RenderMountsSummaryLive(summary string) string {
	t := strings.TrimSpace(summary)
	if t == "" {
		return ""
	}
	return t + "\n"
}

func RenderMountsFindingLive(f prompt.MountFinding, _ int) string {
	title := strings.TrimSpace(f.Title)
	detail := strings.TrimSpace(f.Detail)
	if title == "" && detail == "" {
		return ""
	}
	label := title
	if label == "" {
		label = detail
		detail = ""
	}

	var bullet string
	switch f.Severity {
	case "critical":
		bullet = BulletTitleFail(label)
	case "warning":
		bullet = BulletTitleWarn(label)
	default:
		bullet = BulletOK(label)
	}

	if detail == "" {
		return bullet
	}
	return bullet + indentProse(detail)
}

func UpdatesFindingsHeader() string { return sectionHeader("analysis") }

func RenderUpdatesSummaryLive(summary string) string {
	t := strings.TrimSpace(summary)
	if t == "" {
		return ""
	}
	return t + "\n"
}

func RenderUpdatesFindingLive(u prompt.UpdateFinding, _ int) string {
	title := strings.TrimSpace(u.Package)
	avail := strings.TrimSpace(u.Available)
	if title == "" {
		return ""
	}

	var label strings.Builder
	label.WriteString(title)

	if u.Category != "" && u.Category != "unknown" {
		label.WriteString(" [")
		label.WriteString(u.Category)
		label.WriteString("]")
	}
	if u.RebootRequired {
		label.WriteString(" [reboot required]")
	}

	var bullet string
	switch u.Category {
	case "security":
		bullet = BulletTitleFail(label.String())
	case "major":
		bullet = BulletTitleWarn(label.String())
	default:
		bullet = BulletOK(label.String())
	}

	inst := u.Installed
	if inst == "" {
		inst = "?"
	}

	var out strings.Builder
	out.WriteString(bullet)

	switch {
	case avail != "" && inst != "?":
		out.WriteString("  " + inst + " → " + avail + "\n")
	case avail != "":
		out.WriteString("  → " + avail + "\n")
	default:
		out.WriteString("\n")
	}

	if (u.Category == "security" || u.Category == "major") && strings.TrimSpace(u.Detail) != "" {
		out.WriteString(indentProse(u.Detail))
	}

	return out.String()
}

func HistoryFindingsHeader() string { return sectionHeader("analysis") }

func RenderHistorySummaryLive(summary string) string {
	t := strings.TrimSpace(summary)
	if t == "" {
		return ""
	}
	return t + "\n"
}

func RenderHistoryEntryLive(e prompt.HistoryEntry, _ int) string {
	ts := strings.TrimSpace(e.Timestamp)
	cmd := strings.TrimSpace(e.Command)
	sum := strings.TrimSpace(e.Summary)
	if cmd == "" {
		return ""
	}
	label := fmt.Sprintf("[%s] %s", ts, cmd)
	bullet := BulletOK(label)
	if sum == "" {
		return bullet
	}
	return bullet + indentProse(sum)

}

func FirewallFindingsHeader() string { return sectionHeader("analysis") }

func RenderFirewallSummaryLive(summary string) string {
	t := strings.TrimSpace(summary)
	if t == "" {
		return ""
	}
	return t + "\n"
}

func RenderFirewallFindingLive(f prompt.FirewallFinding, _ int) string {
	title := strings.TrimSpace(f.Title)
	detail := strings.TrimSpace(f.Detail)
	if title == "" && detail == "" {
		return ""
	}
	label := title
	if label == "" {
		label = detail
		detail = ""
	}

	var bullet string
	switch f.Severity {
	case "critical":
		bullet = BulletTitleFail(label)
	case "warning":
		bullet = BulletTitleWarn(label)
	default:
		bullet = BulletOK(label)
	}

	var out strings.Builder
	out.WriteString(bullet)
	if f.Rule != "" {
		out.WriteString(KeyValue("  rule", f.Rule))
	}
	if f.Port != "" {
		out.WriteString(KeyValue("  port", f.Port))
	}
	if detail != "" {
		out.WriteString(indentProse(detail))
	}
	return out.String()
}
