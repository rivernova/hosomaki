// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"

	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt logic for the ports command

type PortsFinding struct {
	Severity string `json:"severity"`
	Port     string `json:"port"` // verbatim from the input
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

type PortsResult struct {
	Summary  string         `json:"summary"`
	Findings []PortsFinding `json:"findings"`
}

type PortsInput struct {
	Environment collector.Environment
	Ports       string
}

func Ports(in PortsInput) string {
	return fmt.Sprintf(`You are a Linux system security expert reviewing the currently listening network ports on a live system.

%s
TASK
Analyse the port list below. Identify ports or services that are unexpected,
potentially insecure, or worth the operator's attention.

A port is worth flagging when any of the following apply:
- It is listening on all interfaces (0.0.0.0 or [::]) on a port associated with
  a sensitive or unusual service (e.g. databases, admin panels, debug servers,
  remote-control agents).
- The process name does not match the expected service for that port (e.g. a
  process other than sshd listening on port 22).
- The port is associated with a service that is commonly exposed accidentally
  in production: development servers, test databases, or services that should
  bind only to localhost.
- The port number is in the high ephemeral range but is consistently open,
  suggesting a persistent service registered on a non-standard port.

Do NOT flag:
- Standard system services on their canonical ports with matching process names
  (e.g. sshd on 22, nginx or apache2 on 80/443, systemd-resolved on 127.0.0.53:53).
- Ports bound exclusively to 127.0.0.1 or [::1] (loopback only), unless the
  service itself is inherently high-risk even locally.
- Empty port lists — return a summary that confirms no ports are listening.

If nothing warrants a finding, return an empty findings array.

OUTPUT
Return ONLY a JSON object — no prose, no markdown fences, no text outside the JSON.
The JSON must use exactly these field names. Do not rename, abbreviate, or add fields.

SCHEMA
%s

FIELD RULES
- "summary": one to two sentences. State the overall picture: how many ports
  are listening, whether anything stands out, and the general exposure level.
  Maximum 40 words.
- "findings": one entry per distinct concern. Do not invent ports not present
  in the input.
- "severity": exactly the string "warning" for something the operator should
  investigate promptly, or "info" for something worth noting but not urgent.
- "port": the protocol and address of the port this finding concerns.
  Copy it verbatim from the input list (e.g. "tcp 0.0.0.0:3306").
- "title": a concise plain-text label, e.g. "MySQL exposed on all interfaces".
  No trailing punctuation.
- "detail": 2–4 sentences. Describe precisely what is unusual and why it may
  be a concern. State what the operator should verify or consider. Do not
  suggest specific remediation commands.
- Do not suggest fixes, commands to run, or remediation steps.

OUTPUT FORMAT
No markdown. No bullet points. No numbered lists. No headers.
All string values are plain prose.

Currently listening ports:
%s`, EnvironmentSection(in.Environment), SchemaPorts, in.Ports)
}
