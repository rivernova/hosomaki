// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package render

import (
	"io"
	"strings"
)

// this file contains the logic for rendering structured analysis data into human-readable reports
type Metric struct {
	Label  string
	Value  string
	Status Status
}

type Finding struct {
	Status Status
	Text   string
}

type Detail struct {
	Key   string
	Value string
}

type Component struct {
	Source               string
	DisplayName          string
	Status               Status
	Details              []Detail
	Suggestion           string
	SuggestionDisruptive bool
}

type SummaryItem struct {
	Text   string
	Status Status
}

type InputInfo struct {
	Origin string
	Detail string
	Lines  int
}

type StatusReport struct {
	Title      string
	Metrics    []Metric
	Services   []Finding
	Components []Component
	RawAI      string
	Summary    []SummaryItem
	Brief      bool
}

type DoctorReport struct {
	Title        string
	Metrics      []Metric
	Findings     []Finding
	ProcessLines []string
	Components   []Component
	RawInsight   string
	Summary      []SummaryItem
	Brief        bool
}

type ExplainReport struct {
	Title        string
	InputInfo    InputInfo
	Context      string
	ProcessLines []string
	Components   []Component
	RawText      string
}

func (r *Renderer) RenderStatus(rep StatusReport) {
	r.Title(rep.Title)
	r.renderStatusBody(rep)
	r.Done()
}

func (r *Renderer) RenderStatusStream(rep StatusReport) io.Writer {
	if rep.Brief {
		return io.Discard
	}
	r.Title(rep.Title)
	r.renderStatusBody(rep)
	r.Section("AI analysis (lightweight)")
	r.Blank()
	for _, p := range []string{
		"analysing general status…",
		"correlating metrics…",
		"detecting patterns…",
	} {
		r.Process(p)
	}
	r.Blank()
	return io.Discard
}

func (r *Renderer) FinaliseStatus(rep StatusReport) {
	if rep.Brief {
		r.renderStatusBrief(rep)
		return
	}

	body, summary := splitSummaryComponent(rep.Components)

	switch {
	case len(body) > 0 || summary != nil:
		for _, c := range body {
			r.Blank()
			r.renderComponent(c, false)
		}
		if summary != nil {
			r.Section("summary")
			r.Blank()
			r.renderComponent(*summary, false)
		} else {
			r.renderStatusMetricSummary(rep)
		}
	case strings.TrimSpace(rep.RawAI) != "":
		r.Blank()
		r.Subject("AI output", Neutral)
		r.Detail("error", "model response was not valid XML")
		r.Detail("raw", rep.RawAI)
		r.renderStatusMetricSummary(rep)
	default:
		r.Blank()
		r.Subject("system", OK)
		r.Detail("detected pattern", "no issues detected")
		r.renderStatusMetricSummary(rep)
	}

	r.Done()
}

func (r *Renderer) renderStatusBrief(rep StatusReport) {
	r.Title(rep.Title)
	r.Section("quick status")
	r.Blank()

	body, summary := splitSummaryComponent(rep.Components)

	switch {
	case len(body) > 0 || summary != nil:
		for _, c := range body {
			r.renderComponent(c, false)
			r.Blank()
		}
		if summary != nil {
			r.Section("summary")
			r.Blank()
			r.renderComponent(*summary, false)
		}
	case strings.TrimSpace(rep.RawAI) != "":
		r.SummaryLine("AI response was not structured XML — run without --brief for details", Neutral)
	default:
		r.SummaryLine("healthy system", OK)
	}

	r.Done()
}

func (r *Renderer) renderStatusBody(rep StatusReport) {
	if len(rep.Metrics) > 0 {
		r.Section("system status")
		r.Blank()
		for _, m := range rep.Metrics {
			r.Metric(m.Label, m.Value, m.Status)
		}
	}
	if len(rep.Services) > 0 {
		r.Section("services")
		r.Blank()
		for _, f := range rep.Services {
			r.Finding(f.Status, f.Text)
		}
	}
}

func (r *Renderer) renderStatusMetricSummary(rep StatusReport) {
	if len(rep.Summary) == 0 {
		return
	}
	r.Section("summary")
	r.Blank()
	for _, s := range rep.Summary {
		r.SummaryLine(s.Text, s.Status)
	}
}

func (r *Renderer) RenderDoctor(rep DoctorReport) {
	r.Title(rep.Title)
	r.renderDoctorPreamble(rep)
	r.Section("AI analysis")
	r.Blank()
	for _, p := range rep.ProcessLines {
		r.Process(p)
	}
	r.renderDoctorComponents(rep)
	r.Done()
}

func (r *Renderer) RenderDoctorStream(rep DoctorReport) io.Writer {
	if rep.Brief {
		return io.Discard
	}
	r.Title(rep.Title)
	r.renderDoctorPreamble(rep)
	r.Section("AI analysis")
	r.Blank()
	for _, p := range rep.ProcessLines {
		r.Process(p)
	}
	return io.Discard
}

func (r *Renderer) FinaliseDoctor(rep DoctorReport) {
	if rep.Brief {
		r.renderDoctorBrief(rep)
		return
	}
	r.renderDoctorComponents(rep)
	r.Done()
}

func (r *Renderer) renderDoctorComponents(rep DoctorReport) {
	body, summary := splitSummaryComponent(rep.Components)

	switch {
	case len(body) > 0 || summary != nil:
		for _, c := range body {
			r.Blank()
			r.renderComponent(c, true)
		}
		if summary != nil {
			r.Section("summary")
			r.Blank()
			r.renderComponent(*summary, true)
		} else {
			r.renderDoctorMetricSummary(rep)
		}
	case strings.TrimSpace(rep.RawInsight) != "":
		r.Blank()
		r.Subject("AI output", Neutral)
		r.Detail("error", "model response was not valid XML")
		r.Detail("raw", rep.RawInsight)
		r.renderDoctorMetricSummary(rep)
	default:
		r.Blank()
		r.Subject("system", OK)
		r.Detail("detected pattern", "no issues detected")
		r.renderDoctorMetricSummary(rep)
	}
}

func (r *Renderer) renderDoctorBrief(rep DoctorReport) {
	r.Title(rep.Title)
	r.Section("quick diagnosis")
	r.Blank()

	body, summary := splitSummaryComponent(rep.Components)

	switch {
	case len(body) > 0 || summary != nil:
		for _, c := range body {
			r.renderComponent(c, true)
			r.Blank()
		}
		if summary != nil {
			r.Section("summary")
			r.Blank()
			r.renderComponent(*summary, true)
		}
	case strings.TrimSpace(rep.RawInsight) != "":
		r.Subject("AI output", Neutral)
		r.Detail("error", "model response was not valid XML — run without --brief for details")
	default:
		r.SummaryLine("healthy system", OK)
	}

	r.Done()
}

func (r *Renderer) renderDoctorMetricSummary(rep DoctorReport) {
	if len(rep.Summary) == 0 {
		return
	}
	r.Section("summary")
	r.Blank()
	for _, s := range rep.Summary {
		r.SummaryLine(s.Text, s.Status)
	}
}

func (r *Renderer) renderDoctorPreamble(rep DoctorReport) {
	if len(rep.Metrics) > 0 {
		r.Section("system analysis")
		r.Blank()
		for _, m := range rep.Metrics {
			r.Metric(m.Label, m.Value, m.Status)
		}
	}
	if len(rep.Findings) > 0 {
		r.Section("local analysis")
		r.Blank()
		for _, f := range rep.Findings {
			r.Finding(f.Status, f.Text)
		}
	}
}

func (r *Renderer) RenderExplain(rep ExplainReport) {
	r.Title(rep.Title)
	r.renderExplainPreamble(rep)
	r.Section("AI analysis")
	r.Blank()
	for _, p := range rep.ProcessLines {
		r.Process(p)
	}
	r.renderExplainComponents(rep)
	r.Done()
}

func (r *Renderer) RenderExplainStream(rep ExplainReport, processLines []string) io.Writer {
	r.Title(rep.Title)
	r.renderExplainPreamble(rep)
	r.Section("AI analysis")
	r.Blank()
	for _, p := range processLines {
		r.Process(p)
	}
	return io.Discard
}

func (r *Renderer) FinaliseExplain(rep ExplainReport) {
	r.renderExplainComponents(rep)
	r.Done()
}

func (r *Renderer) StreamEnd() {
	r.Blank()
}

func (r *Renderer) renderExplainPreamble(rep ExplainReport) {
	if rep.InputInfo.Origin != "" {
		r.Section("entry info")
		r.Blank()
		r.Metric("origin", rep.InputInfo.Origin, Neutral)
		if rep.InputInfo.Detail != "" {
			r.Metric("detail", rep.InputInfo.Detail, Neutral)
		}
		if rep.InputInfo.Lines > 0 {
			r.Metric("lines", itoa(rep.InputInfo.Lines), Neutral)
		}
	}
	if c := strings.TrimSpace(rep.Context); c != "" {
		r.Blank()
		r.Paragraph(c)
	}
}

func (r *Renderer) renderExplainComponents(rep ExplainReport) {
	body, summary := splitSummaryComponent(rep.Components)

	switch {
	case len(body) > 0 || summary != nil:
		for _, c := range body {
			r.Blank()
			r.renderComponent(c, false)
		}
		if summary != nil {
			r.Section("summary")
			r.Blank()
			r.renderComponent(*summary, false)
		}
	case strings.TrimSpace(rep.RawText) != "":
		r.Blank()
		r.Subject("AI output", Neutral)
		r.Detail("error", "model response was not valid XML")
		r.Detail("raw", rep.RawText)
	default:
		r.Blank()
		r.Subject("system", OK)
		r.Detail("detected pattern", "no issues detected in the provided input")
	}
}

func splitSummaryComponent(components []Component) (body []Component, summary *Component) {
	if len(components) == 0 {
		return nil, nil
	}
	last := components[len(components)-1]
	if last.Source == "summary" {
		body = components[:len(components)-1]
		return body, &last
	}
	return components, nil
}

func (r *Renderer) renderComponent(c Component, showSuggestion bool) {
	name := resolveDisplayName(c)
	r.Subject(name, c.Status)

	for _, d := range c.Details {
		r.Detail(d.Key, d.Value)
	}

	if showSuggestion && strings.TrimSpace(c.Suggestion) != "" {
		r.Detail("suggestion", "")
		r.Command(c.Suggestion, c.SuggestionDisruptive)
	}
}

func resolveDisplayName(c Component) string {
	if c.DisplayName != "" {
		return c.DisplayName
	}
	src := strings.TrimSpace(c.Source)
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
	case "summary":
		return "summary"
	}
	return src
}

func plural(n int, one, many string) string {
	if n == 1 {
		return itoa(n) + " " + one
	}
	return itoa(n) + " " + many
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := make([]byte, 0, 10)
	if n < 0 {
		b = append(b, '-')
		n = -n
	}
	tmp := make([]byte, 0, 10)
	for n > 0 {
		tmp = append(tmp, byte('0'+n%10))
		n /= 10
	}
	for i := len(tmp) - 1; i >= 0; i-- {
		b = append(b, tmp[i])
	}
	return string(b)
}
