// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package auditor

import (
	"sort"
	"strings"
	"time"
)

// computes the delta between a stored baseline and a new snapshot

type FileChange struct {
	Path     string
	OldMtime int64
	NewMtime int64
	OldSize  int64
	NewSize  int64
}

type PermChange struct {
	Path     string
	OldMode  string
	NewMode  string
	OldOwner string
	NewOwner string
	OldGroup string
	NewGroup string
}

type PackageChange struct {
	Name       string
	OldVersion string
	NewVersion string
}

type AuditDiff struct {
	BaselineAge        time.Duration
	ServicesAdded      []string
	ServicesRemoved    []string
	FilesAdded         []string
	FilesRemoved       []string
	FilesModified      []FileChange
	PermissionsChanged []PermChange
	PackagesAdded      []string
	PackagesRemoved    []string
	PackagesUpdated    []PackageChange
	PortsOpened        []string
	PortsClosed        []string
	UsersAdded         []string
	UsersRemoved       []string
}

func (d *AuditDiff) IsEmpty() bool {
	return len(d.ServicesAdded) == 0 &&
		len(d.ServicesRemoved) == 0 &&
		len(d.FilesAdded) == 0 &&
		len(d.FilesRemoved) == 0 &&
		len(d.FilesModified) == 0 &&
		len(d.PermissionsChanged) == 0 &&
		len(d.PackagesAdded) == 0 &&
		len(d.PackagesRemoved) == 0 &&
		len(d.PackagesUpdated) == 0 &&
		len(d.PortsOpened) == 0 &&
		len(d.PortsClosed) == 0 &&
		len(d.UsersAdded) == 0 &&
		len(d.UsersRemoved) == 0
}

func (d *AuditDiff) TotalChanges() int {
	return len(d.ServicesAdded) +
		len(d.ServicesRemoved) +
		len(d.FilesAdded) +
		len(d.FilesRemoved) +
		len(d.FilesModified) +
		len(d.PermissionsChanged) +
		len(d.PackagesAdded) +
		len(d.PackagesRemoved) +
		len(d.PackagesUpdated) +
		len(d.PortsOpened) +
		len(d.PortsClosed) +
		len(d.UsersAdded) +
		len(d.UsersRemoved)
}

func Diff(old, current *AuditBaseline) *AuditDiff {
	d := &AuditDiff{
		BaselineAge: time.Since(old.CreatedAt).Truncate(time.Second),
	}

	d.ServicesAdded, d.ServicesRemoved = diffStringSets(old.Services, current.Services)
	d.FilesAdded, d.FilesRemoved, d.FilesModified = diffFiles(old.Files, current.Files)
	d.PermissionsChanged = diffPermissions(old.Permissions, current.Permissions)
	d.PackagesAdded, d.PackagesRemoved, d.PackagesUpdated = diffPackages(old.Packages, current.Packages)
	d.PortsOpened, d.PortsClosed = diffStringSets(old.Ports, current.Ports)
	d.UsersAdded, d.UsersRemoved = diffStringSets(old.Users, current.Users)
	d.ensureNonNil()

	return d
}

func diffStringSets(old, current []string) (added, removed []string) {
	oldSet := stringSet(old)
	curSet := stringSet(current)

	for s := range curSet {
		if _, exists := oldSet[s]; !exists {
			added = append(added, s)
		}
	}
	for s := range oldSet {
		if _, exists := curSet[s]; !exists {
			removed = append(removed, s)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

func diffFiles(old, current []FileEntry) (added []string, removed []string, modified []FileChange) {
	oldMap := make(map[string]FileEntry, len(old))
	for _, f := range old {
		oldMap[f.Path] = f
	}
	curMap := make(map[string]FileEntry, len(current))
	for _, f := range current {
		curMap[f.Path] = f
	}

	for path, cf := range curMap {
		of, exists := oldMap[path]
		if !exists {
			added = append(added, path)
			continue
		}
		if cf.Mtime != of.Mtime || cf.Size != of.Size {
			modified = append(modified, FileChange{
				Path:     path,
				OldMtime: of.Mtime,
				NewMtime: cf.Mtime,
				OldSize:  of.Size,
				NewSize:  cf.Size,
			})
		}
	}
	for path := range oldMap {
		if _, exists := curMap[path]; !exists {
			removed = append(removed, path)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)
	sort.Slice(modified, func(i, j int) bool { return modified[i].Path < modified[j].Path })
	return added, removed, modified
}

func diffPermissions(old, current []PermEntry) []PermChange {
	oldMap := make(map[string]PermEntry, len(old))
	for _, p := range old {
		oldMap[p.Path] = p
	}

	var changes []PermChange
	for _, cp := range current {
		op, exists := oldMap[cp.Path]
		if !exists {
			continue
		}
		if cp.Mode != op.Mode || cp.Owner != op.Owner || cp.Group != op.Group {
			changes = append(changes, PermChange{
				Path:     cp.Path,
				OldMode:  op.Mode,
				NewMode:  cp.Mode,
				OldOwner: op.Owner,
				NewOwner: cp.Owner,
				OldGroup: op.Group,
				NewGroup: cp.Group,
			})
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes
}

func diffPackages(old, current []string) (added []string, removed []string, updated []PackageChange) {
	oldMap := parsePackageLines(old)
	curMap := parsePackageLines(current)

	for name, curVer := range curMap {
		oldVer, exists := oldMap[name]
		if !exists {
			added = append(added, name+" "+curVer)
			continue
		}
		if curVer != oldVer {
			updated = append(updated, PackageChange{
				Name:       name,
				OldVersion: oldVer,
				NewVersion: curVer,
			})
		}
	}
	for name, oldVer := range oldMap {
		if _, exists := curMap[name]; !exists {
			removed = append(removed, name+" "+oldVer)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)
	sort.Slice(updated, func(i, j int) bool { return updated[i].Name < updated[j].Name })
	return added, removed, updated
}

func parsePackageLines(lines []string) map[string]string {
	m := make(map[string]string, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		idx := strings.IndexByte(l, ' ')
		if idx < 0 {
			m[l] = ""
			continue
		}
		m[l[:idx]] = l[idx+1:]
	}
	return m
}

func stringSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func (d *AuditDiff) ensureNonNil() {
	if d.ServicesAdded == nil {
		d.ServicesAdded = []string{}
	}
	if d.ServicesRemoved == nil {
		d.ServicesRemoved = []string{}
	}
	if d.FilesAdded == nil {
		d.FilesAdded = []string{}
	}
	if d.FilesRemoved == nil {
		d.FilesRemoved = []string{}
	}
	if d.FilesModified == nil {
		d.FilesModified = []FileChange{}
	}
	if d.PermissionsChanged == nil {
		d.PermissionsChanged = []PermChange{}
	}
	if d.PackagesAdded == nil {
		d.PackagesAdded = []string{}
	}
	if d.PackagesRemoved == nil {
		d.PackagesRemoved = []string{}
	}
	if d.PackagesUpdated == nil {
		d.PackagesUpdated = []PackageChange{}
	}
	if d.PortsOpened == nil {
		d.PortsOpened = []string{}
	}
	if d.PortsClosed == nil {
		d.PortsClosed = []string{}
	}
	if d.UsersAdded == nil {
		d.UsersAdded = []string{}
	}
	if d.UsersRemoved == nil {
		d.UsersRemoved = []string{}
	}
}
