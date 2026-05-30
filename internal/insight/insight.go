// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package insight

import (
	"strings"
)

// this file contains the logic for insight parsing and normalisation
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
	a := parseAnalysis(raw)
	for i := range a.Components {
		a.Components[i].Severity = ""
		a.Components[i].Suggestion = ""
	}
	return a
}

func ParseStatus(raw string) Analysis {
	a := parseAnalysis(raw)
	for i := range a.Components {
		a.Components[i].Suggestion = ""
	}
	return a
}

func ParseDoctor(raw string) Analysis {
	return parseAnalysis(raw)
}

func parseAnalysis(raw string) Analysis {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return Analysis{}
	}

	if components := parseXMLComponents(clean); len(components) > 0 {
		return Analysis{Components: components}
	}

	if components := parseKeyValueBlocks(clean); len(components) > 0 {
		return Analysis{Components: components}
	}

	return Analysis{Raw: clean}
}

func parseXMLComponents(raw string) []Component {
	chunks := splitTag(raw, "component")
	if len(chunks) == 0 {
		return nil
	}

	var components []Component
	for _, chunk := range chunks {
		source := strings.TrimSpace(extractTag(chunk, "source"))
		pattern := strings.TrimSpace(extractTag(chunk, "pattern"))
		cause := strings.TrimSpace(extractTag(chunk, "cause"))
		severity := normaliseSeverityTag(strings.TrimSpace(extractTag(chunk, "severity")))
		suggestion := strings.TrimSpace(extractTag(chunk, "suggestion"))

		if pattern == "" && cause == "" {
			continue
		}
		if isHealthyContent(source + " " + pattern + " " + cause) {
			continue
		}

		if source == "" {
			source = "pipe"
		}

		components = append(components, Component{
			Source:     source,
			Pattern:    pattern,
			Cause:      cause,
			Severity:   severity,
			Suggestion: suggestion,
		})
	}
	return components
}

func parseKeyValueBlocks(raw string) []Component {
	var components []Component
	var cur map[string]string
	var currentKey string

	flush := func() {
		if len(cur) == 0 {
			return
		}
		for k, v := range cur {
			cur[k] = strings.TrimSpace(v)
		}
		pattern := cur["pattern"]
		cause := cur["cause"]
		if pattern == "" && cause == "" {
			cur = nil
			currentKey = ""
			return
		}
		if isHealthyContent(pattern + " " + cause) {
			cur = nil
			currentKey = ""
			return
		}
		src := cur["source"]
		if src == "" {
			src = "pipe"
		}
		components = append(components, Component{
			Source:     src,
			Pattern:    pattern,
			Cause:      cause,
			Severity:   normaliseSeverityTag(cur["severity"]),
			Suggestion: cur["suggestion"],
		})
		cur = nil
		currentKey = ""
	}

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if cur != nil {
				flush()
			}
			continue
		}

		if isMarkdownNoise(trimmed) {
			continue
		}

		colonIdx := strings.Index(trimmed, ":")
		if colonIdx > 0 && colonIdx <= 40 {
			label := strings.ToLower(strings.TrimSpace(trimmed[:colonIdx]))
			canonical := normFieldLabel(label)
			if canonical != "" {
				if cur == nil {
					cur = make(map[string]string)
				}
				if _, exists := cur[canonical]; exists && canonical == "pattern" {
					flush()
					cur = make(map[string]string)
				}
				cur[canonical] = strings.TrimSpace(trimmed[colonIdx+1:])
				currentKey = canonical
				continue
			}
		}

		if currentKey != "" && cur != nil {
			cur[currentKey] += " " + trimmed
		}
	}
	flush()

	return components
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
			parts = append(parts, rest)
			break
		}
		parts = append(parts, rest[:end])
		rest = rest[end+len(close):]
	}
	return parts
}

func normFieldLabel(label string) string {
	switch label {
	case "source", "origin", "input":
		return "source"
	case "pattern", "symptom", "issue", "problem", "error", "observation",
		"what happened", "description", "details":
		return "pattern"
	case "cause", "root cause", "reason", "why", "explanation":
		return "cause"
	case "severity", "level", "priority":
		return "severity"
	case "suggestion", "action", "fix", "solution", "recommendation",
		"suggested action", "suggested fix", "what to do", "resolution":
		return "suggestion"
	}
	return ""
}

func normaliseSeverityTag(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "low", "minor", "info", "informational":
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

func isHealthyContent(text string) bool {
	lower := strings.ToLower(text)
	healthy := []string{
		"no issues", "no errors", "no failures", "no problems", "no anomalies",
		"healthy", "everything is fine", "all systems operational", "no action required",
		"operating normally", "system is healthy", "no alerts",
	}
	for _, h := range healthy {
		if strings.Contains(lower, h) {
			return true
		}
	}
	return false
}

func isMarkdownNoise(line string) bool {
	if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "---") ||
		strings.HasPrefix(line, "===") || strings.HasPrefix(line, "***") ||
		strings.HasPrefix(line, "#") {
		return true
	}
	if (strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") ||
		strings.HasPrefix(line, "+ ")) && !strings.Contains(line, ":") {
		return true
	}
	return false
}
