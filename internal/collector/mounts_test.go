// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"strings"
	"testing"
)

// unit tests for the mounts collector

func TestUnescape_NoEscape(t *testing.T) {
	if got := unescape("/var/log"); got != "/var/log" {
		t.Errorf("unescape(%q) = %q, want %q", "/var/log", got, "/var/log")
	}
}

func TestUnescape_SpaceEncoded(t *testing.T) {
	got := unescape("/mnt/my\\040drive")
	if got != "/mnt/my drive" {
		t.Errorf("unescape(\\040) = %q, want %q", got, "/mnt/my drive")
	}
}

func TestUnescape_TabEncoded(t *testing.T) {
	got := unescape("foo\\011bar")
	if got != "foo\tbar" {
		t.Errorf("unescape(\\011) = %q, want %q", got, "foo\tbar")
	}
}

func TestUnescape_BackslashEncoded(t *testing.T) {
	got := unescape("/mnt/my\\134path")
	if got != "/mnt/my\\path" {
		t.Errorf("unescape(\\134) = %q, want /mnt/my\\path", got)
	}
}

func TestUnescape_NonZeroPrefixOctal(t *testing.T) {
	got := unescape("/mnt/\\101dir")
	if got != "/mnt/Adir" {
		t.Errorf("unescape(\\101) = %q, want /mnt/Adir", got)
	}
}

func TestUnescape_MultipleEscapes(t *testing.T) {
	got := unescape("foo\\040bar\\011baz")
	if got != "foo bar\tbaz" {
		t.Errorf("unescape(multiple) = %q, want %q", got, "foo bar\tbaz")
	}
}

func TestUnescape_TrailingBackslashPassedThrough(t *testing.T) {
	got := unescape("/path\\")
	if !strings.HasPrefix(got, "/path") {
		t.Errorf("unescape(trailing backslash) = %q, want prefix /path", got)
	}
}

func TestUnescape_NoBackslashFastPath(t *testing.T) {
	s := "/var/log/syslog"
	got := unescape(s)
	if got != s {
		t.Errorf("unescape(no backslash) = %q, want identical string", got)
	}
}

func TestBlocksToHuman_Gigabytes(t *testing.T) {
	got := blocksToHuman("10485760")
	if !strings.HasSuffix(got, "G") {
		t.Errorf("blocksToHuman(10 GiB) = %q, want suffix G", got)
	}
}

func TestBlocksToHuman_Megabytes(t *testing.T) {
	got := blocksToHuman("524288")
	if !strings.HasSuffix(got, "M") {
		t.Errorf("blocksToHuman(512 MiB) = %q, want suffix M", got)
	}
}

func TestBlocksToHuman_Kilobytes(t *testing.T) {
	got := blocksToHuman("512")
	if !strings.HasSuffix(got, "K") {
		t.Errorf("blocksToHuman(512 K) = %q, want suffix K", got)
	}
}

func TestBlocksToHuman_Zero(t *testing.T) {
	got := blocksToHuman("0")
	if got == "" {
		t.Error("blocksToHuman(0) must not return empty string")
	}
}

func TestBlocksToHuman_InvalidString(t *testing.T) {
	got := blocksToHuman("notnumber")
	if got != "notnumber" {
		t.Errorf("blocksToHuman(invalid) = %q, want original string passed through", got)
	}
}

func TestBlocksToHuman_NegativePassedThrough(t *testing.T) {
	got := blocksToHuman("-1")
	if got != "-1" {
		t.Errorf("blocksToHuman(-1) = %q, want original string", got)
	}
}

func TestIsNFS_NFS(t *testing.T) {
	if !IsNFS("nfs") {
		t.Error("IsNFS('nfs') should be true")
	}
}

func TestIsNFS_NFS4(t *testing.T) {
	if !IsNFS("nfs4") {
		t.Error("IsNFS('nfs4') should be true")
	}
}

func TestIsNFS_NFS3(t *testing.T) {
	if !IsNFS("nfs3") {
		t.Error("IsNFS('nfs3') should be true")
	}
}

func TestIsNFS_Ext4(t *testing.T) {
	if IsNFS("ext4") {
		t.Error("IsNFS('ext4') should be false")
	}
}

func TestIsNFS_Tmpfs(t *testing.T) {
	if IsNFS("tmpfs") {
		t.Error("IsNFS('tmpfs') should be false")
	}
}

func TestIsPseudoFS_Tmpfs(t *testing.T) {
	if !IsPseudoFS("tmpfs") {
		t.Error("IsPseudoFS('tmpfs') should be true")
	}
}

func TestIsPseudoFS_Proc(t *testing.T) {
	if !IsPseudoFS("proc") {
		t.Error("IsPseudoFS('proc') should be true")
	}
}

func TestIsPseudoFS_Ext4(t *testing.T) {
	if IsPseudoFS("ext4") {
		t.Error("IsPseudoFS('ext4') should be false")
	}
}

func TestIsPseudoFS_NFS4(t *testing.T) {
	if IsPseudoFS("nfs4") {
		t.Error("IsPseudoFS('nfs4') should be false — NFS is real, not pseudo")
	}
}

func TestParseDfOutput_PopulatesFields(t *testing.T) {
	dfOut := "Filesystem     1024-blocks      Used Available Capacity Mounted on\n" +
		"/dev/sda1         102400000  20480000  81920000      20% /\n"

	entries := []MountEntry{{MountPoint: "/"}}
	idx := map[string]int{"/": 0}
	parseDfOutput(dfOut, entries, idx)

	if entries[0].UsePercent != "20%" {
		t.Errorf("UsePercent = %q, want %q", entries[0].UsePercent, "20%")
	}
	if entries[0].Size == "" {
		t.Error("Size must not be empty after parseDfOutput")
	}
	if entries[0].Used == "" {
		t.Error("Used must not be empty after parseDfOutput")
	}
	if entries[0].Avail == "" {
		t.Error("Avail must not be empty after parseDfOutput")
	}
}

func TestParseDfOutput_IgnoresUnknownMountPoints(t *testing.T) {
	dfOut := "/dev/sdb1  102400000  0  102400000  0% /data\n"
	entries := []MountEntry{{MountPoint: "/"}}
	idx := map[string]int{"/": 0}
	parseDfOutput(dfOut, entries, idx)
	if entries[0].UsePercent != "" {
		t.Errorf("entry for / must stay empty when df output only mentions /data, got %q", entries[0].UsePercent)
	}
}

func TestParseDfOutput_SkipsHeaderLine(t *testing.T) {
	dfOut := "Filesystem     1024-blocks      Used Available Capacity Mounted on\n"
	entries := []MountEntry{{MountPoint: "/"}}
	idx := map[string]int{"/": 0}
	parseDfOutput(dfOut, entries, idx)
}

func TestParseDfOutput_ShortLineSkipped(t *testing.T) {
	dfOut := "/dev/sda1  100000  20000\n"
	entries := []MountEntry{{MountPoint: "/"}}
	idx := map[string]int{"/": 0}
	parseDfOutput(dfOut, entries, idx)
	if entries[0].UsePercent != "" {
		t.Error("short df line must not populate fields")
	}
}

func parseProcMountsLines(text string) []MountEntry {
	seen := make(map[string]struct{})
	var entries []MountEntry
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		mp := unescape(fields[1])
		if _, dup := seen[mp]; dup {
			continue
		}
		seen[mp] = struct{}{}
		entries = append(entries, MountEntry{
			Device:     unescape(fields[0]),
			MountPoint: mp,
			FSType:     fields[2],
			Options:    fields[3],
		})
	}
	return entries
}

func TestParseProcMountsLines_BasicEntry(t *testing.T) {
	text := "/dev/sda1 / ext4 rw,relatime 0 0\n"
	entries := parseProcMountsLines(text)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Device != "/dev/sda1" {
		t.Errorf("Device = %q, want /dev/sda1", entries[0].Device)
	}
	if entries[0].MountPoint != "/" {
		t.Errorf("MountPoint = %q, want /", entries[0].MountPoint)
	}
	if entries[0].FSType != "ext4" {
		t.Errorf("FSType = %q, want ext4", entries[0].FSType)
	}
}

func TestParseProcMountsLines_SkipsComments(t *testing.T) {
	text := "# comment line\n/dev/sda1 / ext4 rw 0 0\n"
	entries := parseProcMountsLines(text)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry (comment skipped), got %d", len(entries))
	}
}

func TestParseProcMountsLines_SkipsShortLines(t *testing.T) {
	text := "/dev/sda1 /\n/dev/sdb1 /data ext4 rw 0 0\n"
	entries := parseProcMountsLines(text)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry (short line skipped), got %d", len(entries))
	}
}

func TestParseProcMountsLines_DeduplicatesMountPoints(t *testing.T) {
	text := "/dev/sda1 / ext4 rw 0 0\n/dev/sda1 / ext4 ro 0 0\n"
	entries := parseProcMountsLines(text)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry (duplicate mount point deduped), got %d", len(entries))
	}
	// First occurrence wins.
	if entries[0].Options != "rw" {
		t.Errorf("Options = %q, want first occurrence 'rw'", entries[0].Options)
	}
}

func TestParseProcMountsLines_UnescapesMountPoint(t *testing.T) {
	text := "/dev/sdb1 /mnt/my\\040data ext4 rw 0 0\n"
	entries := parseProcMountsLines(text)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].MountPoint != "/mnt/my data" {
		t.Errorf("MountPoint = %q, want /mnt/my data", entries[0].MountPoint)
	}
}

func TestParseProcMountsLines_NFSDevice(t *testing.T) {
	text := "fileserver:/export/data /mnt/data nfs4 rw 0 0\n"
	entries := parseProcMountsLines(text)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Device != "fileserver:/export/data" {
		t.Errorf("Device = %q, want fileserver:/export/data", entries[0].Device)
	}
	if entries[0].FSType != "nfs4" {
		t.Errorf("FSType = %q, want nfs4", entries[0].FSType)
	}
}

func TestFormatMountsForPrompt_Empty(t *testing.T) {
	got := FormatMountsForPrompt(nil)
	if !strings.Contains(got, "no mount entries") {
		t.Errorf("FormatMountsForPrompt(nil) = %q, must mention no mounts", got)
	}
}

func TestFormatMountsForPrompt_RealFS_ContainsAllFields(t *testing.T) {
	entries := []MountEntry{
		{
			Device:     "/dev/sda1",
			MountPoint: "/",
			FSType:     "ext4",
			Options:    "rw,relatime",
			Size:       "100.0G",
			Used:       "40.0G",
			Avail:      "60.0G",
			UsePercent: "40%",
		},
	}
	got := FormatMountsForPrompt(entries)
	for _, want := range []string{"/dev/sda1", "ext4", "40%", "100.0G"} {
		if !strings.Contains(got, want) {
			t.Errorf("FormatMountsForPrompt missing %q in:\n%s", want, got)
		}
	}
}

func TestFormatMountsForPrompt_NFS_Stale(t *testing.T) {
	entries := []MountEntry{
		{
			Device:     "fileserver:/export/data",
			MountPoint: "/mnt/data",
			FSType:     "nfs4",
			Options:    "rw",
			NFSStale:   true,
			NFSError:   "timed out after 2s — server may be unreachable",
		},
	}
	got := FormatMountsForPrompt(entries)
	if !strings.Contains(got, "STALE") {
		t.Error("FormatMountsForPrompt must mark stale NFS with STALE")
	}
	if !strings.Contains(got, "timed out") {
		t.Error("FormatMountsForPrompt must include the NFS error message for stale mounts")
	}
}

func TestFormatMountsForPrompt_NFS_Responsive(t *testing.T) {
	entries := []MountEntry{
		{
			Device:     "fileserver:/export/data",
			MountPoint: "/mnt/data",
			FSType:     "nfs",
			Options:    "rw",
			NFSStale:   false,
		},
	}
	got := FormatMountsForPrompt(entries)
	if !strings.Contains(got, "responsive") {
		t.Error("FormatMountsForPrompt must mark responsive NFS mounts")
	}
}

func TestFormatMountsForPrompt_NFS_ProbeSkipped(t *testing.T) {
	entries := []MountEntry{
		{
			Device:          "fileserver:/export/data",
			MountPoint:      "/mnt/data",
			FSType:          "nfs4",
			Options:         "rw",
			NFSStale:        false,
			NFSProbeSkipped: true,
			NFSError:        "probe unavailable: exec: stat not found",
		},
	}
	got := FormatMountsForPrompt(entries)
	if strings.Contains(got, "responsive") {
		t.Error("FormatMountsForPrompt must NOT say 'responsive' when probe was skipped")
	}
	if !strings.Contains(got, "unknown") {
		t.Error("FormatMountsForPrompt must say 'unknown' when probe was skipped")
	}
}

func TestFormatMountsForPrompt_PseudoFSGrouped(t *testing.T) {
	entries := []MountEntry{
		{Device: "tmpfs", MountPoint: "/run", FSType: "tmpfs", Options: "rw"},
		{Device: "proc", MountPoint: "/proc", FSType: "proc", Options: "ro"},
		{Device: "sysfs", MountPoint: "/sys", FSType: "sysfs", Options: "ro"},
	}
	got := FormatMountsForPrompt(entries)
	if strings.Contains(got, "mountpoint:  /run") {
		t.Error("FormatMountsForPrompt must not emit per-entry detail for pseudo-FS")
	}
	if !strings.Contains(got, "pseudo_filesystems") {
		t.Error("FormatMountsForPrompt must include the pseudo_filesystems summary line")
	}
}

func TestFormatMountsForPrompt_PseudoFSSummaryIsDeterministic(t *testing.T) {
	entries := []MountEntry{
		{Device: "tmpfs", MountPoint: "/run", FSType: "tmpfs", Options: "rw"},
		{Device: "proc", MountPoint: "/proc", FSType: "proc", Options: "ro"},
		{Device: "sysfs", MountPoint: "/sys", FSType: "sysfs", Options: "ro"},
		{Device: "cgroup2", MountPoint: "/sys/fs/cgroup", FSType: "cgroup2", Options: "ro"},
	}
	first := FormatMountsForPrompt(entries)
	for i := 0; i < 50; i++ {
		got := FormatMountsForPrompt(entries)
		if got != first {
			t.Errorf("FormatMountsForPrompt output is non-deterministic (iteration %d):\nfirst:  %q\ngot:    %q", i+1, first, got)
			return
		}
	}
}

func TestFormatMountsForPrompt_PseudoFSSummaryIsSorted(t *testing.T) {
	entries := []MountEntry{
		{Device: "tmpfs", MountPoint: "/run", FSType: "tmpfs", Options: "rw"},
		{Device: "proc", MountPoint: "/proc", FSType: "proc", Options: "ro"},
	}
	got := FormatMountsForPrompt(entries)
	procIdx := strings.Index(got, "proc×")
	tmpfsIdx := strings.Index(got, "tmpfs×")
	if procIdx < 0 || tmpfsIdx < 0 {
		t.Fatalf("expected both proc and tmpfs in summary, got: %q", got)
	}
	if procIdx > tmpfsIdx {
		t.Errorf("pseudo_filesystems must be sorted alphabetically; proc should appear before tmpfs in: %q", got)
	}
}

func TestFormatMountsForPrompt_NoTrailingNewline(t *testing.T) {
	entries := []MountEntry{
		{Device: "/dev/sda1", MountPoint: "/", FSType: "ext4", Options: "rw"},
	}
	got := FormatMountsForPrompt(entries)
	if strings.HasSuffix(got, "\n") {
		t.Errorf("FormatMountsForPrompt must not end with a newline, got: %q", got)
	}
}

func TestFormatMountsForPrompt_MixedReal_AndPseudo(t *testing.T) {
	entries := []MountEntry{
		{Device: "/dev/sda1", MountPoint: "/", FSType: "ext4", Options: "rw"},
		{Device: "tmpfs", MountPoint: "/run", FSType: "tmpfs", Options: "rw"},
	}
	got := FormatMountsForPrompt(entries)
	if !strings.Contains(got, "/dev/sda1") {
		t.Error("real filesystem must be listed in detail")
	}
	if !strings.Contains(got, "pseudo_filesystems") {
		t.Error("pseudo filesystem must be in the summary line")
	}
}

func TestMounts_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Mounts() panicked: %v", r)
		}
	}()
	_ = Mounts()
}

func TestMounts_ResultIsStructured(t *testing.T) {
	result := Mounts()
	_ = result.Entries
	_ = result.Warnings
}
