// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//  when a collector's required tool is not installed, it must
// report that it could not read rather than silently pretending nothing was found

func emptyPATH(t *testing.T) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
}

func TestSnapshot_MissingBinariesAreReported(t *testing.T) {
	emptyPATH(t)
	snap, err := Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() error = %v, want nil with collection errors recorded", err)
	}
	if len(snap.Errors) == 0 {
		t.Fatal("Snapshot() recorded no collection errors when every tool was missing")
	}
}

func TestTimers_MissingSystemctlWarns(t *testing.T) {
	emptyPATH(t)
	entries, warn := Timers()
	if warn == "" {
		t.Fatal("Timers() returned no warning when systemctl was missing")
	}
	if len(entries) != 0 {
		t.Fatalf("Timers() returned %d entries when systemctl was missing", len(entries))
	}
}

func TestPorts_MissingSsWarns(t *testing.T) {
	emptyPATH(t)
	entries, warnings := Ports()
	if len(warnings) == 0 {
		t.Fatal("Ports() returned no warnings when ss was missing")
	}
	if len(entries) != 0 {
		t.Fatalf("Ports() returned %d entries when ss was missing", len(entries))
	}
}

func TestReadUserCrontabs_MissingCrontabWarns(t *testing.T) {
	emptyPATH(t)
	jobs, warnings := readUserCrontabs([]string{"root"})
	if len(jobs) != 0 {
		t.Fatalf("readUserCrontabs() returned %d jobs when crontab was missing", len(jobs))
	}
	if len(warnings) == 0 {
		t.Fatal("readUserCrontabs() silently skipped users when crontab was missing")
	}
	if !strings.Contains(warnings[0], binCrontab) {
		t.Fatalf("warning should name the missing tool %q, got: %q", binCrontab, warnings[0])
	}
}

func TestServiceLogs_MissingJournalctlIsDistinct(t *testing.T) {
	emptyPATH(t)
	_, err := ServiceLogs("nginx", LogOptions{})
	if err == nil {
		t.Fatal("ServiceLogs() returned no error when journalctl was missing")
	}
	if strings.Contains(err.Error(), "service name correct") {
		t.Fatalf("missing journalctl was misreported as a wrong service name: %v", err)
	}
	if !strings.Contains(err.Error(), binJournalctl) {
		t.Fatalf("error should name the missing tool %q, got: %v", binJournalctl, err)
	}
}

func TestBootLogs_MissingJournalctlIsDistinct(t *testing.T) {
	emptyPATH(t)
	_, err := BootLogs(0, LogOptions{})
	if err == nil {
		t.Fatal("BootLogs() returned no error when journalctl was missing")
	}
	if strings.Contains(err.Error(), "boot index may be out of range") {
		t.Fatalf("missing journalctl was misreported as a boot range error: %v", err)
	}
	if !strings.Contains(err.Error(), binJournalctl) {
		t.Fatalf("error should name the missing tool %q, got: %v", binJournalctl, err)
	}
}

func TestFileLogs_MissingTailIsDistinct(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	if err := os.WriteFile(path, []byte("line one\nline two\n"), 0o600); err != nil {
		t.Fatalf("write temp log: %v", err)
	}
	emptyPATH(t)
	_, err := FileLogs(path, LogOptions{})
	if err == nil {
		t.Fatal("FileLogs() returned no error when tail was missing")
	}
	if strings.Contains(err.Error(), "empty or unreadable") {
		t.Fatalf("missing tail was misreported as an empty file: %v", err)
	}
	if !strings.Contains(err.Error(), binTail) {
		t.Fatalf("error should name the missing tool %q, got: %v", binTail, err)
	}
}

func TestWhyLogs_MissingJournalctlIsDistinct(t *testing.T) {
	emptyPATH(t)
	_, err := WhyLogs("nginx", LogOptions{})
	if err == nil {
		t.Fatal("WhyLogs() returned no error when journalctl was missing")
	}
	if strings.Contains(err.Error(), "service name correct") {
		t.Fatalf("missing journalctl was misreported as a wrong service name: %v", err)
	}
	if !strings.Contains(err.Error(), "could not collect") {
		t.Fatalf("error should report a collection failure, got: %v", err)
	}
}

func TestFirewallRules_MissingAllBackendsReportsNone(t *testing.T) {
	emptyPATH(t)
	result := FirewallRules()
	if result.ReadStatus != ReadNone {
		t.Fatalf("FirewallRules() ReadStatus = %q, want %q when no backend is available", result.ReadStatus, ReadNone)
	}
	if result.Warning == "" {
		t.Fatal("FirewallRules() returned no warning when no firewall backend was available")
	}
}

func TestEnrichWithDf_MissingDfWarns(t *testing.T) {
	emptyPATH(t)
	warn := enrichWithDf([]MountEntry{{MountPoint: "/", FSType: "ext4"}})
	if warn == "" {
		t.Fatal("enrichWithDf() returned no warning when df was missing")
	}
	if !strings.Contains(warn, binDf) {
		t.Fatalf("warning should name the missing tool %q, got: %q", binDf, warn)
	}
}
