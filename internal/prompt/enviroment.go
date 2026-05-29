// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// this file contains the prompt template for the "environment" section

func EnvironmentSection(e collector.Environment) string {
	body := environmentBody(e)
	if body == "" {
		return "=== HOST ENVIRONMENT ===\n(no data)"
	}
	return "=== HOST ENVIRONMENT ===\n" + body
}

func environmentBody(e collector.Environment) string {
	var b strings.Builder

	add := func(label, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		b.WriteString(label)
		b.WriteString(": ")
		b.WriteString(strings.TrimSpace(value))
		b.WriteByte('\n')
	}

	distro := e.DistroPrettyName
	if distro == "" {
		distro = strings.TrimSpace(e.DistroID + " " + e.DistroVersion)
	}
	add("distro", distro)
	add("distro id", e.DistroID)

	kernel := e.KernelFull
	if kernel == "" {
		kernel = e.Kernel
	}
	add("kernel", kernel)
	add("architecture", e.Architecture)
	add("init system", e.InitSystem)
	add("package manager", e.PackageManager)
	add("shell", e.Shell)
	add("hostname", e.Hostname)
	add("selinux", e.SELinux)
	add("apparmor", e.AppArmor)
	add("virtualisation", e.Virtualisation)

	return strings.TrimRight(b.String(), "\n")
}
