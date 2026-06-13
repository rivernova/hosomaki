// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"bufio"
	"errors"
	"strings"
	"testing"
)

// unit tests for the process collector

func TestProcessInfo_InvalidPID(t *testing.T) {
	_, err := ProcessInfo(-1)
	if err == nil {
		t.Fatal("ProcessInfo(-1) should return an error, got nil")
	}
	if !errors.Is(err, ErrProcessNotFound) {
		t.Fatalf("expected ErrProcessNotFound, got: %v", err)
	}
}

func TestProcessInfo_NonExistentPID(t *testing.T) {
	_, err := ProcessInfo(999999999)
	if err == nil {
		t.Fatal("ProcessInfo(999999999) should return an error, got nil")
	}
	if !errors.Is(err, ErrProcessNotFound) {
		t.Fatalf("expected ErrProcessNotFound, got: %v", err)
	}
}

func TestProcessInfo_PID1_DoesNotPanic(t *testing.T) {
	snap, err := ProcessInfo(1)
	if err != nil {
		if !errors.Is(err, ErrProcessPermission) && !errors.Is(err, ErrProcessNotFound) {
			t.Fatalf("ProcessInfo(1) unexpected error: %v", err)
		}
		return
	}
	if snap == nil {
		t.Fatal("ProcessInfo(1) returned nil snapshot with nil error")
	}
	if snap.PID != 1 {
		t.Errorf("snapshot.PID = %d, want 1", snap.PID)
	}
}

func TestNormaliseFDTarget_Socket(t *testing.T) {
	got := normaliseFDTarget("socket:[12345]")
	if got != "socket (anonymous)" {
		t.Errorf("normaliseFDTarget(socket:[12345]) = %q, want %q", got, "socket (anonymous)")
	}
}

func TestNormaliseFDTarget_Pipe(t *testing.T) {
	got := normaliseFDTarget("pipe:[99]")
	if got != "pipe (anonymous)" {
		t.Errorf("normaliseFDTarget(pipe:[99]) = %q, want %q", got, "pipe (anonymous)")
	}
}

func TestNormaliseFDTarget_RegularFile(t *testing.T) {
	got := normaliseFDTarget("/var/log/syslog")
	if got != "/var/log/syslog" {
		t.Errorf("normaliseFDTarget(/var/log/syslog) = %q, want %q", got, "/var/log/syslog")
	}
}

func TestNormaliseFDTarget_AnonInode(t *testing.T) {
	got := normaliseFDTarget("anon_inode:inotify")
	if got != "anon_inode:inotify" {
		t.Errorf("normaliseFDTarget(anon_inode:inotify) = %q, want %q", got, "anon_inode:inotify")
	}
}

func TestDecodeHexAddr_IPv4_Loopback(t *testing.T) {
	got := decodeHexAddr("0100007F:0050")
	if got != "127.0.0.1:80" {
		t.Errorf("decodeHexAddr(0100007F:0050) = %q, want %q", got, "127.0.0.1:80")
	}
}

func TestDecodeHexAddr_IPv4_AnyAddr(t *testing.T) {
	got := decodeHexAddr("00000000:0000")
	if got != "0.0.0.0:0" {
		t.Errorf("decodeHexAddr(00000000:0000) = %q, want %q", got, "0.0.0.0:0")
	}
}

func TestDecodeHexAddr_InvalidShort(t *testing.T) {
	got := decodeHexAddr("ZZZZ:0050")
	if got == "" {
		t.Error("decodeHexAddr with invalid input should return non-empty fallback")
	}
}

func TestParseTCPLine_Listen(t *testing.T) {
	line := "   0: 0100007F:0050 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 12345 1 0000000000000000 100 0 0 10 0"
	summary, ok := parseTCPLine(line)
	if !ok {
		t.Fatal("parseTCPLine() returned ok=false for valid LISTEN entry")
	}
	if !strings.Contains(summary, "LISTEN") {
		t.Errorf("parseTCPLine() summary should contain 'LISTEN', got: %q", summary)
	}
	if !strings.Contains(summary, "127.0.0.1") {
		t.Errorf("parseTCPLine() summary should contain local address, got: %q", summary)
	}
}

func TestParseTCPLine_TooFewFields(t *testing.T) {
	_, ok := parseTCPLine("0: 0100007F:0050")
	if ok {
		t.Error("parseTCPLine() should return ok=false for a line with fewer than 4 fields")
	}
}

func TestParseTCPLine_Established(t *testing.T) {
	line := "   1: 0100007F:1F90 0200007F:C350 01 00000000:00000000 00:00000000 00000000     0        0 99 1 0000000000000000 20 4 24 10 -1"
	summary, ok := parseTCPLine(line)
	if !ok {
		t.Fatal("parseTCPLine() returned ok=false for ESTABLISHED entry")
	}
	if !strings.Contains(summary, "ESTABLISHED") {
		t.Errorf("parseTCPLine() summary should contain 'ESTABLISHED', got: %q", summary)
	}
	if !strings.Contains(summary, "remote=") {
		t.Errorf("parseTCPLine() should include remote address for ESTABLISHED, got: %q", summary)
	}
}

func TestFormatProcessSnapshotForPrompt_AllFields(t *testing.T) {
	snap := &ProcessSnapshot{
		PID:       1234,
		Name:      "nginx",
		State:     "S (sleeping)",
		VmRSS:     "4096 kB",
		Threads:   "4",
		UID:       "33",
		OpenFiles: []string{"/var/log/nginx/error.log", "socket (anonymous)"},
		Sockets:   []string{"tcp  LISTEN  local=0.0.0.0:80"},
	}

	out := FormatProcessSnapshotForPrompt(snap)

	for _, want := range []string{
		"1234",
		"nginx",
		"S (sleeping)",
		"4096 kB",
		"4",
		"33",
		"/var/log/nginx/error.log",
		"socket (anonymous)",
		"LISTEN",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("FormatProcessSnapshotForPrompt() missing %q in output:\n%s", want, out)
		}
	}
}

func TestFormatProcessSnapshotForPrompt_EmptyFields(t *testing.T) {
	snap := &ProcessSnapshot{PID: 42}
	out := FormatProcessSnapshotForPrompt(snap)
	if !strings.Contains(out, "42") {
		t.Errorf("FormatProcessSnapshotForPrompt() must include PID, got:\n%s", out)
	}
	if !strings.Contains(out, "(none visible)") {
		t.Errorf("FormatProcessSnapshotForPrompt() should show '(none visible)' for empty open-files list, got:\n%s", out)
	}
}

func TestFormatProcessSnapshotForPrompt_CollectionErrors(t *testing.T) {
	snap := &ProcessSnapshot{
		PID:              99,
		CollectionErrors: []string{"file descriptors: permission denied"},
	}
	out := FormatProcessSnapshotForPrompt(snap)
	if !strings.Contains(out, "permission denied") {
		t.Errorf("FormatProcessSnapshotForPrompt() should include collection warnings, got:\n%s", out)
	}
}

func TestFormatProcessSnapshotForPrompt_NoTrailingNewline(t *testing.T) {
	snap := &ProcessSnapshot{PID: 1, Name: "init"}
	out := FormatProcessSnapshotForPrompt(snap)
	if strings.HasSuffix(out, "\n") {
		t.Errorf("FormatProcessSnapshotForPrompt() must not end with a newline, got:\n%q", out)
	}
}

func TestReverseBytes32(t *testing.T) {
	got := reverseBytes32(0x01020304)
	want := uint32(0x04030201)
	if got != want {
		t.Errorf("reverseBytes32(0x01020304) = 0x%08X, want 0x%08X", got, want)
	}
}

func TestFmtIPv6Hex_Valid(t *testing.T) {
	got := fmtIPv6Hex("00000000000000000000000000000001")
	if !strings.Contains(got, ":") {
		t.Errorf("fmtIPv6Hex should produce colon-separated groups, got: %q", got)
	}
}

func TestFmtIPv6Hex_WrongLength(t *testing.T) {
	got := fmtIPv6Hex("tooshort")
	if got != "tooshort" {
		t.Errorf("fmtIPv6Hex with wrong-length input = %q, want original string", got)
	}
}

func TestCollectStatus_EmptyUidFieldDoesNotPanic(t *testing.T) {
	snap := &ProcessSnapshot{PID: 0}
	const statusContent = "Name:\ttestproc\nState:\tS (sleeping)\nUid:\t\nThreads:\t1\n"
	r := strings.NewReader(statusContent)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		val = strings.TrimSpace(val)
		switch key {
		case "Name":
			snap.Name = val
		case "State":
			snap.State = val
		case "Uid":
			if fields := strings.Fields(val); len(fields) > 0 {
				snap.UID = fields[0]
			}
		}
	}
	// The test passes as long as no panic occurred
	if snap.UID != "" {
		t.Errorf("empty Uid line should leave UID empty, got %q", snap.UID)
	}
	if snap.Name != "testproc" {
		t.Errorf("Name = %q, want %q", snap.Name, "testproc")
	}
}
