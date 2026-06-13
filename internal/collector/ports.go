// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// port collection logic for the ports command

type PortEntry struct {
	Protocol string
	Local    string // address:port
	Process  string // "name (pid N)"
}

var reProcField = regexp.MustCompile(`\("([^"]+)",pid=(\d+)`)

func Ports() ([]PortEntry, []string) {
	var entries []PortEntry
	var warnings []string

	for _, proto := range []string{"tcp", "udp"} {
		got, warn := collectProtoEntries(proto)
		if warn != "" {
			warnings = append(warnings, warn)
		}
		entries = append(entries, got...)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Protocol != entries[j].Protocol {
			return entries[i].Protocol < entries[j].Protocol
		}
		return entries[i].Local < entries[j].Local
	})

	return entries, warnings
}

func collectProtoEntries(proto string) ([]PortEntry, string) {
	flag := "-tlnpH"
	if proto == "udp" {
		flag = "-ulnpH"
	}

	out, errMsg := run(binSs, flag)
	if errMsg != "" {
		return nil, fmt.Sprintf("%s ports: %s", proto, errMsg)
	}

	seen := make(map[string]struct{})
	var entries []PortEntry

	for _, line := range nonEmptyLines(out) {
		entry, ok := parseSsLine(proto, line)
		if !ok {
			continue
		}
		key := entry.Protocol + " " + entry.Local
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		entries = append(entries, entry)
	}

	return entries, ""
}

func parseSsLine(proto, line string) (PortEntry, bool) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return PortEntry{}, false
	}

	local := fields[3]
	if local == "" || local == "*" {
		return PortEntry{}, false
	}

	process := ""
	if len(fields) >= 6 {
		process = parseProcess(fields[5])
	}

	return PortEntry{
		Protocol: proto,
		Local:    local,
		Process:  process,
	}, true
}

func parseProcess(raw string) string {
	m := reProcField.FindStringSubmatch(raw)
	if m == nil {
		return ""
	}
	name, pid := m[1], m[2]
	if name == "" {
		return ""
	}
	return fmt.Sprintf("%s (pid %s)", name, pid)
}

func FormatPortsForPrompt(entries []PortEntry) string {
	if len(entries) == 0 {
		return "(no listening ports found)"
	}
	var b strings.Builder
	for _, e := range entries {
		if e.Process != "" {
			_, err := fmt.Fprintf(&b, "%s  %s  — %s\n", e.Protocol, e.Local, e.Process)
			if err != nil {
				return ""
			}
		} else {
			_, err := fmt.Fprintf(&b, "%s  %s  — (process unknown)\n", e.Protocol, e.Local)
			if err != nil {
				return ""
			}
		}
	}
	return strings.TrimSpace(b.String())
}
