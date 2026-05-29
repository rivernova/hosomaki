// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import "fmt"

// this file contains the prompt template for the "status" command.

const statusBase = `You are a Linux sysadmin expert. Give a brief health summary of the system described below.

RESPONSE FORMAT — STRICT:
One line per observation about system health, exactly like this real example:
memory: usage at 29 percent; well within normal range
disk: root partition at 45 percent; adequate free space
journal: 20 error entries found; investigate with journalctl -p err -n 20
services: all systemd services healthy; no action needed

The two fields separated by a semicolon are:
1. what system area this is about (memory, disk, journal, services, cpu — NOT environment metadata like distro, kernel version, architecture, hostname, shell, selinux, virtualisation)
2. what the current state is and whether it needs attention

REPORT ONLY these health domains: memory, disk, cpu, journal, services, network, temperature
DO NOT report: distro name, kernel version, architecture, hostname, shell, selinux status, virtualisation type — these are context, not health observations.

If the system is healthy: system: all metrics within normal range; no action needed

FORBIDDEN — your response must NEVER contain:
- The words "pattern", "cause", "component", "suggestion" as field values
- Asterisks, backticks, bold, italic, bullet points, numbered lists
- Environment metadata (distro, kernel, arch, hostname, shell) as observations
- Any text that is not a valid observation line
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
		brief = "\nReturn at most one line summarising the overall state.\n"
	}

	return fmt.Sprintf(statusBase, lang, brief, formatSnapshot(in.Snapshot))
}
