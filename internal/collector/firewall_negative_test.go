// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import "testing"

// firewall parsers must stay calm on broken output

func TestFirewallParsers_NegativeInputs(t *testing.T) {
	parsers := map[string]func(string) []FirewallRule{
		"firewalld": func(s string) []FirewallRule { return parseFirewalldOutput("public", s) },
		"ufw":       parseUfwOutput,
		"nft":       parseNftOutput,
		"iptables":  parseIptablesOutput,
	}

	inputs := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"whitespace only", "   \n\t\n  "},
		{"garbage lines", "???\nnot a rule\n!!!@@@\n"},
		{"truncated columns", "tcp\nACCEPT\n22/\n/tcp\nChain\n"},
		{"partial header", "Status:\nChain \npkts bytes target\n"},
		{"single brace", "{\n}\n{ }\n"},
	}

	for backend, parse := range parsers {
		for _, tc := range inputs {
			t.Run(backend+"/"+tc.name, func(t *testing.T) {
				rules := parse(tc.in)
				if len(rules) != 0 {
					t.Fatalf("%s parser fabricated %d rule(s) from %s input: %+v", backend, len(rules), tc.name, rules)
				}
			})
		}
	}
}
