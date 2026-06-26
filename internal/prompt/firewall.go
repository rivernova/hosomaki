// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

type FirewallFinding struct {
	Severity string `json:"severity"`
	Rule     string `json:"rule"`
	Port     string `json:"port"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

type FirewallResult struct {
	Summary  string            `json:"summary"`
	Findings []FirewallFinding `json:"findings"`
}

type FirewallInput struct {
	Environment collector.Environment
	Rules       string
	CrossCheck  string
}

func Firewall(in FirewallInput) string {
	extra := ""
	if in.CrossCheck != "" {
		extra = fmt.Sprintf(`
CROSS-REFERENCE DATA
The following ports are currently listening on this system.
Flag any listening port that has NO matching firewall allow rule
as a warning — it means the service is reachable but unprotected.
Also flag any firewall rule that allows a port with NO corresponding
listener — it means the rule may be stale or vestigial.

Listening ports:
%s

`, in.CrossCheck)
	}

	return fmt.Sprintf(`You are a Linux security expert reviewing the active firewall rules on a live system.

%s
IMPORTANT
If read_status is "partial" or "failed", or collection_warning is present,
you MUST mention incomplete collection in the summary. Never claim full
coverage when rules may be missing.

%s
TASK
Analyse the firewall rules below for SECURITY ISSUES. Do NOT list or repeat the
rules in your output. Only flag rules that are problematic.

A rule is worth flagging when any of the following apply:
- A critical service port (SSH:22, HTTPS:443, DNS:53) is exposed to any source (0.0.0.0/0, Anywhere).
- A database or management port (MySQL:3306, PostgreSQL:5432, Redis:6379, MongoDB:27017, Cockpit:9090) is exposed to any source.
- A rule allows ALL ports (-p all, 0-65535) to any source.
- A rule's action is DROP or REJECT for a legitimate service — operator may have locked themselves out.
- A rich/advanced rule has unusual or suspicious criteria.
- A default policy is ACCEPT (should typically be DROP or REJECT for INPUT/FORWARD).

Do NOT flag:
- Standard service ports (22, 80, 443) restricted to specific source IPs or subnets.
- Established/related connections (these are normal session tracking rules).
- Empty findings — if everything looks reasonable, return an empty findings array.
%s
OUTPUT
You MUST return ONLY a JSON object with exactly two fields: "summary" and "findings".
No other fields. No prose before or after. No markdown fences.

VALID example:
{"summary":"nftables active with 10 rules, no issues found.","findings":[]}

SCHEMA (use exactly these field names):
%s

FIELD RULES
- "summary": one to two sentences. State backend, rule count, collection completeness, and whether anything needs attention. Maximum 40 words.
- "findings": one entry per distinct concern.
- "severity": exactly one of "critical", "warning", or "info".
- "rule": rule identifier from input (rule_1, rule_2, ...) or "overall".
- "port": relevant port number or "".
- "title": concise plain-text label.
- "detail": 2–4 sentences describing the issue and recommended action.

Firewall rules:
%s`, EnvironmentSection(in.Environment), collectionWarningSection(in.Rules), extra, SchemaFirewall, in.Rules)
}

func collectionWarningSection(rules string) string {
	if !strings.Contains(rules, "collection_warning:") &&
		!strings.Contains(rules, "read_status: partial") &&
		!strings.Contains(rules, "read_status: failed") {
		return ""
	}
	return "Collection may be incomplete — treat missing-rule conclusions with caution.\n"
}
