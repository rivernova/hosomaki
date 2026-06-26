// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/prompt"
)

func TestFirewallCmd_Registration(t *testing.T) {
	if rootCmd.Commands() == nil {
		t.Fatal("root command has no subcommands")
	}
	var found bool
	for _, c := range rootCmd.Commands() {
		if c.Name() == "firewall" {
			found = true
		}
	}
	if !found {
		t.Fatal("firewall command not registered")
	}
}

func TestFirewallCmd_Flags(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"firewall"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Flags().Lookup("cross-check") == nil {
		t.Fatal("expected --cross-check flag")
	}
	if cmd.Flags().Lookup("debug") == nil {
		t.Fatal("expected --debug flag")
	}
}

func TestValidateFirewallResult_EmptySummary(t *testing.T) {
	if errs := validateFirewallResult(prompt.FirewallResult{}); len(errs) == 0 {
		t.Fatal("expected summary error")
	}
}

func TestValidateFirewallFinding_InvalidSeverity(t *testing.T) {
	errs := validateFirewallFinding(prompt.FirewallFinding{
		Severity: "high", Title: "t", Detail: "d",
	})
	if len(errs) == 0 {
		t.Fatal("expected severity error")
	}
}

func TestValidateFirewallFinding_Valid(t *testing.T) {
	errs := validateFirewallFinding(prompt.FirewallFinding{
		Severity: "warning", Title: "SSH open", Detail: "Port 22 is exposed.",
	})
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestFirewallCmd_RejectsArgs(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"firewall"})
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Args(cmd, []string{"extra"}); err == nil {
		t.Fatal("expected args error")
	}
}

func TestFirewallDetectBackend(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"firewall"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cmd.Long, "Read-only") {
		t.Fatal("expected read-only in long help")
	}
}
