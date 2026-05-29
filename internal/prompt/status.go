// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import "fmt"

// this file contains the status prompt template for "status" command

const statusBase = `You are the status engine inside hosomaki, a Linux CLI tool.
Give a short, calm read on how the system is doing right now.

Return ONLY a single JSON object. No prose around it, no markdown, no code
fences. It must match this schema exactly:

{
  "healthy": true,
  "summary": "",
  "observations": [
    { "text": "", "level": "info" }
  ]
}

Field rules:
- healthy: true when nothing needs attention.
- summary: one or two sentences on the overall state.
- observations: short standalone notes worth surfacing; each has a text field
  and a level of "ok", "info", "warn" or "crit". Use an empty array when there
  is nothing to add.

CRITICAL — every text value is RAW TEXT ONLY. No colours, icons, indentation,
separators, markdown, ANSI escapes or layout of any kind. hosomaki formats the
output; formatting here breaks it.
%s%s
System data:

%s`

func Status(in StatusInput) string {
	if in.Snapshot == nil {
		return fmt.Sprintf(statusBase, "", "", "(no data)")
	}

	lang := ""
	if l := languageLine(in.Language); l != "" {
		lang = "\n" + l
	}

	brief := ""
	if in.Brief {
		brief = "\nBe brief: a single summary sentence and an empty observations array.\n"
	}

	return fmt.Sprintf(statusBase, lang, brief, formatSnapshot(in.Snapshot))
}
