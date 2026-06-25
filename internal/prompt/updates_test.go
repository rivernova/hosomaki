// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit tests for the updates prompt builder

func TestUpdatesPrompt_ContainsSchema(t *testing.T) {
	in := UpdatesInput{
		Environment: collector.Environment{DistroID: "test"},
		Updates:     "",
	}
	result := Updates(in)
	if !strings.Contains(result, SchemaUpdates) {
		t.Error("prompt should contain SchemaUpdates")
	}
}

func TestUpdatesPrompt_RebootRequiredField(t *testing.T) {
	if strings.Contains(SchemaUpdates, `"reboot_required":false`) {
		t.Error("SchemaUpdates should use bool type annotation, not false literal")
	}
	if !strings.Contains(SchemaUpdates, `"reboot_required":bool`) {
		t.Error("SchemaUpdates should declare reboot_required as bool")
	}
}

func TestUpdatesPrompt_ContainsEnvironment(t *testing.T) {
	in := UpdatesInput{
		Environment: collector.Environment{DistroID: "testos"},
		Updates:     "",
	}
	result := Updates(in)
	if !strings.Contains(result, "testos") {
		t.Error("prompt should contain environment string")
	}
}

func TestUpdatesPrompt_NoPendingText(t *testing.T) {
	in := UpdatesInput{
		Environment: collector.Environment{DistroID: "test"},
		Updates:     "",
	}
	result := Updates(in)
	if !strings.Contains(result, "(no pending updates)") {
		t.Error("prompt should say '(no pending updates)' when Updates is empty")
	}
}

func TestUpdatesPrompt_WithUpdates(t *testing.T) {
	in := UpdatesInput{
		Environment: collector.Environment{DistroID: "test"},
		Updates:     "1. nginx  installed: 1.22 -> available: 1.24",
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
		Environment:  collector.Environment{DistroID: "test"},
		Updates:      "",
		SecurityOnly: false,
	}
	result := Updates(in)
	if strings.Contains(result, "Only security-related") {
		t.Error("prompt should not contain security-only note when not requested")
	}
}

func TestUpdatesPrompt_HasFilterNoteWhenSecurityOnly(t *testing.T) {
	in := UpdatesInput{
		Environment:  collector.Environment{DistroID: "test"},
		Updates:      "",
		SecurityOnly: true,
	}
	result := Updates(in)
	if !strings.Contains(result, "security-only") {
		t.Error("prompt should contain security-only note when --security-only is set")
	}
}

func TestUpdatesPrompt_InstalledAndAvailableMustBeCopiedVerbatim(t *testing.T) {
	in := UpdatesInput{
		Environment: collector.Environment{DistroID: "test"},
		Updates:     "",
	}
	result := Updates(in)
	if !strings.Contains(result, "verbatim") {
		t.Error("prompt must instruct the model to copy installed/available verbatim")
	}
}

func TestUpdatesPrompt_DetailRequiredForSecurityAndMajor(t *testing.T) {
	in := UpdatesInput{
		Environment: collector.Environment{DistroID: "test"},
		Updates:     "",
	}
	result := Updates(in)
	if !strings.Contains(result, "REQUIRED") {
		t.Error("prompt must explicitly require detail for security/major findings")
	}
}
