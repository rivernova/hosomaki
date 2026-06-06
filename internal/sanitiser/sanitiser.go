// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package sanitiser

import "strings"

// uses the rules for sanitising the collected data

type Rule interface {
	Name() string
	Apply(input string) string
}

type Sanitiser struct {
	rules []Rule
}

func New(rules ...Rule) *Sanitiser {
	return &Sanitiser{rules: rules}
}

//  1. StripTimestamps:        timestamps and PID tags add noise that would
//     otherwise distort downstream regex matching.
//  2. StripSyslogHostnames:    runs after timestamps so the hostname-process
//     prefix is at the start of the line.
//  3. MaskURLs:                URLs contain path-like sequences; consume
//     them first so the path rules do not partially
//     eat URL paths.
//  4. MaskEmails:              independent of address rules; grouped here
//     with other identifier rules for clarity.
//  5. MaskMACAddresses:        MUST run before MaskIPv6.  The IPv6 grammar
//     can match a hex-pair sequence joined by
//     colons, which would misclassify MACs.
//  6. MaskIPv4, MaskIPv6:      IPv6 uses net.ParseIP for validation, so the
//     order between v4 and v6 does not matter, but
//     we put v4 first as the simpler check.
//  7. NormaliseRepoNames:      libdnf5 cache directories are recognised
//     before the generic /var/cache path rule.
//  8. MaskHomePaths:           home paths are masked before the generic
//     absolute-path rule so the user prefix is
//     consumed as a unit.
//  9. MaskAbsolutePaths:       MUST run before MaskHexAddresses, MaskUUIDs,
//     and NormalisePackageNVR.  Otherwise the
//     inner hex/uuid segments of a path would be
//     partially rewritten.
//  10. MaskUUIDs:              runs after paths so path-embedded UUIDs are
//     not touched.
//  11. MaskHexAddresses:       runs after paths and UUIDs for the same
//     reason.
//  12. NormalisePackageNVR:    after hex masking because hex segments in
//     version-release suffixes would otherwise
//     break the NVR regex.
//  13. ClassifyLines:          second-to-last so each surviving line
//     carries a category tag in the model's view.
//  14. CollapseRepeats:        last because it operates on the final
//     canonical form of each line.
func Default() *Sanitiser {
	return New(
		StripTimestamps{},
		StripSyslogHostnames{},
		MaskURLs{},
		MaskEmails{},
		MaskMACAddresses{},
		MaskIPv4{},
		MaskIPv6{},
		NormaliseRepoNames{},
		MaskHomePaths{},
		MaskAbsolutePaths{},
		MaskUUIDs{},
		MaskHexAddresses{},
		NormalisePackageNVR{},
		ClassifyLines{},
		CollapseRepeats{},
	)
}

func (s *Sanitiser) Sanitise(input string) string {
	if input == "" {
		return ""
	}
	out := input
	for _, r := range s.rules {
		out = r.Apply(out)
	}
	return strings.TrimSpace(out)
}
func (s *Sanitiser) Rules() []Rule {
	out := make([]Rule, len(s.rules))
	copy(out, s.rules)
	return out
}
