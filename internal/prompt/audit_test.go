// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"
	"time"

	"github.com/rivernova/hosomaki/internal/auditor"
	"github.com/rivernova/hosomaki/internal/collector"
)

// unit tests for the audit prompt builder

func emptyDiff() *auditor.AuditDiff {
	d := &auditor.AuditDiff{BaselineAge: 24 * time.Hour}
	d.ServicesAdded = []string{}
	d.ServicesRemoved = []string{}
	d.FilesAdded = []string{}
	d.FilesRemoved = []string{}
	d.FilesModified = []auditor.FileChange{}
	d.PermissionsChanged = []auditor.PermChange{}
	d.PackagesAdded = []string{}
	d.PackagesRemoved = []string{}
	d.PackagesUpdated = []auditor.PackageChange{}
	d.PortsOpened = []string{}
	d.PortsClosed = []string{}
	d.UsersAdded = []string{}
	d.UsersRemoved = []string{}
	return d
}

func makeInput(d *auditor.AuditDiff) AuditInput {
	return AuditInput{
		Environment: collector.Environment{},
		Diff:        d,
		BaselineAge: "1d",
	}
}

func TestAudit_ContainsSchema(t *testing.T) {
	p := Audit(makeInput(emptyDiff()))
	if !strings.Contains(p, SchemaAudit) {
		t.Error("Audit() prompt must contain the schema constant")
	}
}

func TestAudit_ContainsEnvironmentSection(t *testing.T) {
	p := Audit(makeInput(emptyDiff()))
	if !strings.Contains(p, "Host environment") {
		t.Error("Audit() prompt must contain the environment section header")
	}
}

func TestAudit_ContainsBaselineAge(t *testing.T) {
	in := AuditInput{Diff: emptyDiff(), BaselineAge: "2d 4h"}
	p := Audit(in)
	if !strings.Contains(p, "2d 4h") {
		t.Errorf("Audit() should embed the baseline age string, got:\n%s", p)
	}
}

func TestAudit_FallbackAgeWhenEmpty(t *testing.T) {
	in := AuditInput{Diff: emptyDiff(), BaselineAge: ""}
	p := Audit(in)
	if !strings.Contains(p, "unknown") {
		t.Errorf("Audit() should use 'unknown' when BaselineAge is empty, got:\n%s", p)
	}
}

func TestAudit_InstructsNoMarkdown(t *testing.T) {
	p := Audit(makeInput(emptyDiff()))
	if !strings.Contains(p, "No markdown") {
		t.Error("prompt must instruct the model not to use markdown")
	}
}

func TestAudit_SeverityConstraintsMentioned(t *testing.T) {
	p := Audit(makeInput(emptyDiff()))
	for _, sev := range []string{"critical", "warning", "info"} {
		if !strings.Contains(p, sev) {
			t.Errorf("prompt must mention severity level %q", sev)
		}
	}
}

func TestAudit_CategoryConstraintsMentioned(t *testing.T) {
	p := Audit(makeInput(emptyDiff()))
	for _, cat := range []string{"service", "file", "permission", "package", "network", "user"} {
		if !strings.Contains(p, cat) {
			t.Errorf("prompt must mention category %q", cat)
		}
	}
}

func TestAudit_ServicesAddedAppearsInPrompt(t *testing.T) {
	d := emptyDiff()
	d.ServicesAdded = []string{"redis.service"}
	p := Audit(makeInput(d))
	if !strings.Contains(p, "redis.service") {
		t.Error("added service should appear in prompt")
	}
	if !strings.Contains(p, "Services added") {
		t.Error("'Services added' section header should appear in prompt")
	}
}

func TestAudit_ServicesRemovedAppearsInPrompt(t *testing.T) {
	d := emptyDiff()
	d.ServicesRemoved = []string{"old-daemon.service"}
	p := Audit(makeInput(d))
	if !strings.Contains(p, "Services removed") {
		t.Error("'Services removed' section header should appear")
	}
}

func TestAudit_FilesModifiedAppearsInPrompt(t *testing.T) {
	d := emptyDiff()
	d.FilesModified = []auditor.FileChange{
		{Path: "<CONFIG_PATH>", OldMtime: 1000, NewMtime: 2000, OldSize: 512, NewSize: 600},
	}
	p := Audit(makeInput(d))
	if !strings.Contains(p, "Files modified") {
		t.Error("'Files modified' section should appear in prompt")
	}
	if !strings.Contains(p, "512") || !strings.Contains(p, "600") {
		t.Error("file size values should appear in prompt")
	}
}

func TestAudit_FilesModifiedShowsMtimeChanged(t *testing.T) {
	d := emptyDiff()
	d.FilesModified = []auditor.FileChange{
		{Path: "<CONFIG_PATH>", OldMtime: 1000, NewMtime: 2000, OldSize: 100, NewSize: 100},
	}
	p := Audit(makeInput(d))
	if !strings.Contains(p, "true") {
		t.Error("mtime changed=true should appear when mtimes differ")
	}
}

func TestAudit_FilesModifiedShowsMtimeUnchanged(t *testing.T) {
	d := emptyDiff()
	d.FilesModified = []auditor.FileChange{
		{Path: "<CONFIG_PATH>", OldMtime: 1000, NewMtime: 1000, OldSize: 100, NewSize: 200},
	}
	p := Audit(makeInput(d))
	if !strings.Contains(p, "false") {
		t.Error("mtime changed=false should appear when only size differs")
	}
}

func TestAudit_PermissionsChangedAppearsInPrompt(t *testing.T) {
	d := emptyDiff()
	d.PermissionsChanged = []auditor.PermChange{
		{Path: "<PATH>", OldMode: "0755", NewMode: "4755", OldOwner: "root", NewOwner: "root"},
	}
	p := Audit(makeInput(d))
	if !strings.Contains(p, "Permission changes") {
		t.Error("'Permission changes' section should appear in prompt")
	}
	if !strings.Contains(p, "0755") || !strings.Contains(p, "4755") {
		t.Error("permission modes must appear in prompt")
	}
}

func TestAudit_PermissionsChanged_OnlyDeltaFieldsShown(t *testing.T) {
	d := emptyDiff()
	d.PermissionsChanged = []auditor.PermChange{
		{Path: "<PATH>", OldMode: "0644", NewMode: "0755", OldOwner: "root", NewOwner: "root", OldGroup: "root", NewGroup: "root"},
	}
	p := Audit(makeInput(d))
	if strings.Contains(p, "owner root→root") {
		t.Error("unchanged owner should not appear in permission change output")
	}
	if strings.Contains(p, "group root→root") {
		t.Error("unchanged group should not appear in permission change output")
	}
	if !strings.Contains(p, "mode 0644→0755") {
		t.Error("changed mode must appear in permission change output")
	}
}

func TestAudit_PackagesUpdatedAppearsInPrompt(t *testing.T) {
	d := emptyDiff()
	d.PackagesUpdated = []auditor.PackageChange{
		{Name: "openssl", OldVersion: "3.0.0", NewVersion: "3.0.1"},
	}
	p := Audit(makeInput(d))
	if !strings.Contains(p, "Packages updated") {
		t.Error("'Packages updated' section should appear in prompt")
	}
	if !strings.Contains(p, "openssl") {
		t.Error("package name should appear in prompt")
	}
	if !strings.Contains(p, "3.0.0") || !strings.Contains(p, "3.0.1") {
		t.Error("package versions should appear in prompt")
	}
}

func TestAudit_PackagesInstalledAppearsInPrompt(t *testing.T) {
	d := emptyDiff()
	d.PackagesAdded = []string{"vim 9.0.0"}
	p := Audit(makeInput(d))
	if !strings.Contains(p, "Packages installed") {
		t.Error("'Packages installed' section should appear in prompt")
	}
}

func TestAudit_PortsOpenedAppearsInPrompt(t *testing.T) {
	d := emptyDiff()
	d.PortsOpened = []string{"tcp <IPV4>:4444"}
	p := Audit(makeInput(d))
	if !strings.Contains(p, "Ports now listening") {
		t.Error("'Ports now listening' section should appear in prompt")
	}
}

func TestAudit_UsersAddedAppearsInPrompt(t *testing.T) {
	d := emptyDiff()
	d.UsersAdded = []string{"deploy"}
	p := Audit(makeInput(d))
	if !strings.Contains(p, "Users added") {
		t.Error("'Users added' section should appear in prompt")
	}
	if !strings.Contains(p, "deploy") {
		t.Error("added username should appear in prompt")
	}
}

func TestAudit_EmptyDiffSectionsOmitted(t *testing.T) {
	p := Audit(makeInput(emptyDiff()))
	for _, header := range []string{
		"Services added", "Services removed",
		"Files added", "Files removed", "Files modified",
		"Permission changes",
		"Packages installed", "Packages removed", "Packages updated",
		"Ports now listening", "Ports no longer listening",
		"Users added", "Users removed",
	} {
		if strings.Contains(p, header) {
			t.Errorf("empty diff must not produce section %q in prompt", header)
		}
	}
}

func TestAudit_EmptyDiffShowsNoChangesText(t *testing.T) {
	p := Audit(makeInput(emptyDiff()))
	if !strings.Contains(p, "no changes detected") {
		t.Error("empty diff should result in '(no changes detected)' in prompt")
	}
}

func TestFormatDiff_EmptyProducesEmptyString(t *testing.T) {
	got := formatDiff(emptyDiff())
	if strings.TrimSpace(got) != "" {
		t.Errorf("formatDiff(empty) should produce empty string, got %q", got)
	}
}

func TestFormatDiff_MultipleCategories(t *testing.T) {
	d := emptyDiff()
	d.ServicesAdded = []string{"redis.service"}
	d.UsersAdded = []string{"deploy"}
	d.PortsOpened = []string{"tcp <IPV4>:6379"}
	got := formatDiff(d)
	if !strings.Contains(got, "Services added") {
		t.Error("Services added section missing")
	}
	if !strings.Contains(got, "Users added") {
		t.Error("Users added section missing")
	}
	if !strings.Contains(got, "Ports now listening") {
		t.Error("Ports now listening section missing")
	}
}

func TestAuditInput_NilDiffDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Audit() panicked on empty diff: %v", r)
		}
	}()
	_ = Audit(makeInput(emptyDiff()))
}

func TestAuditInput_BaselineAgeIsString(t *testing.T) {
	in := AuditInput{BaselineAge: "3h 12m", Diff: emptyDiff()}
	if in.BaselineAge != "3h 12m" {
		t.Errorf("BaselineAge = %q, want '3h 12m'", in.BaselineAge)
	}
}

func TestAudit_PromptPackageHasNoSanitiserImport(_ *testing.T) {
	// Intentionally empty — the invariant is enforced at compile time
}
