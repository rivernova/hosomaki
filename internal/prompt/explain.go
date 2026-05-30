// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// this file contains the prompt template and builder for the "explain" command.

const explainFull = `You are a Linux sysadmin expert. Analyze the log input and return structured XML.

MANDATORY OUTPUT FORMAT — return ONLY this XML, nothing else:

<analysis>
  <component>
    <source>pipe|service:<name>|boot:<id>|dmesg|file:<name>|inline</source>
    <pattern>Detailed, specific, fully explained description of what is misbehaving and the exact observed failure mode. Do not copy-paste raw log lines — describe the event in clear technical narrative. This field MUST be thorough and complete.</pattern>
    <cause>Detailed, specific, fully explained root cause. Explain the underlying mechanism that produced the failure — not just what happened, but why it happened. This field MUST be thorough and complete.</cause>
  </component>
</analysis>

RULES — every rule is mandatory, none are optional:

Return EXACTLY ONE <component> block per distinct issue, pattern, anomaly, or signal.
NEVER merge multiple issues into a single <component>.
NEVER omit a <component> if an issue exists.
<source> MUST be one of the abstract identifiers listed above. It MUST NOT contain paths, log lines, code, stack traces, or any text from the input.
<pattern> MUST be detailed, specific, and fully explained. Brief or vague entries are not acceptable.
<cause> MUST be detailed, specific, and fully explained. Brief or vague entries are not acceptable.
Do NOT include <severity> — this command does not classify severity.
Do NOT include <suggestion> — this command does not suggest fixes.
Do NOT wrap the output in markdown code fences.
Do NOT use markdown formatting (asterisks, backticks, bullet lists) inside any tag.
Do NOT produce any text outside the <analysis> root element.
If the input contains no issues, return: <analysis></analysis>
` + prohibitions + `
%s%s%s
=== LOG INPUT ===
%s`

const explainBrief = `You are a Linux sysadmin expert. Analyze the log input and return structured XML.

MANDATORY OUTPUT FORMAT — return ONLY this XML, nothing else:

<analysis>
  <component>
    <source>pipe|service:<name>|boot:<id>|dmesg|file:<name>|inline</source>
    <pattern>Concise description of the failure.</pattern>
    <cause>Concise root cause.</cause>
  </component>
</analysis>

RULES — every rule is mandatory:

Return EXACTLY ONE <component> block per distinct issue.
NEVER merge multiple issues into a single <component>.
<source> is a semantic identifier only — never paths, never log content.
Do NOT include <severity> or <suggestion>.
Do NOT produce any text outside the <analysis> root element.
If there are no issues, return: <analysis></analysis>
` + prohibitions + `
%s%s%s
=== LOG INPUT ===
%s`

func Explain(input, source, command string, env collector.Environment, language, brief string) string {
	tmpl := explainFull
	if brief == "brief" {
		tmpl = explainBrief
	}

	lang := ""
	if l := languageLine(language); l != "" {
		lang = "\n" + l + "\n"
	}

	envBlock := EnvironmentSection(env)

	cmdCtx := ""
	if c := strings.TrimSpace(command); c != "" {
		cmdCtx = fmt.Sprintf("\nThese logs were produced by: %s\n", c)
	}

	srcHint := ""
	if s := strings.TrimSpace(source); s != "" {
		srcHint = fmt.Sprintf("\nThe correct <source> value for this input is: %s\n", s)
	}

	body := strings.TrimSpace(input)
	if body == "" {
		body = "(no data)"
	}

	return fmt.Sprintf(tmpl, lang, envBlock, cmdCtx+srcHint, body)
}
