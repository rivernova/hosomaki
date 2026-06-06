// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sanitiser

import (
	"strings"
	"testing"
)

// unit testing for the sanitiser

func TestSanitise_EmptyInput(t *testing.T) {
	got := Default().Sanitise("")
	if got != "" {
		t.Fatalf("empty input should produce empty output, got: %q", got)
	}
}

func TestSanitise_Deterministic(t *testing.T) {
	const input = `2026-06-02T20:08:18+0000 [39054] DEBUG /var/log/test.log https://example.com/foo`
	s := Default()
	first := s.Sanitise(input)
	second := s.Sanitise(input)
	if first != second {
		t.Fatalf("output not deterministic:\n  first:  %q\n  second: %q", first, second)
	}
}

func TestSanitise_DeterministicUnderLoad(t *testing.T) {
	const input = `2026-01-01T00:00:00Z [42] ERROR Failed to bind 0.0.0.0:80 from /home/alice/app /var/cache/x/y`
	s := Default()
	reference := s.Sanitise(input)
	for range 100 {
		if got := s.Sanitise(input); got != reference {
			t.Fatalf("non-deterministic output:\n  reference: %q\n  got:       %q", reference, got)
		}
	}
}

func TestSanitise_NoOpSanitiser(t *testing.T) {
	const input = "any text whatsoever"
	got := New().Sanitise(input)
	if got != input {
		t.Fatalf("no-op sanitiser changed input: %q -> %q", input, got)
	}
}

func TestStripTimestamps_ISO8601(t *testing.T) {
	in := "2026-06-02T20:08:18+0000 hello"
	got := strings.TrimSpace(StripTimestamps{}.Apply(in))
	if got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}

func TestStripTimestamps_Syslog(t *testing.T) {
	in := "Jun  2 20:08:18 hostname sshd[1234]: hello"
	got := StripTimestamps{}.Apply(in)
	if strings.Contains(got, "Jun  2 20:08:18") {
		t.Fatalf("syslog timestamp not stripped, got: %q", got)
	}
}

func TestStripTimestamps_PidTag(t *testing.T) {
	in := "[12345] something"
	got := StripTimestamps{}.Apply(in)
	if strings.Contains(got, "12345") {
		t.Fatalf("PID tag not stripped, got: %q", got)
	}
}

func TestStripSyslogHostnames(t *testing.T) {
	in := "myhost-prod sshd[1234]: Accepted password"
	got := StripSyslogHostnames{}.Apply(in)
	if strings.Contains(got, "myhost-prod") {
		t.Fatalf("hostname not stripped, got: %q", got)
	}
	if !strings.Contains(got, "sshd") || !strings.Contains(got, "Accepted password") {
		t.Fatalf("program name or message lost, got: %q", got)
	}
}

func TestStripSyslogHostnames_NoMatch(t *testing.T) {
	in := "some prose without colons"
	got := StripSyslogHostnames{}.Apply(in)
	if got != in {
		t.Fatalf("unrelated line was modified: %q -> %q", in, got)
	}
}

func TestMaskURLs_HTTP(t *testing.T) {
	in := "fetched from https://download.fedoraproject.org/pub/x.rpm"
	got := MaskURLs{}.Apply(in)
	if !strings.Contains(got, "<URL>") || strings.Contains(got, "fedoraproject") {
		t.Fatalf("URL not masked: %q", got)
	}
}

func TestMaskURLs_Rsync(t *testing.T) {
	in := "via rsync://mirror.example.com/repo"
	got := MaskURLs{}.Apply(in)
	if !strings.Contains(got, "<URL>") {
		t.Fatalf("rsync URL not masked: %q", got)
	}
}

func TestMaskURLs_WithCredentials(t *testing.T) {
	in := "trying https://user:s3cret@host.example.com/path"
	got := MaskURLs{}.Apply(in)
	if strings.Contains(got, "user") || strings.Contains(got, "s3cret") {
		t.Fatalf("credentials leaked through URL mask: %q", got)
	}
}

func TestMaskEmails(t *testing.T) {
	in := "user alice@example.org logged in"
	got := MaskEmails{}.Apply(in)
	if strings.Contains(got, "alice@example.org") {
		t.Fatalf("email not masked: %q", got)
	}
	if !strings.Contains(got, "<EMAIL>") {
		t.Fatalf("missing <EMAIL> placeholder: %q", got)
	}
}

func TestMaskIPv4(t *testing.T) {
	in := "connected to 192.168.1.10 then 10.0.0.5"
	got := MaskIPv4{}.Apply(in)
	if strings.Contains(got, "192.168.1.10") || strings.Contains(got, "10.0.0.5") {
		t.Fatalf("IPv4 not masked: %q", got)
	}
}

func TestMaskIPv6(t *testing.T) {
	in := "addr fe80::1234:5678:9abc:def0 here"
	got := MaskIPv6{}.Apply(in)
	if strings.Contains(got, "fe80::1234:5678:9abc:def0") {
		t.Fatalf("IPv6 not masked: %q", got)
	}
}

func TestMaskIPv6_DoesNotMatchCxxScope(t *testing.T) {
	in := "in std::vector::iterator we crashed"
	got := MaskIPv6{}.Apply(in)
	if strings.Contains(got, "<IPV6>") {
		t.Fatalf("C++ scope resolution misclassified as IPv6: %q", got)
	}
}

func TestMaskIPv6_ValidAddressForms(t *testing.T) {
	cases := []string{
		"2001:db8::1",
		"::1",
		"fe80::1ff:fe23:4567:890a",
		"2001:0db8:0000:0000:0000:ff00:0042:8329",
	}
	for _, addr := range cases {
		in := "addr " + addr + " here"
		got := MaskIPv6{}.Apply(in)
		if !strings.Contains(got, "<IPV6>") {
			t.Errorf("IPv6 %q not masked, got: %q", addr, got)
		}
	}
}

func TestMaskMACAddresses(t *testing.T) {
	in := "device aa:bb:cc:dd:ee:ff registered"
	got := MaskMACAddresses{}.Apply(in)
	if strings.Contains(got, "aa:bb:cc:dd:ee:ff") {
		t.Fatalf("MAC not masked: %q", got)
	}
}

func TestDefault_MACBeforeIPv6(t *testing.T) {
	in := "device aa:bb:cc:dd:ee:ff up"
	got := Default().Sanitise(in)
	if !strings.Contains(got, "<MAC>") {
		t.Fatalf("MAC should be tagged <MAC>, got: %q", got)
	}
	if strings.Contains(got, "<IPV6>") {
		t.Fatalf("MAC was misclassified as IPv6, got: %q", got)
	}
}

func TestMaskHexAddresses_LongHex(t *testing.T) {
	in := "checksum 5a17d225ddb72809e540a2677a5a6edf6c3881324b6089547d11939210fa59ca verified"
	got := MaskHexAddresses{}.Apply(in)
	if strings.Contains(got, "5a17d225ddb72809e540a2677a5a6edf6c3881324b6089547d11939210fa59ca") {
		t.Fatalf("long hex not masked: %q", got)
	}
}

func TestMaskHexAddresses_PrefixedHex(t *testing.T) {
	in := "RIP at 0xffffffff81234567"
	got := MaskHexAddresses{}.Apply(in)
	if strings.Contains(got, "0xffffffff81234567") {
		t.Fatalf("prefixed hex not masked: %q", got)
	}
}

func TestMaskHexAddresses_ShortHexPreserved(t *testing.T) {
	in := "container abc1234 exited"
	got := MaskHexAddresses{}.Apply(in)
	if !strings.Contains(got, "abc1234") {
		t.Fatalf("short hex was masked unexpectedly, got: %q", got)
	}
}

func TestMaskUUIDs(t *testing.T) {
	in := "session abc12345-def0-1234-5678-9abcdef01234 started"
	got := MaskUUIDs{}.Apply(in)
	if !strings.Contains(got, "<UUID>") {
		t.Fatalf("UUID not masked: %q", got)
	}
}

func TestMaskAbsolutePaths_Categories(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"read /etc/passwd", "<CONFIG_PATH>"},
		{"see /var/log/syslog", "<LOG_PATH>"},
		{"in /var/cache/libdnf5/foo", "<CACHE_PATH>"},
		{"from /usr/lib/systemd/system/x.service", "<LIB_PATH>"},
		{"file /opt/myapp/bin/x", "<PATH>"},
	}
	for _, tc := range tests {
		got := MaskAbsolutePaths{}.Apply(tc.in)
		if !strings.Contains(got, tc.want) {
			t.Errorf("%q: expected %q in output, got: %q", tc.in, tc.want, got)
		}
	}
}

func TestMaskAbsolutePaths_BareDirectories(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"changed to /etc directory", "<CONFIG_PATH>"},
		{"logs under /var/log", "<LOG_PATH>"},
		{"cache rooted at /var/cache", "<CACHE_PATH>"},
	}
	for _, tc := range tests {
		got := MaskAbsolutePaths{}.Apply(tc.in)
		if !strings.Contains(got, tc.want) {
			t.Errorf("%q: expected %q, got: %q", tc.in, tc.want, got)
		}
	}
}

func TestMaskHomePaths(t *testing.T) {
	in := "config at /home/alice/.bashrc"
	got := MaskHomePaths{}.Apply(in)
	if strings.Contains(got, "alice") {
		t.Fatalf("home path not masked: %q", got)
	}
	if !strings.Contains(got, "<HOME_PATH>") {
		t.Fatalf("expected <HOME_PATH>, got: %q", got)
	}
}

func TestMaskHomePaths_Root(t *testing.T) {
	in := "config at /root/.bashrc"
	got := MaskHomePaths{}.Apply(in)
	if strings.Contains(got, "/root") {
		t.Fatalf("/root not masked: %q", got)
	}
}

func TestNormalisePackageNVR(t *testing.T) {
	in := "installing kernel-6.10.7-200.fc40.x86_64"
	got := NormalisePackageNVR{}.Apply(in)
	if !strings.Contains(got, "kernel-<VERSION>") {
		t.Fatalf("package NVR not normalised, got: %q", got)
	}
}

func TestClassifyLines_Error(t *testing.T) {
	got := ClassifyLines{}.Apply("could not load module: fatal error")
	if !strings.HasPrefix(got, "<ERROR>") {
		t.Fatalf("expected <ERROR> prefix, got: %q", got)
	}
}

func TestClassifyLines_Warning(t *testing.T) {
	got := ClassifyLines{}.Apply("Warning: deprecated option")
	if !strings.HasPrefix(got, "<WARN>") {
		t.Fatalf("expected <WARN> prefix, got: %q", got)
	}
}

func TestClassifyLines_Transaction(t *testing.T) {
	got := ClassifyLines{}.Apply("Installed: foo-bar-1.0")
	if !strings.HasPrefix(got, "<TRANSACTION>") {
		t.Fatalf("expected <TRANSACTION> prefix, got: %q", got)
	}
}

func TestClassifyLines_Info(t *testing.T) {
	got := ClassifyLines{}.Apply("starting service")
	if !strings.HasPrefix(got, "<INFO>") {
		t.Fatalf("expected <INFO> prefix, got: %q", got)
	}
}

func TestClassifyLines_ErrorBeforeTransaction(t *testing.T) {
	got := ClassifyLines{}.Apply("installed: foo failed to register")
	if !strings.HasPrefix(got, "<ERROR>") {
		t.Fatalf("error keyword must win over transaction keyword, got: %q", got)
	}
}

func TestClassifyLines_PreservesBlankLines(t *testing.T) {
	in := "first\n\nsecond"
	got := ClassifyLines{}.Apply(in)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	if lines[1] != "" {
		t.Fatalf("blank line not preserved, got: %q", lines[1])
	}
}

func TestCollapseRepeats(t *testing.T) {
	in := "<INFO> a\n<INFO> a\n<INFO> a\n<INFO> b"
	got := CollapseRepeats{}.Apply(in)
	if !strings.Contains(got, "[x3]") {
		t.Fatalf("repeats not collapsed, got: %q", got)
	}
	if !strings.Contains(got, "<INFO> b") {
		t.Fatalf("non-repeated lines lost: %q", got)
	}
}

func TestCollapseRepeats_SingleLine(t *testing.T) {
	got := CollapseRepeats{}.Apply("only one")
	if got != "only one" {
		t.Fatalf("single line corrupted: %q", got)
	}
}

func TestDefault_RuleOrder_PathsBeforeHex(t *testing.T) {
	rules := Default().Rules()
	pathIdx, hexIdx := -1, -1
	for i, r := range rules {
		switch r.Name() {
		case "mask-paths":
			pathIdx = i
		case "mask-hex":
			hexIdx = i
		}
	}
	if pathIdx == -1 || hexIdx == -1 {
		t.Fatalf("expected mask-paths and mask-hex rules in Default, got: %v", ruleNames(rules))
	}
	if pathIdx >= hexIdx {
		t.Fatalf("mask-paths (index %d) must come before mask-hex (index %d)", pathIdx, hexIdx)
	}
}

func TestDefault_RuleOrder_MACBeforeIPv6(t *testing.T) {
	rules := Default().Rules()
	macIdx, ipv6Idx := -1, -1
	for i, r := range rules {
		switch r.Name() {
		case "mask-mac":
			macIdx = i
		case "mask-ipv6":
			ipv6Idx = i
		}
	}
	if macIdx == -1 || ipv6Idx == -1 {
		t.Fatalf("expected mask-mac and mask-ipv6 rules in Default")
	}
	if macIdx >= ipv6Idx {
		t.Fatalf("mask-mac (index %d) must come before mask-ipv6 (index %d)", macIdx, ipv6Idx)
	}
}

func ruleNames(rules []Rule) []string {
	out := make([]string, len(rules))
	for i, r := range rules {
		out[i] = r.Name()
	}
	return out
}

const realDNF5Sample = `2026-06-02T20:08:18+0000 [39054] DEBUG [librepo]   rsync://rpmfusion.ip-connect.info/rpmfusion/free/fedora/updates/44/x86_64/repodata/repomd.xml
2026-06-02T20:08:18+0000 [39054] DEBUG [librepo]   http://ftp-stud.hs-esslingen.de/pub/Mirrors/rpmfusion.org/free/fedora/updates/44/x86_64/repodata/repomd.xml
2026-06-02T20:08:18+0000 [39054] DEBUG [librepo] lr_yum_check_checksum_of_md_record: Checking checksum of /var/cache/libdnf5/rpmfusion-free-updates-880b8bb66393fb1c/repodata/5a17d225ddb72809e540a2677a5a6edf6c3881324b6089547d11939210fa59ca-comps-f44.xml.xz (expected: 5a17d225ddb72809e540a2677a5a6edf6c3881324b6089547d11939210fa59ca [sha256])
2026-06-02T20:08:19+0000 [39054] DEBUG Solvfile's repomd checksum doesn't match, read: "568f3fec82ff0e3ba4123f93376535875c2d57799cf8ba6eb0e0712a101f4936" vs. expected repomd checksum: "617ad626e387f6cef480c7ba0fff614707899d0a5066a55bbb83417ffacebad5" for: /var/cache/libdnf5/updates-3cc07c89a20302f2/solv/updates-updateinfo.solvx
2026-06-02T20:08:20+0000 [39054] INFO DNF5 finished`

func TestDefault_StripsAllSensitiveTokens(t *testing.T) {
	got := Default().Sanitise(realDNF5Sample)
	mustNotContain := []string{
		"2026-06-02T20:08:18+0000",
		"rsync://", "http://", "rpmfusion.ip-connect.info",
		"39054",
		"5a17d225ddb72809", "568f3fec82ff0e3b",
		"880b8bb66393fb1c", "3cc07c89a20302f2",
	}
	for _, token := range mustNotContain {
		if strings.Contains(got, token) {
			t.Errorf("sensitive token %q leaked into sanitised output:\n%s", token, got)
		}
	}
}

func TestDefault_PreservesSemanticStructure(t *testing.T) {
	got := Default().Sanitise(realDNF5Sample)
	mustContain := []string{"<URL>", "<HEX>", "DNF5 finished", "<DEBUG>", "<INFO>"}
	for _, token := range mustContain {
		if !strings.Contains(got, token) {
			t.Errorf("expected sanitised output to contain %q, got:\n%s", token, got)
		}
	}
}

func FuzzSanitise(f *testing.F) {
	seeds := []string{
		"",
		"a",
		"::::::",
		"std::vector::iterator",
		"/etc/passwd",
		"aa:bb:cc:dd:ee:ff",
		"2026-01-01T00:00:00Z [42] ERROR something",
		strings.Repeat("a:", 1000),
	}
	for _, s := range seeds {
		f.Add(s)
	}
	s := Default()
	f.Fuzz(func(t *testing.T, in string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input %q: %v", in, r)
			}
		}()
		_ = s.Sanitise(in)
	})
}
