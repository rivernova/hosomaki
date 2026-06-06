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

func ExplainIssueHeader(index int) string {
	return sectionHeader(fmt.Sprintf("issue %d", index))
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
