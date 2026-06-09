// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package auditor

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rivernova/hosomaki/internal/collector"
)

// sets a system snapshot for later diffing

const baselineVersion = 1

type FileEntry struct {
	Path  string `json:"path"`
	Mtime int64  `json:"mtime"` // unix seconds
	Size  int64  `json:"size"`
}

type PermEntry struct {
	Path  string `json:"path"`
	Mode  string `json:"mode"`  // octal
	Owner string `json:"owner"` // username
	Group string `json:"group"` // group name
}

type AuditBaseline struct {
	Version          int                   `json:"version"`
	CreatedAt        time.Time             `json:"created_at"`
	Environment      collector.Environment `json:"environment"`
	Services         []string              `json:"services"`
	Files            []FileEntry           `json:"files"`
	Permissions      []PermEntry           `json:"permissions"`
	Ports            []string              `json:"ports"`
	Packages         []string              `json:"packages"`
	Users            []string              `json:"users"`
	CollectionErrors []string              `json:"collection_errors,omitempty"`
}

type CollectOptions struct {
	WatchDirs   []string
	Environment collector.Environment
}

var defaultWatchDirs = []string{
	"/etc",
	"/usr/local/bin",
	"/usr/local/sbin",
}

func Collect(opts CollectOptions) *AuditBaseline {
	if len(opts.WatchDirs) == 0 {
		opts.WatchDirs = defaultWatchDirs
	}

	b := &AuditBaseline{
		Version:     baselineVersion,
		CreatedAt:   time.Now(),
		Environment: opts.Environment,
	}

	record := func(errs ...string) {
		for _, e := range errs {
			if e != "" {
				b.CollectionErrors = append(b.CollectionErrors, e)
			}
		}
	}

	var err string

	b.Services, err = collectServices()
	record(err)

	b.Files, err = collectFiles(opts.WatchDirs)
	record(err)

	b.Permissions, err = collectPermissions(opts.WatchDirs)
	record(err)

	b.Ports, err = collectPorts()
	record(err)

	b.Packages, err = collectPackages(opts.Environment.PackageManager)
	record(err)

	b.Users, err = collectUsers()
	record(err)

	return b
}

func collectServices() ([]string, string) {
	out, err := runShell(
		"systemctl list-units --type=service --all --no-legend --no-pager --plain 2>/dev/null | awk '{print $1}'",
	)
	if err != "" {
		return nil, fmt.Sprintf("services: %s", err)
	}
	return sortedLines(out), ""
}

func collectFiles(dirs []string) ([]FileEntry, string) {
	if len(dirs) == 0 {
		return nil, ""
	}
	args := append(dirs,
		"-maxdepth", "5",
		"-type", "f",
		"-printf", `%T@ %s %p\n`,
	)
	out, err := run("find", args...)
	if err != "" {
		return nil, fmt.Sprintf("files: %s", err)
	}

	var entries []FileEntry
	for _, line := range nonEmptyLines(out) {
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			continue
		}

		mf, parseErr := strconv.ParseFloat(parts[0], 64)
		if parseErr != nil {
			continue
		}
		sz, parseErr := strconv.ParseInt(parts[1], 10, 64)
		if parseErr != nil {
			continue
		}
		entries = append(entries, FileEntry{
			Path:  parts[2],
			Mtime: int64(mf),
			Size:  sz,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	return entries, ""
}

func collectPermissions(dirs []string) ([]PermEntry, string) {
	if len(dirs) == 0 {
		return nil, ""
	}
	args := append(dirs,
		"-maxdepth", "5",
		"-printf", `%#m %U %G %p\n`,
	)
	out, err := run("find", args...)
	if err != "" {
		return nil, fmt.Sprintf("permissions: %s", err)
	}

	var entries []PermEntry
	for _, line := range nonEmptyLines(out) {
		parts := strings.SplitN(line, " ", 4)
		if len(parts) != 4 {
			continue
		}
		entries = append(entries, PermEntry{
			Mode:  parts[0],
			Owner: parts[1],
			Group: parts[2],
			Path:  parts[3],
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	return entries, ""
}

func collectPorts() ([]string, string) {
	out, err := runShell("ss -tlnpH 2>/dev/null; ss -ulnpH 2>/dev/null")
	if err != "" {
		return nil, fmt.Sprintf("ports: %s", err)
	}
	seen := make(map[string]struct{})
	var ports []string
	for _, line := range nonEmptyLines(out) {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		protocol := strings.ToLower(fields[0]) // "tcp" or "udp"
		local := fields[3]
		key := protocol + " " + local
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			ports = append(ports, key)
		}
	}
	sort.Strings(ports)
	return ports, ""
}

func collectPackages(mgr string) ([]string, string) {
	cmd := packageListCommand(mgr)
	if cmd == "" {
		return nil, fmt.Sprintf("packages: unsupported package manager %q", mgr)
	}
	out, err := runShell(cmd)
	if err != "" {
		return nil, fmt.Sprintf("packages (%s): %s", mgr, err)
	}
	return sortedLines(out), ""
}

func packageListCommand(mgr string) string {
	switch mgr {
	case "apt":
		return `dpkg-query -W -f='${Package} ${Version}\n' 2>/dev/null`
	case "dnf", "yum":
		return `rpm -qa --qf '%{NAME} %{VERSION}-%{RELEASE}\n' 2>/dev/null`
	case "pacman":
		return `pacman -Q 2>/dev/null`
	case "apk":
		return `apk info -v 2>/dev/null`
	case "zypper":
		return `rpm -qa --qf '%{NAME} %{VERSION}-%{RELEASE}\n' 2>/dev/null`
	case "xbps":
		return `xbps-query -l 2>/dev/null | awk '{print $2}' | sed 's/-[^-]*$//'`
	default:
		return ""
	}
}

func collectUsers() ([]string, string) {
	out, err := runShell(`awk -F: '{print $1}' /etc/passwd 2>/dev/null`)
	if err != "" {
		return nil, fmt.Sprintf("users: %s", err)
	}
	return sortedLines(out), ""
}

func run(name string, args ...string) (string, string) {
	out, execErr := exec.Command(name, args...).Output()
	if execErr != nil {
		return "", fmt.Sprintf("%s: %v", name, execErr)
	}
	return strings.TrimSpace(string(out)), ""
}

func runShell(cmd string) (string, string) {
	out, execErr := exec.Command("sh", "-c", cmd).Output()
	if execErr != nil {
		return "", fmt.Sprintf("sh -c %q: %v", cmd, execErr)
	}
	return strings.TrimSpace(string(out)), ""
}

func sortedLines(s string) []string {
	lines := nonEmptyLines(s)
	sort.Strings(lines)
	return lines
}

func nonEmptyLines(s string) []string {
	raw := strings.Split(s, "\n")
	out := make([]string, 0, len(raw))
	for _, l := range raw {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}
