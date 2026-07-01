// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/sanitiser"
)

// secrets planted in each collector's prompt-formatted
// output must be scrubbed by the sanitiser before the text could reach the LLM

func TestFormatters_NoSecretSurvivesSanitiser(t *testing.T) {
	san := sanitiser.Default()

	cases := []struct {
		name      string
		formatted string
		secrets   []string
	}{
		{
			name: "ports",
			formatted: FormatPortsForPrompt([]PortEntry{
				{Protocol: "tcp", Local: "203.0.113.45:8443", Process: "app (pid 1)"},
			}),
			secrets: []string{"203.0.113.45"},
		},
		{
			name: "crons",
			formatted: FormatCronsForPrompt([]CronJob{
				{Source: "user:alice", Schedule: "@daily", User: "alice", Command: "/home/alice/run.sh --host 10.1.2.3"},
			}),
			secrets: []string{"/home/alice/run.sh", "10.1.2.3"},
		},
		{
			name: "mounts",
			formatted: FormatMountsForPrompt([]MountEntry{
				{Device: "192.0.2.50:/export/data", MountPoint: "/home/bob/mnt", FSType: "nfs4", Options: "rw"},
			}),
			secrets: []string{"192.0.2.50", "/home/bob/mnt"},
		},
		{
			name: "firewall",
			formatted: FormatFirewallForPrompt(FirewallResult{
				Backend:    BackendIptables,
				ReadStatus: ReadOK,
				Rules: []FirewallRule{
					{Backend: BackendIptables, Action: "ACCEPT", Protocol: "tcp", Port: "22", Source: "198.51.100.7"},
				},
			}),
			secrets: []string{"198.51.100.7"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cleaned := san.Sanitise(tc.formatted)
			for _, secret := range tc.secrets {
				if strings.Contains(cleaned, secret) {
					t.Errorf("%s formatter leaked %q through the sanitiser:\n%s", tc.name, secret, cleaned)
				}
			}
		})
	}
}
