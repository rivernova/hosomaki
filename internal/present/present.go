// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package present

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/analysis"
	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/insight"
	"github.com/rivernova/hosomaki/internal/render"
)

// this file contains the logic for turning analysis reports and parsed insights
// into view models that the renderer transforms into terminal output
func AnalysisInput(s *collector.SystemSnapshot) analysis.Input {
	if s == nil {
		return analysis.Input{}
	}
	kernel := s.Environment.KernelFull
	if kernel == "" {
		kernel = s.Environment.Kernel
	}
	return analysis.Input{
		Kernel:         kernel,
		Uptime:         s.Uptime,
		Memory:         s.Memory,
		Disk:           s.Disk,
		FailedServices: s.FailedServices,
		RecentErrors:   s.RecentErrors,
	}
}

func ContextLine(command string) string {
	if command == "" {
		return ""
	}
	return "produced by: " + command
}

func DoctorReport(rep analysis.Report, ai insight.Doctor, brief bool) render.DoctorReport {
	out := render.DoctorReport{
		Title:        "hosomaki doctor",
		Metrics:      toMetrics(rep),
		Findings:     toFindings(rep),
		ProcessLines: doctorProcessLines(),
		RawInsight:   ai.Raw,
		Brief:        brief,
	}

	for _, iss := range ai.Issues {
		out.Issues = append(out.Issues, toIssue(iss))
	}

	out.Summary = doctorSummary(rep, ai)

	if brief {
		out.ProcessLines = nil
	}
	return out
}

func StatusReport(rep analysis.Report, ai insight.Status, brief bool) render.StatusReport {
	return render.StatusReport{
		Title:    "hosomaki status",
		Metrics:  toMetrics(rep),
		Services: toFindings(rep),
		Summary:  statusSummary(rep, ai),
		Brief:    brief,
	}
}

func StatusReportWithObservations(rep analysis.Report, ai insight.Status, brief bool) render.StatusReport {
	sr := StatusReport(rep, ai, brief)
	sr.RawAI = strings.TrimSpace(ai.Raw)
	sr.BriefText = buildBriefTextFromObservations(ai)
	sr.Summary = statusHealthSummary(rep, len(ai.Observations))

	for _, obs := range ai.Observations {
		if obs.Text == "" {
			continue
		}
		text := stripObservationPrefix(obs.Text)
		pattern := text
		cause := ""
		if idx := strings.Index(text, " — "); idx >= 0 {
			pattern = text[:idx]
			cause = text[idx+3:]
		}
		iss := render.Issue{
			Subject: observationSubject(obs.Text),
			Status:  sevStatus(obs.Level),
		}
		if pattern != "" {
			iss.Details = append(iss.Details, render.Detail{Key: "detected pattern", Value: pattern})
		}
		if cause != "" {
			iss.Details = append(iss.Details, render.Detail{Key: "probable cause", Value: cause})
		}
		sr.Issues = append(sr.Issues, iss)
	}
	return sr
}

func observationSubject(text string) string {
	text = strings.TrimSpace(text)
	if len(text) > 0 && text[0] == '[' {
		end := strings.Index(text, "]")
		if end > 0 {
			return strings.TrimSpace(text[1:end])
		}
	}
	return "system"
}

func buildBriefTextFromObservations(ai insight.Status) string {
	if strings.TrimSpace(ai.Raw) != "" {
		text := strings.TrimSpace(ai.Raw)
		if idx := strings.IndexAny(text, ".!?"); idx >= 0 && idx < len(text)-1 {
			text = text[:idx+1]
		}
		return text
	}
	if len(ai.Observations) > 0 {
		return stripObservationPrefix(ai.Observations[0].Text)
	}
	return "system operating normally"
}

func statusHealthSummary(rep analysis.Report, aiPatterns int) []render.SummaryItem {
	var items []render.SummaryItem

	critCount := countCritFindings(rep)
	if critCount > 0 {
		items = append(items, render.SummaryItem{
			Text:   plural(critCount, "service failing", "services failing"),
			Status: render.Crit,
		})
	}
	if rep.FailedCount > 0 {
		items = append(items, render.SummaryItem{
			Text:   plural(rep.FailedCount, "service degraded", "services degraded"),
			Status: render.Warn,
		})
	}
	if aiPatterns > 0 {
		items = append(items, render.SummaryItem{
			Text:   plural(aiPatterns, "pattern detected by AI", "patterns detected by AI"),
			Status: render.Info,
		})
	}
	if len(items) == 0 {
		items = append(items, render.SummaryItem{Text: "healthy system", Status: render.OK})
	}
	return items
}

func stripObservationPrefix(text string) string {
	text = strings.TrimSpace(text)
	if len(text) == 0 || text[0] != '[' {
		return text
	}
	end := strings.Index(text, "]")
	if end < 0 {
		return text
	}
	return strings.TrimSpace(text[end+1:])
}

func ExplainReportFromIssues(inputInfo render.InputInfo, command string, issues []insight.Issue, raw string) render.ExplainReport {
	rep := render.ExplainReport{
		Title:     "hosomaki explain",
		InputInfo: inputInfo,
		Context:   ContextLine(command),
		RawText:   raw,
	}
	for _, iss := range issues {
		rep.Issues = append(rep.Issues, toIssue(iss))
	}
	return rep
}

func rstatus(l analysis.Level) render.Status {
	switch l {
	case analysis.OK:
		return render.OK
	case analysis.Info:
		return render.Info
	case analysis.Warn:
		return render.Warn
	case analysis.Crit:
		return render.Crit
	default:
		return render.Neutral
	}
}

func sevStatus(sev string) render.Status {
	switch insight.NormalizeSeverity(sev) {
	case "ok":
		return render.OK
	case "warn":
		return render.Warn
	case "crit":
		return render.Crit
	case "info":
		return render.Info
	default:
		return render.Neutral
	}
}

func toMetrics(rep analysis.Report) []render.Metric {
	out := make([]render.Metric, 0, len(rep.Metrics))
	for _, m := range rep.Metrics {
		out = append(out, render.Metric{Label: m.Label, Value: m.Value, Status: rstatus(m.Level)})
	}
	return out
}

func toFindings(rep analysis.Report) []render.Finding {
	out := make([]render.Finding, 0, len(rep.Findings))
	for _, f := range rep.Findings {
		out = append(out, render.Finding{Status: rstatus(f.Level), Text: f.Text})
	}
	return out
}

func toIssue(iss insight.Issue) render.Issue {
	v := render.Issue{
		Subject: iss.Subject,
		Status:  sevStatus(iss.Severity),
	}
	if iss.Pattern != "" {
		v.Details = append(v.Details, render.Detail{Key: "detected pattern", Value: iss.Pattern})
	}
	if iss.Cause != "" {
		v.Details = append(v.Details, render.Detail{Key: "probable cause", Value: iss.Cause})
	}
	for _, d := range iss.Details {
		if d != "" {
			v.Details = append(v.Details, render.Detail{Value: d})
		}
	}
	for _, a := range iss.Actions {
		v.Actions = append(v.Actions, render.Action{
			Description: a.Description,
			Command:     a.Command,
			Disruptive:  a.Disruptive,
		})
	}
	return v
}

func doctorProcessLines() []string {
	return []string{
		"analysing logs…",
		"detecting patterns…",
		"correlating events…",
	}
}

func doctorSummary(rep analysis.Report, ai insight.Doctor) []render.SummaryItem {
	var items []render.SummaryItem

	if rep.FailedCount > 0 {
		items = append(items, render.SummaryItem{
			Text:   plural(rep.FailedCount, "service degraded", "services degraded"),
			Status: render.Warn,
		})
	}
	if rep.Anomalies > 0 {
		items = append(items, render.SummaryItem{
			Text:   plural(rep.Anomalies, "anomaly detected", "anomalies detected"),
			Status: render.Warn,
		})
	}
	actions := 0
	for _, iss := range ai.Issues {
		actions += len(iss.Actions)
	}
	if actions > 0 {
		items = append(items, render.SummaryItem{
			Text:   plural(actions, "action suggested", "actions suggested"),
			Status: render.Info,
		})
	}
	if len(items) == 0 {
		items = append(items, render.SummaryItem{Text: "healthy system", Status: render.OK})
	}
	return items
}

func statusSummary(rep analysis.Report, ai insight.Status) []render.SummaryItem {
	var items []render.SummaryItem

	for _, o := range ai.Observations {
		if o.Text == "" {
			continue
		}
		items = append(items, render.SummaryItem{
			Text:   o.Text,
			Status: sevStatus(o.Level),
		})
	}

	if len(items) > 0 {
		return items
	}

	return statusFallbackSummary(rep)
}

func statusFallbackSummary(rep analysis.Report) []render.SummaryItem {
	var items []render.SummaryItem
	if rep.FailedCount > 0 {
		items = append(items, render.SummaryItem{
			Text:   plural(rep.FailedCount, "service with warnings", "services with warnings"),
			Status: render.Warn,
		})
	}
	critCount := countCritFindings(rep)
	if critCount > 0 {
		items = append(items, render.SummaryItem{
			Text:   plural(critCount, "service failing", "services failing"),
			Status: render.Crit,
		})
	}
	if len(items) == 0 {
		items = append(items, render.SummaryItem{Text: "0 services failing", Status: render.OK})
	}
	return items
}

func countCritFindings(rep analysis.Report) int {
	n := 0
	for _, f := range rep.Findings {
		if f.Level == analysis.Crit {
			n++
		}
	}
	return n
}

func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}

func ExplainReport(command string, ai insight.Doctor) render.ExplainReport {
	rep := render.ExplainReport{
		Title:   "hosomaki explain",
		Context: ContextLine(command),
	}
	if len(ai.Issues) > 0 {
		for _, iss := range ai.Issues {
			rep.Issues = append(rep.Issues, toIssue(iss))
		}
	} else {
		rep.RawText = ai.Raw
		if rep.RawText == "" {
			rep.RawText = ai.Summary
		}
	}
	return rep
}

func Rstatus(l analysis.Level) render.Status { return rstatus(l) }
func SevStatus(sev string) render.Status     { return sevStatus(sev) }

func Plural(n int, one, many string) string { return plural(n, one, many) }
