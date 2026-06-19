// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt logic for the updates command

type UpdateFinding struct {
	Package        string `json:"package"`
	Installed      string `json:"installed"`
	Available      string `json:"available"`
	Category       string `json:"category"` // "security", "major", "minor", "unknown"
	RebootRequired bool   `json:"reboot_required"`
}

type UpdatesResult struct {
	Summary string          `json:"summary"`
	Updates []UpdateFinding `json:"updates"`
}

type UpdatesInput struct {
	Environment  collector.Environment
	Updates      string
	SecurityOnly bool
}

func Updates(in UpdatesInput) string {
	pkgsText := strings.TrimSpace(in.Updates)
	if pkgsText == "" {
		pkgsText = "(no pending updates)"
	}

	filterNote := ""
	if in.SecurityOnly {
		filterNote = "\nOnly security-related updates were requested (--security-only)."
	}

	return fmt.Sprintf(`You are a Linux security and operations expert reviewing pending system package updates.

%s
TASK
Below is a list of pending package updates available on this system.%s
Analyse each update and help the operator understand what is changing, what the risk is,
and which updates should be prioritised.

Return ONLY a JSON object -- no prose, no markdown fences, no text outside the JSON.
The JSON must use exactly these field names. Do not rename, abbreviate, or add fields.

SCHEMA
%s

FIELD RULES
- "summary": one to three plain-prose sentences. State the overall update posture.
  Mention the total number of pending updates and any critical security concerns.
  Do not list individual updates here.
- "updates": array of objects. Each object has these fields:
  - "package": the exact package name from the list (string).
  - "installed": currently installed version string, or "" if unknown (string).
  - "available": version string that would be installed after update (string).
  - "category": MUST be exactly one of: "security", "major", "minor", "unknown".
    Use "security" when flagged as a security fix or advisory (CVE, DSA, RHSA, etc).
    Use "major" for significant version jumps (e.g. 1.x to 2.x) with likely
    breaking changes. Use "minor" for routine patch or feature releases.
    Use "unknown" only when you truly cannot determine the category.
  - "reboot_required": boolean. True when the package is a kernel, init system,
    graphics driver (nvidia), libc/glibc, or firmware update. Default to false.
- If there are no pending updates, return {"summary":"...","updates":[]}.
- Do not invent updates not present in the list.

OUTPUT FORMAT
No markdown. No bullet points. No headers.
All string values are plain prose.

Pending updates:
%s`, EnvironmentSection(in.Environment), filterNote, SchemaUpdates, pkgsText)
}
