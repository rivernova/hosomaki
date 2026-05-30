// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt template and builder for the "explain" command.

const explainFull = `You are a Linux sysadmin expert. Analyze the log input and return structured XML.

MANDATORY OUTPUT FORMAT — return ONLY this XML, nothing else:

<analysis>
  <component>
    <source>pipe|service:<name>|boot:<id>|dmesg|file:<name>|inline</source>
    <pattern>Detailed, specific, fully explained description of what is misbehaving and the exact observed failure mode. Do not copy-paste raw log lines — describe the event in clear technical narrative. This field MUST be thorough and complete. Never truncate, never use ellipsis, never end mid-sentence.</pattern>
    <cause>Detailed, specific, fully explained root cause. Explain the underlying mechanism that produced the failure — not just what happened, but why it happened. This field MUST be thorough and complete. Never truncate, never use ellipsis, never end mid-sentence.</cause>
  </component>
  <component>
    <source>summary</source>
    <pattern>Complete synthesis of all findings across the analysis above. This is the final component. MUST be complete. Never truncate.</pattern>
    <cause>Complete synthesis of all root causes identified above. MUST be complete. Never truncate.</cause>
  </component>
</analysis>

THE SCHEMA ABOVE IS THE ONLY VALID OUTPUT STRUCTURE. Any other XML tags are FORBIDDEN and must never appear in the output.

RULES — every rule is mandatory, none are optional:

Return EXACTLY ONE <component> block per distinct issue, pattern, anomaly, or signal.
NEVER merge multiple issues into a single <component>.
NEVER omit a <component> if an issue exists.
The LAST <component> MUST have <source>summary</source> and synthesise all findings. It is ALWAYS required.
<source> MUST be one of the abstract identifiers listed above. It MUST NOT contain paths, log lines, code, stack traces, or any text from the input.
<pattern> MUST be detailed, specific, and fully explained. Brief or vague entries are not acceptable.
<cause> MUST be detailed, specific, and fully explained. Brief or vague entries are not acceptable.
Every field MUST be written in full. Never truncate. Never use "...", "…", "[...]", "etc.", "and more", or any shortening device.
Do NOT include <severity> — this command does not classify severity.
Do NOT include <suggestion> — this command does not suggest fixes.
Do NOT wrap the output in markdown code fences.
Do NOT use markdown formatting (asterisks, backticks, bullet lists) inside any tag.
Do NOT produce any text outside the <analysis> root element.
If the input contains no issues, return a single summary component: <analysis><component><source>summary</source><pattern>No issues detected in the provided input.</pattern><cause>The log input contains no errors, failures, or anomalies requiring attention.</cause></component></analysis>
%s%s%s
=== LOG INPUT ===
%s
` + prohibitions + summaryRule

const explainBrief = `You are a Linux sysadmin expert. Analyze the log input and return structured XML.

MANDATORY OUTPUT FORMAT — return ONLY this XML, nothing else:

<analysis>
  <component>
    <source>pipe|service:<name>|boot:<id>|dmesg|file:<name>|inline</source>
    <pattern>Concise but complete description of the failure. MUST NOT be truncated or end mid-sentence.</pattern>
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

Return EXACTLY ONE <component> block per distinct issue.
NEVER merge multiple issues into a single <component>.
The LAST <component> MUST have <source>summary</source>. It is ALWAYS required.
<source> is a semantic identifier only — never paths, never log content.
Every field MUST be complete. Never truncate. Never use "...", "…", "[...]", or any shortening device.
Do NOT include <severity> or <suggestion>.
Do NOT produce any text outside the <analysis> root element.
If there are no issues, return a single summary component: <analysis><component><source>summary</source><pattern>No issues detected.</pattern><cause>The log input contains no errors or anomalies.</cause></component></analysis>
%s%s%s
=== LOG INPUT ===
%s
` + prohibitions + summaryRule

func Explain(input, source, command string, env collector.Environment, language string, brief bool) string {
	tmpl := explainFull
	if brief {
		tmpl = explainBrief
	}

	lang := ""
	if l := languageLine(language); l != "" {
		lang = "\n" + l + "\n"
	}

	envBlock := EnvironmentSection(env)

	var hints strings.Builder
	if c := strings.TrimSpace(command); c != "" {
		fmt.Fprintf(&hints, "\nThese logs were produced by: %s\n", c)
	}
	if s := strings.TrimSpace(source); s != "" {
		fmt.Fprintf(&hints, "\nThe correct <source> value for this input is: %s\n", s)
	}

	body := strings.TrimSpace(input)
	if body == "" {
		body = "(no data)"
	}

	return fmt.Sprintf(tmpl, lang, envBlock, hints.String(), body)
}
