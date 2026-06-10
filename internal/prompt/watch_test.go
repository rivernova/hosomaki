// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit tests for the watch prompt builder

func makeWatchInput(batch string) WatchInput {
	return WatchInput{
		Service:     "nginx.service",
		Batch:       batch,
		Environment: collector.Environment{DistroID: "debian", InitSystem: "systemd"},
	}
}

func TestWatch_ContainsSchema(t *testing.T) {
	p := Watch(makeWatchInput("<ERROR> connection refused"))
	if !strings.Contains(p, SchemaWatch) {
		t.Error("Watch() prompt must contain the schema constant")
	}
}

func TestWatch_ContainsServiceName(t *testing.T) {
	p := Watch(makeWatchInput("<ERROR> something"))
	if !strings.Contains(p, "nginx.service") {
		t.Error("Watch() prompt must contain the service name")
	}
}

func TestWatch_ContainsEnvironmentSection(t *testing.T) {
	p := Watch(makeWatchInput("<ERROR> something"))
	if !strings.Contains(p, "Host environment") {
		t.Error("Watch() prompt must contain the environment section")
	}
}

func TestWatch_ContainsBatch(t *testing.T) {
	batch := "<ERROR> disk full\n<WARN> inode exhaustion"
	p := Watch(makeWatchInput(batch))
	if !strings.Contains(p, batch) {
		t.Error("Watch() prompt must embed the batch text verbatim")
	}
}

func TestWatch_InstructsNoMarkdown(t *testing.T) {
	p := Watch(makeWatchInput("<ERROR> test"))
	if !strings.Contains(p, "no markdown") && !strings.Contains(p, "no prose") {
		t.Error("Watch() prompt must instruct the model to return only JSON")
	}
}

func TestWatch_InstructsNoFixes(t *testing.T) {
	p := Watch(makeWatchInput("<ERROR> test"))
	if !strings.Contains(p, "Do not suggest fixes") {
		t.Error("Watch() prompt must forbid fix suggestions")
	}
}

func TestWatch_InstructsEmptyOnNoAlerts(t *testing.T) {
	p := Watch(makeWatchInput("<INFO> all is well"))
	if !strings.Contains(p, `{"issues": []}`) {
		t.Error(`Watch() prompt must instruct model to return {"issues": []} when no alerts`)
	}
}

func TestWatch_InstructsNoDuplication(t *testing.T) {
	p := Watch(makeWatchInput("<ERROR> test"))
	if !strings.Contains(p, "not repeat") && !strings.Contains(p, "Do not repeat") {
		t.Error("Watch() prompt must tell the model not to repeat the same issue")
	}
}

func TestWatch_InstructsConciseness(t *testing.T) {
	p := Watch(makeWatchInput("<ERROR> test"))
	if !strings.Contains(p, "concise") {
		t.Error("Watch() prompt must instruct the model to be concise")
	}
}

func TestWatch_MentionsBothWhatAndWhy(t *testing.T) {
	p := Watch(makeWatchInput("<ERROR> test"))
	if !strings.Contains(p, `"what"`) || !strings.Contains(p, `"why"`) {
		t.Error("Watch() prompt must describe both 'what' and 'why' fields")
	}
}

func TestWatch_MentionsPlaceholders(t *testing.T) {
	p := Watch(makeWatchInput("<ERROR> test"))
	for _, ph := range []string{"<ERROR>", "<WARN>", "<IPV4>", "<PATH>"} {
		if !strings.Contains(p, ph) {
			t.Errorf("Watch() prompt must document placeholder %q", ph)
		}
	}
}

func TestWatchInput_DifferentServices(t *testing.T) {
	in := WatchInput{
		Service:     "postgresql.service",
		Batch:       "<ERROR> could not connect to server",
		Environment: collector.Environment{},
	}
	p := Watch(in)
	if !strings.Contains(p, "postgresql.service") {
		t.Error("Watch() must include the correct service name")
	}
}

func TestWatch_EmptyBatchStillBuildsValidPrompt(t *testing.T) {
	in := WatchInput{
		Service:     "nginx.service",
		Batch:       "",
		Environment: collector.Environment{},
	}
	p := Watch(in)
	if !strings.Contains(p, SchemaWatch) {
		t.Error("Watch() with empty batch must still produce a valid prompt with schema")
	}
}
