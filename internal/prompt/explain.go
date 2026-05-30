// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt template and builder for the "explain" command

const explainInstructions = `STRUCTURED OUTPUT MODE.
You are a data emitter. You do not converse. You do not explain. You do not use markdown.
Your entire response MUST be a single XML document.
It MUST begin with <analysis> as the very first characters.
It MUST end with </analysis> as the very last characters.
There MUST be zero characters before <analysis> and zero characters after </analysis>.

You are a Linux sysadmin expert. Analyze the log input and return structured XML.

MANDATORY OUTPUT FORMAT — return ONLY this XML, nothing else:

<analysis>
  <component>
    <source>pipe|service:<name>|boot:<id>|dmesg|file:<name>|inline</source>
    <pattern>Detailed, specific, fully explained description of what is misbehaving and the exact observed failure mode. Do not copy-paste raw log lines — describe the event in clear technical narrative. MUST be thorough and complete. Never truncate, never use ellipsis, never end mid-sentence.</pattern>
    <cause>Detailed, specific, fully explained root cause. Explain the underlying mechanism — not just what happened, but why. MUST be thorough and complete. Never truncate, never use ellipsis, never end mid-sentence.</cause>
  </component>
  <component>
    <source>summary</source>
    <pattern>Complete synthesis of all findings. MUST be complete. Never truncate.</pattern>
    <cause>Complete synthesis of all root causes. MUST be complete. Never truncate.</cause>
  </component>
</analysis>

THE SCHEMA ABOVE IS THE ONLY VALID OUTPUT STRUCTURE. Any other XML tags are FORBIDDEN.

RULES — every rule is mandatory:

Return EXACTLY ONE <component> per distinct issue, pattern, anomaly, or signal.
NEVER merge multiple issues into one <component>.
NEVER omit a <component> if an issue exists.
The LAST <component> MUST have <source>summary</source>. Always required.
<source> MUST be one of the abstract identifiers listed above. Never paths, log lines, or input text.
Do NOT include <severity> or <suggestion>.
Do NOT wrap in markdown code fences.
Do NOT use markdown formatting anywhere.
If there are no issues: <analysis><component><source>summary</source><pattern>No issues detected.</pattern><cause>The log input contains no errors or anomalies.</cause></component></analysis>`

const explainInstructionsBrief = `STRUCTURED OUTPUT MODE.
You are a data emitter. You do not converse. You do not explain. You do not use markdown.
Your entire response MUST be a single XML document.
It MUST begin with <analysis> as the very first characters.
It MUST end with </analysis> as the very last characters.
There MUST be zero characters before <analysis> and zero characters after </analysis>.

You are a Linux sysadmin expert. Analyze the log input and return structured XML — one <component> per distinct issue.

MANDATORY OUTPUT FORMAT — return ONLY this XML, nothing else:

<analysis>
  <component>
    <source>pipe|service:<name>|boot:<id>|dmesg|file:<name>|inline</source>
    <pattern>Concise but complete description of the issue. MUST NOT be truncated or end mid-sentence.</pattern>
    <cause>Concise but complete root cause. MUST NOT be truncated or end mid-sentence.</cause>
  </component>
  <component>
    <source>summary</source>
    <pattern>Concise but complete synthesis of all findings. MUST NOT be truncated.</pattern>
    <cause>Concise but complete synthesis of all root causes. MUST NOT be truncated.</cause>
  </component>
</analysis>

THE SCHEMA ABOVE IS THE ONLY VALID OUTPUT STRUCTURE. Any other XML tags are FORBIDDEN.

RULES — every rule is mandatory:

Return EXACTLY ONE <component> per distinct issue.
NEVER merge multiple issues into one <component>.
The LAST <component> MUST have <source>summary</source>. Always required.
<source> MUST be one of the abstract identifiers listed above. Never paths or log content.
Do NOT include <severity> or <suggestion>.
Do NOT wrap in markdown code fences.
Do NOT use markdown formatting anywhere.
If there are no issues: <analysis><component><source>summary</source><pattern>No issues detected.</pattern><cause>The log input contains no errors or anomalies.</cause></component></analysis>`

func Explain(input, source, command string, env collector.Environment, language string, brief bool) string {
	instructions := explainInstructions
	if brief {
		instructions = explainInstructionsBrief
	}

	var extra strings.Builder
	extra.WriteString("\n")
	extra.WriteString(EnvironmentSection(env))
	if c := strings.TrimSpace(command); c != "" {
		fmt.Fprintf(&extra, "\nThese logs were produced by: %s\n", c)
	}
	if s := strings.TrimSpace(source); s != "" {
		fmt.Fprintf(&extra, "\nThe correct <source> value for this input is: %s\n", s)
	}
	instructions += extra.String()

	if lang := languageLine(language); lang != "" {
		instructions += "\n" + lang
	}

	body := strings.TrimSpace(input)
	if body == "" {
		body = "(no data)"
	}

	return assemblePrompt(instructions, "=== LOG INPUT ===", body)
}
