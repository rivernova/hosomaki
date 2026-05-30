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

// this file contains presentation for converting insights to renderable output

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

func DoctorReport(rep analysis.Report, ai insight.Analysis, brief bool) render.DoctorReport {
	out := render.DoctorReport{
		Title:        "hosomaki doctor",
		Metrics:      toMetrics(rep),
		Findings:     toFindings(rep),
		ProcessLines: doctorProcessLines(),
		RawInsight:   ai.Raw,
		Brief:        brief,
	}

	for _, c := range ai.Components {
		out.Components = append(out.Components, toComponent(c, true))
	}

	out.Summary = doctorSummary(rep, ai)

	if brief {
		out.ProcessLines = nil
	}
	return out
}

func StatusReport(rep analysis.Report, brief bool) render.StatusReport {
	return render.StatusReport{
		Title:    "hosomaki status",
		Metrics:  toMetrics(rep),
		Services: toFindings(rep),
		Summary:  statusFallbackSummary(rep),
		Brief:    brief,
	}
}

func StatusReportWithAnalysis(rep analysis.Report, ai insight.Analysis, brief bool) render.StatusReport {
	sr := StatusReport(rep, brief)
	sr.RawAI = strings.TrimSpace(ai.Raw)
	sr.BriefText = buildBriefText(ai)
	sr.Summary = statusHealthSummary(rep, len(ai.Components))

	for _, c := range ai.Components {
		sr.Components = append(sr.Components, toComponent(c, false))
	}
	return sr
}

func ExplainReport(inputInfo render.InputInfo, command string, ai insight.Analysis) render.ExplainReport {
	rep := render.ExplainReport{
		Title:     "hosomaki explain",
		InputInfo: inputInfo,
		Context:   ContextLine(command),
		RawText:   ai.Raw,
	}
	for _, c := range ai.Components {
		rep.Components = append(rep.Components, toComponent(c, false))
	}
	return rep
}

func toComponent(c insight.Component, includeSuggestion bool) render.Component {
	v := render.Component{
		Source:      c.Source,
		DisplayName: sourceDisplayName(c.Source),
		Status:      severityToStatus(c.Severity),
	}

	if c.Pattern != "" {
		v.Details = append(v.Details, render.Detail{Key: "detected pattern", Value: c.Pattern})
	}
	if c.Cause != "" {
		v.Details = append(v.Details, render.Detail{Key: "probable cause", Value: c.Cause})
	}

	if includeSuggestion && strings.TrimSpace(c.Suggestion) != "" {
		v.Suggestion = c.Suggestion
		v.SuggestionDisruptive = isDisruptive(c.Suggestion)
	}

	return v
}

func sourceDisplayName(source string) string {
	src := strings.TrimSpace(source)
	if src == "" {
		return "system"
	}
	for _, prefix := range []string{"service:", "file:", "boot:"} {
		if strings.HasPrefix(src, prefix) {
			name := src[len(prefix):]
			if name != "" {
				return name
			}
		}
	}
	switch src {
	case "dmesg":
		return "kernel"
	case "pipe":
		return "system"
	case "inline":
		return "input"
	}
	return src
}

func isDisruptive(suggestion string) bool {
	lower := strings.ToLower(suggestion)
	for _, kw := range []string{
		"disruptive", "irreversible", "destructive", "data loss",
		"caution", "warning:", "careful", "backup first",
	} {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
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

func severityToStatus(sev string) render.Status {
	switch insight.NormaliseSeverity(sev) {
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

func doctorProcessLines() []string {
	return []string{
		"analysing logs…",
		"detecting patterns…",
		"correlating events…",
	}
}

func doctorSummary(rep analysis.Report, ai insight.Analysis) []render.SummaryItem {
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
	suggestions := 0
	for _, c := range ai.Components {
		if strings.TrimSpace(c.Suggestion) != "" {
			suggestions++
		}
	}
	if suggestions > 0 {
		items = append(items, render.SummaryItem{
			Text:   plural(suggestions, "action suggested", "actions suggested"),
			Status: render.Info,
		})
	}
	if len(items) == 0 {
		items = append(items, render.SummaryItem{Text: "healthy system", Status: render.OK})
	}
	return items
}

func statusHealthSummary(rep analysis.Report, aiComponentCount int) []render.SummaryItem {
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
	if aiComponentCount > 0 {
		items = append(items, render.SummaryItem{
			Text:   plural(aiComponentCount, "pattern detected by AI", "patterns detected by AI"),
			Status: render.Info,
		})
	}
	if len(items) == 0 {
		items = append(items, render.SummaryItem{Text: "healthy system", Status: render.OK})
	}
	return items
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

func buildBriefText(ai insight.Analysis) string {
	if strings.TrimSpace(ai.Raw) != "" {
		text := strings.TrimSpace(ai.Raw)
		if idx := strings.IndexAny(text, ".!?"); idx >= 0 && idx < len(text)-1 {
			text = text[:idx+1]
		}
		return text
	}
	if len(ai.Components) > 0 {
		c := ai.Components[0]
		if c.Pattern != "" {
			return c.Pattern
		}
	}
	return "system operating normally"
}

func Rstatus(l analysis.Level) render.Status    { return rstatus(l) }
func SeverityToStatus(sev string) render.Status { return severityToStatus(sev) }
func Plural(n int, one, many string) string     { return plural(n, one, many) }

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
