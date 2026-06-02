// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// template for the prompt for the "explain" command
type ExplainEntry struct {
	What string `json:"what"`
	Why  string `json:"why"`
}
type ExplainResult struct {
	Issues []ExplainEntry `json:"issues"`
}

func Explain(input, cmd string, env collector.Environment) string {
	var cmdContext string
	if c := strings.TrimSpace(cmd); c != "" {
		cmdContext = fmt.Sprintf("\nThe output below was produced by running: %s\n", c)
	}

	return fmt.Sprintf(`You are a Linux system expert. You will be given log output or an error message.

%s
TASK
Analyse the input below. Respond with ONLY a JSON object — no other text, no markdown, no explanation.

Identify every distinct error pattern or issue in the input. Return one entry per issue.

The JSON must follow this exact structure:
{"issues":[{"what":"<string>","why":"<string>"},{"what":"<string>","why":"<string>"}]}

Rules for each entry:
- "what": a prose string describing this specific error or event. Be precise and reference the actual log lines.
- "why": a prose string explaining the root cause of this specific issue. Do not suggest fixes.
- Both values must be plain strings, not arrays or nested objects.
- If there is only one issue, the array has one entry.
- Group related log lines into a single entry. Do not create one entry per log line.
%s
Input:
%s`, EnvironmentSection(env), cmdContext, input)
}
