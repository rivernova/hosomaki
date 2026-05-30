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

	clean = stripMarkdownFence(clean)

	body, found := extractTagContent(clean, "analysis")
	if !found {
		return Analysis{Raw: clean}
	}

	return Analysis{Components: parseComponents(body)}
}

func stripMarkdownFence(s string) string {
	for _, fence := range []string{"```xml", "```XML", "```"} {
		if strings.HasPrefix(s, fence) {
			s = s[len(fence):]
			s = strings.TrimPrefix(s, "\n")
			if idx := strings.LastIndex(s, "```"); idx >= 0 {
				s = s[:idx]
			}
			return strings.TrimSpace(s)
		}
	}
	return s
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
		return ""
	}
	return s[start : start+end]
}

func splitTag(s, tag string) []string {
	open := "<" + tag + ">"
	close := "</" + tag + ">"

	var chunks []string
	for {
		start := strings.Index(s, open)
		if start < 0 {
			break
		}
		s = s[start+len(open):]
		end := strings.Index(s, close)
		if end < 0 {
			chunks = append(chunks, strings.TrimSpace(s))
			break
		}
		chunks = append(chunks, strings.TrimSpace(s[:end]))
		s = s[end+len(close):]
	}
	return chunks
}

func NormaliseSeverity(s string) string {
	return normaliseSeverityTag(s)
}

func normaliseSeverityTag(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "low", "minor":
		return "ok"
	case "medium", "warn", "warning", "moderate":
		return "warn"
	case "high", "critical", "crit", "fatal", "error":
		return "crit"
	default:
		return "info"
	}
}
