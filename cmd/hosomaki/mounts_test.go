// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/prompt"
)

// unit tests for the mounts command

func TestMountsCmd_HasDebugFlag(t *testing.T) {
	cmd := newMountsCmd()
	f := cmd.Flags().Lookup("debug")
	if f == nil {
		t.Fatal("mounts command is missing the --debug flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--debug default = %q, want %q", f.DefValue, "false")
	}
}

func TestMountsCmd_NoArgs(t *testing.T) {
	cmd := newMountsCmd()
	if cmd.Args == nil {
		t.Fatal("mounts command must define an Args validator")
	}
	if err := cmd.Args(cmd, []string{"unexpected"}); err == nil {
		t.Error("mounts command must reject positional arguments")
	}
}

func TestMountsCmd_AcceptsNoPositionalArgs(t *testing.T) {
	cmd := newMountsCmd()
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("mounts command must accept zero positional arguments, got error: %v", err)
	}
}

func TestValidateMountsResult_ValidResult(t *testing.T) {
	r := prompt.MountsResult{
		Summary: "Three filesystems mounted; one approaching capacity.",
		Findings: []prompt.MountFinding{
			{
				Severity:   "warning",
				MountPoint: "/var",
				Title:      "Disk usage approaching capacity",
				Detail:     "/var is at 87% capacity and may fill up soon.",
			},
		},
	}
	errs := validateMountsResult(r)
	if len(errs) != 0 {
		t.Errorf("validateMountsResult() returned unexpected errors: %v", errs)
	}
}

func TestValidateMountsResult_EmptySummary(t *testing.T) {
	r := prompt.MountsResult{Summary: "", Findings: nil}
	errs := validateMountsResult(r)
	if len(errs) == 0 {
		t.Error("validateMountsResult() must reject empty summary")
	}
	if !strings.Contains(strings.Join(errs, " "), "summary") {
		t.Errorf("error must mention 'summary', got: %v", errs)
	}
}

func TestValidateMountsResult_EmptyFindingsAccepted(t *testing.T) {
	r := prompt.MountsResult{Summary: "All mount points are healthy.", Findings: nil}
	errs := validateMountsResult(r)
	if len(errs) != 0 {
		t.Errorf("validateMountsResult() must accept empty findings, got: %v", errs)
	}
}

func TestValidateMountsResult_EmptySeverityRejected(t *testing.T) {
	r := prompt.MountsResult{
		Summary: "One issue found.",
		Findings: []prompt.MountFinding{
			{Severity: "", MountPoint: "/var", Title: "Something", Detail: "Some detail."},
		},
	}
	errs := validateMountsResult(r)
	if len(errs) == 0 {
		t.Error("validateMountsResult() must reject empty severity")
	}
	joined := strings.Join(errs, " ")
	if !strings.Contains(joined, "severity") {
		t.Errorf("error must mention 'severity', got: %v", errs)
	}
}

func TestValidateMountsResult_InvalidSeverityRejected(t *testing.T) {
	r := prompt.MountsResult{
		Summary: "One issue found.",
		Findings: []prompt.MountFinding{
			{Severity: "medium", MountPoint: "/var", Title: "Something", Detail: "Some detail."},
		},
	}
	errs := validateMountsResult(r)
	if len(errs) == 0 {
		t.Error("validateMountsResult() must reject severity 'medium'")
	}
}

func TestValidateMountsResult_AllSeverityValuesAccepted(t *testing.T) {
	for _, sev := range []string{"critical", "warning", "info"} {
		r := prompt.MountsResult{
			Summary: "Found something.",
			Findings: []prompt.MountFinding{
				{Severity: sev, MountPoint: "/mnt/x", Title: "T", Detail: "D"},
			},
		}
		errs := validateMountsResult(r)
		if len(errs) != 0 {
			t.Errorf("validateMountsResult() must accept severity %q, got: %v", sev, errs)
		}
	}
}

func TestValidateMountsResult_EmptyMountPointRejected(t *testing.T) {
	r := prompt.MountsResult{
		Summary: "One issue found.",
		Findings: []prompt.MountFinding{
			{Severity: "warning", MountPoint: "", Title: "Something", Detail: "Some detail."},
		},
	}
	errs := validateMountsResult(r)
	if len(errs) == 0 {
		t.Error("validateMountsResult() must reject empty mount_point")
	}
}

func TestValidateMountsResult_EmptyTitleRejected(t *testing.T) {
	r := prompt.MountsResult{
		Summary: "One issue found.",
		Findings: []prompt.MountFinding{
			{Severity: "warning", MountPoint: "/var", Title: "", Detail: "Some detail."},
		},
	}
	errs := validateMountsResult(r)
	if len(errs) == 0 {
		t.Error("validateMountsResult() must reject empty title")
	}
}

func TestValidateMountsResult_EmptyDetailRejected(t *testing.T) {
	r := prompt.MountsResult{
		Summary: "One issue found.",
		Findings: []prompt.MountFinding{
			{Severity: "warning", MountPoint: "/var", Title: "Something", Detail: ""},
		},
	}
	errs := validateMountsResult(r)
	if len(errs) == 0 {
		t.Error("validateMountsResult() must reject empty detail")
	}
}

func TestValidateMountsResult_MultipleErrorsPerEntry(t *testing.T) {
	r := prompt.MountsResult{
		Summary: "Issues.",
		Findings: []prompt.MountFinding{
			{Severity: "", MountPoint: "", Title: "", Detail: ""},
		},
	}
	errs := validateMountsResult(r)
	if len(errs) < 3 {
		t.Errorf("validateMountsResult() must produce one error per invalid field, got: %v", errs)
	}
}

func TestCountMountKinds_CountsRealAndNFS(t *testing.T) {
	entries := []collector.MountEntry{
		{FSType: "ext4", NFSStale: false},  // real
		{FSType: "ext4", NFSStale: false},  // real
		{FSType: "nfs4", NFSStale: true},   // real + NFS + stale
		{FSType: "tmpfs", NFSStale: false}, // pseudo
		{FSType: "proc", NFSStale: false},  // pseudo
	}

	mountKinds, nfs, stale := countMountKinds(entries)

	if mountKinds != 3 {
		t.Errorf("real = %d, want 3 (ext4×2 + nfs4×1)", mountKinds)
	}
	if nfs != 1 {
		t.Errorf("nfs = %d, want 1", nfs)
	}
	if stale != 1 {
		t.Errorf("staleNFS = %d, want 1", stale)
	}
}

func TestCountMountKinds_AllPseudo(t *testing.T) {
	entries := []collector.MountEntry{
		{FSType: "tmpfs"},
		{FSType: "proc"},
		{FSType: "sysfs"},
	}
	mountKinds, nfs, stale := countMountKinds(entries)
	if mountKinds != 0 || nfs != 0 || stale != 0 {
		t.Errorf("all-pseudo entries: real=%d nfs=%d stale=%d, want 0 0 0", mountKinds, nfs, stale)
	}
}

func TestCountMountKinds_NoStale(t *testing.T) {
	entries := []collector.MountEntry{
		{FSType: "nfs4", NFSStale: false},
		{FSType: "nfs", NFSStale: false},
	}
	_, nfs, stale := countMountKinds(entries)
	if nfs != 2 {
		t.Errorf("nfs = %d, want 2", nfs)
	}
	if stale != 0 {
		t.Errorf("staleNFS = %d, want 0", stale)
	}
}

func TestCountMountKinds_NFS3Counted(t *testing.T) {
	entries := []collector.MountEntry{
		{FSType: "nfs3", NFSStale: false},
	}
	mountKinds, nfs, stale := countMountKinds(entries)
	if mountKinds != 1 {
		t.Errorf("real = %d, want 1", mountKinds)
	}
	if nfs != 1 {
		t.Errorf("nfs = %d, want 1 (nfs3 must count as NFS)", nfs)
	}
	if stale != 0 {
		t.Errorf("staleNFS = %d, want 0", stale)
	}
}
