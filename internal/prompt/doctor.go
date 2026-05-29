// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import "fmt"

// this file contains the prompt template for the "doctor" command.

const doctorBase = `You are a Linux sysadmin expert. Diagnose the system described below.

RESPONSE FORMAT — STRICT:
One line per problem, exactly like this real example:
r8169: firmware failed to load at boot; r8169 module missing firmware file; sudo dnf install linux-firmware && sudo dracut -f
kernel: ACPI could not resolve symbol _SB.LPCB.EC0; outdated BIOS DSDT table; update BIOS from manufacturer website
sudo: authentication failed for user rivernova; wrong password or PAM misconfiguration; check /etc/pam.d/sudo and run passwd username

The three fields separated by semicolons are:
1. what component has the problem (a real name like kernel, nginx, r8169 — NOT the word "component")
2. what symptom was observed in plain words (NOT the word "pattern")
3. what to do about it (a concrete action or command — NOT the word "suggestion" or "cause: something")

If the system is healthy: system: all services and resources are healthy; no issues detected; no action needed

FORBIDDEN — your response must NEVER contain:
- The words "component", "pattern", "cause", "suggestion" as field values
- Asterisks, backticks, bold, italic, bullet points, numbered lists
- More than one line per distinct issue
- Any text that is not a valid problem line
%s%s%s
System data:

%s`

func Doctor(in DoctorInput) string {
	if in.Snapshot == nil {
		return fmt.Sprintf(doctorBase, "", "", "", "(no data)")
	}

	mac := ""
	if in.Snapshot.Environment.SELinux != "" || in.Snapshot.Environment.AppArmor != "" {
		mac = "\nMAC system active (SELinux or AppArmor) — account for it in suggestions.\n"
	}

	lang := ""
	if l := languageLine(in.Language); l != "" {
		lang = "\n" + l
	}

	brief := ""
	if in.Brief {
		brief = "\nOnly include issues that genuinely need action. Skip minor observations.\n"
	}

	return fmt.Sprintf(doctorBase, mac, lang, brief, formatSnapshot(in.Snapshot))
}
