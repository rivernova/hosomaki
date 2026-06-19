// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

func TestUpdatesPrompt_ContainsSchema(t *testing.T) {
	in := UpdatesInput{
		Environment:    collector.Environment{DistroID: "test", PackageManager: "apt"},
		PendingUpdates: []collector.PendingUpdate{},
	}
	result := Updates(in)
	if !strings.Contains(result, SchemaUpdates) {
		t.Error("prompt should contain SchemaUpdates")
	}
}

func TestUpdatesPrompt_ContainsEnvironment(t *testing.T) {
	in := UpdatesInput{
		Environment:    collector.Environment{DistroID: "testos", PackageManager: "apt"},
		PendingUpdates: []collector.PendingUpdate{},
	}
	result := Updates(in)
	if !strings.Contains(result, "testos") {
		t.Error("prompt should contain environment distro ID")
	}
}

func TestUpdatesPrompt_NoPendingText(t *testing.T) {
	in := UpdatesInput{
		Environment:    collector.Environment{PackageManager: "apt"},
		PendingUpdates: []collector.PendingUpdate{},
	}
	result := Updates(in)
	if !strings.Contains(result, "(no pending updates)") {
		t.Error("prompt should say '(no pending updates)' when list is empty")
	}
}

func TestUpdatesPrompt_WithUpdates(t *testing.T) {
	in := UpdatesInput{
		Environment: collector.Environment{PackageManager: "apt"},
		PendingUpdates: []collector.PendingUpdate{
			{Package: "nginx", Installed: "1.22", Available: "1.24"},
		},
	}
	result := Updates(in)
	if !strings.Contains(result, "nginx") {
		t.Error("prompt should contain package name 'nginx'")
	}
	if !strings.Contains(result, "1.22") {
		t.Error("prompt should contain installed version")
	}
	if !strings.Contains(result, "1.24") {
		t.Error("prompt should contain available version")
	}
}

func TestUpdatesPrompt_NoFilterNoteByDefault(t *testing.T) {
	in := UpdatesInput{
		Environment:    collector.Environment{PackageManager: "apt"},
		PendingUpdates: []collector.PendingUpdate{},
		SecurityOnly:   false,
	}
	result := Updates(in)
	if strings.Contains(result, "Only security-related") {
		t.Error("prompt should not contain security-only note when not requested")
	}
}

func TestUpdatesPrompt_HasFilterNoteWhenSecurityOnly(t *testing.T) {
	in := UpdatesInput{
		Environment:    collector.Environment{PackageManager: "apt"},
		PendingUpdates: []collector.PendingUpdate{},
		SecurityOnly:   true,
	}
	result := Updates(in)
	if !strings.Contains(result, "security-only") {
		t.Error("prompt should contain security-only note when --security-only is set")
	}
}

func TestUpdatesPrompt_UpdatesListedInOrder(t *testing.T) {
	in := UpdatesInput{
		Environment: collector.Environment{PackageManager: "apt"},
		PendingUpdates: []collector.PendingUpdate{
			{Package: "aaa", Available: "1.0"},
			{Package: "bbb", Available: "2.0"},
		},
	}
	result := Updates(in)
	if !strings.Contains(result, "1. aaa") {
		t.Error("first package should be numbered '1.'")
	}
	if !strings.Contains(result, "2. bbb") {
		t.Error("second package should be numbered '2.'")
	}
}

func TestUpdatesPrompt_FormatPendingUpdatesEmpty(t *testing.T) {
	result := formatPendingUpdates(nil)
	if result != "(no pending updates)" {
		t.Errorf("nil input should return '(no pending updates)', got %q", result)
	}

	result = formatPendingUpdates([]collector.PendingUpdate{})
	if result != "(no pending updates)" {
		t.Errorf("empty input should return '(no pending updates)', got %q", result)
	}
}

func TestUpdatesPrompt_FormatPendingUpdatesSecurityTag(t *testing.T) {
	updates := []collector.PendingUpdate{
		{Package: "openssl", Available: "3.0", Security: true},
	}
	result := formatPendingUpdates(updates)
	if !strings.Contains(result, "[SECURITY]") {
		t.Error("security update should have [SECURITY] tag")
	}
}

func TestUpdatesPrompt_FormatPendingUpdatesRebootTag(t *testing.T) {
	updates := []collector.PendingUpdate{
		{Package: "linux-image-x86", Available: "6.1", RebootRequired: true},
	}
	result := formatPendingUpdates(updates)
	if !strings.Contains(result, "[REBOOT]") {
		t.Error("reboot-required update should have [REBOOT] tag")
	}
}

func TestUpdatesPrompt_FormatPendingUpdatesUnknownInstalled(t *testing.T) {
	updates := []collector.PendingUpdate{
		{Package: "pkg", Available: "2.0"}, // Installed is ""
	}
	result := formatPendingUpdates(updates)
	if !strings.Contains(result, "(unknown)") {
		t.Error("update with empty Installed should show '(unknown)'")
	}
}
