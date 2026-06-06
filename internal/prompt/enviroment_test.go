// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit testing for the enviroment passing to the LLM

func TestEnvironmentSectionContainsKeyFields(t *testing.T) {
	env := collector.Environment{
		DistroID:         "fedora",
		DistroLike:       "",
		DistroVersion:    "40",
		DistroPrettyName: "Fedora Linux 40 (Workstation Edition)",
		Kernel:           "6.8.5-301.fc40.x86_64",
		KernelFull:       "Linux 6.8.5-301.fc40.x86_64 x86_64",
		Architecture:     "x86_64",
		InitSystem:       "systemd",
		PackageManager:   "dnf",
		Shell:            "zsh",
		Hostname:         "workstation",
		SELinux:          "Enforcing",
		Virtualisation:   "none",
	}

	out := EnvironmentSection(env)

	for _, want := range []string{
		"Fedora Linux 40",
		"fedora",
		"x86_64",
		"systemd",
		"dnf",
		"zsh",
		"Enforcing",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("EnvironmentSection missing %q in output:\n%s", want, out)
		}
	}

	if strings.Contains(out, "Virtualisation: none") {
		t.Error("EnvironmentSection should hide 'none' virtualisation")
	}
}

func TestEnvironmentSectionInstructsSilently(t *testing.T) {
	out := EnvironmentSection(collector.Environment{DistroID: "ubuntu"})

	if !strings.Contains(out, "Do not repeat any of this environment information back to the user") {
		t.Error("EnvironmentSection must instruct the model not to repeat the environment back")
	}
}

func TestEnvironmentSectionUnknownFields(t *testing.T) {
	out := EnvironmentSection(collector.Environment{})

	if !strings.Contains(out, "(unknown)") {
		t.Error("EnvironmentSection should mark missing fields as (unknown)")
	}

	if strings.Contains(out, "SELinux:") {
		t.Error("EnvironmentSection should omit SELinux line when value is empty")
	}
	if strings.Contains(out, "AppArmor:") {
		t.Error("EnvironmentSection should omit AppArmor line when value is empty")
	}
	if strings.Contains(out, "User shell:") {
		t.Error("EnvironmentSection should omit shell line when value is empty")
	}
}

func TestEnvironmentSectionPrettyNameFallback(t *testing.T) {
	env := collector.Environment{DistroID: "arch", DistroVersion: "rolling"}
	out := EnvironmentSection(env)

	if !strings.Contains(out, "arch rolling") {
		t.Errorf("EnvironmentSection should fall back to ID+version, got:\n%s", out)
	}
}

func TestEnvironmentSectionShowsVirtualisationWhenSet(t *testing.T) {
	env := collector.Environment{DistroID: "ubuntu", Virtualisation: "docker"}
	out := EnvironmentSection(env)

	if !strings.Contains(out, "Virtualisation: docker") {
		t.Error("EnvironmentSection should expose virtualisation when not 'none'")
	}
}
