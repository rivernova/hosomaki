// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit tests for the explain --pid prompt builder

func TestExplainProcess_ContainsSchema(t *testing.T) {
	p := ExplainProcess("PID: 1\nProcess name: init", 1, collector.Environment{})
	if !strings.Contains(p, SchemaExplain) {
		t.Error("ExplainProcess() must embed the SchemaExplain constant")
	}
}

func TestExplainProcess_ContainsPID(t *testing.T) {
	p := ExplainProcess("PID: 42\nProcess name: nginx", 42, collector.Environment{})
	if !strings.Contains(p, "42") {
		t.Error("ExplainProcess() must reference the PID in the prompt")
	}
}

func TestExplainProcess_ContainsSnapshot(t *testing.T) {
	snapshot := "PID: 99\nProcess name: sshd\nState: S (sleeping)"
	p := ExplainProcess(snapshot, 99, collector.Environment{})
	if !strings.Contains(p, snapshot) {
		t.Error("ExplainProcess() must embed the sanitised snapshot verbatim")
	}
}

func TestExplainProcess_ContainsEnvironmentSection(t *testing.T) {
	env := collector.Environment{
		DistroID:         "ubuntu",
		DistroPrettyName: "Ubuntu 24.04 LTS",
		InitSystem:       "systemd",
		PackageManager:   "apt",
	}
	p := ExplainProcess("PID: 1", 1, env)
	for _, want := range []string{"Ubuntu 24.04 LTS", "systemd", "apt"} {
		if !strings.Contains(p, want) {
			t.Errorf("ExplainProcess() missing environment field %q in prompt", want)
		}
	}
}

func TestExplainProcess_InstructsPureJSON(t *testing.T) {
	p := ExplainProcess("PID: 1", 1, collector.Environment{})
	if !strings.Contains(p, "Return ONLY a JSON object") {
		t.Error("ExplainProcess() must instruct the model to return only JSON")
	}
}

func TestExplainProcess_NoFixSuggestions(t *testing.T) {
	p := ExplainProcess("PID: 1", 1, collector.Environment{})
	if !strings.Contains(p, "Do not suggest actions") {
		t.Error("ExplainProcess() must instruct the model not to suggest fixes")
	}
}

func TestExplainProcess_MentionsProcfsFields(t *testing.T) {
	p := ExplainProcess("PID: 1", 1, collector.Environment{})
	for _, want := range []string{"/proc/", "open file", "socket"} {
		if !strings.Contains(p, want) {
			t.Errorf("ExplainProcess() prompt should describe procfs input format, missing %q", want)
		}
	}
}

func TestExplainProcess_UsesSchemaExplainNotANewSchema(t *testing.T) {
	p := ExplainProcess("PID: 5", 5, collector.Environment{})
	if !strings.Contains(p, `"issues"`) {
		t.Error("ExplainProcess() must produce a prompt that references the 'issues' key from SchemaExplain")
	}
}
