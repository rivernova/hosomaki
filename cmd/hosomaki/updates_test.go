// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/prompt"
)

// unit tests for the updates command

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

func TestValidateUpdatesResult_EmptySummaryFails(t *testing.T) {
	r := prompt.UpdatesResult{Summary: "", Updates: nil}
	errs := validateUpdatesResult(r)
	if !containsSubstring(errs, "summary must not be empty") {
		t.Errorf("validateUpdatesResult() = %v, want an error about empty summary", errs)
	}
}

func TestValidateUpdatesResult_EmptyPackageFails(t *testing.T) {
	r := prompt.UpdatesResult{
		Summary: "one update pending",
		Updates: []prompt.UpdateFinding{{Package: "", Category: "minor"}},
	}
	errs := validateUpdatesResult(r)
	if !containsSubstring(errs, "package must not be empty") {
		t.Errorf("validateUpdatesResult() = %v, want an error about empty package", errs)
	}
}

func TestValidateUpdatesResult_InvalidCategoryFails(t *testing.T) {
	r := prompt.UpdatesResult{
		Summary: "one update pending",
		Updates: []prompt.UpdateFinding{{Package: "nginx", Category: "critical"}},
	}
	errs := validateUpdatesResult(r)
	if !containsSubstring(errs, "category must be") {
		t.Errorf("validateUpdatesResult() = %v, want an error about invalid category", errs)
	}
}

func TestValidateUpdatesResult_SecurityWithoutDetailFails(t *testing.T) {
	r := prompt.UpdatesResult{
		Summary: "one security update pending",
		Updates: []prompt.UpdateFinding{{Package: "openssl", Category: "security", Detail: ""}},
	}
	errs := validateUpdatesResult(r)
	if !containsSubstring(errs, "detail must not be empty") {
		t.Errorf("validateUpdatesResult() = %v, want an error requiring detail for a security finding", errs)
	}
}

func TestValidateUpdatesResult_MajorWithoutDetailFails(t *testing.T) {
	r := prompt.UpdatesResult{
		Summary: "one major update pending",
		Updates: []prompt.UpdateFinding{{Package: "docker-ce", Category: "major", Detail: ""}},
	}
	errs := validateUpdatesResult(r)
	if !containsSubstring(errs, "detail must not be empty") {
		t.Errorf("validateUpdatesResult() = %v, want an error requiring detail for a major finding", errs)
	}
}

func TestValidateUpdatesResult_MinorWithoutDetailIsFine(t *testing.T) {
	r := prompt.UpdatesResult{
		Summary: "one minor update pending",
		Updates: []prompt.UpdateFinding{{Package: "libuuid", Category: "minor", Detail: ""}},
	}
	errs := validateUpdatesResult(r)
	if len(errs) != 0 {
		t.Errorf("validateUpdatesResult() = %v, want no errors for a minor finding with empty detail", errs)
	}
}

func TestValidateUpdatesResult_UnknownWithoutDetailIsFine(t *testing.T) {
	r := prompt.UpdatesResult{
		Summary: "one update pending",
		Updates: []prompt.UpdateFinding{{Package: "mystery-pkg", Category: "unknown", Detail: ""}},
	}
	errs := validateUpdatesResult(r)
	if len(errs) != 0 {
		t.Errorf("validateUpdatesResult() = %v, want no errors for an unknown finding with empty detail", errs)
	}
}

func TestValidateUpdatesResult_SecurityWithDetailPasses(t *testing.T) {
	r := prompt.UpdatesResult{
		Summary: "one security update pending",
		Updates: []prompt.UpdateFinding{{Package: "openssl", Category: "security", Detail: "Fixes a buffer overflow."}},
	}
	errs := validateUpdatesResult(r)
	if len(errs) != 0 {
		t.Errorf("validateUpdatesResult() = %v, want no errors when detail is present", errs)
	}
}

func containsSubstring(errs []string, substr string) bool {
	for _, e := range errs {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}
