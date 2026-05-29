// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// this file contains the prompt template for the "explain" command.

const explainBase = `You are a Linux sysadmin expert. Analyse the log input below and identify problems.

RESPONSE FORMAT — STRICT:
One line per problem, exactly like this real example:
r8169: firmware failed to load at boot; r8169 module missing firmware file; sudo dnf install linux-firmware && sudo dracut -f
kernel: ACPI could not resolve symbol _SB.LPCB.EC0; outdated BIOS DSDT table; update BIOS from manufacturer website
sshd: too many failed logins from 1.2.3.4; brute force attempt on SSH; add fail2ban or block IP with firewall-cmd

The three fields separated by semicolons are:
1. what component had the problem (a real name like kernel, nginx, r8169 — NOT the word "component")
2. what symptom was observed in plain words (NOT the word "pattern")
3. what to do about it (a concrete action or command — NOT the word "suggestion" or "cause: something")

If the logs show no real errors: system: logs appear clean; no actionable issues found; no action required

FORBIDDEN — your response must NEVER contain:
- The words "component", "pattern", "cause", "suggestion" as field values
- Asterisks, backticks, bold, italic, bullet points, numbered lists
- More than one line per distinct issue
- Any text that is not a valid problem line
%s
%s
%s=== LOG INPUT ===
%s`

func Explain(input, command string, env collector.Environment, language string) string {
	lang := ""
	if l := languageLine(language); l != "" {
		lang = "\n" + l
	}

	envBlock := EnvironmentSection(env)

	cmdCtx := ""
	if c := strings.TrimSpace(command); c != "" {
		cmdCtx = fmt.Sprintf("These logs were produced by: %s\n\n", c)
	}

	body := strings.TrimSpace(input)
	if body == "" {
		body = "(no data)"
	}

	return fmt.Sprintf(explainBase, lang, envBlock, cmdCtx, body)
}
