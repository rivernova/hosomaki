// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"fmt"
	"os/exec"
	"strings"
)

// PendingUpdate represents a single pending package update.
type PendingUpdate struct {
	Package        string `json:"package"`
	Installed      string `json:"installed"`
	Available      string `json:"available"`
	Security       bool   `json:"security"`
	RebootRequired bool   `json:"reboot_required"`
}

// PendingUpdates collects pending package updates from the system's
// package manager. Accepts the detected environment (caller must obtain
// it via Env() or inject a mock in tests).
func PendingUpdates(env Environment) ([]PendingUpdate, error) {
	mgr := env.PackageManager
	if mgr == "" {
		return nil, fmt.Errorf("no supported package manager found")
	}
	return collectPendingUpdates(mgr)
}

func collectPendingUpdates(mgr string) ([]PendingUpdate, error) {
	cmd := pendingUpdatesCommand(mgr)
	if cmd == "" {
		return nil, fmt.Errorf("pending updates not supported for %q", mgr)
	}

	raw, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return nil, fmt.Errorf("collect pending updates (%s): %w", mgr, err)
	}

	return parsePendingOutput(mgr, strings.TrimSpace(string(raw)))
}

// pendingUpdatesCommand returns a shell command that lists pending updates
// for the given package manager. LC_ALL=C ensures locale-independent output.
func pendingUpdatesCommand(mgr string) string {
	switch mgr {
	case "apt":
		return "LC_ALL=C apt list --upgradable 2>/dev/null | tail -n +2"
	case "dnf", "yum":
		return "LC_ALL=C dnf list updates 2>/dev/null | tail -n +3 | head -n -1"
	case "pacman":
		return "LC_ALL=C pacman -Qu 2>/dev/null"
	case "zypper":
		return "LC_ALL=C zypper list-updates 2>/dev/null | tail -n +5 | head -n -1"
	case "apk":
		return "LC_ALL=C apk list --upgradable 2>/dev/null"
	case "xbps":
		return "LC_ALL=C xbps-install -nuM 2>/dev/null"
	default:
		return ""
	}
}

// isRebootRequired returns true if the package name suggests a reboot
// is needed after update. Uses package-name heuristics shared by
// unattended-upgrades and other tooling.
func isRebootRequired(pkg string) bool {
	pkg = strings.ToLower(pkg)
	return strings.HasPrefix(pkg, "linux-image") ||
 	strings.HasPrefix(pkg, "linux-headers") ||
 	strings.HasPrefix(pkg, "systemd") ||
 	strings.HasPrefix(pkg, "nvidia-") ||
 	strings.HasPrefix(pkg, "libnvidia") ||
 	strings.HasPrefix(pkg, "firmware") ||
 	strings.Contains(pkg, "-modules-") ||
 	strings.Contains(pkg, "-firmware") ||
		pkg == "glibc" ||
		pkg == "libc6" ||
		pkg == "dbus" ||
		pkg == "udev"
}

// parsePendingOutput parses the raw output from a package manager's
// pending-updates command into structured PendingUpdate entries.
func parsePendingOutput(mgr, raw string) ([]PendingUpdate, error) {
	lines := nonEmptyLines(raw)
	if len(lines) == 0 {
		return []PendingUpdate{}, nil
	}

	updates := make([]PendingUpdate, 0, len(lines))
	for _, line := range lines {
		u := parseLine(mgr, line)
		if u != nil {
			updates = append(updates, *u)
		}
	}
	return updates, nil
}

// parseLine dispatches to the correct parser for the given package manager.
func parseLine(mgr, line string) *PendingUpdate {
	switch mgr {
	case "apt":
		return parseAptLine(line)
	case "dnf", "yum":
		return parseDnfLine(line)
	case "pacman":
		return parsePacmanLine(line)
	default:
		return &PendingUpdate{Package: strings.TrimSpace(line)}
	}
}

// apt line format:
//   pkg/stable 1.2.3 amd64 [upgradable from: 1.2.2]
func parseAptLine(line string) *PendingUpdate {
	line = strings.TrimSpace(line)
	if line == "" || strings.Contains(line, "Listing...") {
		return nil
	}

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil
	}

	pkgName := parts[0]
	if idx := strings.IndexByte(pkgName, '/'); idx >= 0 {
		pkgName = pkgName[:idx]
	}
	available := parts[1]

	installed := ""
	rest := strings.Join(parts[2:], " ")
	if idx := strings.Index(rest, "upgradable from:"); idx >= 0 {
		r := rest[idx+17:]
		if end := strings.IndexByte(r, ']'); end >= 0 {
			installed = strings.TrimSpace(r[:end])
		}
	}

	security := strings.Contains(line, "security") ||
		strings.Contains(line, "~deb") ||
		strings.Contains(line, "+security")

	return &PendingUpdate{
		Package:        pkgName,
		Installed:      installed,
		Available:      available,
		Security:       security,
		RebootRequired: isRebootRequired(pkgName),
	}
}

// dnf line format:
//   pkg.x86_64  1.2.3-1  repo
func parseDnfLine(line string) *PendingUpdate {
	line = strings.TrimSpace(line)
	if line == "" || strings.Contains(line, "Last metadata") ||
		strings.Contains(line, "Available Upgrades") || strings.Contains(line, "---") {
		return nil
	}

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil
	}

	pkgName := parts[0]
	if idx := strings.IndexByte(pkgName, '.'); idx >= 0 {
		pkgName = pkgName[:idx]
	}

	security := strings.Contains(strings.ToLower(line), "security")

	return &PendingUpdate{
		Package:        pkgName,
		Available:      parts[1],
		Security:       security,
		RebootRequired: isRebootRequired(pkgName),
	}
}

// pacman line format:
//   pkg 1.2.2-1 -> 1.2.3-1
func parsePacmanLine(line string) *PendingUpdate {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil
	}

	pkgName := parts[0]
	installed := parts[1]
	available := ""
	if len(parts) >= 4 && parts[2] == "->" {
		available = strings.TrimRight(parts[3], "-")
	}

	return &PendingUpdate{
		Package:        pkgName,
		Installed:      installed,
		Available:      available,
		RebootRequired: isRebootRequired(pkgName),
	}
}
