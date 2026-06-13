// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// mount collection logic for the mounts command

type MountEntry struct {
	Device          string
	MountPoint      string
	FSType          string
	Options         string
	Size            string
	Used            string
	Avail           string
	UsePercent      string
	NFSStale        bool
	NFSProbeSkipped bool
	NFSError        string
}

type MountsResult struct {
	Entries  []MountEntry
	Warnings []string
}

const nfsProbeTimeout = 2 * time.Second

var pseudoFSTypes = map[string]bool{
	"proc":       true,
	"sysfs":      true,
	"devtmpfs":   true,
	"devpts":     true,
	"tmpfs":      true,
	"cgroup":     true,
	"cgroup2":    true,
	"pstore":     true,
	"bpf":        true,
	"debugfs":    true,
	"tracefs":    true,
	"securityfs": true,
	"hugetlbfs":  true,
	"mqueue":     true,
	"fusectl":    true,
	"configfs":   true,
	"autofs":     true,
	"efivarfs":   true,
	"squashfs":   true,
}

func IsNFS(fsType string) bool {
	return fsType == "nfs" || fsType == "nfs4" || fsType == "nfs3"
}

func IsPseudoFS(fsType string) bool {
	return pseudoFSTypes[fsType]
}

func Mounts() MountsResult {
	raw, warn := readProcMounts()
	result := MountsResult{Entries: raw}
	if warn != "" {
		result.Warnings = append(result.Warnings, warn)
	}

	if len(raw) == 0 {
		return result
	}

	if dfWarn := enrichWithDf(result.Entries); dfWarn != "" {
		result.Warnings = append(result.Warnings, dfWarn)
	}

	for i := range result.Entries {
		if !IsNFS(result.Entries[i].FSType) {
			continue
		}
		result.Entries[i].NFSStale,
			result.Entries[i].NFSProbeSkipped,
			result.Entries[i].NFSError = probeNFSStaleness(result.Entries[i].MountPoint)
	}

	return result
}

func readProcMounts() ([]MountEntry, string) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, fmt.Sprintf("/proc/mounts: %v", err)
	}
	defer func() { _ = f.Close() }()

	seen := make(map[string]struct{})
	var entries []MountEntry

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		mountPoint := unescape(fields[1])

		if _, dup := seen[mountPoint]; dup {
			continue
		}
		seen[mountPoint] = struct{}{}

		entries = append(entries, MountEntry{
			Device:     unescape(fields[0]),
			MountPoint: mountPoint,
			FSType:     fields[2],
			Options:    fields[3],
		})
	}
	if err := scanner.Err(); err != nil {
		return entries, fmt.Sprintf("/proc/mounts scan error: %v", err)
	}
	return entries, ""
}

func enrichWithDf(entries []MountEntry) string {
	var targets []string
	idx := make(map[string]int, len(entries))

	for i, e := range entries {
		if IsPseudoFS(e.FSType) || IsNFS(e.FSType) {
			continue
		}
		targets = append(targets, e.MountPoint)
		idx[e.MountPoint] = i
	}

	if len(targets) == 0 {
		return ""
	}

	args := append([]string{"-Pl"}, targets...)
	out, err := exec.Command(binDf, args...).Output()
	if err != nil {
		return fmt.Sprintf("df: %v", err)
	}

	parseDfOutput(string(out), entries, idx)
	return ""
}

func parseDfOutput(out string, entries []MountEntry, idx map[string]int) {
	for _, line := range nonEmptyLines(out) {
		fields := strings.Fields(line)
		if len(fields) < 6 || fields[0] == "Filesystem" {
			continue
		}

		mountPoint := fields[5]
		i, ok := idx[mountPoint]
		if !ok {
			continue
		}

		entries[i].Size = blocksToHuman(fields[1])
		entries[i].Used = blocksToHuman(fields[2])
		entries[i].Avail = blocksToHuman(fields[3])
		entries[i].UsePercent = fields[4]
	}
}

func probeNFSStaleness(mountPoint string) (stale, skipped bool, errMsg string) {
	cmd := exec.Command("stat", "--file-system", mountPoint)
	if startErr := cmd.Start(); startErr != nil {
		return false, true, fmt.Sprintf("probe unavailable: %v", startErr)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	timer := time.NewTimer(nfsProbeTimeout)
	defer timer.Stop()

	select {
	case err := <-done:
		if err != nil {
			return true, false, fmt.Sprintf("stat error: %v", err)
		}
		return false, false, ""

	case <-timer.C:
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		<-done
		return true, false, fmt.Sprintf("timed out after %s — server may be unreachable", nfsProbeTimeout)
	}
}

func unescape(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] == '\\' && i+4 <= len(s) &&
			s[i+1] >= '0' && s[i+1] <= '7' &&
			s[i+2] >= '0' && s[i+2] <= '7' &&
			s[i+3] >= '0' && s[i+3] <= '7' {
			v := (s[i+1]-'0')<<6 | (s[i+2]-'0')<<3 | (s[i+3] - '0')
			b.WriteByte(v)
			i += 4
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func blocksToHuman(s string) string {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return s
	}
	kib := n
	switch {
	case kib >= 1024*1024:
		return fmt.Sprintf("%.1fG", float64(kib)/1024/1024)
	case kib >= 1024:
		return fmt.Sprintf("%.1fM", float64(kib)/1024)
	default:
		return fmt.Sprintf("%dK", kib)
	}
}

func FormatMountsForPrompt(entries []MountEntry) string {
	if len(entries) == 0 {
		return "(no mount entries found)"
	}

	var mountEntries, pseudo []MountEntry
	for _, e := range entries {
		if IsPseudoFS(e.FSType) {
			pseudo = append(pseudo, e)
		} else {
			mountEntries = append(mountEntries, e)
		}
	}

	var b strings.Builder

	for _, e := range mountEntries {
		_, _ = fmt.Fprintf(&b, "device:      %s\n", e.Device)
		_, _ = fmt.Fprintf(&b, "mountpoint:  %s\n", e.MountPoint)
		_, _ = fmt.Fprintf(&b, "fstype:      %s\n", e.FSType)
		_, _ = fmt.Fprintf(&b, "options:     %s\n", e.Options)
		if e.Size != "" {
			_, _ = fmt.Fprintf(&b, "size:        %s\n", e.Size)
			_, _ = fmt.Fprintf(&b, "used:        %s  (%s)\n", e.Used, e.UsePercent)
			_, _ = fmt.Fprintf(&b, "available:   %s\n", e.Avail)
		}
		if IsNFS(e.FSType) {
			switch {
			case e.NFSStale:
				_, _ = fmt.Fprintf(&b, "nfs_status:  STALE — %s\n", e.NFSError)
			case e.NFSProbeSkipped:
				_, _ = fmt.Fprintf(&b, "nfs_status:  unknown — %s\n", e.NFSError)
			default:
				b.WriteString("nfs_status:  responsive\n")
			}
		}
		b.WriteString("\n")
	}

	if len(pseudo) > 0 {
		typeCounts := make(map[string]int, len(pseudo))
		for _, e := range pseudo {
			typeCounts[e.FSType]++
		}

		types := make([]string, 0, len(typeCounts))
		for t := range typeCounts {
			types = append(types, t)
		}
		sort.Strings(types)

		b.WriteString("pseudo_filesystems: ")
		for i, t := range types {
			if i > 0 {
				b.WriteString(", ")
			}
			_, _ = fmt.Fprintf(&b, "%s×%d", t, typeCounts[t])
		}
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
