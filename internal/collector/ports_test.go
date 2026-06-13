// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"strings"
	"testing"
)

// unit tests for the ports collector

func TestParseSsLine_TCPWithProcess(t *testing.T) {
	line := `LISTEN 0 128 0.0.0.0:22 0.0.0.0:* users:(("sshd",pid=1234,fd=3))`
	entry, ok := parseSsLine("tcp", line)
	if !ok {
		t.Fatal("parseSsLine() returned ok=false, want true")
	}
	if entry.Protocol != "tcp" {
		t.Errorf("Protocol = %q, want %q", entry.Protocol, "tcp")
	}
	if entry.Local != "0.0.0.0:22" {
		t.Errorf("Local = %q, want %q", entry.Local, "0.0.0.0:22")
	}
	if !strings.Contains(entry.Process, "sshd") {
		t.Errorf("Process = %q, want it to contain 'sshd'", entry.Process)
	}
	if !strings.Contains(entry.Process, "1234") {
		t.Errorf("Process = %q, want it to contain pid '1234'", entry.Process)
	}
}

func TestParseSsLine_TCPWithoutProcess(t *testing.T) {
	line := `LISTEN 0 128 0.0.0.0:80 0.0.0.0:*`
	entry, ok := parseSsLine("tcp", line)
	if !ok {
		t.Fatal("parseSsLine() returned ok=false, want true")
	}
	if entry.Local != "0.0.0.0:80" {
		t.Errorf("Local = %q, want %q", entry.Local, "0.0.0.0:80")
	}
	if entry.Process != "" {
		t.Errorf("Process = %q, want empty string when not available", entry.Process)
	}
}

func TestParseSsLine_IPv6(t *testing.T) {
	line := `LISTEN 0 128 [::]:443 [::]:* users:(("nginx",pid=5678,fd=7))`
	entry, ok := parseSsLine("tcp", line)
	if !ok {
		t.Fatal("parseSsLine() returned ok=false, want true")
	}
	if entry.Local != "[::]:443" {
		t.Errorf("Local = %q, want %q", entry.Local, "[::]:443")
	}
	if !strings.Contains(entry.Process, "nginx") {
		t.Errorf("Process = %q, want it to contain 'nginx'", entry.Process)
	}
}

func TestParseSsLine_UDP(t *testing.T) {
	line := `UNCONN 0 0 0.0.0.0:68 0.0.0.0:* users:(("dhclient",pid=999,fd=5))`
	entry, ok := parseSsLine("udp", line)
	if !ok {
		t.Fatal("parseSsLine() returned ok=false, want true")
	}
	if entry.Protocol != "udp" {
		t.Errorf("Protocol = %q, want %q", entry.Protocol, "udp")
	}
	if !strings.Contains(entry.Process, "dhclient") {
		t.Errorf("Process = %q, want it to contain 'dhclient'", entry.Process)
	}
}

func TestParseSsLine_TooFewFields(t *testing.T) {
	_, ok := parseSsLine("tcp", "LISTEN 0 128")
	if ok {
		t.Error("parseSsLine() should return ok=false for lines with fewer than 4 fields")
	}
}

func TestParseSsLine_WildcardLocal(t *testing.T) {
	line := `LISTEN 0 128 * 0.0.0.0:*`
	_, ok := parseSsLine("tcp", line)
	if ok {
		t.Error("parseSsLine() should return ok=false when local address is '*'")
	}
}

func TestParseSsLine_EmptyLine(t *testing.T) {
	_, ok := parseSsLine("tcp", "")
	if ok {
		t.Error("parseSsLine() should return ok=false for empty input")
	}
}

func TestParseProcess_ValidInput(t *testing.T) {
	raw := `users:(("nginx",pid=1234,fd=6))`
	got := parseProcess(raw)
	if !strings.Contains(got, "nginx") {
		t.Errorf("parseProcess() = %q, want it to contain 'nginx'", got)
	}
	if !strings.Contains(got, "1234") {
		t.Errorf("parseProcess() = %q, want it to contain '1234'", got)
	}
}

func TestParseProcess_MultipleProcesses(t *testing.T) {
	raw := `users:(("nginx",pid=100,fd=6),("nginx",pid=101,fd=6))`
	got := parseProcess(raw)
	if !strings.Contains(got, "nginx") {
		t.Errorf("parseProcess() = %q, want it to contain 'nginx'", got)
	}
}

func TestParseProcess_EmptyString(t *testing.T) {
	got := parseProcess("")
	if got != "" {
		t.Errorf("parseProcess(%q) = %q, want empty string", "", got)
	}
}

func TestParseProcess_MalformedInput(t *testing.T) {
	got := parseProcess("garbage-data-no-match")
	if got != "" {
		t.Errorf("parseProcess(%q) = %q, want empty string for malformed input", "garbage-data-no-match", got)
	}
}

func TestParseProcess_Format(t *testing.T) {
	raw := `users:(("sshd",pid=42,fd=3))`
	got := parseProcess(raw)
	want := "sshd (pid 42)"
	if got != want {
		t.Errorf("parseProcess() = %q, want %q", got, want)
	}
}

func TestFormatPortsForPrompt_Nil(t *testing.T) {
	got := FormatPortsForPrompt(nil)
	if !strings.Contains(got, "no listening ports") {
		t.Errorf("FormatPortsForPrompt(nil) = %q, want it to mention no ports", got)
	}
}

func TestFormatPortsForPrompt_Empty(t *testing.T) {
	got := FormatPortsForPrompt([]PortEntry{})
	if !strings.Contains(got, "no listening ports") {
		t.Errorf("FormatPortsForPrompt([]) = %q, want it to mention no ports", got)
	}
}

func TestFormatPortsForPrompt_WithKnownProcess(t *testing.T) {
	entries := []PortEntry{
		{Protocol: "tcp", Local: "0.0.0.0:22", Process: "sshd (pid 1234)"},
	}
	got := FormatPortsForPrompt(entries)
	if !strings.Contains(got, "sshd") {
		t.Errorf("FormatPortsForPrompt() missing process name, got:\n%s", got)
	}
	if !strings.Contains(got, "0.0.0.0:22") {
		t.Errorf("FormatPortsForPrompt() missing address, got:\n%s", got)
	}
	if !strings.Contains(got, "tcp") {
		t.Errorf("FormatPortsForPrompt() missing protocol, got:\n%s", got)
	}
}

func TestFormatPortsForPrompt_WithUnknownProcess(t *testing.T) {
	entries := []PortEntry{
		{Protocol: "tcp", Local: "0.0.0.0:80", Process: ""},
	}
	got := FormatPortsForPrompt(entries)
	if !strings.Contains(got, "process unknown") {
		t.Errorf("FormatPortsForPrompt() should label unknown process, got:\n%s", got)
	}
}

func TestFormatPortsForPrompt_NoTrailingNewline(t *testing.T) {
	entries := []PortEntry{
		{Protocol: "tcp", Local: "0.0.0.0:22", Process: "sshd (pid 1)"},
	}
	got := FormatPortsForPrompt(entries)
	if strings.HasSuffix(got, "\n") {
		t.Errorf("FormatPortsForPrompt() output must not end with a newline, got:\n%q", got)
	}
}

func TestPortsCollect_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Ports() panicked: %v", r)
		}
	}()
	_, _ = Ports()
}
