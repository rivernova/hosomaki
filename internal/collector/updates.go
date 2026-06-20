// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// pending package updates collected

type Update struct {
	Package        string
	Installed      string
	Available      string
	Security       bool
	RebootRequired bool
}

type updateCollector func() ([]Update, error)

var updateCollectors = map[string]updateCollector{
	"apt":    collectAptUpdates,
	"dnf":    collectDnfUpdates,
	"yum":    collectDnfUpdates,
	"pacman": collectPacmanUpdates,
	"zypper": collectZypperUpdates,
	"apk":    collectApkUpdates,
	"xbps":   collectXbpsUpdates,
	"emerge": collectEmergeUpdates,
	"nix":    collectNixUpdates,
}

func Updates(env Environment) ([]Update, error) {
	mgr := env.PackageManager
	if mgr == "" {
		return nil, fmt.Errorf("no supported package manager found")
	}
	collect, ok := updateCollectors[mgr]
	if !ok {
		return nil, fmt.Errorf("pending updates not supported for %q", mgr)
	}
	updates, err := collect()
	if err != nil {
		return nil, fmt.Errorf("collect updates (%s): %w", mgr, err)
	}
	if updates == nil {
		updates = []Update{}
	}
	return updates, nil
}

// Most managers' update-listing commands are pipelines (filtering
// with tail/head/grep)
func shellLines(cmd string) ([]string, error) {
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return nil, err
	}
	return nonEmptyLines(strings.TrimSpace(string(out))), nil
}

func collectAptUpdates() ([]Update, error) {
	lines, err := shellLines("LC_ALL=C apt list --upgradable 2>/dev/null | tail -n +2")
	if err != nil {
		return nil, err
	}
	updates := make([]Update, 0, len(lines))
	for _, line := range lines {
		if u := parseAptLine(line); u != nil {
			updates = append(updates, *u)
		}
	}
	return updates, nil
}

func collectPacmanUpdates() ([]Update, error) {
	lines, err := shellLines("LC_ALL=C pacman -Qu 2>/dev/null")
	if err != nil {
		return nil, err
	}
	updates := make([]Update, 0, len(lines))
	for _, line := range lines {
		if u := parsePacmanLine(line); u != nil {
			updates = append(updates, *u)
		}
	}
	return updates, nil
}

func collectZypperUpdates() ([]Update, error) {
	lines, err := shellLines("LC_ALL=C zypper list-updates 2>/dev/null | tail -n +5 | head -n -1")
	if err != nil {
		return nil, err
	}
	return packageNameUpdates(lines), nil
}

func collectApkUpdates() ([]Update, error) {
	lines, err := shellLines("LC_ALL=C apk list --upgradable 2>/dev/null")
	if err != nil {
		return nil, err
	}
	return packageNameUpdates(lines), nil
}

func collectXbpsUpdates() ([]Update, error) {
	lines, err := shellLines("LC_ALL=C xbps-install -nuM 2>/dev/null")
	if err != nil {
		return nil, err
	}
	return packageNameUpdates(lines), nil
}

func collectEmergeUpdates() ([]Update, error) {
	lines, err := shellLines("LC_ALL=C emerge -up --ask=n @world 2>/dev/null | grep -E '^\\[ebuild' | sed 's/.*\\] //' | head -n -1")
	if err != nil {
		return nil, err
	}
	return packageNameUpdates(lines), nil
}

func collectNixUpdates() ([]Update, error) {
	lines, err := shellLines("LC_ALL=C nix-env -q --outdated 2>/dev/null | grep -v '^$'")
	if err != nil {
		return nil, err
	}
	return packageNameUpdates(lines), nil
}

func packageNameUpdates(lines []string) []Update {
	updates := make([]Update, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		updates = append(updates, Update{Package: line})
	}
	return updates
}

func collectDnfUpdates() ([]Update, error) {
	cmd := exec.Command("dnf", "check-update")
	cmd.Env = append(cmd.Environ(), "LC_ALL=C")
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 100 {
			return nil, err
		}
	}

	installed := rpmInstalledVersions()

	updates := make([]Update, 0)
	for _, line := range nonEmptyLines(strings.TrimSpace(string(out))) {
		u := parseDnfLine(line)
		if u == nil {
			continue
		}
		u.Installed = installed[u.Package]
		updates = append(updates, *u)
	}
	return updates, nil
}

func rpmInstalledVersions() map[string]string {
	cmd := exec.Command("rpm", "-qa", "--queryformat", "%{NAME} %{VERSION}-%{RELEASE}\n")
	cmd.Env = append(cmd.Environ(), "LC_ALL=C")
	out, err := cmd.Output()
	if err != nil {
		return map[string]string{}
	}
	versions := make(map[string]string)
	for _, line := range nonEmptyLines(strings.TrimSpace(string(out))) {
		parts := strings.Fields(line)
		if len(parts) == 2 {
			versions[parts[0]] = parts[1]
		}
	}
	return versions
}

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
	parts := strings.Fields(line)
	if len(parts) != 3 || !strings.Contains(parts[0], ".") {
		return nil
	}
	pkgName := parts[0]

	if idx := strings.LastIndexByte(pkgName, '.'); idx >= 0 {
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
