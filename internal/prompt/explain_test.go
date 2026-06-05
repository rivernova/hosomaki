// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit test for explain

func TestExplainWithoutCmd(t *testing.T) {
	p := Explain("some error output", "", collector.Environment{})
	if strings.Contains(p, "produced by running") {
		t.Error("Explain() with empty cmd should not include command context line")
	}
	if !strings.Contains(p, "some error output") {
		t.Error("Explain() should include the input")
	}
}

func TestExplainWithCmd(t *testing.T) {
	p := Explain("no configuration file provided: not found", "docker compose up", collector.Environment{})
	if !strings.Contains(p, "docker compose up") {
		t.Error("Explain() with cmd should include the command")
	}
	if !strings.Contains(p, "produced by running") {
		t.Error("Explain() with cmd should include context sentence")
	}
	if !strings.Contains(p, "no configuration file provided: not found") {
		t.Error("Explain() should include the input")
	}
}

func TestExplainCmdWhitespaceOnlyIgnored(t *testing.T) {
	p := Explain("some error", "   ", collector.Environment{})
	if strings.Contains(p, "produced by running") {
		t.Error("Explain() with whitespace-only cmd should not include command context line")
	}
}

func TestExplainIncludesEnvironmentContext(t *testing.T) {
	env := collector.Environment{
		DistroID:         "fedora",
		DistroPrettyName: "Fedora Linux 40",
		PackageManager:   "dnf",
		InitSystem:       "systemd",
	}
	p := Explain("some error", "", env)
	for _, want := range []string{"Fedora Linux 40", "dnf", "systemd"} {
		if !strings.Contains(p, want) {
			t.Errorf("Explain() should include environment field %q", want)
		}
	}
}

func TestExplainInstructsModelToUseEnvironment(t *testing.T) {
	p := Explain("some error", "", collector.Environment{DistroID: "arch"})
	if !strings.Contains(p, "Host environment") {
		t.Error("Explain() should instruct the model to use the host environment")
	}
}

func TestExplainPromptRequestsJSON(t *testing.T) {
	p := Explain("some error", "", collector.Environment{})
	if !strings.Contains(p, `"what"`) {
		t.Error("Explain() prompt must reference the 'what' JSON field")
	}
	if !strings.Contains(p, `"why"`) {
		t.Error("Explain() prompt must reference the 'why' JSON field")
	}
}

func TestExplainPromptPureJSONInstruction(t *testing.T) {
	p := Explain("some error", "", collector.Environment{})
	if !strings.Contains(p, "Return ONLY a JSON object") {
		t.Error("Explain() prompt must instruct the model to return pure JSON")
	}
}

func TestExplainPromptNoFixSuggestions(t *testing.T) {
	p := Explain("some error", "", collector.Environment{})
	if !strings.Contains(p, "Do not suggest fixes") {
		t.Error("Explain() prompt must forbid fix suggestions")
	}
}
