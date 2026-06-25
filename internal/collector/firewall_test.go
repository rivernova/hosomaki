// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"strings"
	"testing"
)

func TestParseFirewalldOutput_BasicPorts(t *testing.T) {
	output := "ports: 22/tcp 80/tcp 443/tcp\nservices: ssh\nsources:"
	rules := parseFirewalldOutput("public", output)
	if len(rules) != 4 {
		t.Fatalf("expected 4 rules, got %d", len(rules))
	}
}

func TestParseUfwOutput_SkipsHeader(t *testing.T) {
	output := "Status: active\nTo                         Action      From\n[ 1] 22/tcp                     ALLOW IN    Anywhere"
	rules := parseUfwOutput(output)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Port != "22" {
		t.Fatalf("expected port 22, got %q", rules[0].Port)
	}
}

func TestParseNftOutput_ChainContext(t *testing.T) {
	output := "chain INPUT {\n  tcp dport 22 accept\n}"
	rules := parseNftOutput(output)
	if len(rules) != 1 || rules[0].Chain != "INPUT" || rules[0].Port != "22" {
		t.Fatalf("unexpected nft parse: %+v", rules)
	}
}

func TestParseNftOutput_SetSyntax(t *testing.T) {
	output := "chain INPUT {\n  tcp dport { 80, 443 } accept\n}"
	rules := parseNftOutput(output)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Port != "80,443" {
		t.Fatalf("expected ports \"80,443\", got %q", rules[0].Port)
	}
	if rules[0].Action != "ACCEPT" {
		t.Fatalf("expected ACCEPT, got %q", rules[0].Action)
	}
}

func TestParseNftOutput_SetSyntaxWithRange(t *testing.T) {
	output := "chain INPUT {\n  tcp dport { 80-90 } accept\n}"
	rules := parseNftOutput(output)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Port != "80-90" {
		t.Fatalf("expected port range \"80-90\", got %q", rules[0].Port)
	}
	if rules[0].Action != "ACCEPT" {
		t.Fatalf("expected ACCEPT, got %q", rules[0].Action)
	}
}

func TestParseNftOutput_EmptySet(t *testing.T) {
	output := "chain INPUT {\n  tcp dport { } accept\n}"
	rules := parseNftOutput(output)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Port != "" {
		t.Fatalf("expected empty port for empty set, got %q", rules[0].Port)
	}
}

func TestParseIptablesOutput_DefaultPolicy(t *testing.T) {
	output := "Chain INPUT (policy DROP)\ntarget     prot opt source               destination"
	rules := parseIptablesOutput(output)
	if len(rules) != 1 || rules[0].Action != "DROP" {
		t.Fatalf("expected default policy DROP, got %+v", rules)
	}
}

func TestParseIptablesOutput_RealRuleRow(t *testing.T) {
	output := "Chain INPUT (policy DROP)\ntarget     prot opt source               destination\nACCEPT     all  --  0.0.0.0/0            0.0.0.0/0           \nACCEPT     tcp  --  0.0.0.0/0            0.0.0.0/0            tcp dpt:22"
	rules := parseIptablesOutput(output)
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules (1 policy + 2 real), got %d", len(rules))
	}
	if rules[0].Action != "DROP" || rules[0].Comment != "default policy" {
		t.Fatalf("rule[0] expected DROP policy, got %+v", rules[0])
	}
	if rules[1].Protocol != "all" || rules[1].Action != "ACCEPT" {
		t.Fatalf("rule[1] expected all/ACCEPT, got %+v", rules[1])
	}
	if rules[2].Port != "22" || rules[2].Protocol != "tcp" {
		t.Fatalf("rule[2] expected port 22/tcp, got %+v", rules[2])
	}
}

func TestFormatFirewallForPrompt_IncludesWarning(t *testing.T) {
	result := FirewallResult{
		Backend:    BackendNftables,
		ReadStatus: ReadPartial,
		Warning:    "only INPUT chain could be read",
		Rules: []FirewallRule{{
			Backend: BackendNftables, Chain: "INPUT", Action: "ACCEPT", Port: "22", Protocol: "tcp",
		}},
	}
	formatted := FormatFirewallForPrompt(result)
	for _, want := range []string{"read_status: partial", "collection_warning:", "rule_1:"} {
		if !strings.Contains(formatted, want) {
			t.Fatalf("expected %q in formatted output, got:\n%s", want, formatted)
		}
	}
}

func TestFormatFirewallForPrompt_FailedRead(t *testing.T) {
	result := FirewallResult{
		Backend:    BackendUfw,
		ReadStatus: ReadFailed,
		Warning:    "ufw is installed but inactive",
	}
	formatted := FormatFirewallForPrompt(result)
	if !strings.Contains(formatted, "read_status: failed") {
		t.Fatalf("expected failed status in output, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "rules: (none collected)") {
		t.Fatalf("expected none collected, got:\n%s", formatted)
	}
}

func TestFinalizeFirewallResult_Partial(t *testing.T) {
	result := finalizeFirewallResult(FirewallResult{
		Backend: BackendFirewalld,
		Warning: "zone read failed",
		Rules:   []FirewallRule{{Backend: BackendFirewalld, Port: "22"}},
	})
	if result.ReadStatus != ReadPartial {
		t.Fatalf("expected partial, got %q", result.ReadStatus)
	}
}

func TestDetectFirewallBackend_SystemProbe(t *testing.T) {
	backend := DetectFirewallBackend()
	t.Logf("detected backend: %q", backend)
}
