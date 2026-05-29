// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package insight

import (
	"encoding/json"
	"strings"
)

// this file contains logic for parsing and normalising insights output into structured data

type Action struct {
	Description string `json:"description"`
	Command     string `json:"command"`
	Disruptive  bool   `json:"disruptive"`
}

type Issue struct {
	Subject  string   `json:"subject"`
	Severity string   `json:"severity"`
	Pattern  string   `json:"pattern"`
	Cause    string   `json:"cause"`
	Details  []string `json:"details"`
	Actions  []Action `json:"actions"`
}

type Doctor struct {
	Healthy bool    `json:"healthy"`
	Summary string  `json:"summary"`
	Issues  []Issue `json:"issues"`
	Raw     string  `json:"-"`
}

type Observation struct {
	Level string `json:"level"`
	Text  string `json:"text"`
}

type Status struct {
	Healthy      bool          `json:"healthy"`
	Summary      string        `json:"summary"`
	Observations []Observation `json:"observations"`

	Raw string `json:"-"`
}

func ParseDoctor(raw string) Doctor {
	clean := cleanText(raw)
	obj, found := extractJSON(clean)
	if found {
		var d Doctor
		if err := json.Unmarshal([]byte(obj), &d); err == nil {
			d.normalize()
			if len(d.Issues) > 0 || strings.TrimSpace(d.Summary) != "" {
				return d
			}
		}
	}
	return Doctor{Raw: clean}
}

func ParseStatus(raw string) Status {
	clean := cleanText(raw)
	obj, found := extractJSON(clean)
	if found {
		var s Status
		if err := json.Unmarshal([]byte(obj), &s); err == nil {
			s.normalize()
			if strings.TrimSpace(s.Summary) != "" || len(s.Observations) > 0 {
				return s
			}
		}
	}
	return Status{Raw: clean, Summary: clean}
}

func (d *Doctor) normalize() {
	d.Summary = strings.TrimSpace(d.Summary)
	kept := d.Issues[:0]
	for _, iss := range d.Issues {
		iss.Subject = strings.TrimSpace(iss.Subject)
		iss.Pattern = strings.TrimSpace(iss.Pattern)
		iss.Cause = strings.TrimSpace(iss.Cause)
		iss.Severity = NormalizeSeverity(iss.Severity)
		if iss.Subject == "" && iss.Cause == "" && iss.Pattern == "" &&
			len(iss.Details) == 0 && len(iss.Actions) == 0 {
			continue
		}
		if iss.Subject == "" {
			iss.Subject = "issue"
		}
		kept = append(kept, iss)
	}
	d.Issues = kept
}

func (s *Status) normalize() {
	s.Summary = strings.TrimSpace(s.Summary)
	for i := range s.Observations {
		s.Observations[i].Level = NormalizeSeverity(s.Observations[i].Level)
		s.Observations[i].Text = strings.TrimSpace(s.Observations[i].Text)
	}
}

func NormalizeSeverity(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ok", "good", "healthy", "pass", "fine":
		return "ok"
	case "warn", "warning", "degraded", "minor", "medium":
		return "warn"
	case "crit", "critical", "error", "severe", "high", "fatal", "fail", "failed":
		return "crit"
	case "":
		return "info"
	default:
		return "info"
	}
}

func cleanText(raw string) string {
	s := strings.TrimSpace(raw)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	if nl := strings.IndexByte(s, '\n'); nl >= 0 {
		s = s[nl+1:]
	}
	s = strings.TrimSuffix(strings.TrimRight(s, "\n"), "```")
	return strings.TrimSpace(s)
}

func extractJSON(s string) (string, bool) {
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')
	if start < 0 || end <= start {
		return "", false
	}
	return s[start : end+1], true
}
