// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package output

import (
	"encoding/json"
	"io"

	"github.com/rivernova/hosomaki/internal/analysis"
	"github.com/rivernova/hosomaki/internal/insight"
)

// this file contains the logic for converting analysis reports and insights into JSON output for the --output flag
type ComponentJSON struct {
	Source     string `json:"source"`
	Pattern    string `json:"pattern"`
	Cause      string `json:"cause"`
	Severity   string `json:"severity,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}
type DoctorJSON struct {
	Healthy    bool            `json:"healthy"`
	Components []ComponentJSON `json:"components"`
	Metrics    []MetricJSON    `json:"metrics"`
	Findings   []FindingJSON   `json:"findings"`
	Raw        string          `json:"raw,omitempty"`
}

type StatusJSON struct {
	Healthy    bool            `json:"healthy"`
	Components []ComponentJSON `json:"components"`
	Metrics    []MetricJSON    `json:"metrics"`
	Raw        string          `json:"raw,omitempty"`
}

type ExplainJSON struct {
	Source     string          `json:"source"`
	Command    string          `json:"command,omitempty"`
	Components []ComponentJSON `json:"components"`
	Raw        string          `json:"raw,omitempty"`
}

type MetricJSON struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Level string `json:"level"`
}

type FindingJSON struct {
	Level string `json:"level"`
	Text  string `json:"text"`
}

func WriteDoctor(w io.Writer, rep analysis.Report, ai insight.Analysis) error {
	out := DoctorJSON{
		Healthy:  isHealthyAnalysis(ai),
		Metrics:  metricsJSON(rep),
		Findings: findingsJSON(rep),
		Raw:      ai.Raw,
	}
	for _, c := range ai.Components {
		out.Components = append(out.Components, componentJSON(c))
	}
	return writeJSON(w, out)
}

func WriteStatus(w io.Writer, rep analysis.Report, ai insight.Analysis) error {
	out := StatusJSON{
		Healthy: isHealthyAnalysis(ai),
		Metrics: metricsJSON(rep),
		Raw:     ai.Raw,
	}
	for _, c := range ai.Components {
		cj := componentJSON(c)
		cj.Suggestion = ""
		out.Components = append(out.Components, cj)
	}
	return writeJSON(w, out)
}

func WriteExplain(w io.Writer, source, command string, ai insight.Analysis) error {
	out := ExplainJSON{
		Source:  source,
		Command: command,
		Raw:     ai.Raw,
	}
	for _, c := range ai.Components {
		cj := componentJSON(c)
		cj.Severity = ""
		cj.Suggestion = ""
		out.Components = append(out.Components, cj)
	}
	return writeJSON(w, out)
}

func isHealthyAnalysis(ai insight.Analysis) bool {
	if ai.Raw != "" {
		return false
	}
	for _, c := range ai.Components {
		if c.Source != "summary" {
			return false
		}
	}
	return true
}

func componentJSON(c insight.Component) ComponentJSON {
	return ComponentJSON{
		Source:     c.Source,
		Pattern:    c.Pattern,
		Cause:      c.Cause,
		Severity:   c.Severity,
		Suggestion: c.Suggestion,
	}
}

func metricsJSON(rep analysis.Report) []MetricJSON {
	out := make([]MetricJSON, 0, len(rep.Metrics))
	for _, m := range rep.Metrics {
		out = append(out, MetricJSON{Label: m.Label, Value: m.Value, Level: levelString(m.Level)})
	}
	return out
}

func findingsJSON(rep analysis.Report) []FindingJSON {
	out := make([]FindingJSON, 0, len(rep.Findings))
	for _, f := range rep.Findings {
		out = append(out, FindingJSON{Level: levelString(f.Level), Text: f.Text})
	}
	return out
}

func levelString(l analysis.Level) string {
	switch l {
	case analysis.OK:
		return "ok"
	case analysis.Info:
		return "info"
	case analysis.Warn:
		return "warn"
	case analysis.Crit:
		return "crit"
	default:
		return "neutral"
	}
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
