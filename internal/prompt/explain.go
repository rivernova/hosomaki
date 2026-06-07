// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt logic for the explain command

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

	return fmt.Sprintf(`You are a Linux system expert. You will be given log output that has been pre-processed for safety.

%s
INPUT FORMAT
The log has been sanitised. You will see:
  - Line categories: <ERROR>, <WARN>, <INFO>, <DEBUG>, <TRANSACTION>, <SCRIPTLET>.
  - Placeholders: <URL>, <PATH>, <CONFIG_PATH>, <LOG_PATH>, <CACHE_PATH>,
    <LIB_PATH>, <HOME_PATH>, <HEX>, <UUID>, <IPV4>, <IPV6>, <MAC>,
    <EMAIL>, <REPO_CACHE>, <VERSION>.
Treat placeholders as opaque identifiers. Do not invent real values.
%s
TASK
Examine the input. Identify only lines tagged <ERROR>, <WARN>, or <SCRIPTLET>,
plus any <INFO>/<TRANSACTION>/<DEBUG> lines that are clearly part of the same
incident. Group related lines into distinct issues.

If the input contains NO <ERROR> and NO <WARN> lines and no failed transactions
or scriptlets, return EXACTLY {"issues": []} and nothing else. A healthy log
is a valid input — do not invent issues.

OUTPUT
Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.

SCHEMA
%s

FIELD RULES
- "what": 2–4 sentences. Describe precisely what is happening for this issue.
  Reference the specific line categories or placeholders that show it.
  Explain the observable behaviour and its immediate effect on the system.
- "why": 2–4 sentences. Explain the root cause of this specific issue.
  Reference system state, configuration, hardware, or software factors.
  If the cause cannot be determined from the input alone, state what is most
  likely and what evidence supports that conclusion.
- Both values must be plain strings. Do not use arrays or nested objects.
- Do not suggest fixes, commands to run, or remediation steps in either field.
- Group related log lines into a single entry. Do not create one entry per line.

Input:
%s`, EnvironmentSection(env), cmdContext, SchemaExplain, input)
}

func ExplainDiff(fromLogs, toLogs string, from, to int, env collector.Environment) string {
	fromLabel := BootLabel(from)
	toLabel := BootLabel(to)

	return fmt.Sprintf(`You are a Linux system expert. You will be given sanitised log output from two separate boots of the same system.

%s
INPUT FORMAT
Both logs have been sanitised. You will see:
  - Line categories: <ERROR>, <WARN>, <INFO>, <DEBUG>, <TRANSACTION>, <SCRIPTLET>.
  - Placeholders: <URL>, <PATH>, <CONFIG_PATH>, <LOG_PATH>, <CACHE_PATH>,
    <LIB_PATH>, <HOME_PATH>, <HEX>, <UUID>, <IPV4>, <IPV6>, <MAC>,
    <EMAIL>, <REPO_CACHE>, <VERSION>.
Treat placeholders as opaque identifiers. Do not invent real values.

TASK
Compare the two boots and identify what changed between them.
Focus only on meaningful differences: errors or warnings that appear in one boot but not the other,
recurring issues that have changed in frequency or severity, or patterns that suggest something
changed in the system between the two boots.

Do NOT report issues that are identical in both boots — only differences matter here.
If both boots are identical in terms of errors and warnings, return EXACTLY {"issues": []} and nothing else.

OUTPUT
Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.

SCHEMA
%s

FIELD RULES
- "what": 2–4 sentences. Describe the difference precisely: what appears in %s that was not in %s,
  or vice versa. State which boot the change belongs to. Reference the specific line categories or
  placeholders that evidence it.
- "why": 2–4 sentences. Explain what likely caused this change between boots. Reference system state,
  configuration changes, hardware factors, or software updates where relevant. If the cause cannot be
  determined, state what is most likely and what evidence supports that conclusion.
- Both values must be plain strings. Do not use arrays or nested objects.
- Do not suggest fixes, commands to run, or remediation steps in either field.
- Group related differences into a single entry. Do not create one entry per line.

=== %s ===
%s

=== %s ===
%s`, EnvironmentSection(env), SchemaExplain, toLabel, fromLabel, fromLabel, fromLogs, toLabel, toLogs)
}

func BootLabel(index int) string {
	switch index {
	case 0:
		return "current boot (0)"
	case -1:
		return "previous boot (-1)"
	default:
		return fmt.Sprintf("boot %d", index)
	}
}
