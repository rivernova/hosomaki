// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import "fmt"

// this file contains the prompt templates and builders for the doctor command

const doctorFull = `You are a Linux sysadmin expert. Read the system data, diagnose every problem, and return structured XML with concrete remediation steps.

MANDATORY OUTPUT FORMAT — return ONLY this XML, nothing else:

<analysis>
  <component>
    <source>pipe</source>
    <pattern>Detailed, specific, fully explained description of what is failing and the exact observed failure mode. Include concrete metrics, error messages, and operational impact. Do not copy raw data verbatim — synthesise it into clear technical narrative. This field MUST be thorough and complete.</pattern>
    <cause>Detailed, specific, fully explained root cause. Explain the underlying mechanism that produced the failure — not just what happened, but why it happened. This field MUST be thorough and complete.</cause>
    <severity>low|medium|high|critical</severity>
    <suggestion>Concrete, specific, actionable steps to resolve the issue. Include exact commands to run, files to inspect, and configuration values to change. If this action is potentially disruptive or irreversible, state that warning explicitly at the very beginning of this field before describing the steps.</suggestion>
  </component>
</analysis>

THE SCHEMA ABOVE IS THE ONLY VALID OUTPUT STRUCTURE. Any other XML tags (<processes>, <errors>, <top_process>, or any tag not listed above) are FORBIDDEN and must never appear in the output.

RULES — every rule is mandatory, none are optional:

Return EXACTLY ONE <component> block per distinct issue, pattern, anomaly, or signal.
NEVER merge multiple issues into a single <component>.
NEVER omit a <component> if an issue exists.
<source> MUST always be "pipe" for this command.
<pattern> MUST be detailed, specific, and fully explained.
<cause> MUST be detailed, specific, and fully explained.
<severity> MUST be exactly one of: low, medium, high, critical — plain text, no symbols, no colour codes.
<suggestion> MUST be detailed and specific. Vague suggestions like "restart the service" are not acceptable — provide exact commands and explain what each step does.
If any suggested step is potentially disruptive or irreversible, state that warning explicitly at the START of the <suggestion> field.
This tool never modifies the system — only suggest actions; do not claim to perform them.
Do NOT wrap the output in markdown code fences.
Do NOT use markdown formatting (asterisks, backticks, bullet lists) inside any tag.
Do NOT produce any text outside the <analysis> root element.
If the system is completely healthy with no issues, return: <analysis></analysis>
%s%s
System data:

%s
` + prohibitions

const doctorBrief = `You are a Linux sysadmin expert. Read the system data and return structured XML — one <component> per problem found.

MANDATORY OUTPUT FORMAT — return ONLY this XML, nothing else:

<analysis>
  <component>
    <source>pipe</source>
    <pattern>Concise description of the failure.</pattern>
    <cause>Concise root cause.</cause>
    <severity>low|medium|high|critical</severity>
    <suggestion>Concise fix. If disruptive or irreversible, state that warning first.</suggestion>
  </component>
</analysis>

THE SCHEMA ABOVE IS THE ONLY VALID OUTPUT STRUCTURE. Any other XML tags are FORBIDDEN.

RULES — every rule is mandatory:

Return EXACTLY ONE <component> per distinct issue.
NEVER merge multiple issues into a single <component>.
<source> is always "pipe".
<severity> is exactly one of: low, medium, high, critical — plain text only.
<suggestion> MUST include concrete steps. If disruptive or irreversible, warn at the start.
Do NOT produce any text outside the <analysis> root element.
%s%s
System data:

%s
` + prohibitions

func Doctor(in DoctorInput) string {
	if in.Snapshot == nil {
		if in.Brief {
			return fmt.Sprintf(doctorBrief, "", "", "(no data)")
		}
		return fmt.Sprintf(doctorFull, "", "", "(no data)")
	}

	mac := ""
	if in.Snapshot.Environment.SELinux != "" || in.Snapshot.Environment.AppArmor != "" {
		mac = "\nMAC security module active (SELinux or AppArmor) — account for it in every <suggestion>.\n"
	}

	lang := ""
	if l := languageLine(in.Language); l != "" {
		lang = "\n" + l + "\n"
	}

	if in.Brief {
		return fmt.Sprintf(doctorBrief, mac, lang, formatSnapshot(in.Snapshot))
	}
	return fmt.Sprintf(doctorFull, mac, lang, formatSnapshotFull(in.Snapshot))
}
