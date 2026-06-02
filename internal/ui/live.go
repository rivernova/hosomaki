// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/prompt"
)

// render prompt results into terminal output

func DoctorIssuesHeader() string {
	return sectionHeader("issues")
}

func DoctorActionsHeader() string {
	return sectionHeader("suggested actions")
}

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
	var b strings.Builder
	switch iss.Severity {
	case "failed":
		b.WriteString(BulletTitleFail(label))
	default:
		b.WriteString(BulletTitleWarn(label))
	}
	if detail != "" {
		b.WriteString(indentProse(detail))
	}
	return b.String()
}

func RenderDoctorActionLive(act prompt.DoctorAction, _ int) string {
	desc := strings.TrimSpace(act.Description)
	if desc == "" {
		return ""
	}
	var b strings.Builder
	if act.Disruptive {
		b.WriteString(BulletFail(fmt.Sprintf("[disruptive] %s", desc)))
	} else {
		b.WriteString(BulletOK(desc))
	}
	return b.String()
}

func StatusOverviewHeader() string {
	return sectionHeader("system overview")
}

func StatusAnomaliesHeader() string {
	return sectionHeader("anomalies")
}

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
	var b strings.Builder
	switch a.Severity {
	case "failed":
		b.WriteString(BulletTitleFail(label))
	default:
		b.WriteString(BulletTitleWarn(label))
	}
	if detail != "" {
		b.WriteString(indentProse(detail))
	}
	return b.String()
}

func ExplainIssueHeader(index int) string {
	return sectionHeader(fmt.Sprintf("issue %d", index))
}

func RenderExplainEntryLive(entry prompt.ExplainEntry, index int, multi bool) string {
	what := strings.TrimSpace(entry.What)
	why := strings.TrimSpace(entry.Why)
	if what == "" && why == "" {
		return ""
	}
	if what == "" {
		what = "(no information)"
	}
	if why == "" {
		why = "(no information)"
	}

	var b strings.Builder
	if multi {
		b.WriteString(Section(fmt.Sprintf("issue %d — what is happening", index), what))
		b.WriteString(Section(fmt.Sprintf("issue %d — why it is happening", index), why))
	} else {
		b.WriteString(Section("what is happening", what))
		b.WriteString(Section("why it is happening", why))
	}
	return b.String()
}
