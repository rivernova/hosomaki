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

// this file contains logic for turning structured insights into JSON output
type DoctorJSON struct {
	Healthy  bool          `json:"healthy"`
	Summary  string        `json:"summary"`
	Issues   []IssueJSON   `json:"issues"`
	Metrics  []MetricJSON  `json:"metrics"`
	Findings []FindingJSON `json:"findings"`
}

type IssueJSON struct {
	Subject  string       `json:"subject"`
	Severity string       `json:"severity"`
	Pattern  string       `json:"pattern"`
	Cause    string       `json:"cause"`
	Details  []string     `json:"details"`
	Actions  []ActionJSON `json:"actions"`
}

type ActionJSON struct {
	Description string `json:"description"`
	Command     string `json:"command"`
	Disruptive  bool   `json:"disruptive"`
}

type StatusJSON struct {
	Healthy      bool              `json:"healthy"`
	Summary      string            `json:"summary"`
	Observations []ObservationJSON `json:"observations"`
	Metrics      []MetricJSON      `json:"metrics"`
}

type ObservationJSON struct {
	Level string `json:"level"`
	Text  string `json:"text"`
}

type ExplainJSON struct {
	Input       string `json:"input"`
	Command     string `json:"command,omitempty"`
	Explanation string `json:"explanation"`
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

func WriteDoctor(w io.Writer, rep analysis.Report, ai insight.Doctor) error {
	out := DoctorJSON{
		Healthy:  ai.Healthy,
		Summary:  ai.Summary,
		Metrics:  metricsJSON(rep),
		Findings: findingsJSON(rep),
	}
	if ai.Raw != "" && out.Summary == "" {
		out.Summary = ai.Raw
	}
	for _, iss := range ai.Issues {
		out.Issues = append(out.Issues, issueJSON(iss))
	}
	return writeJSON(w, out)
}

func WriteStatus(w io.Writer, rep analysis.Report, ai insight.Status) error {
	out := StatusJSON{
		Healthy: ai.Healthy,
		Summary: ai.Summary,
		Metrics: metricsJSON(rep),
	}
	if ai.Raw != "" && out.Summary == "" {
		out.Summary = ai.Raw
	}
	for _, o := range ai.Observations {
		out.Observations = append(out.Observations, ObservationJSON{Level: o.Level, Text: o.Text})
	}
	return writeJSON(w, out)
}

func WriteExplain(w io.Writer, input, command, explanation string) error {
	return writeJSON(w, ExplainJSON{
		Input:       input,
		Command:     command,
		Explanation: explanation,
	})
}

func issueJSON(iss insight.Issue) IssueJSON {
	out := IssueJSON{
		Subject:  iss.Subject,
		Severity: iss.Severity,
		Pattern:  iss.Pattern,
		Cause:    iss.Cause,
		Details:  iss.Details,
	}
	for _, a := range iss.Actions {
		out.Actions = append(out.Actions, ActionJSON{
			Description: a.Description,
			Command:     a.Command,
			Disruptive:  a.Disruptive,
		})
	}
	return out
}

func metricsJSON(rep analysis.Report) []MetricJSON {
	out := make([]MetricJSON, 0, len(rep.Metrics))
	for _, m := range rep.Metrics {
		out = append(out, MetricJSON{Label: m.Label, Value: m.Value, Level: levelName(m.Level)})
	}
	return out
}

func findingsJSON(rep analysis.Report) []FindingJSON {
	out := make([]FindingJSON, 0, len(rep.Findings))
	for _, f := range rep.Findings {
		out = append(out, FindingJSON{Level: levelName(f.Level), Text: f.Text})
	}
	return out
}

func levelName(l analysis.Level) string {
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
