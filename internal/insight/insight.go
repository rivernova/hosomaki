// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package insight

import (
	"strings"
)

// this file contains logic for parsing and normalising raw AI insights

type Action struct {
	Description string
	Command     string
	Disruptive  bool
}

type Issue struct {
	Subject  string
	Pattern  string
	Cause    string
	Severity string
	Details  []string
	Actions  []Action
}

type Doctor struct {
	Healthy bool
	Summary string
	Issues  []Issue
	Raw     string
}

type Observation struct {
	Level string
	Text  string
}

type Status struct {
	Healthy      bool
	Summary      string
	Observations []Observation
	Raw          string
}

type Explain struct {
	Issues []Issue
	Raw    string
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

func splitIssues(raw string) []string {
	var parts []string
	rest := raw
	for {
		start := strings.Index(rest, "<issue>")
		if start < 0 {
			break
		}
		rest = rest[start+len("<issue>"):]
		end := strings.Index(rest, "</issue>")
		if end < 0 {
			parts = append(parts, rest)
			break
		}
		parts = append(parts, rest[:end])
		rest = rest[end+len("</issue>"):]
	}
	return parts
}

func parseXMLIssues(raw string) []Issue {
	chunks := splitIssues(raw)
	if len(chunks) == 0 {
		return nil
	}
	var issues []Issue
	for _, chunk := range chunks {
		comp := normaliseComponent(extractTag(chunk, "component"))
		symptom := strings.TrimSpace(extractTag(chunk, "symptom"))
		cause := strings.TrimSpace(extractTag(chunk, "cause"))
		action := strings.TrimSpace(extractTag(chunk, "action"))

		if comp == "" || symptom == "" {
			continue
		}
		if isHealthyContent(comp + " " + symptom + " " + cause) {
			continue
		}
		iss := Issue{
			Subject: comp,
			Pattern: symptom,
			Cause:   cause,
		}
		if action != "" {
			iss.Actions = []Action{{Description: action}}
		}
		issues = append(issues, iss)
	}
	return issues
}

func parseXMLObservations(raw string) []Observation {
	var obs []Observation
	rest := raw
	for {
		start := strings.Index(rest, "<observation>")
		if start < 0 {
			break
		}
		rest = rest[start+len("<observation>"):]
		end := strings.Index(rest, "</observation>")
		chunk := rest
		if end >= 0 {
			chunk = rest[:end]
			rest = rest[end+len("</observation>"):]
		} else {
			rest = ""
		}
		area := strings.TrimSpace(extractTag(chunk, "area"))
		text := strings.TrimSpace(extractTag(chunk, "text"))
		if area == "" || text == "" {
			continue
		}
		if isHealthyContent(area + " " + text) {
			continue
		}
		obs = append(obs, Observation{Level: "info", Text: "[" + area + "] " + text})
	}
	return obs
}

func normLabel(label string) string {
	label = strings.ToLower(strings.TrimSpace(label))
	switch label {
	case "component", "module", "service", "process", "package", "subsystem":
		return "component"
	case "area", "domain", "section":
		return "area"
	case "symptom", "what happened", "what i observed", "observation", "observed", "error", "issue", "problem", "details", "description":
		return "symptom"
	case "cause", "root cause", "why", "reason", "explanation":
		return "cause"
	case "action", "solution", "fix", "recommendation", "resolution", "suggested fix", "suggested action", "what to do":
		return "action"
	case "text", "status", "state", "health":
		return "text"
	}
	return ""
}

type rawBlock struct {
	fields map[string]string
}

func parseBlocks(raw string) []rawBlock {
	var blocks []rawBlock
	var current map[string]string
	var currentKey string

	flush := func() {
		if len(current) > 0 {
			for k, v := range current {
				current[k] = strings.TrimSpace(v)
			}
			blocks = append(blocks, rawBlock{fields: current})
		}
		current = nil
		currentKey = ""
	}

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if current != nil {
				flush()
			}
			continue
		}

		if strings.HasPrefix(trimmed, "```") ||
			strings.HasPrefix(trimmed, "---") ||
			strings.HasPrefix(trimmed, "===") ||
			strings.HasPrefix(trimmed, "***") {
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		colonIdx := strings.Index(trimmed, ":")
		if colonIdx > 0 && colonIdx <= 40 {
			rawLabel := trimmed[:colonIdx]
			hasDigit := false
			for _, r := range rawLabel {
				if r >= '0' && r <= '9' {
					hasDigit = true
					break
				}
			}
			if !hasDigit {
				canonical := normLabel(rawLabel)
				if canonical != "" {
					value := strings.TrimSpace(trimmed[colonIdx+1:])
					if current == nil {
						current = make(map[string]string)
					}
					if _, exists := current[canonical]; exists && canonical == "component" {
						flush()
						current = make(map[string]string)
					}
					current[canonical] = value
					currentKey = canonical
					continue
				}
			}
		}

		if currentKey != "" && current != nil {
			current[currentKey] += " " + trimmed
			continue
		}
	}
	flush()
	return blocks
}

func blocksToIssues(blocks []rawBlock, includeAction bool) []Issue {
	var issues []Issue
	for _, b := range blocks {
		comp := normaliseComponent(b.fields["component"])
		symptom := strings.TrimSpace(b.fields["symptom"])
		if comp == "" || symptom == "" {
			continue
		}
		cause := strings.TrimSpace(b.fields["cause"])
		if isHealthyContent(comp + " " + symptom + " " + cause) {
			continue
		}
		iss := Issue{Subject: comp, Pattern: symptom, Cause: cause}
		if includeAction {
			if act := strings.TrimSpace(b.fields["action"]); act != "" {
				iss.Actions = []Action{{Description: act}}
			}
		}
		issues = append(issues, iss)
	}
	return issues
}

func blocksToObservations(blocks []rawBlock) []Observation {
	var obs []Observation
	for _, b := range blocks {
		area := strings.TrimSpace(b.fields["area"])
		if area == "" {
			area = strings.TrimSpace(b.fields["component"])
		}
		text := strings.TrimSpace(b.fields["text"])
		if text == "" {
			text = strings.TrimSpace(b.fields["symptom"])
		}
		if area == "" || text == "" {
			continue
		}
		if isHealthyContent(area + " " + text) {
			continue
		}
		obs = append(obs, Observation{Level: "info", Text: "[" + area + "] " + text})
	}
	return obs
}

var knownComponents = map[string]string{
	"r8169": "r8169", "realtek": "r8169", "nic": "r8169", "network interface": "r8169",
	"firewalld": "firewalld", "firewall": "firewalld",
	"gdm": "gdm", "gdm-password": "gdm-password", "gnome display": "gdm",
	"gkr-pam": "gdm-password", "keyring": "gdm-password",
	"dbus": "dbus-broker", "dbus-broker": "dbus-broker", "d-bus": "dbus-broker",
	"bluetooth": "bluetooth", "hci0": "bluetooth",
	"acpi":   "acpi",
	"sudo":   "sudo",
	"polkit": "polkitd", "polkitd": "polkitd",
	"networkmanager": "networkmanager", "network manager": "networkmanager",
	"systemd": "systemd",
	"sshd":    "sshd", "ssh": "sshd",
	"kernel":     "kernel",
	"avahi":      "avahi",
	"docker":     "docker",
	"podman":     "podman",
	"nginx":      "nginx",
	"apache":     "apache",
	"postgresql": "postgresql",
	"mysql":      "mysql",
	"snap":       "snapd", "snapd": "snapd",
	"ollama": "ollama",
}

func componentFromText(text string) string {
	lower := strings.ToLower(text)
	for keyword, comp := range knownComponents {
		if strings.Contains(lower, keyword) {
			return comp
		}
	}
	return ""
}

func extractNumberedSections(raw string) []Issue {
	lines := strings.Split(raw, "\n")
	var issues []Issue
	seen := map[string]bool{}

	type section struct {
		heading string
		body    []string
	}

	var sections []section
	var cur *section

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		stripped := stripNumericPrefix(trimmed)
		isHeading := stripped != trimmed

		if !isHeading {
			lower := strings.ToLower(trimmed)
			for _, word := range []string{"problem", "issue", "error", "bug"} {
				if strings.HasPrefix(lower, word) {
					isHeading = true
					stripped = trimmed
					break
				}
			}
			if strings.HasPrefix(trimmed, "#") {
				isHeading = true
				stripped = strings.TrimLeft(trimmed, "# ")
			}
		}

		if isHeading {
			if cur != nil {
				sections = append(sections, *cur)
			}
			cur = &section{heading: stripped}
			continue
		}

		if cur != nil {
			cur.body = append(cur.body, trimmed)
		}
	}
	if cur != nil {
		sections = append(sections, *cur)
	}

	for _, sec := range sections {
		if len(sec.body) == 0 {
			continue
		}

		comp := componentFromText(sec.heading)
		if comp == "" {
			comp = componentFromText(strings.Join(sec.body, " "))
		}
		if comp == "" || seen[comp] {
			continue
		}

		var descLines, solutionLines []string
		inSolution := false
		for _, bl := range sec.body {
			lower := strings.ToLower(bl)
			isSolutionLabel := false
			for _, word := range []string{"solution:", "fix:", "resolution:", "recommendation:", "to resolve", "to fix", "suggested action:", "action:"} {
				if strings.HasPrefix(lower, word) {
					inSolution = true
					isSolutionLabel = true
					idx := strings.Index(bl, ":")
					if idx >= 0 {
						rest := strings.TrimSpace(bl[idx+1:])
						if rest != "" {
							solutionLines = append(solutionLines, rest)
						}
					}
					break
				}
			}
			if isSolutionLabel {
				continue
			}
			if inSolution {
				solutionLines = append(solutionLines, bl)
			} else {
				descLines = append(descLines, bl)
			}
		}

		symptom := strings.TrimSpace(strings.Join(descLines, " "))
		symptom = firstSentences(symptom, 2)
		if symptom == "" {
			symptom = strings.TrimSpace(sec.heading)
		}

		cause := ""
		action := ""
		if len(solutionLines) > 0 {
			action = strings.TrimSpace(strings.Join(solutionLines, " "))
		}

		_ = cause

		if isHealthyContent(comp + " " + symptom) {
			continue
		}

		seen[comp] = true
		iss := Issue{Subject: comp, Pattern: symptom}
		if action != "" {
			iss.Actions = []Action{{Description: action}}
		}
		issues = append(issues, iss)
	}
	return issues
}

func firstSentences(text string, n int) string {
	count := 0
	for i, r := range text {
		if r == '.' || r == '!' || r == '?' {
			count++
			if count == n {
				return strings.TrimSpace(text[:i+1])
			}
		}
	}
	return text
}

func ParseDoctor(raw string) Doctor {
	clean := strings.TrimSpace(raw)

	issues := parseXMLIssues(clean)

	if len(issues) == 0 {
		blocks := parseBlocks(clean)
		issues = blocksToIssues(blocks, true)
	}

	if len(issues) == 0 {
		issues = parseInsightLines(clean)
	}

	if len(issues) == 0 {
		issues = extractNumberedSections(clean)
	}

	if len(issues) == 0 {
		return Doctor{Raw: clean}
	}
	return Doctor{Healthy: false, Issues: issues}
}

func ParseStatus(raw string) Status {
	clean := strings.TrimSpace(raw)

	obs := parseXMLObservations(clean)

	if len(obs) == 0 {
		blocks := parseBlocks(clean)
		obs = blocksToObservations(blocks)
	}

	if len(obs) == 0 {
		issues := parseInsightLines(clean)
		for _, iss := range issues {
			text := iss.Pattern
			if iss.Cause != "" {
				text += " — " + iss.Cause
			}
			obs = append(obs, Observation{Level: "info", Text: "[" + iss.Subject + "] " + text})
		}
	}

	if len(obs) == 0 {
		return Status{Raw: clean, Summary: clean}
	}
	return Status{Healthy: false, Observations: obs}
}

func ParseExplain(raw string) Explain {
	clean := strings.TrimSpace(raw)

	issues := parseXMLIssues(clean)
	for i := range issues {
		issues[i].Actions = nil
	}

	if len(issues) == 0 {
		blocks := parseBlocks(clean)
		issues = blocksToIssues(blocks, false)
	}

	if len(issues) == 0 {
		issues = parseInsightLines(clean)
		for i := range issues {
			issues[i].Actions = nil
		}
	}

	if len(issues) == 0 {
		issues = extractNumberedSections(clean)
		for i := range issues {
			issues[i].Actions = nil
		}
	}

	if len(issues) == 0 {
		return Explain{Raw: clean}
	}
	return Explain{Issues: issues}
}

func isHealthyContent(text string) bool {
	lower := strings.ToLower(text)
	for _, phrase := range []string{
		"no issues", "no errors", "no problems", "all healthy", "all services healthy",
		"within normal", "no action needed", "no action required", "logs appear clean",
		"no failed", "no degraded", "all metrics", "functioning normally",
		"no issues detected", "healthy system", "system is healthy",
	} {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

func normaliseComponent(s string) string {
	s = strings.TrimSpace(s)
	if lb := strings.Index(s, "["); lb >= 0 {
		s = strings.TrimSpace(s[:lb])
	}
	s = strings.TrimSuffix(s, ":")
	return strings.TrimSpace(s)
}

var envMetadataSubjects = []string{
	"distro", "kernel", "architecture", "arch", "hostname",
	"shell", "selinux", "apparmor", "virtualisation", "virtualization",
	"init", "package-manager", "packagemanager",
}

func isEnvMetadataSubject(subject string) bool {
	lower := strings.ToLower(strings.TrimSpace(subject))
	for _, m := range envMetadataSubjects {
		if lower == m {
			return true
		}
	}
	return false
}

func stripMarkdown(s string) string {
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	s = strings.ReplaceAll(s, "`", "")
	for _, pfx := range []string{"- ", "* ", "+ "} {
		s = strings.TrimPrefix(s, pfx)
	}
	return strings.TrimSpace(s)
}

func isMarkdownLine(line string) bool {
	if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "```") {
		return true
	}
	if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "===") || strings.HasPrefix(line, "***") {
		return true
	}
	if strings.HasPrefix(line, "{") || strings.HasPrefix(line, "[") {
		return true
	}
	if (strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "+ ")) &&
		!strings.Contains(line, ";") {
		return true
	}
	return false
}

func parseInsightLines(raw string) []Issue {
	var issues []Issue
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || isMarkdownLine(line) {
			continue
		}
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "__", "")
		line = strings.TrimSpace(line)
		line = stripNumericPrefix(line)

		colonIdx := strings.Index(line, ":")
		if colonIdx <= 0 {
			continue
		}
		component := stripMarkdown(line[:colonIdx])
		if lb := strings.Index(component, "["); lb >= 0 {
			component = strings.TrimSpace(component[:lb])
		}
		rest := strings.TrimSpace(line[colonIdx+1:])
		if !strings.Contains(rest, ";") {
			continue
		}
		if strings.Contains(component, " ") || len(component) == 0 || len(component) > 60 {
			continue
		}
		if isEnvMetadataSubject(component) {
			continue
		}
		parts := splitSemicolon(rest)
		iss := Issue{Subject: component}
		if len(parts) >= 1 {
			iss.Pattern = stripMarkdown(parts[0])
		}
		if len(parts) >= 2 {
			iss.Cause = stripMarkdown(parts[1])
		}
		if len(parts) >= 3 {
			if s := stripMarkdown(parts[2]); s != "" {
				iss.Actions = []Action{{Description: s}}
			}
		}
		if iss.Subject != "" && iss.Pattern != "" {
			issues = append(issues, iss)
		}
	}
	return issues
}

func stripNumericPrefix(line string) string {
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	if i > 0 && i < len(line) && (line[i] == '.' || line[i] == ')') {
		if rest := strings.TrimSpace(line[i+1:]); rest != "" {
			return rest
		}
	}
	return line
}

func splitSemicolon(s string) []string {
	var parts []string
	var cur strings.Builder
	for _, r := range s {
		if r == ';' {
			parts = append(parts, cur.String())
			cur.Reset()
		} else {
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
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
