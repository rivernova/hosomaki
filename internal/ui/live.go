// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/prompt"
)

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

func AuditFindingsHeader() string { return sectionHeader("ai analysis") }

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
