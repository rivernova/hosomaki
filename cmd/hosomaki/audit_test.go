// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/rivernova/hosomaki/internal/auditor"
	"github.com/rivernova/hosomaki/internal/sanitiser"
)

// unit tests for the audit command

func TestAuditCmdRegistered(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "audit" {
			found = true
			break
		}
	}
	if !found {
		t.Error("audit command is not registered on the root command")
	}
}

func TestAuditCmdHasInitFlag(t *testing.T) {
	cmd := newAuditCmd()
	f := cmd.Flags().Lookup("init")
	if f == nil {
		t.Fatal("audit command is missing the --init flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--init default = %q, want 'false'", f.DefValue)
	}
}

func TestAuditCmdHasBaselineFlag(t *testing.T) {
	cmd := newAuditCmd()
	f := cmd.Flags().Lookup("baseline")
	if f == nil {
		t.Fatal("audit command is missing the --baseline flag")
	}
	if f.DefValue != "" {
		t.Errorf("--baseline default = %q, want empty string", f.DefValue)
	}
}

func TestAuditCmdHasDirsFlag(t *testing.T) {
	cmd := newAuditCmd()
	f := cmd.Flags().Lookup("dirs")
	if f == nil {
		t.Fatal("audit command is missing the --dirs flag")
	}
	if f.DefValue != "" {
		t.Errorf("--dirs default = %q, want empty string", f.DefValue)
	}
}

func TestAuditCmdHasDebugFlag(t *testing.T) {
	cmd := newAuditCmd()
	f := cmd.Flags().Lookup("debug")
	if f == nil {
		t.Fatal("audit command is missing the --debug flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--debug default = %q, want 'false'", f.DefValue)
	}
}

func TestAuditCmdHasNoWatchFlag(t *testing.T) {
	cmd := newAuditCmd()
	if f := cmd.Flags().Lookup("watch"); f != nil {
		t.Error("audit command must not register --watch; use --dirs instead")
	}
}

func TestAuditCmdRejectsPositionalArgs(t *testing.T) {
	cmd := newAuditCmd()
	if err := cmd.Args(cmd, []string{"unexpected"}); err == nil {
		t.Error("audit command should reject positional arguments")
	}
}

func TestAuditCmdShortDescription(t *testing.T) {
	if newAuditCmd().Short == "" {
		t.Error("audit command must have a non-empty Short description")
	}
}

func TestAuditCmdLongContainsKeyPhrases(t *testing.T) {
	long := newAuditCmd().Long
	for _, phrase := range []string{"baseline", "--init", "never modifies", "read-only"} {
		if !strings.Contains(long, phrase) {
			t.Errorf("audit Long help text is missing expected phrase %q", phrase)
		}
	}
}

func TestAuditCmdLongMentionsStoragePath(t *testing.T) {
	if !strings.Contains(newAuditCmd().Long, ".local/share/hosomaki") {
		t.Error("audit Long help text should mention the default baseline path")
	}
}

func TestAuditCmdHelp_DoesNotPanic(t *testing.T) {
	cmd := newAuditCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("help panicked: %v", r)
		}
	}()
	_ = cmd.Help()
}

func TestParseDirs_SingleDir(t *testing.T) {
	got := parseDirs("/etc")
	if len(got) != 1 || got[0] != "/etc" {
		t.Errorf("parseDirs('/etc') = %v, want [/etc]", got)
	}
}

func TestParseDirs_MultipleDirs(t *testing.T) {
	got := parseDirs("/etc,/usr/local/bin, /usr/local/sbin")
	if len(got) != 3 {
		t.Fatalf("parseDirs() len = %d, want 3", len(got))
	}
	if got[2] != "/usr/local/sbin" {
		t.Errorf("got[2] = %q, want /usr/local/sbin", got[2])
	}
}

func TestParseDirs_EmptyReturnsNil(t *testing.T) {
	if got := parseDirs(""); got != nil {
		t.Errorf("parseDirs('') = %v, want nil", got)
	}
}

func TestParseDirs_WhitespaceOnlyReturnsNil(t *testing.T) {
	if got := parseDirs("   "); got != nil {
		t.Errorf("parseDirs('   ') = %v, want nil", got)
	}
}

func TestParseDirs_SkipsEmptySegments(t *testing.T) {
	got := parseDirs("/etc,,/tmp,")
	if len(got) != 2 {
		t.Errorf("parseDirs should skip empty segments, got %v", got)
	}
}

func TestResolveBaselinePath_FlagValueTakesPrecedence(t *testing.T) {
	got, err := resolveBaselinePath("/custom/path/baseline.json")
	if err != nil {
		t.Fatalf("resolveBaselinePath() error: %v", err)
	}
	if got != "/custom/path/baseline.json" {
		t.Errorf("resolveBaselinePath() = %q, want /custom/path/baseline.json", got)
	}
}

func TestResolveBaselinePath_EmptyFlagUsesDefault(t *testing.T) {
	got, err := resolveBaselinePath("")
	if err != nil {
		t.Fatalf("resolveBaselinePath('') error: %v", err)
	}
	if got == "" {
		t.Error("resolveBaselinePath('') returned empty string")
	}
	if !strings.Contains(got, "audit-baseline.json") {
		t.Errorf("default path should contain audit-baseline.json, got %q", got)
	}
}

func TestHumanDuration_DaysAndHours(t *testing.T) {
	if got := humanDuration(2*24*time.Hour + 4*time.Hour); got != "2d 4h" {
		t.Errorf("humanDuration(2d4h) = %q, want '2d 4h'", got)
	}
}

func TestHumanDuration_HoursAndMinutes(t *testing.T) {
	if got := humanDuration(3*time.Hour + 12*time.Minute); got != "3h 12m" {
		t.Errorf("humanDuration(3h12m) = %q, want '3h 12m'", got)
	}
}

func TestHumanDuration_MinutesOnly(t *testing.T) {
	if got := humanDuration(45 * time.Minute); got != "45m" {
		t.Errorf("humanDuration(45m) = %q, want '45m'", got)
	}
}

func TestHumanDuration_DaysNoMinutes(t *testing.T) {
	if got := humanDuration(1*24*time.Hour + 30*time.Minute); got != "1d" {
		t.Errorf("humanDuration(1d30m) = %q, want '1d'", got)
	}
}

func TestHumanDuration_Zero(t *testing.T) {
	if got := humanDuration(0); got != "just now" {
		t.Errorf("humanDuration(0) = %q, want 'just now'", got)
	}
}

func TestHumanDuration_SubMinute(t *testing.T) {
	if got := humanDuration(30 * time.Second); got != "just now" {
		t.Errorf("humanDuration(30s) = %q, want 'just now'", got)
	}
}

func TestHumanDuration_ExactlyOneHour(t *testing.T) {
	if got := humanDuration(time.Hour); got != "1h" {
		t.Errorf("humanDuration(1h) = %q, want '1h'", got)
	}
}

func makeDiff() *auditor.AuditDiff {
	return &auditor.AuditDiff{
		BaselineAge:     24 * time.Hour,
		ServicesAdded:   []string{"redis.service"},
		ServicesRemoved: []string{"old.service"},
		FilesAdded:      []string{"/etc/cron.d/job"},
		FilesRemoved:    []string{"/etc/ssh/sshd_config.bak"},
		FilesModified: []auditor.FileChange{
			{Path: "/etc/sudoers", OldMtime: 1000, NewMtime: 2000, OldSize: 512, NewSize: 600},
		},
		PermissionsChanged: []auditor.PermChange{
			{Path: "/usr/local/bin/tool", OldMode: "0755", NewMode: "4755", OldOwner: "root", NewOwner: "alice", OldGroup: "root", NewGroup: "staff"},
		},
		PackagesAdded:   []string{"vim 9.0.0"},
		PackagesRemoved: []string{"nano 6.0"},
		PackagesUpdated: []auditor.PackageChange{
			{Name: "openssl", OldVersion: "3.0.0", NewVersion: "3.0.1"},
		},
		PortsOpened:  []string{"tcp 10.0.0.1:4444"},
		PortsClosed:  []string{"tcp 0.0.0.0:8080"},
		UsersAdded:   []string{"deploy"},
		UsersRemoved: []string{"olduser"},
	}
}

func TestSanitiseDiff_PreservesBaselineAge(t *testing.T) {
	d := makeDiff()
	out := sanitiseDiff(d, sanitiser.Default())
	if out.BaselineAge != d.BaselineAge {
		t.Errorf("BaselineAge = %v, want %v", out.BaselineAge, d.BaselineAge)
	}
}

func TestSanitiseDiff_MasksIPsInPorts(t *testing.T) {
	d := makeDiff()
	out := sanitiseDiff(d, sanitiser.Default())
	for _, p := range out.PortsOpened {
		if strings.Contains(p, "10.0.0.1") {
			t.Errorf("PortsOpened contains unsanitised IP: %q", p)
		}
	}
}

func TestSanitiseDiff_MasksFilePaths(t *testing.T) {
	d := makeDiff()
	out := sanitiseDiff(d, sanitiser.Default())
	for _, f := range out.FilesAdded {
		if strings.Contains(f, "/etc/cron.d/job") {
			t.Errorf("FilesAdded contains unsanitised path: %q", f)
		}
	}
	for _, fc := range out.FilesModified {
		if strings.Contains(fc.Path, "/etc/sudoers") {
			t.Errorf("FilesModified.Path contains unsanitised path: %q", fc.Path)
		}
	}
}

func TestSanitiseDiff_MasksPermPaths(t *testing.T) {
	d := makeDiff()
	out := sanitiseDiff(d, sanitiser.Default())
	for _, pc := range out.PermissionsChanged {
		if strings.Contains(pc.Path, "/usr/local/bin/tool") {
			t.Errorf("PermissionsChanged.Path contains unsanitised path: %q", pc.Path)
		}
	}
}

func TestSanitiseDiff_PreservesPermissionModes(t *testing.T) {
	d := makeDiff()
	out := sanitiseDiff(d, sanitiser.Default())
	if len(out.PermissionsChanged) == 0 {
		t.Fatal("expected PermissionsChanged to be non-empty")
	}
	pc := out.PermissionsChanged[0]
	if pc.OldMode != "0755" {
		t.Errorf("OldMode = %q, want '0755'", pc.OldMode)
	}
	if pc.NewMode != "4755" {
		t.Errorf("NewMode = %q, want '4755'", pc.NewMode)
	}
}

func TestSanitiseDiff_PreservesFileSizesAndMtimes(t *testing.T) {
	d := makeDiff()
	out := sanitiseDiff(d, sanitiser.Default())
	if len(out.FilesModified) == 0 {
		t.Fatal("expected FilesModified to be non-empty")
	}
	fc := out.FilesModified[0]
	if fc.OldSize != 512 || fc.NewSize != 600 {
		t.Errorf("sizes = %d/%d, want 512/600", fc.OldSize, fc.NewSize)
	}
	if fc.OldMtime != 1000 || fc.NewMtime != 2000 {
		t.Errorf("mtimes = %d/%d, want 1000/2000", fc.OldMtime, fc.NewMtime)
	}
}

func TestSanitiseDiff_DoesNotMutateInput(t *testing.T) {
	d := makeDiff()
	origPath := d.FilesModified[0].Path
	_ = sanitiseDiff(d, sanitiser.Default())
	if d.FilesModified[0].Path != origPath {
		t.Error("sanitiseDiff must not mutate the input diff")
	}
}

func TestSanitiseDiff_EmptySlicesRemainEmpty(t *testing.T) {
	d := &auditor.AuditDiff{
		BaselineAge:     time.Hour,
		ServicesAdded:   []string{},
		FilesModified:   []auditor.FileChange{},
		PackagesUpdated: []auditor.PackageChange{},
	}
	out := sanitiseDiff(d, sanitiser.Default())
	if len(out.ServicesAdded) != 0 {
		t.Errorf("ServicesAdded should remain empty, got %v", out.ServicesAdded)
	}
	if len(out.FilesModified) != 0 {
		t.Errorf("FilesModified should remain empty, got %v", out.FilesModified)
	}
}
