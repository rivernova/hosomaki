// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import "fmt"

// this file contains the prompt template for the "doctor" command

const doctorBase = `You are the diagnostic engine inside hosomaki, a Linux CLI tool.
Analyse the system data below and report what a careful sysadmin would.

Return ONLY a single JSON object. No prose before or after it, no markdown, no
code fences. The JSON must match this schema exactly:

{
  "healthy": false,
  "summary": "",
  "issues": [
    {
      "subject": "",
      "severity": "warn",
      "pattern": "",
      "cause": "",
      "details": [],
      "actions": [
        { "description": "", "command": "", "disruptive": false }
      ]
    }
  ]
}

Field rules:
- healthy: true only when nothing needs attention.
- summary: one or two sentences stating the overall verdict.
- issues: one object per DISTINCT problem; empty array when healthy.
- subject: the unit, device or component the issue concerns.
- severity: one of "crit", "warn", "info".
- pattern: the observed symptom, stated plainly.
- cause: the most probable underlying cause.
- details: extra plain-text observations; omit or leave empty if none.
- actions: concrete remediation steps with an optional command and
  disruptive=true when the command can cause downtime, data loss or a reboot.

CRITICAL — every text value is RAW TEXT ONLY. Do NOT add colours, indentation,
separators, icons, bullet characters, markdown, ANSI escapes or any layout.
hosomaki formats everything itself; formatting here corrupts the output.

Commands must be correct for THIS host: honour the package manager, init system
and security modules shown in the host environment below, and never invent units
or paths that the data does not support.
%s%s%s
System data:

%s`

func Doctor(in DoctorInput) string {
	if in.Snapshot == nil {
		return fmt.Sprintf(doctorBase, "", "", "", "(no data)")
	}

	mac := ""
	if in.Snapshot.Environment.SELinux != "" || in.Snapshot.Environment.AppArmor != "" {
		mac = "\nA mandatory-access-control system is active; account for SELinux or " +
			"AppArmor when proposing fixes.\n"
	}

	lang := ""
	if l := languageLine(in.Language); l != "" {
		lang = "\n" + l
	}

	brief := ""
	if in.Brief {
		brief = "\nBe brief: a single concise summary sentence per issue. Include only " +
			"issues that genuinely need action; omit minor observations.\n"
	}

	return fmt.Sprintf(doctorBase, mac, lang, brief, formatSnapshot(in.Snapshot))
}
