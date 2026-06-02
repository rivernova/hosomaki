// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/prompt"
)

// rendering functions for the structured outputs from the prompt package

func ParseJSON(raw string, v interface{}) error {
	if s, ok := extractJSONObject(raw); ok {
		if err := json.Unmarshal([]byte(s), v); err == nil {
			return nil
		}
	}
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	return json.Unmarshal([]byte(strings.TrimSpace(s)), v)
}

func ParseExplainJSON(raw string, result *prompt.ExplainResult) error {
	objStr, ok := extractJSONObject(raw)
	if !ok {
		return fmt.Errorf("no JSON object found in model response")
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(objStr), &m); err != nil {
		return fmt.Errorf("could not parse model response as JSON object: %w", err)
	}

	coerce := func(msg json.RawMessage) string {
		var s string
		if err := json.Unmarshal(msg, &s); err == nil {
			return strings.TrimSpace(s)
		}
		var arr []string
		if err := json.Unmarshal(msg, &arr); err == nil {
			parts := make([]string, 0, len(arr))
			for _, item := range arr {
				if t := strings.TrimSpace(item); t != "" {
					parts = append(parts, t)
				}
			}
			return strings.Join(parts, " ")
		}
		return ""
	}

	whatAliases := []string{
		"what", "what_is_happening", "whats_happening",
		"analysis", "description", "event", "events", "error", "errors",
	}
	whyAliases := []string{
		"why", "why_it_is_happening", "whys_happening",
		"cause", "causes", "reason", "reasons", "explanation", "root_cause",
	}

	for _, k := range whatAliases {
		if msg, ok := m[k]; ok {
			if v := coerce(msg); v != "" {
				result.What = v
				break
			}
		}
	}
	for _, k := range whyAliases {
		if msg, ok := m[k]; ok {
			if v := coerce(msg); v != "" {
				result.Why = v
				break
			}
		}
	}

	if result.What == "" || result.Why == "" {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sortStrings(keys)

		var vals []string
		for _, k := range keys {
			if v := coerce(m[k]); v != "" {
				vals = append(vals, v)
				if len(vals) == 2 {
					break
				}
			}
		}
		if result.What == "" && len(vals) > 0 {
			result.What = vals[0]
		}
		if result.Why == "" && len(vals) > 1 {
			result.Why = vals[1]
		}
	}

	return nil
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func extractJSONObject(s string) (string, bool) {
	start := strings.Index(s, "{")
	if start == -1 {
		return "", false
	}
	depth := 0
	inString := false
	escape := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if escape {
			escape = false
			continue
		}
		if ch == '\\' && inString {
			escape = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1], true
			}
		}
	}
	return "", false
}

func RenderDoctor(result prompt.DoctorResult) string {
	var b strings.Builder

	// ── issues section ──────────────────────────────────────────────────────
	issueBody := renderIssues(result.Issues, "no issues detected")
	b.WriteString(Section("issues", issueBody))

	// ── actions section ─────────────────────────────────────────────────────
	actionBody := renderActions(result.Actions, "no actions required")
	b.WriteString(Section("suggested actions", actionBody))

	return b.String()
}

func RenderDoctorBrief(result prompt.DoctorBriefResult) string {
	return RenderDoctor(result)
}

func RenderDoctorSummary(result prompt.DoctorResult) string {
	var b strings.Builder
	disruptive := 0
	for _, act := range result.Actions {
		if act.Disruptive {
			disruptive++
		}
	}
	b.WriteString(SummaryLine(plural(len(result.Issues), "issue found", "issues found")))
	b.WriteString(SummaryLine(plural(len(result.Actions), "action suggested", "actions suggested")))
	if disruptive > 0 {
		b.WriteString(SummaryLine(plural(disruptive, "action flagged as disruptive", "actions flagged as disruptive")))
	}
	return SectionSummary(b.String())
}

func RenderStatus(result prompt.StatusResult) string {
	var b strings.Builder

	overview := strings.TrimSpace(result.Overview)
	if overview == "" {
		overview = "(no overview)"
	}
	b.WriteString(Section("system overview", overview))

	anomalyBody := renderAnomalies(result.Anomalies, "no anomalies detected")
	b.WriteString(Section("anomalies", anomalyBody))

	return b.String()
}

func RenderStatusBrief(result prompt.StatusBriefResult) string {
	summary := strings.TrimSpace(result.Summary)
	if summary == "" {
		summary = "(no summary)"
	}
	return Section("summary", summary)
}

func RenderStatusSummary(result prompt.StatusResult) string {
	var b strings.Builder
	critical, warnings := 0, 0
	for _, a := range result.Anomalies {
		if a.Severity == "failed" {
			critical++
		} else {
			warnings++
		}
	}
	b.WriteString(SummaryLine(plural(critical, "critical issue", "critical issues")))
	b.WriteString(SummaryLine(plural(warnings, "warning", "warnings")))
	return SectionSummary(b.String())
}

func RenderExplain(result prompt.ExplainResult) string {
	var b strings.Builder

	what := strings.TrimSpace(result.What)
	if what == "" {
		what = "(no information)"
	}
	b.WriteString(Section("what is happening", what))

	why := strings.TrimSpace(result.Why)
	if why == "" {
		why = "(no information)"
	}
	b.WriteString(Section("why it is happening", why))

	return b.String()
}

func renderIssues(issues []prompt.DoctorIssue, emptyMsg string) string {
	if len(issues) == 0 {
		return BulletOK(emptyMsg)
	}
	var b strings.Builder
	for _, iss := range issues {
		summary := strings.TrimSpace(iss.Summary)
		if summary == "" {
			continue
		}
		switch iss.Severity {
		case "failed":
			b.WriteString(BulletFail(summary))
		default:
			b.WriteString(BulletWarn(summary))
		}
	}
	if b.Len() == 0 {
		return BulletOK(emptyMsg)
	}
	return b.String()
}

func renderAnomalies(anomalies []prompt.StatusAnomaly, emptyMsg string) string {
	if len(anomalies) == 0 {
		return BulletOK(emptyMsg)
	}
	var b strings.Builder
	for _, a := range anomalies {
		summary := strings.TrimSpace(a.Summary)
		if summary == "" {
			continue
		}
		switch a.Severity {
		case "failed":
			b.WriteString(BulletFail(summary))
		default:
			b.WriteString(BulletWarn(summary))
		}
	}
	if b.Len() == 0 {
		return BulletOK(emptyMsg)
	}
	return b.String()
}

func renderActions(actions []prompt.DoctorAction, emptyMsg string) string {
	if len(actions) == 0 {
		return BulletOK(emptyMsg)
	}
	var b strings.Builder
	for _, act := range actions {
		desc := strings.TrimSpace(act.Description)
		if desc == "" {
			continue
		}
		if act.Disruptive {
			b.WriteString(BulletFail(fmt.Sprintf("[disruptive] %s", desc)))
		} else {
			b.WriteString(BulletOK(desc))
		}
	}
	if b.Len() == 0 {
		return BulletOK(emptyMsg)
	}
	return b.String()
}

func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, pluralForm)
}
