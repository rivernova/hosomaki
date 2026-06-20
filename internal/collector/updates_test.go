// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"os"
	"strings"
	"testing"
)

// unit tests for updates collection

func TestParseAptLine(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		match func(*Update) bool
	}{
		{
			name: "normal upgrade",
			line: "nginx/stable 1.24.0-1 amd64 [upgradable from: 1.22.0-1]",
			match: func(u *Update) bool {
				return u.Package == "nginx" && u.Installed == "1.22.0-1" && u.Available == "1.24.0-1" && !u.Security
			},
		},
		{
			name: "security upgrade",
			line: "libssl3/stable-security 3.0.2-1 amd64 [upgradable from: 3.0.1-1]",
			match: func(u *Update) bool {
				return u.Package == "libssl3" && u.Available == "3.0.2-1" && u.Security
			},
		},
		{
			name:  "listing header ignored",
			line:  "Listing... Done",
			match: func(u *Update) bool { return u == nil },
		},
		{
			name:  "empty line ignored",
			line:  "",
			match: func(u *Update) bool { return u == nil },
		},
		{
			name: "package with epoch version (not security)",
			line: "libc6/stable 2.36-9+deb12u7 amd64 [upgradable from: 2.36-9+deb12u4]",
			match: func(u *Update) bool {
				return u.Package == "libc6" && u.Installed == "2.36-9+deb12u4" && u.Available == "2.36-9+deb12u7" && !u.Security
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAptLine(tt.line)
			if !tt.match(got) {
				t.Errorf("parseAptLine(%q) = %+v, want match", tt.line, got)
			}
		})
	}
}

func TestParseDnfLine(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		match func(*Update) bool
	}{
		{
			name: "normal update",
			line: "nginx.x86_64 1.24.0-1 upstream",
			match: func(u *Update) bool {
				return u.Package == "nginx" && u.Available == "1.24.0-1"
			},
		},
		{
			name: "security update",
			line: "openssl.x86_64 3.0.2-1 updates-security",
			match: func(u *Update) bool {
				return u.Package == "openssl" && u.Available == "3.0.2-1" && u.Security
			},
		},
		{
			name: "package name containing a dot is preserved (containerd.io)",
			line: "containerd.io.x86_64 2.2.5-1.fc44 docker-ce-stable",
			match: func(u *Update) bool {
				return u.Package == "containerd.io" && u.Available == "2.2.5-1.fc44"
			},
		},
		{
			name: "epoch-prefixed version is kept verbatim",
			line: "docker-ce.x86_64 3:29.6.0-1.fc44 docker-ce-stable",
			match: func(u *Update) bool {
				return u.Package == "docker-ce" && u.Available == "3:29.6.0-1.fc44"
			},
		},
		{
			name:  "DNF4 header line ignored",
			line:  "Last metadata expiration check: 0:30:00 ago",
			match: func(u *Update) bool { return u == nil },
		},
		{
			name:  "DNF4 footer header ignored",
			line:  "Available Upgrades",
			match: func(u *Update) bool { return u == nil },
		},
		{
			name:  "DNF5 banner line 1 ignored",
			line:  "Updating and loading repositories:",
			match: func(u *Update) bool { return u == nil },
		},
		{
			name:  "DNF5 banner line 2 ignored",
			line:  "Repositories loaded.",
			match: func(u *Update) bool { return u == nil },
		},
		{
			name:  "DNF5 'no matching packages' ignored",
			line:  "No matching packages to list",
			match: func(u *Update) bool { return u == nil },
		},
		{
			name:  "DNF5 'Upgrades (...)' header ignored",
			line:  "Upgrades (available for reinstall, available for upgrade)",
			match: func(u *Update) bool { return u == nil },
		},
		{
			name:  "empty line ignored",
			line:  "",
			match: func(u *Update) bool { return u == nil },
		},
		{
			name:  "line without an arch-style dot ignored",
			line:  "somepackage 1.0.0 repo",
			match: func(u *Update) bool { return u == nil },
		},
		{
			name:  "line with wrong field count ignored",
			line:  "nginx.x86_64 1.24.0-1",
			match: func(u *Update) bool { return u == nil },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDnfLine(tt.line)
			if !tt.match(got) {
				t.Errorf("parseDnfLine(%q) = %+v, want match", tt.line, got)
			}
		})
	}
}

func TestParsePacmanLine(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		match func(*Update) bool
	}{
		{
			name: "normal update",
			line: "nginx 1.22.0-1 -> 1.24.0-1",
			match: func(u *Update) bool {
				return u.Package == "nginx" && u.Installed == "1.22.0-1" && u.Available == "1.24.0-1"
			},
		},
		{
			name:  "empty line ignored",
			line:  "",
			match: func(u *Update) bool { return u == nil },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePacmanLine(tt.line)
			if !tt.match(got) {
				t.Errorf("parsePacmanLine(%q) = %+v, want match", tt.line, got)
			}
		})
	}
}

func TestPackageNameUpdates(t *testing.T) {
	lines := []string{"pkg-a 1.0 -> 1.1", "", "  pkg-b 2.0 -> 2.1  "}
	updates := packageNameUpdates(lines)
	if len(updates) != 2 {
		t.Fatalf("got %d updates, want 2 (blank line should be skipped): %+v", len(updates), updates)
	}
	if updates[0].Package != "pkg-a 1.0 -> 1.1" {
		t.Errorf("updates[0].Package = %q", updates[0].Package)
	}
	if updates[1].Package != "pkg-b 2.0 -> 2.1" {
		t.Errorf("updates[1].Package = %q, want trimmed", updates[1].Package)
	}
}

func fakeBinEnv(t *testing.T, scripts map[string]string) {
	t.Helper()
	dir := t.TempDir()
	for name, script := range scripts {
		if err := os.WriteFile(dir+"/"+name, []byte(script), 0o755); err != nil {
			t.Fatalf("failed to write fake %s: %v", name, err)
		}
	}
	t.Setenv("PATH", dir)
}

func TestCollectDnfUpdatesExitCodes(t *testing.T) {
	noInstalled := "#!/bin/sh\nexit 0\n"

	tests := []struct {
		name      string
		dnf       string
		wantErr   bool
		wantCount int
	}{
		{
			name: "exit 100 with updates is not an error",
			dnf: "#!/bin/sh\n" +
				"echo 'Updating and loading repositories:'\n" +
				"echo 'Repositories loaded.'\n" +
				"echo ''\n" +
				"echo 'nginx.x86_64 1.24.0-1 updates'\n" +
				"exit 100\n",
			wantCount: 1,
		},
		{
			name: "exit 0 with no updates is not an error",
			dnf: "#!/bin/sh\n" +
				"echo 'Updating and loading repositories:'\n" +
				"echo 'Repositories loaded.'\n" +
				"echo 'No matching packages to list'\n" +
				"exit 0\n",
			wantCount: 0,
		},
		{
			name:    "exit 1 is a real error",
			dnf:     "#!/bin/sh\necho boom >&2\nexit 1\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeBinEnv(t, map[string]string{"dnf": tt.dnf, "rpm": noInstalled})

			updates, err := Updates(Environment{PackageManager: "dnf"})
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Updates() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Updates() error = %v, want nil", err)
			}
			if len(updates) != tt.wantCount {
				t.Errorf("Updates() returned %d updates, want %d: %+v", len(updates), tt.wantCount, updates)
			}
		})
	}
}

func TestCollectDnfUpdatesFillsInstalledVersion(t *testing.T) {
	dnf := "#!/bin/sh\n" +
		"echo 'Updating and loading repositories:'\n" +
		"echo 'Repositories loaded.'\n" +
		"echo ''\n" +
		"echo 'nginx.x86_64 1.24.0-1 updates'\n" +
		"echo 'containerd.io.x86_64 2.2.5-1.fc44 docker-ce-stable'\n" +
		"exit 100\n"
	rpm := "#!/bin/sh\n" +
		"echo 'nginx 1.22.0-1'\n" +
		"echo 'containerd.io 2.2.4-1.fc44'\n" +
		"exit 0\n"
	fakeBinEnv(t, map[string]string{"dnf": dnf, "rpm": rpm})

	updates, err := collectDnfUpdates()
	if err != nil {
		t.Fatalf("collectDnfUpdates() error = %v, want nil", err)
	}
	if len(updates) != 2 {
		t.Fatalf("got %d updates, want 2: %+v", len(updates), updates)
	}
	if updates[0].Package != "nginx" || updates[0].Installed != "1.22.0-1" {
		t.Errorf("updates[0] = %+v, want Installed filled in from rpm", updates[0])
	}
	if updates[1].Package != "containerd.io" || updates[1].Installed != "2.2.4-1.fc44" {
		t.Errorf("updates[1] = %+v, want Installed filled in from rpm", updates[1])
	}
}

func TestCollectDnfUpdatesSurvivesMissingRpm(t *testing.T) {
	dnf := "#!/bin/sh\n" +
		"echo 'Updating and loading repositories:'\n" +
		"echo 'Repositories loaded.'\n" +
		"echo ''\n" +
		"echo 'nginx.x86_64 1.24.0-1 updates'\n" +
		"exit 100\n"
	fakeBinEnv(t, map[string]string{"dnf": dnf})

	updates, err := collectDnfUpdates()
	if err != nil {
		t.Fatalf("collectDnfUpdates() error = %v, want nil", err)
	}
	if len(updates) != 1 || updates[0].Package != "nginx" || updates[0].Installed != "" {
		t.Errorf("collectDnfUpdates() = %+v, want one nginx update with blank Installed", updates)
	}
}

func TestUpdatesEmptyManager(t *testing.T) {
	if _, err := Updates(Environment{PackageManager: ""}); err == nil {
		t.Error("expected error for empty package manager")
	}
}

func TestUpdatesUnsupportedManager(t *testing.T) {
	if _, err := Updates(Environment{PackageManager: "nonexistent"}); err == nil {
		t.Error("expected error for unsupported package manager")
	}
}

func TestUpdateCollectorsTableComplete(t *testing.T) {
	managers := []string{"apt", "dnf", "yum", "pacman", "zypper", "apk", "xbps", "emerge", "nix"}
	for _, mgr := range managers {
		if _, ok := updateCollectors[mgr]; !ok {
			t.Errorf("updateCollectors is missing an entry for %q", mgr)
		}
	}
}

func TestIsRebootRequired(t *testing.T) {
	tests := []struct {
		pkg  string
		want bool
	}{
		{"linux-image-6.1.0", true},
		{"linux-headers-6.1.0", true},
		{"systemd", true},
		{"systemd-timesyncd", true},
		{"nvidia-driver-535", true},
		{"libnvidia-compute-535", true},
		{"nginx", false},
		{"openssl", false},
		{"python3", false},
		{"glibc", true},
		{"libc6", true},
		{"dbus", true},
		{"udev", true},
		{"firmware-iwlwifi", true},
		{"alsa-modules-6.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			if got := isRebootRequired(tt.pkg); got != tt.want {
				t.Errorf("isRebootRequired(%q) = %v, want %v", tt.pkg, got, tt.want)
			}
		})
	}
}

func TestFormatUpdatesForPrompt(t *testing.T) {
	updates := []Update{
		{Package: "nginx", Installed: "1.22.0-1", Available: "1.24.0-1"},
		{Package: "libssl3", Installed: "3.0.1-1", Available: "3.0.2-1", Security: true},
		{Package: "linux-image-6.1.0", Installed: "6.1.0-1", Available: "6.1.0-2", RebootRequired: true},
	}

	got := FormatUpdatesForPrompt(updates)

	if !strings.Contains(got, "nginx") || !strings.Contains(got, "1.22.0-1") || !strings.Contains(got, "1.24.0-1") {
		t.Errorf("FormatUpdatesForPrompt() missing nginx details: %q", got)
	}
	if !strings.Contains(got, "[SECURITY]") {
		t.Errorf("FormatUpdatesForPrompt() missing [SECURITY] tag: %q", got)
	}
	if !strings.Contains(got, "[REBOOT]") {
		t.Errorf("FormatUpdatesForPrompt() missing [REBOOT] tag: %q", got)
	}
}

func TestFormatUpdatesForPromptEmpty(t *testing.T) {
	got := FormatUpdatesForPrompt(nil)
	if got != "(no pending updates)" {
		t.Errorf("FormatUpdatesForPrompt(nil) = %q, want %q", got, "(no pending updates)")
	}
}

func TestFormatUpdatesForPromptRebootTag(t *testing.T) {
	updates := []Update{
		{Package: "systemd", Installed: "255-1", Available: "255-2", RebootRequired: true},
	}
	got := FormatUpdatesForPrompt(updates)
	if !strings.Contains(got, "[REBOOT]") {
		t.Errorf("FormatUpdatesForPrompt() = %q, want it to contain [REBOOT]", got)
	}
}

func TestFormatUpdatesForPromptUnknownInstalled(t *testing.T) {
	updates := []Update{
		{Package: "nginx", Installed: "", Available: "1.24.0-1"},
	}
	got := FormatUpdatesForPrompt(updates)
	if !strings.Contains(got, "(unknown)") {
		t.Errorf("FormatUpdatesForPrompt() = %q, want it to contain (unknown) for empty Installed", got)
	}
}
