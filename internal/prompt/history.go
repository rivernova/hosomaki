// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// SchemaHistory is the JSON schema for the history command output.
const SchemaHistory = `{"summary":"...","entries":[{"timestamp":"...","command":"...","summary":"..."}]}`

// HistoryEntry represents one past diagnostic in the AI response.
type HistoryEntry struct {
	Timestamp string `json:"timestamp"`
	Command   string `json:"command"`
	Summary   string `json:"summary"` // one-line preview of the result
}

// HistoryResult is the full AI response for the history command.
type HistoryResult struct {
	Summary string         `json:"summary"`
	Entries []HistoryEntry `json:"entries"`
}

// HistoryInput carries the data needed to build the history prompt.
type HistoryInput struct {
	Environment collector.Environment
	History     string // pre-formatted, pre-sanitised text of past entries
	FilterDesc  string // e.g. "last 10 entries" or "explain entries from last 7 days"
}

// History builds the AI prompt for the history command.
func History(in HistoryInput) string {
	historyText := strings.TrimSpace(in.History)
	if historyText == "" {
		historyText = "(no previous diagnostic results found)"
	}

	filterDesc := strings.TrimSpace(in.FilterDesc)
	if filterDesc == "" {
		filterDesc = "all available entries"
	}

	return fmt.Sprintf(`You are a Linux operations expert reviewing past system diagnostics.

%s
TASK
Below is a log of past diagnostic results from this system (%s).
Summarise what was investigated, identify recurring issues or patterns,
and help the operator understand what has been happening on this system
over time.

Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.
The JSON must use exactly these field names. Do not rename, abbreviate, or add fields.

SCHEMA
%s

FIELD RULES
- "summary": two to four plain-prose sentences. Summarise the overall picture:
  how many diagnostics were run, what areas were investigated, any recurring
  problems or patterns worth flagging. Do not list individual entries here.
- "entries": array of objects, one per past diagnostic. Each object has:
  - "timestamp": ISO 8601 time string from the input.
  - "command": the source command name from the input.
  - "summary": one plain-prose sentence summarising that specific diagnostic
    result. Capture the key finding — was it healthy, what was the issue,
    what action was suggested.
- If there are no history entries, return {"summary":"...","entries":[]}.
- Do not invent entries not present in the input.

OUTPUT FORMAT
No markdown. No bullet points. No headers.
All string values are plain prose.

Diagnostic history (%s):
%s`, EnvironmentSection(in.Environment), filterDesc, SchemaHistory, filterDesc, historyText)
}
