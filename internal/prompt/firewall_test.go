// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFirewall_ContainsSchema(t *testing.T) {
	result := Firewall(FirewallInput{Rules: "backend: iptables\nread_status: ok\nrules:\n  rule_1:"})
	if !strings.Contains(result, SchemaFirewall) {
		t.Fatal("expected SchemaFirewall in prompt")
	}
}

func TestFirewall_CrossCheckIncluded(t *testing.T) {
	result := Firewall(FirewallInput{
		Rules:      "backend: nftables\nread_status: ok\nrules:\n  rule_1:",
		CrossCheck: "tcp 0.0.0.0:22 ssh",
	})
	if !strings.Contains(result, "CROSS-REFERENCE DATA") {
		t.Fatal("expected cross-check section")
	}
}

func TestFirewall_WarningTriggersCaution(t *testing.T) {
	result := Firewall(FirewallInput{
		Rules: "backend: nftables\nread_status: partial\ncollection_warning: incomplete\nrules:\n  rule_1:",
	})
	if !strings.Contains(result, "Collection may be incomplete") {
		t.Fatal("expected incomplete collection caution")
	}
}

func TestFirewallResult_JSONRoundTrip(t *testing.T) {
	original := FirewallResult{
		Summary: "nftables active, 1 warning",
		Findings: []FirewallFinding{{
			Severity: "warning", Rule: "rule_1", Port: "22",
			Title: "SSH exposed", Detail: "Port 22 open to any source.",
		}},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded FirewallResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Findings[0].Port != "22" {
		t.Fatalf("got %q", decoded.Findings[0].Port)
	}
}
