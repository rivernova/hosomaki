// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package auditor

import (
	"testing"
	"time"
)

// unit tests for diff computation

func makeBaseline(t *testing.T) *AuditBaseline {
	t.Helper()
	return &AuditBaseline{
		Version:     baselineVersion,
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		Services:    []string{"nginx.service", "ssh.service"},
		Files:       []FileEntry{{Path: "/etc/hosts", Mtime: 1000, Size: 200}},
		Permissions: []PermEntry{{Path: "/etc/hosts", Mode: "0644", Owner: "root", Group: "root"}},
		Packages:    []string{"curl 7.68.0", "nginx 1.18.0"},
		Ports:       []string{"tcp 0.0.0.0:22", "tcp 0.0.0.0:80"},
		Users:       []string{"alice", "bob", "root"},
	}
}

func TestDiff_IdenticalSnapshotsProduceEmptyDiff(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.CreatedAt = time.Now()

	d := Diff(old, current)
	if !d.IsEmpty() {
		t.Errorf("expected empty diff for identical snapshots, got %d changes", d.TotalChanges())
	}
}

func TestDiff_IsEmptyReturnsFalseWhenChangesExist(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Services = append(current.Services, "redis.service")

	d := Diff(old, current)
	if d.IsEmpty() {
		t.Error("IsEmpty() should return false when there are changes")
	}
}

func TestDiff_TotalChangesCountsAllCategories(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Services = append(current.Services, "redis.service")
	current.Users = append(current.Users, "charlie")

	d := Diff(old, current)
	if d.TotalChanges() != 2 {
		t.Errorf("TotalChanges() = %d, want 2", d.TotalChanges())
	}
}

func TestDiff_ServiceAdded(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Services = append(current.Services, "redis.service")

	d := Diff(old, current)
	if len(d.ServicesAdded) != 1 || d.ServicesAdded[0] != "redis.service" {
		t.Errorf("ServicesAdded = %v, want [redis.service]", d.ServicesAdded)
	}
	if len(d.ServicesRemoved) != 0 {
		t.Errorf("ServicesRemoved should be empty, got %v", d.ServicesRemoved)
	}
}

func TestDiff_ServiceRemoved(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Services = []string{"nginx.service"} // ssh.service removed

	d := Diff(old, current)
	if len(d.ServicesRemoved) != 1 || d.ServicesRemoved[0] != "ssh.service" {
		t.Errorf("ServicesRemoved = %v, want [ssh.service]", d.ServicesRemoved)
	}
}

func TestDiff_FileAdded(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Files = append(current.Files, FileEntry{Path: "/etc/nginx/nginx.conf", Mtime: 2000, Size: 512})

	d := Diff(old, current)
	if len(d.FilesAdded) != 1 || d.FilesAdded[0] != "/etc/nginx/nginx.conf" {
		t.Errorf("FilesAdded = %v, want [/etc/nginx/nginx.conf]", d.FilesAdded)
	}
}

func TestDiff_FileRemoved(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Files = []FileEntry{}

	d := Diff(old, current)
	if len(d.FilesRemoved) != 1 || d.FilesRemoved[0] != "/etc/hosts" {
		t.Errorf("FilesRemoved = %v, want [/etc/hosts]", d.FilesRemoved)
	}
}

func TestDiff_FileModifiedByMtime(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Files = []FileEntry{{Path: "/etc/hosts", Mtime: 9999, Size: 200}}

	d := Diff(old, current)
	if len(d.FilesModified) != 1 {
		t.Fatalf("FilesModified len = %d, want 1", len(d.FilesModified))
	}
	fc := d.FilesModified[0]
	if fc.Path != "/etc/hosts" {
		t.Errorf("FilesModified[0].Path = %q, want /etc/hosts", fc.Path)
	}
	if fc.OldMtime != 1000 || fc.NewMtime != 9999 {
		t.Errorf("mtime: old=%d new=%d, want 1000→9999", fc.OldMtime, fc.NewMtime)
	}
}

func TestDiff_FileModifiedBySize(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Files = []FileEntry{{Path: "/etc/hosts", Mtime: 1000, Size: 999}}

	d := Diff(old, current)
	if len(d.FilesModified) != 1 {
		t.Fatalf("FilesModified len = %d, want 1", len(d.FilesModified))
	}
	if d.FilesModified[0].OldSize != 200 || d.FilesModified[0].NewSize != 999 {
		t.Errorf("size: old=%d new=%d, want 200→999", d.FilesModified[0].OldSize, d.FilesModified[0].NewSize)
	}
}

func TestDiff_FileUnchangedNotReported(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)

	d := Diff(old, current)
	if len(d.FilesModified) != 0 {
		t.Errorf("FilesModified should be empty for unchanged file, got %v", d.FilesModified)
	}
}
func TestDiff_PermissionModeChanged(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Permissions = []PermEntry{{Path: "/etc/hosts", Mode: "0777", Owner: "root", Group: "root"}}

	d := Diff(old, current)
	if len(d.PermissionsChanged) != 1 {
		t.Fatalf("PermissionsChanged len = %d, want 1", len(d.PermissionsChanged))
	}
	pc := d.PermissionsChanged[0]
	if pc.OldMode != "0644" || pc.NewMode != "0777" {
		t.Errorf("mode: %q → %q, want 0644 → 0777", pc.OldMode, pc.NewMode)
	}
}

func TestDiff_PermissionOwnerChanged(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Permissions = []PermEntry{{Path: "/etc/hosts", Mode: "0644", Owner: "alice", Group: "root"}}

	d := Diff(old, current)
	if len(d.PermissionsChanged) != 1 {
		t.Fatalf("PermissionsChanged len = %d, want 1", len(d.PermissionsChanged))
	}
	if d.PermissionsChanged[0].NewOwner != "alice" {
		t.Errorf("NewOwner = %q, want alice", d.PermissionsChanged[0].NewOwner)
	}
}

func TestDiff_PermissionNewPathNotReported(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Permissions = append(current.Permissions,
		PermEntry{Path: "/etc/new-file", Mode: "0644", Owner: "root", Group: "root"},
	)

	d := Diff(old, current)
	if len(d.PermissionsChanged) != 0 {
		t.Errorf("new path should not appear in PermissionsChanged, got %v", d.PermissionsChanged)
	}
}

func TestDiff_PackageAdded(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Packages = append(current.Packages, "vim 9.0.0")

	d := Diff(old, current)
	if len(d.PackagesAdded) != 1 {
		t.Fatalf("PackagesAdded len = %d, want 1", len(d.PackagesAdded))
	}
	if d.PackagesAdded[0] != "vim 9.0.0" {
		t.Errorf("PackagesAdded[0] = %q, want 'vim 9.0.0'", d.PackagesAdded[0])
	}
}

func TestDiff_PackageRemoved(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Packages = []string{"nginx 1.18.0"} // curl removed

	d := Diff(old, current)
	if len(d.PackagesRemoved) != 1 {
		t.Fatalf("PackagesRemoved len = %d, want 1", len(d.PackagesRemoved))
	}
	if d.PackagesRemoved[0] != "curl 7.68.0" {
		t.Errorf("PackagesRemoved[0] = %q, want 'curl 7.68.0'", d.PackagesRemoved[0])
	}
}

func TestDiff_PackageUpdated(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Packages = []string{"curl 8.0.0", "nginx 1.18.0"} // curl upgraded

	d := Diff(old, current)
	if len(d.PackagesUpdated) != 1 {
		t.Fatalf("PackagesUpdated len = %d, want 1", len(d.PackagesUpdated))
	}
	u := d.PackagesUpdated[0]
	if u.Name != "curl" || u.OldVersion != "7.68.0" || u.NewVersion != "8.0.0" {
		t.Errorf("update: %+v, want curl 7.68.0→8.0.0", u)
	}
}

func TestDiff_PackageUpdatedNotReportedAsAddedOrRemoved(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Packages = []string{"curl 8.0.0", "nginx 1.18.0"}

	d := Diff(old, current)
	for _, a := range d.PackagesAdded {
		if a == "curl 8.0.0" || a == "curl 7.68.0" {
			t.Errorf("updated package curl should not appear in PackagesAdded, got %v", d.PackagesAdded)
		}
	}
	for _, r := range d.PackagesRemoved {
		if r == "curl 7.68.0" || r == "curl 8.0.0" {
			t.Errorf("updated package curl should not appear in PackagesRemoved, got %v", d.PackagesRemoved)
		}
	}
}
func TestDiff_PortOpened(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Ports = append(current.Ports, "tcp 0.0.0.0:443")

	d := Diff(old, current)
	if len(d.PortsOpened) != 1 || d.PortsOpened[0] != "tcp 0.0.0.0:443" {
		t.Errorf("PortsOpened = %v, want [tcp 0.0.0.0:443]", d.PortsOpened)
	}
}

func TestDiff_PortClosed(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Ports = []string{"tcp 0.0.0.0:22"} // port 80 closed

	d := Diff(old, current)
	if len(d.PortsClosed) != 1 || d.PortsClosed[0] != "tcp 0.0.0.0:80" {
		t.Errorf("PortsClosed = %v, want [tcp 0.0.0.0:80]", d.PortsClosed)
	}
}

func TestDiff_UserAdded(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Users = append(current.Users, "charlie")

	d := Diff(old, current)
	if len(d.UsersAdded) != 1 || d.UsersAdded[0] != "charlie" {
		t.Errorf("UsersAdded = %v, want [charlie]", d.UsersAdded)
	}
}

func TestDiff_UserRemoved(t *testing.T) {
	old := makeBaseline(t)
	current := makeBaseline(t)
	current.Users = []string{"alice", "root"} // bob removed

	d := Diff(old, current)
	if len(d.UsersRemoved) != 1 || d.UsersRemoved[0] != "bob" {
		t.Errorf("UsersRemoved = %v, want [bob]", d.UsersRemoved)
	}
}

func TestDiff_AllSlicesNonNilOnEmptyDiff(t *testing.T) {
	old := &AuditBaseline{Version: baselineVersion, CreatedAt: time.Now()}
	current := &AuditBaseline{Version: baselineVersion, CreatedAt: time.Now()}

	d := Diff(old, current)

	if d.ServicesAdded == nil || d.ServicesRemoved == nil {
		t.Error("Services slices must not be nil")
	}
	if d.FilesAdded == nil || d.FilesRemoved == nil || d.FilesModified == nil {
		t.Error("Files slices must not be nil")
	}
	if d.PermissionsChanged == nil {
		t.Error("PermissionsChanged must not be nil")
	}
	if d.PackagesAdded == nil || d.PackagesRemoved == nil || d.PackagesUpdated == nil {
		t.Error("Packages slices must not be nil")
	}
	if d.PortsOpened == nil || d.PortsClosed == nil {
		t.Error("Ports slices must not be nil")
	}
	if d.UsersAdded == nil || d.UsersRemoved == nil {
		t.Error("Users slices must not be nil")
	}
}

func TestParsePackageLines_NameAndVersion(t *testing.T) {
	m := parsePackageLines([]string{"curl 7.68.0", "nginx 1.18.0"})
	if m["curl"] != "7.68.0" {
		t.Errorf("curl version = %q, want 7.68.0", m["curl"])
	}
	if m["nginx"] != "1.18.0" {
		t.Errorf("nginx version = %q, want 1.18.0", m["nginx"])
	}
}

func TestParsePackageLines_VersionWithSpaces(t *testing.T) {
	m := parsePackageLines([]string{"foo 1.0 extra"})
	if m["foo"] != "1.0 extra" {
		t.Errorf("foo version = %q, want '1.0 extra'", m["foo"])
	}
}

func TestParsePackageLines_LineWithoutVersion(t *testing.T) {
	m := parsePackageLines([]string{"orphan"})
	if _, ok := m["orphan"]; !ok {
		t.Error("package with no version should still be tracked")
	}
	if m["orphan"] != "" {
		t.Errorf("orphan version = %q, want empty", m["orphan"])
	}
}

func TestParsePackageLines_EmptyLines(t *testing.T) {
	m := parsePackageLines([]string{"", "  ", "curl 7.68.0"})
	if len(m) != 1 {
		t.Errorf("expected 1 entry, got %d", len(m))
	}
}

func TestDiffStringSets_AddedAndRemoved(t *testing.T) {
	old := []string{"a", "b", "c"}
	cur := []string{"b", "c", "d"}

	added, removed := diffStringSets(old, cur)
	if len(added) != 1 || added[0] != "d" {
		t.Errorf("added = %v, want [d]", added)
	}
	if len(removed) != 1 || removed[0] != "a" {
		t.Errorf("removed = %v, want [a]", removed)
	}
}

func TestDiffStringSets_NoDifference(t *testing.T) {
	ss := []string{"a", "b", "c"}
	added, removed := diffStringSets(ss, ss)
	if len(added) != 0 || len(removed) != 0 {
		t.Errorf("expected no diff, got added=%v removed=%v", added, removed)
	}
}

func TestDiffStringSets_BothEmpty(t *testing.T) {
	added, removed := diffStringSets(nil, nil)
	if len(added) != 0 || len(removed) != 0 {
		t.Errorf("both nil: expected no diff, got added=%v removed=%v", added, removed)
	}
}
