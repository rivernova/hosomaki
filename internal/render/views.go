// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package render

import (
	"fmt"
	"io"
	"strings"
)

// this file contains the viewer models for the different report types.
// these are the "view models" that the renderer turns into terminal output

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

type Action struct {
	Description string
	Command     string
	Disruptive  bool
}

type Issue struct {
	Subject string
	Status  Status
	Details []Detail
	Actions []Action
}

type SummaryItem struct {
	Text   string
	Status Status
}

type StatusReport struct {
	Title    string
	Metrics  []Metric
	Services []Finding
	Summary  []SummaryItem
}

type DoctorReport struct {
	Title        string
	Metrics      []Metric
	Findings     []Finding
	ProcessLines []string
	Issues       []Issue
	RawInsight   string
	Summary      []SummaryItem
}

type ExplainReport struct {
	Title       string
	ServiceInfo []Metric
	Context     string
	Issues      []Issue
	RawText     string
}

func (r *Renderer) RenderStatus(rep StatusReport) {
	r.Title(rep.Title)
	r.renderStatusBody(rep)
	if len(rep.Summary) > 0 {
		r.Section("summary")
		r.Blank()
		for _, s := range rep.Summary {
			r.SummaryLine(s.Text, s.Status)
		}
	}
	r.Done()
}

func (r *Renderer) RenderStatusStream(rep StatusReport) io.Writer {
	r.Title(rep.Title)
	r.renderStatusBody(rep)
	return r.StreamStart("summary")
}

func (r *Renderer) FinaliseStatus() {
	r.StreamEnd()
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

func (r *Renderer) RenderDoctor(rep DoctorReport) {
	r.Title(rep.Title)
	r.renderDoctorPreamble(rep)
	r.Section("ai analysis")
	r.Blank()
	for _, p := range rep.ProcessLines {
		r.Process(p)
	}
	r.renderDoctorAI(rep)
	r.renderDoctorSummary(rep)
	r.Done()
}

func (r *Renderer) RenderDoctorStream(rep DoctorReport) io.Writer {
	r.Title(rep.Title)
	r.renderDoctorPreamble(rep)
	r.Section("ai analysis")
	r.Blank()
	for _, p := range rep.ProcessLines {
		r.Process(p)
	}
	return io.Discard
}

func (r *Renderer) FinaliseDoctor(rep DoctorReport) {
	r.renderDoctorAI(rep)
	r.renderDoctorSummary(rep)
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
	case len(rep.Issues) > 0:
		for _, iss := range rep.Issues {
			r.Blank()
			r.renderIssue(iss)
		}
	case strings.TrimSpace(rep.RawInsight) != "":
		r.Blank()
		r.Paragraph(rep.RawInsight)
	default:
		r.Blank()
		r.Paragraph("No further insight was produced.")
	}
}

func (r *Renderer) renderDoctorSummary(rep DoctorReport) {
	if len(rep.Summary) > 0 {
		r.Section("summary")
		r.Blank()
		for _, s := range rep.Summary {
			r.SummaryLine(s.Text, s.Status)
		}
	}
}

func (r *Renderer) renderIssue(iss Issue) {
	r.Subject(iss.Subject, iss.Status)
	for _, d := range iss.Details {
		r.Detail(d.Key, d.Value)
	}
	for _, a := range iss.Actions {
		desc := strings.TrimSpace(a.Description)
		if desc != "" {
			r.Detail("suggestion", desc)
		}
		if cmd := strings.TrimSpace(a.Command); cmd != "" {
			r.Command(cmd, a.Disruptive)
		}
	}
}

func (r *Renderer) RenderExplain(rep ExplainReport) {
	r.Title(rep.Title)
	r.renderExplainPreamble(rep)
	r.Section("ai analysis")
	r.Blank()
	switch {
	case len(rep.Issues) > 0:
		for _, iss := range rep.Issues {
			r.Blank()
			r.renderIssue(iss)
		}
	case strings.TrimSpace(rep.RawText) != "":
		r.Paragraph(rep.RawText)
	default:
		r.Paragraph("No explanation was produced.")
	}
	r.Done()
}

func (r *Renderer) RenderExplainStream(rep ExplainReport, processLines []string) io.Writer {
	r.Title(rep.Title)
	r.renderExplainPreamble(rep)
	r.Section("ai analysis")
	r.Blank()
	for _, p := range processLines {
		r.Process(p)
	}
	fmt.Fprint(r.w, indent(1))
	return &streamWriter{r: r}
}

func (r *Renderer) renderExplainPreamble(rep ExplainReport) {
	if len(rep.ServiceInfo) > 0 {
		r.Section("service info")
		r.Blank()
		for _, m := range rep.ServiceInfo {
			r.Metric(m.Label, m.Value, m.Status)
		}
	}
	if c := strings.TrimSpace(rep.Context); c != "" {
		r.Blank()
		r.Process(c)
	}
}
