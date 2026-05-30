// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import "fmt"

// prompt templates and builders for the status command

const statusFull = `You are a Linux sysadmin expert. Read the system snapshot and return structured XML describing the health of the system.

MANDATORY OUTPUT FORMAT — return ONLY this XML, nothing else:

<analysis>
  <component>
    <source>pipe</source>
    <pattern>Detailed, specific, fully explained description of what is observed in this health domain — include concrete metrics, error counts, and operational state. Do not copy raw data verbatim; synthesise it into clear technical narrative. This field MUST be thorough and complete. Never truncate, never use ellipsis, never end mid-sentence.</pattern>
    <cause>Detailed, specific, fully explained root cause or contributing factor behind the observed state. Explain why the metrics show this state — not just what the metrics show. This field MUST be thorough and complete. Never truncate, never use ellipsis, never end mid-sentence.</cause>
    <severity>low|medium|high|critical</severity>
  </component>
  <component>
    <source>summary</source>
    <pattern>Complete synthesis of all health domains observed above. MUST be complete. Never truncate.</pattern>
    <cause>Complete synthesis of all contributing factors identified above. MUST be complete. Never truncate.</cause>
    <severity>low|medium|high|critical</severity>
  </component>
</analysis>

THE SCHEMA ABOVE IS THE ONLY VALID OUTPUT STRUCTURE. Any other XML tags (<suggestion>, <processes>, <errors>, <top_process>, <bluetooth_errors>, or any tag not listed above) are FORBIDDEN and must never appear in the output.

RULES — every rule is mandatory, none are optional:

Return EXACTLY ONE <component> block per distinct health domain or detected issue.
You MUST always produce a <component> for each of these mandatory domains that has data: memory, disk, cpu, services.
For every distinct error, failure, or anomaly found in the journal or log fields, produce a SEPARATE <component> block.
NEVER merge multiple issues into a single <component>.
The LAST <component> MUST have <source>summary</source> and synthesise all findings. It is ALWAYS required.
<source> MUST always be "pipe" for all non-summary components.
<pattern> MUST be detailed, specific, and fully explained.
<cause> MUST be detailed, specific, and fully explained.
<severity> MUST be exactly one of: low, medium, high, critical — plain text, no symbols, no colour codes.
Every field MUST be written in full. Never truncate. Never use "...", "…", "[...]", "etc.", "and more", or any shortening device.
Do NOT include <suggestion> — status does not suggest fixes.
Do NOT wrap the output in markdown code fences.
Do NOT use markdown formatting (asterisks, backticks, bullet lists) inside any tag.
Do NOT produce any text outside the <analysis> root element.
If the system is completely healthy with no issues, return a single summary component: <analysis><component><source>summary</source><pattern>All monitored domains are operating within normal parameters. No failures or anomalies detected.</pattern><cause>All metrics are within expected ranges and no service failures have been recorded.</cause><severity>low</severity></component></analysis>
%s
System snapshot:

%s
` + prohibitions + summaryRule

const statusBrief = `You are a Linux sysadmin expert. Read the system snapshot and return structured XML — one <component> per health domain or issue.

MANDATORY OUTPUT FORMAT — return ONLY this XML, nothing else:

<analysis>
  <component>
    <source>pipe</source>
    <pattern>Concise but complete description of the observed state for this domain. MUST NOT be truncated or end mid-sentence.</pattern>
    <cause>Concise but complete explanation of why the system is in this state. MUST NOT be truncated or end mid-sentence.</cause>
    <severity>low|medium|high|critical</severity>
  </component>
  <component>
    <source>summary</source>
    <pattern>Concise but complete synthesis of all health domains. MUST NOT be truncated.</pattern>
    <cause>Concise but complete synthesis of all contributing factors. MUST NOT be truncated.</cause>
    <severity>low|medium|high|critical</severity>
  </component>
</analysis>

THE SCHEMA ABOVE IS THE ONLY VALID OUTPUT STRUCTURE. Any other XML tags are FORBIDDEN.

RULES — every rule is mandatory:

Return EXACTLY ONE <component> per distinct domain or issue.
Mandatory domains: memory, disk, cpu, services — produce a block for each if data is available.
The LAST <component> MUST have <source>summary</source>. It is ALWAYS required.
<source> is always "pipe" for all non-summary components.
<severity> is exactly one of: low, medium, high, critical — plain text only.
Every field MUST be complete. Never truncate. Never use "...", "…", "[...]", or any shortening device.
Do NOT include <suggestion>.
Do NOT produce any text outside the <analysis> root element.
If the system is healthy, return a single summary component with <source>summary</source>.
%s
System snapshot:

%s
` + prohibitions + summaryRule

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
