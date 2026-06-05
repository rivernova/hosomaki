// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sanitiser

import (
	"regexp"
	"strings"
)

// sanitisation rules to apply to collected data before it goes into prompts.
// each rule is a struct implementing the Rule interface, and they are applied in sequence by the sanitiser.
// each rule should be designed to be as precise as possible in what it matches and replaces,
// to avoid over-sanitisation which can remove important context from the data.

type StripTimestamps struct{}

func (StripTimestamps) Name() string { return "strip-timestamps" }

var (
	reISO8601    = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:[.,]\d+)?(?:[+-]\d{2}:?\d{2}|Z)?`)
	reSyslogTime = regexp.MustCompile(`(?m)^[A-Z][a-z]{2} +\d{1,2} \d{2}:\d{2}:\d{2} `)
	reJournalTS  = regexp.MustCompile(`(?m)^-- (?:Boot|Logs begin|No entries|Reboot)[^\n]*`)
	rePidTag     = regexp.MustCompile(`\[\d{2,}\]\s*`)
)

func (StripTimestamps) Apply(input string) string {
	input = reISO8601.ReplaceAllString(input, "")
	input = reSyslogTime.ReplaceAllString(input, "")
	input = reJournalTS.ReplaceAllString(input, "")
	input = rePidTag.ReplaceAllString(input, "")
	return input
}

type MaskURLs struct{}

func (MaskURLs) Name() string { return "mask-urls" }

var reURL = regexp.MustCompile(`(?i)\b(?:https?|ftp|rsync|file)://[^\s,;)\]>"']+`)

func (MaskURLs) Apply(input string) string {
	return reURL.ReplaceAllString(input, "<URL>")
}

type MaskIPv4 struct{}

func (MaskIPv4) Name() string { return "mask-ipv4" }

var reIPv4 = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}(?:/\d{1,2})?\b`)

func (MaskIPv4) Apply(input string) string {
	return reIPv4.ReplaceAllString(input, "<IPV4>")
}

type MaskIPv6 struct{}

func (MaskIPv6) Name() string { return "mask-ipv6" }

var reIPv6 = regexp.MustCompile(
	`\b(?:` +
		`(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}` +
		`|` +
		`(?:[0-9a-fA-F]{1,4}:){1,7}:` +
		`|` +
		`(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}` +
		`|::(?:[0-9a-fA-F]{1,4}:){0,6}[0-9a-fA-F]{0,4}` +
		`)\b`,
)

func (MaskIPv6) Apply(input string) string {
	return reIPv6.ReplaceAllString(input, "<IPV6>")
}

type MaskMACAddresses struct{}

func (MaskMACAddresses) Name() string { return "mask-mac" }

var reMAC = regexp.MustCompile(`\b(?:[0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}\b`)

func (MaskMACAddresses) Apply(input string) string {
	return reMAC.ReplaceAllString(input, "<MAC>")
}

type MaskHexAddresses struct{}

func (MaskHexAddresses) Name() string { return "mask-hex" }

var (
	rePrefixedHex = regexp.MustCompile(`\b0x[0-9a-fA-F]{6,}\b`)
	reBareHex     = regexp.MustCompile(`\b[0-9a-fA-F]{12,}\b`)
)

func (MaskHexAddresses) Apply(input string) string {
	input = rePrefixedHex.ReplaceAllString(input, "<HEX>")
	input = reBareHex.ReplaceAllString(input, "<HEX>")
	return input
}

type MaskUUIDs struct{}

func (MaskUUIDs) Name() string { return "mask-uuid" }

var reUUID = regexp.MustCompile(`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`)

func (MaskUUIDs) Apply(input string) string {
	return reUUID.ReplaceAllString(input, "<UUID>")
}

type MaskAbsolutePaths struct{}

func (MaskAbsolutePaths) Name() string { return "mask-paths" }

var (
	reConfigPath = regexp.MustCompile(`(?:/etc(?:/[^\s,;)\]>"']+)+)`)
	reLogPath    = regexp.MustCompile(`(?:/var/log(?:/[^\s,;)\]>"']+)+)`)
	reCachePath  = regexp.MustCompile(`(?:/var/cache(?:/[^\s,;)\]>"']+)+)`)
	reLibPath    = regexp.MustCompile(`(?:/usr/(?:lib|share|local)(?:/[^\s,;)\]>"']+)+)`)
)

func (MaskAbsolutePaths) Apply(input string) string {
	input = reConfigPath.ReplaceAllString(input, "<CONFIG_PATH>")
	input = reLogPath.ReplaceAllString(input, "<LOG_PATH>")
	input = reCachePath.ReplaceAllString(input, "<CACHE_PATH>")
	input = reLibPath.ReplaceAllString(input, "<LIB_PATH>")
	input = maskRemainingPaths(input)
	return input
}

var reAnyAbsolutePath = regexp.MustCompile(`(^|[\s(])/[A-Za-z0-9_.][^\s,;)\]>"']*`)

func maskRemainingPaths(input string) string {
	return reAnyAbsolutePath.ReplaceAllStringFunc(input, func(m string) string {
		if len(m) == 0 {
			return m
		}
		head := m[0]
		if head == '/' {
			return "<PATH>"
		}
		return string(head) + "<PATH>"
	})
}

type MaskHomePaths struct{}

func (MaskHomePaths) Name() string { return "mask-home" }

var reHomePath = regexp.MustCompile(`/home/[^/\s,;)\]>"']+(?:/[^\s,;)\]>"']*)*`)

func (MaskHomePaths) Apply(input string) string {
	return reHomePath.ReplaceAllString(input, "<HOME_PATH>")
}

type NormalisePackageNVR struct{}

func (NormalisePackageNVR) Name() string { return "normalise-package-nvr" }

var reRPMNVR = regexp.MustCompile(
	`\b([a-zA-Z][a-zA-Z0-9_+.]*(?:-[a-zA-Z][a-zA-Z0-9_+.]*)*)` +
		`-\d+(?:\.\d+){0,3}` +
		`(?:-[\w.]+)?` +
		`(?:\.(?:fc|el|rhel|al|amzn|suse|deb|ubuntu)\d+)?` +
		`(?:\.(?:x86_64|i[3-6]86|aarch64|armv\d\w*|noarch|src))\b`,
)

func (NormalisePackageNVR) Apply(input string) string {
	return reRPMNVR.ReplaceAllString(input, "$1-<VERSION>")
}

type NormaliseRepoNames struct{}

func (NormaliseRepoNames) Name() string { return "normalise-repo" }

var reRepoCacheDir = regexp.MustCompile(`/var/cache/(?:libdnf5|dnf|yum)/([a-zA-Z0-9._-]+)-[0-9a-fA-F]{6,}`)
var reRepoCtx = regexp.MustCompile(`(?:repo|repository)[\s:=]+"?([a-zA-Z0-9._-]+)"?`)

func (NormaliseRepoNames) Apply(input string) string {
	input = reRepoCacheDir.ReplaceAllString(input, "<REPO_CACHE>")
	input = reRepoCtx.ReplaceAllString(input, "repo <REPO>")
	return input
}

type CollapseRepeats struct{}

func (CollapseRepeats) Name() string { return "collapse-repeats" }

func (CollapseRepeats) Apply(input string) string {
	lines := strings.Split(input, "\n")
	if len(lines) <= 1 {
		return input
	}
	var out []string
	var prev string
	count := 0
	flush := func() {
		if count == 0 {
			return
		}
		if count == 1 {
			out = append(out, prev)
		} else {
			out = append(out, prev+" (repeated "+itoa(count)+" times)")
		}
	}
	for _, line := range lines {
		if line == prev {
			count++
			continue
		}
		flush()
		prev = line
		count = 1
	}
	flush()
	return strings.Join(out, "\n")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
