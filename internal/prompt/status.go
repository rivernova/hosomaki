// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import "fmt"

// this file contains the prompt templates and builders for the status command.

const statusFull = `You are a Linux sysadmin expert. Read the system snapshot and return structured XML describing the health of the system.

MANDATORY OUTPUT FORMAT — return ONLY this XML, nothing else:

<analysis>
  <component>
    <source>pipe</source>
    <pattern>Detailed, specific, fully explained description of what is observed in this health domain — include concrete metrics, error counts, and operational state. Do not copy raw data verbatim; synthesise it into clear technical narrative. This field MUST be thorough and complete.</pattern>
    <cause>Detailed, specific, fully explained root cause or contributing factor behind the observed state. If the state is healthy, explain why the metrics indicate normal operation. This field MUST be thorough and complete.</cause>
    <severity>low|medium|high|critical</severity>
  </component>
</analysis>

RULES — every rule is mandatory, none are optional:

Return EXACTLY ONE <component> block per distinct health domain or detected issue.
You MUST always produce a <component> for each of these mandatory domains that has data: memory, disk, cpu, services.
For every distinct error, failure, or anomaly found in the journal or log fields, produce a SEPARATE <component> block.
NEVER merge multiple issues into a single <component>.
<source> MUST always be "pipe" for this command.
<pattern> MUST be detailed, specific, and fully explained.
<cause> MUST be detailed, specific, and fully explained.
<severity> MUST be exactly one of: low, medium, high, critical — plain text, no symbols, no colour codes.
Do NOT include <suggestion> — status does not suggest fixes.
Do NOT wrap the output in markdown code fences.
Do NOT use markdown formatting (asterisks, backticks, bullet lists) inside any tag.
Do NOT produce any text outside the <analysis> root element.
If the system is completely healthy with no issues, return <analysis></analysis>.
` + prohibitions + `
%s
System snapshot:

%s`

const statusBrief = `You are a Linux sysadmin expert. Read the system snapshot and return structured XML — one <component> per health domain or issue.

MANDATORY OUTPUT FORMAT — return ONLY this XML, nothing else:

<analysis>
  <component>
    <source>pipe</source>
    <pattern>Concise description of the observed state for this domain.</pattern>
    <cause>Concise explanation of the underlying cause or reason.</cause>
    <severity>low|medium|high|critical</severity>
  </component>
</analysis>

RULES — every rule is mandatory:

Return EXACTLY ONE <component> per distinct domain or issue.
Mandatory domains: memory, disk, cpu, services — produce a block for each if data is available.
<source> is always "pipe".
<severity> is exactly one of: low, medium, high, critical — plain text only.
Do NOT include <suggestion>.
Do NOT produce any text outside the <analysis> root element.
If entirely healthy, return: <analysis></analysis>
` + prohibitions + `
%s
System snapshot:

%s`

func Status(in StatusInput) string {
	if in.Snapshot == nil {
		if in.Brief {
			return fmt.Sprintf(statusBrief, "", "(no data)")
		}
		return fmt.Sprintf(statusFull, "", "(no data)")
	}

	lang := ""
	if l := languageLine(in.Language); l != "" {
		lang = "\n" + l + "\n"
	}

	if in.Brief {
		return fmt.Sprintf(statusBrief, lang, formatSnapshot(in.Snapshot))
	}
	return fmt.Sprintf(statusFull, lang, formatSnapshotFull(in.Snapshot))
}
