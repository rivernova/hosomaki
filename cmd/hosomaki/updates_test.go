// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"testing"
)

func TestUpdatesCmdRegistered(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "updates" {
			found = true
			break
		}
	}
	if !found {
		t.Error("updates command is not registered on rootCmd")
	}
}

func TestUpdatesCmd_HasSecurityOnlyFlag(t *testing.T) {
	cmd := newUpdatesCmd()
	f := cmd.Flags().Lookup("security-only")
	if f == nil {
		t.Fatal("updates command is missing the --security-only flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--security-only default should be false, got %q", f.DefValue)
	}
}

func TestUpdatesCmd_HasDebugFlag(t *testing.T) {
	cmd := newUpdatesCmd()
	f := cmd.Flags().Lookup("debug")
	if f == nil {
		t.Fatal("updates command is missing the --debug flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--debug default should be false, got %q", f.DefValue)
	}
}

func TestUpdatesCmd_RejectsArgs(t *testing.T) {
	cmd := newUpdatesCmd()
	cmd.SetArgs([]string{"extra-arg"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for positional args, got nil")
	}
}

func TestUpdatesCmd_DescriptionsNonEmpty(t *testing.T) {
	cmd := newUpdatesCmd()
	if cmd.Short == "" {
		t.Error("updates command Short description is empty")
	}
	if cmd.Long == "" {
		t.Error("updates command Long description is empty")
	}
	if cmd.Use != "updates" {
		t.Errorf("updates command Use should be 'updates', got %q", cmd.Use)
	}
}
