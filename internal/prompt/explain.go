// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// template for the explain command prompt

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
Analyse the input below. Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.

Group the input into distinct issues. Each issue that shares a root cause or component belongs in one entry.

The JSON must use exactly these field names:

SCHEMA
{"issues":[{"what":"string","why":"string"}]}

FIELD RULES
- "what": 2–4 sentences. Describe precisely what is happening for this issue. Reference the specific
  log lines, service names, error codes, or kernel messages that show it. Explain the observable
  behaviour and its immediate effect on the system.
- "why": 2–4 sentences. Explain the root cause of this specific issue. Reference system state,
  configuration, hardware, or software factors that produced it. If the cause cannot be determined
  from the input alone, state what is most likely and what evidence supports that conclusion.
- Both values must be plain strings. Do not use arrays or nested objects.
- Do not suggest fixes, commands to run, or remediation steps in either field.
- Group related log lines into a single entry. Do not create one entry per log line.
- If there is only one issue the array has one entry.
%s
Input:
%s`, EnvironmentSection(env), cmdContext, input)
}
