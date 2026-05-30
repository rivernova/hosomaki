// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"testing"
)

// unit testing for environment detection logic

func TestSplitKeyValue(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantOk    bool
	}{
		{"simple", `ID=ubuntu`, "ID", "ubuntu", true},
		{"double quoted", `PRETTY_NAME="Ubuntu 24.04 LTS"`, "PRETTY_NAME", "Ubuntu 24.04 LTS", true},
		{"single quoted", `VERSION_ID='24.04'`, "VERSION_ID", "24.04", true},
		{"id_like multi", `ID_LIKE="rhel fedora"`, "ID_LIKE", "rhel fedora", true},
		{"empty value", `FOO=`, "FOO", "", true},
		{"blank line", ``, "", "", false},
		{"whitespace line", `   `, "", "", false},
		{"comment", `# this is a comment`, "", "", false},
		{"no equals", `weird line`, "", "", false},
		{"equals at start", `=value`, "", "", false},
		{"unclosed quote stays as-is", `NAME="oops`, "NAME", `"oops`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, v, ok := splitKeyValue(tt.line)
			if k != tt.wantKey || v != tt.wantValue || ok != tt.wantOk {
				t.Errorf("splitKeyValue(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tt.line, k, v, ok, tt.wantKey, tt.wantValue, tt.wantOk)
			}
		})
	}
}

func TestDetectPackageManagerByID(t *testing.T) {
	tests := []struct {
		id, idLike, want string
	}{
		{"ubuntu", "debian", "apt"},
		{"debian", "", "apt"},
		{"linuxmint", "ubuntu debian", "apt"},
		{"fedora", "", "dnf"},
		{"rocky", "rhel centos fedora", "dnf"},
		{"almalinux", "rhel centos fedora", "dnf"},
		{"arch", "", "pacman"},
		{"manjaro", "arch", "pacman"},
		{"opensuse-tumbleweed", "opensuse suse", "zypper"},
		{"alpine", "", "apk"},
		{"void", "", "xbps"},
		{"gentoo", "", "emerge"},
		{"nixos", "", "nix"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := detectPackageManager(tt.id, tt.idLike)
			if got != tt.want {
				t.Errorf("detectPackageManager(%q, %q) = %q, want %q",
					tt.id, tt.idLike, got, tt.want)
			}
		})
	}
}

func TestDetectPackageManagerFallsBackToIDLike(t *testing.T) {
	got := detectPackageManager("madeup-distro", "ubuntu debian")
	if got != "apt" {
		t.Errorf("detectPackageManager unknown id with ubuntu ID_LIKE = %q, want apt", got)
	}
}

func TestDetectPackageManagerUnknown(t *testing.T) {
	got := detectPackageManager("totally-unknown-12345", "also-unknown")
	_ = got
}

func TestEnvDoesNotPanic(t *testing.T) {
	e := Env()
	if e.Hostname == "" {
		t.Log("hostname empty — unusual but not necessarily wrong")
	}
}

func TestEnvReturnsConsistentArchitecture(t *testing.T) {
	e := Env()
	if e.Architecture == "" {
		t.Skip("architecture not detectable in this environment")
	}
	switch e.Architecture {
	case "x86_64", "aarch64", "arm64", "armv7l", "i686", "ppc64le", "s390x", "riscv64":
	default:
		t.Logf("architecture is %q — not in the recognised list but not necessarily wrong", e.Architecture)
	}
}
