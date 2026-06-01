// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"strings"
	"testing"
)

// unit tests for layout functions for generating sections

func TestFormatUptimeHoursMinutes(t *testing.T) {
	got := formatUptime("up 5 hours, 55 minutes")
	want := "5h 55m"
	if got != want {
		t.Errorf("formatUptime() = %q, want %q", got, want)
	}
}

func TestFormatUptimeDaysHoursMinutes(t *testing.T) {
	got := formatUptime("up 1 day, 3 hours, 2 minutes")
	want := "1d 3h 2m"
	if got != want {
		t.Errorf("formatUptime() = %q, want %q", got, want)
	}
}

func TestFormatUptimeEmpty(t *testing.T) {
	got := formatUptime("")
	if got != "(none)" {
		t.Errorf("formatUptime('') = %q, want (none)", got)
	}
}

func TestFormatUptimeHoursOnly(t *testing.T) {
	got := formatUptime("up 2 hours")
	want := "2h"
	if got != want {
		t.Errorf("formatUptime() = %q, want %q", got, want)
	}
}

func TestFormatMemoryParsesMemLine(t *testing.T) {
	raw := "              total        used        free      shared  buff/cache   available\n" +
		"Mem:            29Gi       7.3Gi       9.5Gi       1.3Gi        13Gi        21Gi\n" +
		"Swap:          8.0Gi          0B       8.0Gi\n"
	lines := formatMemory(raw)
	if len(lines) == 0 {
		t.Fatal("formatMemory() returned no lines")
	}
	mem := lines[0]
	if !strings.Contains(mem, "7.3G") {
		t.Errorf("memory line should contain used value, got %q", mem)
	}
	if !strings.Contains(mem, "29G") {
		t.Errorf("memory line should contain total value, got %q", mem)
	}
	if !strings.Contains(mem, "21G") {
		t.Errorf("memory line should contain available value, got %q", mem)
	}
}

func TestFormatMemorySwapInactiveWhenZero(t *testing.T) {
	raw := "Mem:            16Gi       4Gi       8Gi       1Gi        4Gi        10Gi\n" +
		"Swap:          8.0Gi          0B       8.0Gi\n"
	lines := formatMemory(raw)
	found := false
	for _, l := range lines {
		if strings.Contains(l, "swap") && strings.Contains(l, "inactive") {
			found = true
		}
	}
	if !found {
		t.Errorf("swap with 0B used should show 'inactive', got %v", lines)
	}
}

func TestFormatMemorySwapShownWhenUsed(t *testing.T) {
	raw := "Mem:            16Gi       4Gi       8Gi       1Gi        4Gi        10Gi\n" +
		"Swap:          8.0Gi       2.0Gi       6.0Gi\n"
	lines := formatMemory(raw)
	found := false
	for _, l := range lines {
		if strings.Contains(l, "swap") && !strings.Contains(l, "inactive") {
			found = true
		}
	}
	if !found {
		t.Errorf("swap with usage should show values, got %v", lines)
	}
}

func TestFormatMemoryEmptyInput(t *testing.T) {
	lines := formatMemory("")
	if len(lines) == 0 {
		t.Fatal("formatMemory('') should return fallback line")
	}
	if !strings.Contains(lines[0], "(none)") {
		t.Errorf("formatMemory('') should return '(none)', got %q", lines[0])
	}
}

func TestFormatDiskDeduplicatesDevice(t *testing.T) {
	raw := "Filesystem      Size  Used Avail Use% Mounted on\n" +
		"/dev/nvme0n1p3  952G   49G  901G   6% /\n" +
		"/dev/nvme0n1p3  952G   49G  901G   6% /home\n"
	lines := formatDisk(raw)
	count := 0
	for _, l := range lines {
		if strings.Contains(l, "nvme0n1p3") || strings.Contains(l, "disk /") {
			count++
		}
	}
	if count > 1 {
		t.Errorf("same device should appear only once, got %d lines: %v", count, lines)
	}
}

func TestFormatDiskSkipsNonDevEntries(t *testing.T) {
	raw := "Filesystem      Size  Used Avail Use% Mounted on\n" +
		"efivarfs        192K  137K   51K  73% /sys/firmware/efi/efivars\n" +
		"/dev/sda1        50G   10G   40G  20% /\n"
	lines := formatDisk(raw)
	for _, l := range lines {
		if strings.Contains(l, "efivarfs") {
			t.Errorf("non-/dev/ entry should be skipped, got %q", l)
		}
	}
}

func TestFormatDiskEmptyInput(t *testing.T) {
	lines := formatDisk("")
	if len(lines) == 0 {
		t.Fatal("formatDisk('') should return fallback line")
	}
	if !strings.Contains(lines[0], "(none)") {
		t.Errorf("formatDisk('') should return '(none)', got %q", lines[0])
	}
}

func TestCleanUnit(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"29Gi", "29G"},
		{"7.3Gi", "7.3G"},
		{"837Mi", "837M"},
		{"192Ki", "192K"},
		{"8.0G", "8G"},
		{"2.0M", "2M"},
		{"952G", "952G"},
		{"0B", "0B"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanUnit(tt.input)
			if got != tt.want {
				t.Errorf("cleanUnit(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
