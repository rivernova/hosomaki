// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"bytes"
	"strings"
	"testing"
)

// unit testing for doctor command setup

func TestDoctorCmdRegistered(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "doctor" {
			found = true
			break
		}
	}
	if !found {
		t.Error("doctor command is not registered on the root command")
	}
}

func TestDoctorCmdHasBriefFlag(t *testing.T) {
	cmd := newDoctorCmd()
	f := cmd.Flags().Lookup("brief")
	if f == nil {
		t.Fatal("doctor command is missing the --brief flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--brief default = %q, want %q", f.DefValue, "false")
	}
}

func TestDoctorCmdRejectsArgs(t *testing.T) {
	cmd := newDoctorCmd()
	cmd.SetArgs([]string{"unexpected-arg"})
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)

	err := cmd.Args(cmd, []string{"unexpected-arg"})
	if err == nil {
		t.Error("doctor command should reject positional arguments")
	}
}

func TestDoctorCmdHelpContainsKeyPhrases(t *testing.T) {
	cmd := newDoctorCmd()
	long := cmd.Long

	for _, phrase := range []string{
		"diagnosis",
		"suggested actions",
		"potentially disruptive",
		"never modifies the system",
	} {
		if !strings.Contains(long, phrase) {
			t.Errorf("doctor Long help text is missing expected phrase %q", phrase)
		}
	}
}

func TestDoctorCmdShortDescription(t *testing.T) {
	cmd := newDoctorCmd()
	if cmd.Short == "" {
		t.Error("doctor command must have a non-empty Short description")
	}
	if !strings.Contains(strings.ToLower(cmd.Short), "diagnosis") &&
		!strings.Contains(strings.ToLower(cmd.Short), "diagnos") {
		t.Errorf("doctor Short description should mention diagnosis, got: %q", cmd.Short)
	}
}
