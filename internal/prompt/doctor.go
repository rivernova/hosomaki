// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at [https://mozilla.org/MPL/2.0/](https://mozilla.org/MPL/2.0/).

package prompt

import "fmt"

// this file contains the prompt templates and builders for the doctor command

const doctorFull = `<system>You are a Linux sysadmin expert. Read the system data and output exactly one <issue> block per unique problem found.

CRITICAL FORMATTING CONTRAINTS (CLI PARSER SAFE):
- Your entire response must conform strictly to the XML schema below.
- DO NOT wrap the output in markdown code blocks (e.g., do not use ` + "`" + `xml ... ` + "`" + `).
- DO NOT use any markdown formatting within the text fields (no asterisks for bolding, no backticks for inline code, no markdown tables, no markdown bullet lists). 
- Text fields must contain plain, unformatted human narrative only.
- DO NOT simply repeat or copy-paste raw logs into the tags. Explain the data natively.

<issues>
  <issue>
    <component>[The specific systemd service, kernel module, package name, or hardware element]</component>
    <symptom>[Elaborated plain-text explanation of the error state. Detail what is failing and the operational impact on the running system. No markdown formatting.]</symptom>
    <cause>[Elaborated plain-text explanation of the underlying root cause mechanism. Explain why this happened conceptually. No markdown formatting.]</cause>
    <action>[The exact plain-text bash command string or step sequence to cleanly fix the problem. No markdown formatting.]</action>
  </issue>
</issues>

If an action is potentially disruptive or irreversible, explicitly state that warning clearly inside the narrative of the <action> tag.</system>
%s%s%s
System data:

%s`

// doctor brief: compact semicolon one-liners.
const doctorBrief = `<system>You are a Linux sysadmin expert. List the system problems that need action. Provide exactly ONE plain-text line per distinct issue using this strict format:

component: what is wrong; exact command or steps to fix it

CRITICAL CONSTRAINT: Do not include any markdown bolding, backticks, or lists. Output only the pure matching lines. No intro, no conclusion.</system>
%s%s
System data:

%s`

func Doctor(in DoctorInput) string {
	if in.Snapshot == nil {
		if in.Brief {
			return fmt.Sprintf(doctorBrief, "", "", "(no data)")
		}
		return fmt.Sprintf(doctorFull, "", "", "", "(no data)")
	}

	mac := ""
	if in.Snapshot.Environment.SELinux != "" || in.Snapshot.Environment.AppArmor != "" {
		mac = "\nMAC system active (SELinux or AppArmor) — account for it in action suggestions.\n"
	}

	lang := ""
	if l := languageLine(in.Language); l != "" {
		lang = "\n" + l
	}

	if in.Brief {
		return fmt.Sprintf(doctorBrief, mac, lang, formatSnapshot(in.Snapshot))
	}
	return fmt.Sprintf(doctorFull, mac, lang, "", formatSnapshotFull(in.Snapshot))
}
