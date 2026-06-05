// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sanitiser

import "strings"

type Sanitiser struct {
	rules []Rule
}

type Rule interface {
	Name() string
	Apply(input string) string
}

func New(rules ...Rule) *Sanitiser {
	return &Sanitiser{rules: rules}
}

func Default() *Sanitiser {
	return New(
		StripTimestamps{}, // drop timestamps and PID tags (noise without semantic value)
		MaskURLs{},        // URLs first — they contain path-like sequences
		MaskIPv6{},        // IPv6 before IPv4 (IPv6 syntax is a superset)
		MaskIPv4{},
		MaskMACAddresses{},
		NormaliseRepoNames{}, // recognise repo cache directories before path masking
		MaskHomePaths{},      // home paths before the generic absolute-path rule
		MaskAbsolutePaths{},  // paths BEFORE hex/uuid/nvr so the path is consumed whole
		MaskUUIDs{},
		MaskHexAddresses{},
		NormalisePackageNVR{},
		CollapseRepeats{}, // collapse runs of identical sanitised lines
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
