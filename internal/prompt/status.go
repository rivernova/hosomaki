// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at [https://mozilla.org/MPL/2.0/](https://mozilla.org/MPL/2.0/).

package prompt

import "fmt"

// this file contains the prompt templates and builders for the status command

const statusFull = `<system>You are a Linux sysadmin expert. Read the system snapshot and output one <observation> block per health domain.

CRITICAL FORMATTING CONTRAINTS (CLI PARSER SAFE):
- Your entire response must conform strictly to the XML schema below.
- DO NOT wrap the output in markdown blocks or use formatting flags like **, ` + "`" + `, or markdown tables.
- DO NOT echo raw system log entries inside the text block. Describe the status cleanly.

<observations>
  <observation>
    <area>[The health domain name: memory, disk, cpu, services, or journal]</area>
    <text>[A highly descriptive overview containing relevant metrics, capacities, error counts, or operational health status for this specific domain. Write a clean narrative text block.]</text>
  </observation>
</observations>

REQUIREMENTS:
1. You must always generate exactly one <observation> block for each mandatory area: memory, disk, cpu, and services.
2. For each unique error component noticed in the system log fields, generate a separate <observation> block where the area is "journal", summarizing the failure clearly in the text narrative without markdown elements.</system>
%s
System data:

%s`

// status brief: one sentence only.
const statusBrief = `<system>You are a Linux sysadmin expert. Write EXACTLY ONE concise sentence describing the current structural health of this system. Include key metrics or major alerts if found. 

CRITICAL: Do not include any XML, markdown syntax, or mitigation recommendations. Plain text only.</system>

%s
System data:

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
		lang = "\n" + l
	}

	if in.Brief {
		return fmt.Sprintf(statusBrief, lang, formatSnapshot(in.Snapshot))
	}
	return fmt.Sprintf(statusFull, lang, formatSnapshotFull(in.Snapshot))
}
