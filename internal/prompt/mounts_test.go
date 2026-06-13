// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit tests for the mounts prompt builder

func makeMountsInput(mounts string) MountsInput {
	return MountsInput{
		Environment: collector.Environment{
			DistroID:   "ubuntu",
			InitSystem: "systemd",
		},
		Mounts: mounts,
	}
}

func TestMounts_ContainsSchema(t *testing.T) {
	p := Mounts(makeMountsInput("device: /dev/sda1"))
	if !strings.Contains(p, SchemaMounts) {
		t.Error("Mounts() prompt must contain the schema constant")
	}
}

func TestMounts_ContainsEnvironmentSection(t *testing.T) {
	p := Mounts(makeMountsInput("device: /dev/sda1"))
	if !strings.Contains(p, "Host environment") {
		t.Error("Mounts() prompt must contain the environment section")
	}
}

func TestMounts_ContainsMountData(t *testing.T) {
	data := "device:      fileserver:/export/data\nmountpoint:  /mnt/data\nfstype:      nfs4"
	p := Mounts(makeMountsInput(data))
	if !strings.Contains(p, "fileserver:/export/data") {
		t.Errorf("Mounts() prompt must embed the mount list verbatim")
	}
}

func TestMounts_InstructsPureJSON(t *testing.T) {
	p := Mounts(makeMountsInput("device: /dev/sda1"))
	if !strings.Contains(p, "Return ONLY a JSON object") {
		t.Error("Mounts() prompt must instruct the model to return only JSON")
	}
}

func TestMounts_InstructsNoMarkdown(t *testing.T) {
	p := Mounts(makeMountsInput("device: /dev/sda1"))
	if !strings.Contains(p, "No markdown") {
		t.Error("Mounts() prompt must forbid markdown output")
	}
}

func TestMounts_DefinesNFSStaleCritical(t *testing.T) {
	p := Mounts(makeMountsInput("nfs_status: STALE"))
	if !strings.Contains(p, "STALE") {
		t.Error("Mounts() prompt must define NFS STALE as a critical finding")
	}
	if !strings.Contains(p, "critical") {
		t.Error("Mounts() prompt must use the word 'critical' for NFS STALE")
	}
}

func TestMounts_DefinesDiskUsageThresholds(t *testing.T) {
	p := Mounts(makeMountsInput("used: 90%"))
	if !strings.Contains(p, "85") {
		t.Error("Mounts() prompt must specify the 85%% warning threshold")
	}
	if !strings.Contains(p, "95") {
		t.Error("Mounts() prompt must specify the 95%% critical threshold")
	}
}

func TestMounts_InstructsEmptyFindingsWhenHealthy(t *testing.T) {
	p := Mounts(makeMountsInput("device: /dev/sda1"))
	if !strings.Contains(p, "empty findings array") {
		t.Error("Mounts() prompt must tell the model to return empty findings when nothing is wrong")
	}
}

func TestMounts_InstructsNoPseudoFSFindings(t *testing.T) {
	p := Mounts(makeMountsInput("device: tmpfs"))
	if !strings.Contains(p, "Pseudo-filesystems") || !strings.Contains(p, "Do NOT flag") {
		t.Error("Mounts() prompt must explicitly exempt pseudo-filesystems from findings")
	}
}

func TestMounts_InstructsNoInventedMountPoints(t *testing.T) {
	p := Mounts(makeMountsInput("device: /dev/sda1"))
	if !strings.Contains(p, "Do not invent mount points") {
		t.Error("Mounts() prompt must instruct the model not to invent mount points")
	}
}

func TestMounts_SeverityFieldReferenced(t *testing.T) {
	p := Mounts(makeMountsInput("device: /dev/sda1"))
	if !strings.Contains(p, `"severity"`) {
		t.Error("Mounts() prompt must reference the 'severity' JSON field")
	}
}

func TestMounts_MountPointFieldReferenced(t *testing.T) {
	p := Mounts(makeMountsInput("device: /dev/sda1"))
	if !strings.Contains(p, `"mount_point"`) {
		t.Error("Mounts() prompt must reference the 'mount_point' JSON field")
	}
}

func TestMounts_AllSeverityValuesReferenced(t *testing.T) {
	p := Mounts(makeMountsInput("device: /dev/sda1"))
	for _, sev := range []string{"critical", "warning", "info"} {
		if !strings.Contains(p, `"`+sev+`"`) {
			t.Errorf("Mounts() prompt must define severity value %q", sev)
		}
	}
}

func TestSchemaMounts_ContainsSummary(t *testing.T) {
	if !strings.Contains(SchemaMounts, "summary") {
		t.Error("SchemaMounts must contain 'summary' field")
	}
}

func TestSchemaMounts_ContainsFindings(t *testing.T) {
	if !strings.Contains(SchemaMounts, "findings") {
		t.Error("SchemaMounts must contain 'findings' field")
	}
}

func TestSchemaMounts_ContainsSeverity(t *testing.T) {
	if !strings.Contains(SchemaMounts, "severity") {
		t.Error("SchemaMounts must contain 'severity' field")
	}
}

func TestSchemaMounts_ContainsMountPoint(t *testing.T) {
	if !strings.Contains(SchemaMounts, `"mount_point"`) {
		t.Error("SchemaMounts must contain 'mount_point' field")
	}
}
