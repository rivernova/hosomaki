// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"
)

// SchemaUpdates is the JSON schema for the updates command output.
const SchemaUpdates = `{"summary":"...","updates":[{"package":"...","installed":"...","available":"...","category":"...","reboot_required":false}]}`

// UpdateFinding represents a single pending update with AI analysis.
type UpdateFinding struct {
	Package        string `json:"package"`
	Installed      string `json:"installed"`
	Available      string `json:"available"`
	Category       string `json:"category"`       // "security", "major", "minor", "unknown"
	RebootRequired bool   `json:"reboot_required"` // heuristic, may be false
}

// UpdatesResult is the full AI response for the updates command.
type UpdatesResult struct {
	Summary string          `json:"summary"`
	Updates []UpdateFinding `json:"updates"`
}

// UpdatesInput carries the data needed to build the updates prompt.
// Updates is a pre-formatted, pre-sanitised plain-text string.
type UpdatesInput struct {
	Environment  string
	Updates      string
	SecurityOnly bool
}

// Updates builds the AI prompt for the updates command.
func Updates(in UpdatesInput) string {
	pkgsText := strings.TrimSpace(in.Updates)
	if pkgsText == "" {
		pkgsText = "(no pending updates)"
	}

	filterNote := ""
	if in.SecurityOnly {
		filterNote = "\nOnly security-related updates were requested (--security-only)."
	}

	envText := ""
	if in.Environment != "" {
		envText = fmt.Sprintf("System: %s\n", in.Environment)
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
  - "reboot_required": boolean.
- If there are no pending updates, return {"summary":"...","updates":[]}.
- Do not invent updates not present in the list.

OUTPUT FORMAT
No markdown. No bullet points. No headers.
All string values are plain prose.

Pending updates:
%s`, envText, filterNote, SchemaUpdates, pkgsText)
}
