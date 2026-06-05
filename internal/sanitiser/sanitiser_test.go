// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sanitiser

import (
	"strings"
	"testing"
)

// uni tests for the sanitiser

func TestSanitise_EmptyInput(t *testing.T) {
	if got := Default().Sanitise(""); got != "" {
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

func TestSanitise_NoOpSanitiser(t *testing.T) {
	const input = "any text whatsoever"
	if got := New().Sanitise(input); got != input {
		t.Fatalf("no-op sanitiser changed input: %q -> %q", input, got)
	}
}

func TestStripTimestamps_ISO8601(t *testing.T) {
	in := "2026-06-02T20:08:18+0000 hello"
	want := "hello"
	got := strings.TrimSpace(StripTimestamps{}.Apply(in))
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
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

func TestMaskMACAddresses(t *testing.T) {
	in := "device aa:bb:cc:dd:ee:ff registered"
	got := MaskMACAddresses{}.Apply(in)
	if strings.Contains(got, "aa:bb:cc:dd:ee:ff") {
		t.Fatalf("MAC not masked: %q", got)
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
	in := "version abc1"
	got := MaskHexAddresses{}.Apply(in)
	if !strings.Contains(got, "abc1") {
		t.Fatalf("short hex should not be masked, got: %q", got)
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

func TestNormalisePackageNVR(t *testing.T) {
	in := "installing kernel-6.10.7-200.fc40.x86_64"
	got := NormalisePackageNVR{}.Apply(in)
	if !strings.Contains(got, "kernel-<VERSION>") {
		t.Fatalf("package NVR not normalised, got: %q", got)
	}
}

func TestCollapseRepeats(t *testing.T) {
	in := "<INFO> a\n<INFO> a\n<INFO> a\n<INFO> b"
	got := CollapseRepeats{}.Apply(in)
	if !strings.Contains(got, "(repeated 3 times)") {
		t.Fatalf("repeats not collapsed, got: %q", got)
	}
	if !strings.Contains(got, "<INFO> b") {
		t.Fatalf("non-repeated lines lost: %q", got)
	}
}

func TestCollapseRepeats_SingleLine(t *testing.T) {
	in := "only one"
	got := CollapseRepeats{}.Apply(in)
	if got != "only one" {
		t.Fatalf("single line corrupted: %q", got)
	}
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

	mustContain := []string{
		"<URL>", "<HEX>", "DNF5 finished",
		"<DEBUG>", "<INFO>",
	}
	for _, token := range mustContain {
		if !strings.Contains(got, token) {
			t.Errorf("expected sanitised output to contain %q, got:\n%s", token, got)
		}
	}
}
