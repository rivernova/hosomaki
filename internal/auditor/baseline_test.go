// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package auditor

import (
	"sort"
	"testing"
)

// unit tests for baseline

func TestSortedLines_Basic(t *testing.T) {
	got := sortedLines("zebra\nalpha\nmango\n")
	want := []string{"alpha", "mango", "zebra"}
	if len(got) != len(want) {
		t.Fatalf("sortedLines len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("sortedLines[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSortedLines_EmptyInput(t *testing.T) {
	got := sortedLines("")
	if len(got) != 0 {
		t.Fatalf("sortedLines('') = %v, want empty", got)
	}
}

func TestSortedLines_WhitespaceOnlyLines(t *testing.T) {
	got := sortedLines("a\n   \nb\n\n")
	if len(got) != 2 {
		t.Fatalf("sortedLines should skip blank lines, got %v", got)
	}
}

func TestNonEmptyLines_TrimsSpaces(t *testing.T) {
	got := nonEmptyLines("  hello  \n  world  \n")
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	if got[0] != "hello" || got[1] != "world" {
		t.Fatalf("unexpected lines: %v", got)
	}
}

func TestPackageListCommand_KnownManagers(t *testing.T) {
	cases := []struct {
		mgr  string
		want string
	}{
		{"apt", "dpkg-query"},
		{"dnf", "rpm"},
		{"yum", "rpm"},
		{"pacman", "pacman"},
		{"apk", "apk"},
		{"zypper", "rpm"},
	}
	for _, tc := range cases {
		t.Run(tc.mgr, func(t *testing.T) {
			cmd := packageListCommand(tc.mgr)
			if cmd == "" {
				t.Fatalf("packageListCommand(%q) returned empty string", tc.mgr)
			}
			if tc.want != "" {
				found := false
				// cmd is a shell string, check substring
				for i := 0; i < len(cmd)-len(tc.want)+1; i++ {
					if cmd[i:i+len(tc.want)] == tc.want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("packageListCommand(%q) = %q, expected to contain %q", tc.mgr, cmd, tc.want)
				}
			}
		})
	}
}

func TestPackageListCommand_UnknownManager(t *testing.T) {
	cmd := packageListCommand("totally-unknown")
	if cmd != "" {
		t.Errorf("packageListCommand(unknown) = %q, want empty string", cmd)
	}
}

func TestCollect_ReturnsVersionAndTimestamp(t *testing.T) {
	b := Collect(CollectOptions{WatchDirs: []string{"/tmp"}})
	if b.Version != baselineVersion {
		t.Errorf("Version = %d, want %d", b.Version, baselineVersion)
	}
	if b.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestCollect_ResultIsAlwaysSorted(t *testing.T) {
	b := Collect(CollectOptions{WatchDirs: []string{"/tmp"}})

	if !sort.StringsAreSorted(b.Services) {
		t.Error("Services should be sorted")
	}
	if !sort.StringsAreSorted(b.Ports) {
		t.Error("Ports should be sorted")
	}
	if !sort.StringsAreSorted(b.Users) {
		t.Error("Users should be sorted")
	}

	paths := make([]string, len(b.Files))
	for i, f := range b.Files {
		paths[i] = f.Path
	}
	if !sort.StringsAreSorted(paths) {
		t.Error("Files should be sorted by path")
	}

	ppaths := make([]string, len(b.Permissions))
	for i, p := range b.Permissions {
		ppaths[i] = p.Path
	}
	if !sort.StringsAreSorted(ppaths) {
		t.Error("Permissions should be sorted by path")
	}
}

func TestCollect_NoWatchDirsUsesDefaults(t *testing.T) {
	b := Collect(CollectOptions{})
	if b.Version != baselineVersion {
		t.Errorf("Version = %d, want %d", b.Version, baselineVersion)
	}
}
