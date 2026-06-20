// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit test for history prompt logic

func TestHistoryPrompt_ContainsSchema(t *testing.T) {
	in := HistoryInput{
		Environment: collector.Environment{DistroID: "test"},
		History:     "",
		FilterDesc:  "all entries",
	}
	result := History(in)
	if !strings.Contains(result, SchemaHistory) {
		t.Error("prompt should contain SchemaHistory")
	}
}

func TestHistoryPrompt_ContainsEnvironment(t *testing.T) {
	in := HistoryInput{
		Environment: collector.Environment{DistroID: "testos", PackageManager: "apt"},
		History:     "",
		FilterDesc:  "",
	}
	result := History(in)
	if !strings.Contains(result, "testos") {
		t.Error("prompt should contain environment info")
	}
}

func TestHistoryPrompt_EmptyHistory(t *testing.T) {
	in := HistoryInput{
		Environment: collector.Environment{DistroID: "test"},
		History:     "",
		FilterDesc:  "test filter",
	}
	result := History(in)
	if !strings.Contains(result, "(no previous diagnostic results found)") {
		t.Error("prompt should say '(no previous diagnostic results found)' when History is empty")
	}
}

func TestHistoryPrompt_ContainsEntries(t *testing.T) {
	in := HistoryInput{
		Environment: collector.Environment{DistroID: "test"},
		History:     "1. [2026-06-19T20:00:00Z] explain: nginx was down",
		FilterDesc:  "last 1 entry",
	}
	result := History(in)
	if !strings.Contains(result, "nginx") {
		t.Error("prompt should contain entry text 'nginx'")
	}
	if !strings.Contains(result, "last 1 entry") {
		t.Error("prompt should contain filter description")
	}
}

func TestHistoryPrompt_FilterDesc(t *testing.T) {
	in := HistoryInput{
		Environment: collector.Environment{DistroID: "test"},
		History:     "",
		FilterDesc:  "explain entries from last 7 days",
	}
	result := History(in)
	if !strings.Contains(result, "explain entries from last 7 days") {
		t.Error("prompt should contain the filter description")
	}
}

func TestHistoryPrompt_DefaultFilterDesc(t *testing.T) {
	in := HistoryInput{
		Environment: collector.Environment{DistroID: "test"},
		History:     "some entries",
		FilterDesc:  "",
	}
	result := History(in)
	if !strings.Contains(result, "all available entries") {
		t.Error("prompt should use 'all available entries' when FilterDesc is empty")
	}
}
