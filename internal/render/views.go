// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package render

import (
	"io"
	"strings"
)

// this file contains the logic for rendering the various report types to the terminal

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
	BriefText  string
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
		"analyzing general status…",
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
	r.renderStatusAI(rep)
	r.renderStatusSummary(rep)
	r.Done()
}

func (r *Renderer) renderStatusBrief(rep StatusReport) {
	r.Blank()
	text := rep.BriefText
	if text == "" && strings.TrimSpace(rep.RawAI) != "" {
		text = strings.SplitN(strings.TrimSpace(rep.RawAI), "\n", 2)[0]
	}
	if text == "" {
		text = "system operating normally"
	}
	r.line(indent(1) + r.paint(r.pal.Text, text))
	r.Done()
}

func (r *Renderer) renderStatusAI(rep StatusReport) {
	switch {
	case len(rep.Components) > 0:
		for _, c := range rep.Components {
			r.Blank()
			r.renderComponent(c, false)
		}
	case strings.TrimSpace(rep.RawAI) != "":
		r.Blank()
		r.Paragraph(rep.RawAI)
	}
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

func (r *Renderer) renderStatusSummary(rep StatusReport) {
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
	r.renderDoctorAI(rep)
	r.renderDoctorSummary(rep)
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
	r.renderDoctorAI(rep)
	r.renderDoctorSummary(rep)
	r.Done()
}

func (r *Renderer) renderDoctorBrief(rep DoctorReport) {
	r.Title(rep.Title)
	r.Section("quick diagnosis")
	r.Blank()
	if len(rep.Components) == 0 {
		r.SummaryLine("healthy system", OK)
	} else {
		for _, c := range rep.Components {
			name := resolveDisplayName(c)
			var cause string
			for _, d := range c.Details {
				if d.Key == "probable cause" {
					cause = d.Value
					break
				}
			}
			line := name
			if cause != "" {
				line += ": " + cause
			}
			if c.Suggestion != "" {
				short := c.Suggestion
				if len(short) > 100 {
					short = short[:100] + "…"
				}
				line += " → " + short
			}
			r.Detail("", line)
		}
	}
	r.Done()
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

func (r *Renderer) renderDoctorAI(rep DoctorReport) {
	switch {
	case len(rep.Components) > 0:
		for _, c := range rep.Components {
			r.Blank()
			r.renderComponent(c, true)
		}
	default:
		r.Blank()
		r.Subject("system", Neutral)
		r.Detail("detected pattern", "AI response could not be parsed into structured output")
		r.Detail("probable cause", "the local model did not follow the expected XML format")
	}
}

func (r *Renderer) renderDoctorSummary(rep DoctorReport) {
	if len(rep.Summary) == 0 {
		return
	}
	r.Section("summary")
	r.Blank()
	for _, s := range rep.Summary {
		r.SummaryLine(s.Text, s.Status)
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
	r.renderExplainAI(rep)
	r.renderExplainSummary(rep)
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
	r.renderExplainAI(rep)
	r.renderExplainSummary(rep)
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

func (r *Renderer) renderExplainAI(rep ExplainReport) {
	switch {
	case len(rep.Components) > 0:
		for _, c := range rep.Components {
			r.Blank()
			r.renderComponent(c, false)
		}
	case strings.TrimSpace(rep.RawText) != "":
		r.Blank()
		r.Paragraph(rep.RawText)
	default:
		r.Blank()
		r.Subject("system", Neutral)
		r.Detail("detected pattern", "AI response could not be parsed into structured output")
		r.Detail("probable cause", "the local model did not follow the expected XML format")
	}
}

func (r *Renderer) renderExplainSummary(rep ExplainReport) {
	if len(rep.Components) == 0 {
		return
	}

	patterns := 0
	causes := 0
	for _, c := range rep.Components {
		for _, d := range c.Details {
			switch d.Key {
			case "detected pattern":
				patterns++
			case "probable cause":
				causes++
			}
		}
	}
	if patterns == 0 {
		patterns = len(rep.Components)
	}
	if causes == 0 {
		causes = len(rep.Components)
	}

	r.Section("summary")
	r.Blank()
	r.SummaryLine(plural(patterns, "detected pattern", "detected patterns"), Info)
	if causes > 0 {
		r.SummaryLine(plural(causes, "probable cause", "probable causes"), Info)
	}
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
