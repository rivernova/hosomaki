// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"testing"
)

func TestParseAptLine(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		want  *Update
		match func(*Update) bool
	}{
		{
			name: "normal upgrade",
			line: "nginx/stable 1.24.0-1 amd64 [upgradable from: 1.22.0-1]",
			match: func(u *Update) bool {
				return u.Package == "nginx" &&
					u.Installed == "1.22.0-1" &&
					u.Available == "1.24.0-1" &&
					!u.Security
			},
		},
		{
			name: "security upgrade",
			line: "libssl3/stable-security 3.0.2-1 amd64 [upgradable from: 3.0.1-1]",
			match: func(u *Update) bool {
				return u.Package == "libssl3" &&
					u.Available == "3.0.2-1" &&
					u.Security
			},
		},
		{
			name: "listing header ignored",
			line: "Listing... Done",
			match: func(u *Update) bool {
				return u == nil
			},
		},
		{
			name: "empty line ignored",
			line: "",
			match: func(u *Update) bool {
				return u == nil
			},
		},
		{
			name: "package with epoch version (not security)",
			line: "libc6/stable 2.36-9+deb12u7 amd64 [upgradable from: 2.36-9+deb12u4]",
			match: func(u *Update) bool {
				return u.Package == "libc6" &&
					u.Installed == "2.36-9+deb12u4" &&
					u.Available == "2.36-9+deb12u7" &&
					!u.Security // +deb is a normal backport, not security
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
				return u.Package == "nginx" &&
					u.Available == "1.24.0-1"
			},
		},
		{
			name: "header line ignored",
			line: "Last metadata expiration check: 0:30:00 ago",
			match: func(u *Update) bool {
				return u == nil
			},
		},
		{
			name: "empty line ignored",
			line: "",
			match: func(u *Update) bool {
				return u == nil
			},
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
				return u.Package == "nginx" &&
					u.Installed == "1.22.0-1" &&
					u.Available == "1.24.0-1"
			},
		},
		{
			name: "empty line ignored",
			line: "",
			match: func(u *Update) bool {
				return u == nil
			},
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

func TestParsePendingOutput(t *testing.T) {
	input := `nginx/stable 1.24.0-1 amd64 [upgradable from: 1.22.0-1]
libssl3/stable-security 3.0.2-1 amd64 [upgradable from: 3.0.1-1]`

	updates, err := parseUpdatesOutput("apt", input)
	if err != nil {
		t.Fatalf("parseUpdatesOutput() error = %v", err)
	}

	if len(updates) != 2 {
		t.Fatalf("got %d updates, want 2", len(updates))
	}

	if updates[0].Package != "nginx" || updates[0].Available != "1.24.0-1" {
		t.Errorf("updates[0] = %+v", updates[0])
	}
	if updates[1].Package != "libssl3" || !updates[1].Security {
		t.Errorf("updates[1] = %+v", updates[1])
	}
}

func TestParsePendingOutputEmpty(t *testing.T) {
	updates, err := parseUpdatesOutput("apt", "")
	if err != nil {
		t.Fatalf("parseUpdatesOutput() error = %v", err)
	}
	if len(updates) != 0 {
		t.Errorf("got %d updates, want 0", len(updates))
	}
}

func TestParsePendingOutputDnf(t *testing.T) {
	input := `nginx.x86_64 1.24.0-1 upstream
openssl.x86_64 3.0.2-1 updates`

	updates, err := parseUpdatesOutput("dnf", input)
	if err != nil {
		t.Fatalf("parseUpdatesOutput() error = %v", err)
	}

	if len(updates) != 2 {
		t.Fatalf("got %d updates, want 2", len(updates))
	}

	if updates[0].Package != "nginx" || updates[0].Available != "1.24.0-1" {
		t.Errorf("updates[0] = %+v", updates[0])
	}
}

func TestParsePendingOutputPacman(t *testing.T) {
	input := `nginx 1.22.0-1 -> 1.24.0-1
openssl 3.0.1-1 -> 3.0.2-1`

	updates, err := parseUpdatesOutput("pacman", input)
	if err != nil {
		t.Fatalf("parseUpdatesOutput() error = %v", err)
	}

	if len(updates) != 2 {
		t.Fatalf("got %d updates, want 2", len(updates))
	}

	if updates[0].Package != "nginx" || updates[0].Available != "1.24.0-1" {
		t.Errorf("updates[0] = %+v", updates[0])
	}
}

func TestParsePendingOutputNoiseOnly(t *testing.T) {
	input := "Listing... Done"
	updates, err := parseUpdatesOutput("apt", input)
	if err != nil {
		t.Fatalf("parseUpdatesOutput() error = %v", err)
	}
	if len(updates) != 0 {
		t.Errorf("noise-only input should yield 0 updates, got %d", len(updates))
	}
}

func TestUpdatesEmptyManager(t *testing.T) {
	env := Environment{PackageManager: ""}
	_, err := Updates(env)
	if err == nil {
		t.Error("expected error for empty package manager")
	}
}

func TestUpdatesUnsupportedManager(t *testing.T) {
	env := Environment{PackageManager: "nonexistent"}
	_, err := Updates(env)
	if err == nil {
		t.Error("expected error for unsupported package manager")
	}
}

func TestUpdatesCommandAllManagers(t *testing.T) {
	managers := []string{"apt", "dnf", "yum", "pacman", "zypper", "apk", "xbps", "emerge", "nix"}
	for _, mgr := range managers {
		cmd := updatesCommand(mgr)
		if cmd == "" {
			t.Errorf("updatesCommand(%q) returned empty string", mgr)
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
		{"firmware-iwlwifi", true},   // contains "firmware" prefix
		{"alsa-modules-6.1", true},    // contains "-modules-"
	}
	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			got := isRebootRequired(tt.pkg)
			if got != tt.want {
				t.Errorf("isRebootRequired(%q) = %v, want %v", tt.pkg, got, tt.want)
			}
		})
	}
}

func TestFormatUpdatesForPrompt(t *testing.T) {
	updates := []Update{
		{Package: "nginx", Installed: "1.22", Available: "1.24"},
		{Package: "openssl", Available: "3.0", Security: true},
	}
	result := FormatUpdatesForPrompt(updates)
	if result == "" {
		t.Fatal("FormatUpdatesForPrompt returned empty string")
	}
	if !contains(result, "nginx") {
		t.Error("result should contain 'nginx'")
	}
	if !contains(result, "1.22") {
		t.Error("result should contain '1.22'")
	}
	if !contains(result, "[SECURITY]") {
		t.Error("security update should have [SECURITY] tag")
	}
}

func TestFormatUpdatesForPromptEmpty(t *testing.T) {
	result := FormatUpdatesForPrompt(nil)
	if result != "(no pending updates)" {
		t.Errorf("nil input should return '(no pending updates)', got %q", result)
	}

	result = FormatUpdatesForPrompt([]Update{})
	if result != "(no pending updates)" {
		t.Errorf("empty input should return '(no pending updates)', got %q", result)
	}
}

func TestFormatUpdatesForPromptRebootTag(t *testing.T) {
	updates := []Update{
		{Package: "linux-image-x86", Available: "6.1", RebootRequired: true},
	}
	result := FormatUpdatesForPrompt(updates)
	if !contains(result, "[REBOOT]") {
		t.Error("reboot-required update should have [REBOOT] tag")
	}
}

func TestFormatUpdatesForPromptUnknownInstalled(t *testing.T) {
	updates := []Update{
		{Package: "pkg", Available: "2.0"}, // Installed is ""
	}
	result := FormatUpdatesForPrompt(updates)
	if !contains(result, "(unknown)") {
		t.Error("update with empty Installed should show '(unknown)'")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}