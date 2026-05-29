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

var placeholderWords = []string{
	"pattern",
	"component",
	"suggestion",
}

var envMetadataSubjects = []string{
	"distro", "kernel", "architecture", "arch", "hostname",
	"shell", "selinux", "apparmor", "virtualisation", "virtualization",
	"init", "package-manager", "packagemanager",
}

var componentKeywords = []string{
	"dbus", "pam", "sudo", "gdm", "polkit", "kernel", "r8169",
	"bluetooth", "bluetoothd", "firewall", "firewalld", "networkmanager",
	"systemd", "journald", "sshd", "nginx", "apache", "postgresql",
	"mysql", "docker", "podman", "selinux", "apparmor", "acpi",
	"snapd", "flatpak", "dnf", "apt", "pacman", "grub", "udev",
	"spd5118", "thermal", "cpu", "memory", "oom", "disk",
}

func isPlaceholderValue(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	for _, p := range placeholderWords {
		if lower == p {
			return true
		}
	}
	if strings.HasPrefix(lower, "cause:") {
		return true
	}
	return false
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

func extractFromProse(raw string) []Issue {
	seen := map[string]bool{}
	var issues []Issue

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "`", "")
		lower := strings.ToLower(line)

		for _, kw := range componentKeywords {
			if !strings.Contains(lower, kw) {
				continue
			}
			if seen[kw] {
				continue
			}
			desc := strings.TrimSpace(line)
			if len(desc) < 15 {
				continue
			}
			seen[kw] = true
			issues = append(issues, Issue{
				Subject: kw,
				Pattern: desc,
			})
			break
		}
	}
	return issues
}

func ParseDoctor(raw string) Doctor {
	clean := strings.TrimSpace(raw)
	issues := parseInsightLines(clean)
	if len(issues) == 0 {
		issues = extractFromProse(clean)
	}
	if len(issues) == 0 {
		return Doctor{Raw: clean}
	}
	return Doctor{Healthy: false, Issues: issues}
}

func ParseStatus(raw string) Status {
	clean := strings.TrimSpace(raw)
	issues := parseInsightLines(clean)
	if len(issues) == 0 {
		return Status{Raw: clean, Summary: clean}
	}
	var obs []Observation
	for _, iss := range issues {
		text := iss.Pattern
		if iss.Cause != "" {
			text += " — " + iss.Cause
		}
		obs = append(obs, Observation{Level: "info", Text: "[" + iss.Subject + "] " + text})
	}
	return Status{Healthy: false, Observations: obs}
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
	if strings.HasPrefix(line, "#") {
		return true
	}
	if strings.HasPrefix(line, "```") {
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
		if line == "" {
			continue
		}
		if isMarkdownLine(line) {
			continue
		}

		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "__", "")
		line = strings.TrimSpace(line)

		colonIdx := strings.Index(line, ":")
		if colonIdx <= 0 {
			continue
		}

		component := stripMarkdown(line[:colonIdx])
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
			suggestion := stripMarkdown(parts[2])
			if suggestion != "" {
				iss.Actions = []Action{{Description: suggestion}}
			}
		}

		if isPlaceholderValue(iss.Pattern) || isPlaceholderValue(iss.Cause) {
			continue
		}

		if iss.Subject != "" && iss.Pattern != "" {
			issues = append(issues, iss)
		}
	}
	return issues
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
