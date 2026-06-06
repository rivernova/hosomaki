// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"fmt"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
)

// prompt for passing the environment info to the LLM

func EnvironmentSection(e collector.Environment) string {
	var b strings.Builder

	field := func(label, value string) {
		if strings.TrimSpace(value) == "" {
			value = "(unknown)"
		}
		fmt.Fprintf(&b, "%s: %s\n", label, value)
	}

	b.WriteString("=== Host environment ===\n")

	switch {
	case e.DistroPrettyName != "":
		field("Distribution", e.DistroPrettyName)
	case e.DistroID != "" && e.DistroVersion != "":
		field("Distribution", e.DistroID+" "+e.DistroVersion)
	case e.DistroID != "":
		field("Distribution", e.DistroID)
	default:
		field("Distribution", "")
	}
	if e.DistroID != "" {
		field("Distro ID", e.DistroID)
	}
	if e.DistroLike != "" {
		field("Distro family (ID_LIKE)", e.DistroLike)
	}

	if e.KernelFull != "" {
		field("Kernel", e.KernelFull)
	} else if e.Kernel != "" {
		field("Kernel", e.Kernel)
	}
	if e.Architecture != "" {
		field("Architecture", e.Architecture)
	}

	field("Init system", e.InitSystem)
	field("Package manager", e.PackageManager)
	if e.Shell != "" {
		field("User shell", e.Shell)
	}

	if e.SELinux != "" {
		field("SELinux", e.SELinux)
	}
	if e.AppArmor != "" {
		field("AppArmor", e.AppArmor)
	}
	if e.Virtualisation != "" && e.Virtualisation != "none" {
		field("Virtualisation", e.Virtualisation)
	}

	b.WriteString(`
Use this environment information silently to make every answer correct for this exact system:
- When suggesting or referring to package operations, use the package manager listed above (never assume apt on a non-Debian system, never assume dnf on a non-Red-Hat system).
- When referring to services, paths or log locations, use the conventions of the distribution and init system listed above.
- If SELinux is enforcing or AppArmor is enabled, account for that when explaining permission-denied or access errors.
- Do not repeat any of this environment information back to the user. The user already knows what system they are on. This block is for your reasoning only.

`)

	return b.String()
}
