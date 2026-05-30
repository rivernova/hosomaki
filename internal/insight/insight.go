// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package insight

import (
	"strings"
)

// this file contains the logic for parsing the XML output from the analysis scripts into structured data
type Component struct {
	Source     string
	Pattern    string
	Cause      string
	Severity   string
	Suggestion string
}

type Analysis struct {
	Components []Component
	Raw        string
}

func ParseExplain(raw string) Analysis {
	a := parseXMLAnalysis(raw)
	for i := range a.Components {
		a.Components[i].Severity = ""
		a.Components[i].Suggestion = ""
	}
	return a
}

func ParseStatus(raw string) Analysis {
	a := parseXMLAnalysis(raw)
	for i := range a.Components {
		a.Components[i].Suggestion = ""
	}
	return a
}

func ParseDoctor(raw string) Analysis {
	return parseXMLAnalysis(raw)
}

func parseXMLAnalysis(raw string) Analysis {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return Analysis{}
	}

	body, found := extractTagContent(clean, "analysis")
	if !found {
		return Analysis{Raw: clean}
	}

	return Analysis{Components: parseComponents(body)}
}

func parseComponents(body string) []Component {
	chunks := splitTag(body, "component")
	if len(chunks) == 0 {
		return nil
	}

	out := make([]Component, 0, len(chunks))
	for _, chunk := range chunks {
		source := strings.TrimSpace(extractTag(chunk, "source"))
		pattern := strings.TrimSpace(extractTag(chunk, "pattern"))
		cause := strings.TrimSpace(extractTag(chunk, "cause"))
		severity := normaliseSeverityTag(strings.TrimSpace(extractTag(chunk, "severity")))
		suggestion := strings.TrimSpace(extractTag(chunk, "suggestion"))

		if pattern == "" {
			continue
		}
		if source == "" {
			source = "pipe"
		}

		out = append(out, Component{
			Source:     source,
			Pattern:    pattern,
			Cause:      cause,
			Severity:   severity,
			Suggestion: suggestion,
		})
	}
	return out
}

func extractTagContent(s, tag string) (string, bool) {
	open := "<" + tag
	close := "</" + tag + ">"

	start := strings.Index(s, open)
	if start < 0 {
		return "", false
	}
	gt := strings.Index(s[start:], ">")
	if gt < 0 {
		return "", false
	}
	contentStart := start + gt + 1

	end := strings.LastIndex(s, close)
	if end < contentStart {
		return strings.TrimSpace(s[contentStart:]), true
	}
	return strings.TrimSpace(s[contentStart:end]), true
}

func extractTag(s, tag string) string {
	open := "<" + tag + ">"
	close := "</" + tag + ">"

	start := strings.Index(s, open)
	if start < 0 {
		return ""
	}
	start += len(open)
	end := strings.Index(s[start:], close)
	if end < 0 {
		return strings.TrimSpace(s[start:])
	}
	return strings.TrimSpace(s[start : start+end])
}

func splitTag(raw, tag string) []string {
	open := "<" + tag + ">"
	close := "</" + tag + ">"

	var parts []string
	rest := raw
	for {
		start := strings.Index(rest, open)
		if start < 0 {
			break
		}
		rest = rest[start+len(open):]
		end := strings.Index(rest, close)
		if end < 0 {
			if trimmed := strings.TrimSpace(rest); trimmed != "" {
				parts = append(parts, trimmed)
			}
			break
		}
		parts = append(parts, rest[:end])
		rest = rest[end+len(close):]
	}
	return parts
}

func normaliseSeverityTag(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "low", "minor", "info", "informational", "ok":
		return "low"
	case "medium", "warn", "warning", "moderate", "degraded":
		return "medium"
	case "high", "error", "severe":
		return "high"
	case "critical", "crit", "fatal", "fail", "failed":
		return "critical"
	}
	return ""
}

func NormaliseSeverity(s string) string {
	switch normaliseSeverityTag(s) {
	case "low":
		return "ok"
	case "medium":
		return "warn"
	case "high", "critical":
		return "crit"
	}
	return "info"
}
