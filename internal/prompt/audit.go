// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/auditor"
	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt logic for the audit command

const maxDiffItemsPerCategory = 20

type AuditFinding struct {
	Severity string `json:"severity"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

type AuditResult struct {
	Summary  string         `json:"summary"`
	Findings []AuditFinding `json:"findings"`
}

type AuditInput struct {
	Environment collector.Environment
	Diff        *auditor.AuditDiff
	BaselineAge string
}

func Audit(in AuditInput) string {
	diffText := formatDiff(in.Diff)
	if strings.TrimSpace(diffText) == "" {
		diffText = "(no changes detected)"
	}

	age := in.BaselineAge
	if age == "" {
		age = "unknown"
	}

	return fmt.Sprintf(`You are a Linux security and operations expert reviewing a system change report.

%s
TASK
A baseline snapshot of this system was taken previously. The diff below shows the changes
detected since then. Large categories are truncated; a note states how many were omitted.
Analyse the changes and assess their security and operational significance.

Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.
The JSON must use exactly these field names. Do not rename, abbreviate, or add fields.

SCHEMA
%s

FIELD RULES
- "summary": one to three plain-prose sentences. State the overall risk posture of the changes.
  If nothing is concerning, say so clearly. Do not list individual findings here.
- "findings": one entry per meaningful change or group of closely related changes.
  Group routine bulk changes (e.g. many package upgrades) into a single finding.
- "severity": MUST be exactly one of: "critical", "warning", "info".
  Use "critical" only for changes that pose an immediate security or availability risk
  (e.g. a new SUID binary, a root-owned file made world-writable, an unexpected new user,
  a new privileged listening port below 1024).
  Use "warning" for changes that warrant review but are not immediately dangerous.
  Use "info" for routine operational changes (package upgrades, expected service restarts).
- "category": MUST be exactly one of: "service", "file", "permission", "package", "network", "user".
- "title": a concise plain-text label. Reference the specific path, service name, port, or
  package involved, e.g. "sshd_config modified", "tcp port 4444 now listening", "user mallory added".
- "detail": 2–4 sentences. Explain exactly what changed, why it is or is not concerning, and
  what the operator should verify or do next. Reference specific values from the diff.
- If the diff contains no changes that warrant a finding, return {"summary":"...","findings":[]}.
- Do not invent changes not present in the diff. Do not speculate beyond what is shown.

OUTPUT FORMAT
No markdown. No bullet points. No numbered lists. No headers.
All string values are plain prose. "title" is a short label, not a full sentence.

System change diff (baseline age: %s):
%s`, EnvironmentSection(in.Environment), SchemaAudit, age, diffText)
}

func formatDiff(d *auditor.AuditDiff) string {
	var b strings.Builder

	section := func(title string, items []string) {
		if len(items) == 0 {
			return
		}
		b.WriteString("=== " + title + " ===\n")
		shown, extra := capDiffItems(items)
		for _, item := range shown {
			b.WriteString("  " + item + "\n")
		}
		if extra > 0 {
			b.WriteString(fmt.Sprintf("  … and %d more not shown\n", extra))
		}
		b.WriteByte('\n')
	}

	section("Services added", d.ServicesAdded)
	section("Services removed", d.ServicesRemoved)
	section("Files added", d.FilesAdded)
	section("Files removed", d.FilesRemoved)
	section("Files modified", filesModifiedLines(d.FilesModified))
	section("Permission changes", permissionLines(d.PermissionsChanged))
	section("Packages installed", d.PackagesAdded)
	section("Packages removed", d.PackagesRemoved)
	section("Packages updated", packagesUpdatedLines(d.PackagesUpdated))
	section("Ports now listening", d.PortsOpened)
	section("Ports no longer listening", d.PortsClosed)
	section("Users added", d.UsersAdded)
	section("Users removed", d.UsersRemoved)

	return strings.TrimSpace(b.String())
}

func capDiffItems(items []string) (shown []string, extra int) {
	if len(items) <= maxDiffItemsPerCategory {
		return items, 0
	}
	return items[:maxDiffItemsPerCategory], len(items) - maxDiffItemsPerCategory
}

func filesModifiedLines(changes []auditor.FileChange) []string {
	out := make([]string, len(changes))
	for i, fc := range changes {
		out[i] = fmt.Sprintf("%s  (size: %d → %d bytes, mtime changed: %v)",
			fc.Path, fc.OldSize, fc.NewSize, fc.OldMtime != fc.NewMtime)
	}
	return out
}

func permissionLines(changes []auditor.PermChange) []string {
	out := make([]string, len(changes))
	for i, pc := range changes {
		var parts []string
		if pc.OldMode != pc.NewMode {
			parts = append(parts, fmt.Sprintf("mode %s→%s", pc.OldMode, pc.NewMode))
		}
		if pc.OldOwner != pc.NewOwner {
			parts = append(parts, fmt.Sprintf("owner %s→%s", pc.OldOwner, pc.NewOwner))
		}
		if pc.OldGroup != pc.NewGroup {
			parts = append(parts, fmt.Sprintf("group %s→%s", pc.OldGroup, pc.NewGroup))
		}
		out[i] = fmt.Sprintf("%s  (%s)", pc.Path, strings.Join(parts, ", "))
	}
	return out
}

func packagesUpdatedLines(changes []auditor.PackageChange) []string {
	out := make([]string, len(changes))
	for i, pu := range changes {
		out[i] = fmt.Sprintf("%s  (%s → %s)", pu.Name, pu.OldVersion, pu.NewVersion)
	}
	return out
}
