// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"fmt"
	"os/exec"
	"strings"
)

// Update represents a single pending package update.
type Update struct {
	Package        string
	Installed      string
	Available      string
	Security       bool
	RebootRequired bool
}

// Updates collects pending package updates from the system's
// package manager. Accepts the detected environment (caller must obtain
// it via Env() or inject a mock in tests).
func Updates(env Environment) ([]Update, error) {
	mgr := env.PackageManager
	if mgr == "" {
		return nil, fmt.Errorf("no supported package manager found")
	}
	return collectUpdates(mgr)
}

func collectUpdates(mgr string) ([]Update, error) {
	cmd := updatesCommand(mgr)
	if cmd == "" {
		return nil, fmt.Errorf("pending updates not supported for %q", mgr)
	}
	raw, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return nil, fmt.Errorf("collect updates (%s): %w", mgr, err)
	}
	return parseUpdatesOutput(mgr, strings.TrimSpace(string(raw)))
}

// updatesCommand returns a shell command that lists pending updates
// for the given package manager. LC_ALL=C ensures locale-independent output.
func updatesCommand(mgr string) string {
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
	case "emerge":
		return "LC_ALL=C emerge -up --ask=n @world 2>/dev/null | grep -E '^\\[ebuild' | sed 's/.*\\] //' | head -n -1"
	case "nix":
		return "LC_ALL=C nix-env -q --outdated 2>/dev/null | grep -v '^$'"
	default:
		return ""
	}
}

// FormatUpdatesForPrompt formats pending updates as a plain-text string
// suitable for the AI prompt. Sanitise the return value before passing
// it to prompt.Updates.
func FormatUpdatesForPrompt(updates []Update) string {
	if len(updates) == 0 {
		return "(no pending updates)"
	}
	var b strings.Builder
	for i, u := range updates {
		tag := ""
		if u.Security {
			tag = " [SECURITY]"
		}
		if u.RebootRequired {
			tag += " [REBOOT]"
		}
		inst := u.Installed
		if inst == "" {
			inst = "(unknown)"
		}
		_, _ = fmt.Fprintf(&b, "%d. %s%s  installed: %s -> available: %s\n",
			i+1, u.Package, tag, inst, u.Available)
	}
	return strings.TrimSpace(b.String())
}

// isRebootRequired returns true if the package name suggests a reboot.
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

func parseUpdatesOutput(mgr, raw string) ([]Update, error) {
	lines := nonEmptyLines(raw)
	if len(lines) == 0 {
		return []Update{}, nil
	}
	updates := make([]Update, 0, len(lines))
	for _, line := range lines {
		u := parseLine(mgr, line)
		if u != nil {
			updates = append(updates, *u)
		}
	}
	return updates, nil
}

func parseLine(mgr, line string) *Update {
	switch mgr {
	case "apt":
		return parseAptLine(line)
	case "dnf", "yum":
		return parseDnfLine(line)
	case "pacman":
		return parsePacmanLine(line)
	default:
		return &Update{Package: strings.TrimSpace(line)}
	}
}

func parseAptLine(line string) *Update {
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
	return &Update{
		Package:        pkgName,
		Installed:      installed,
		Available:      available,
		Security:       security,
		RebootRequired: isRebootRequired(pkgName),
	}
}

func parseDnfLine(line string) *Update {
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
	return &Update{
		Package:        pkgName,
		Available:      parts[1],
		Security:       security,
		RebootRequired: isRebootRequired(pkgName),
	}
}

func parsePacmanLine(line string) *Update {
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
	return &Update{
		Package:        pkgName,
		Installed:      installed,
		Available:      available,
		RebootRequired: isRebootRequired(pkgName),
	}
}