// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit test for explain prompt logic

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

func TestExplainDiffIncludesBothBootLabels(t *testing.T) {
	p := ExplainDiff("log from boot -1", "log from boot 0", -1, 0, collector.Environment{})
	if !strings.Contains(p, "previous boot (-1)") {
		t.Error("ExplainDiff() prompt must include the from-boot label")
	}
	if !strings.Contains(p, "current boot (0)") {
		t.Error("ExplainDiff() prompt must include the to-boot label")
	}
}

func TestExplainDiffIncludesBothLogSections(t *testing.T) {
	p := ExplainDiff("FROM_LOG_SENTINEL", "TO_LOG_SENTINEL", -1, 0, collector.Environment{})
	if !strings.Contains(p, "FROM_LOG_SENTINEL") {
		t.Error("ExplainDiff() prompt must include the from-boot logs")
	}
	if !strings.Contains(p, "TO_LOG_SENTINEL") {
		t.Error("ExplainDiff() prompt must include the to-boot logs")
	}
}

func TestExplainDiffRequestsJSON(t *testing.T) {
	p := ExplainDiff("a", "b", -1, 0, collector.Environment{})
	if !strings.Contains(p, `"what"`) {
		t.Error("ExplainDiff() prompt must reference the 'what' JSON field")
	}
	if !strings.Contains(p, `"why"`) {
		t.Error("ExplainDiff() prompt must reference the 'why' JSON field")
	}
	if !strings.Contains(p, "Return ONLY a JSON object") {
		t.Error("ExplainDiff() prompt must instruct the model to return pure JSON")
	}
}

func TestExplainDiffNoFixSuggestions(t *testing.T) {
	p := ExplainDiff("a", "b", -1, 0, collector.Environment{})
	if !strings.Contains(p, "Do not suggest fixes") {
		t.Error("ExplainDiff() prompt must forbid fix suggestions")
	}
}

func TestExplainDiffFocusesOnDifferences(t *testing.T) {
	p := ExplainDiff("a", "b", -1, 0, collector.Environment{})
	if !strings.Contains(p, "changed") {
		t.Error("ExplainDiff() prompt must instruct the model to focus on what changed")
	}
	if !strings.Contains(p, "identical") {
		t.Error("ExplainDiff() prompt must instruct the model to return empty issues when both boots are identical")
	}
}

func TestExplainDiffArbitraryBootIndices(t *testing.T) {
	p := ExplainDiff("log a", "log b", -3, -2, collector.Environment{})
	if !strings.Contains(p, "boot -3") {
		t.Error("ExplainDiff() must include arbitrary from-boot index in prompt")
	}
	if !strings.Contains(p, "boot -2") {
		t.Error("ExplainDiff() must include arbitrary to-boot index in prompt")
	}
}

func TestBootLabelCurrent(t *testing.T) {
	if got := BootLabel(0); got != "current boot (0)" {
		t.Errorf("BootLabel(0) = %q, want %q", got, "current boot (0)")
	}
}

func TestBootLabelPrevious(t *testing.T) {
	if got := BootLabel(-1); got != "previous boot (-1)" {
		t.Errorf("BootLabel(-1) = %q, want %q", got, "previous boot (-1)")
	}
}

func TestBootLabelArbitrary(t *testing.T) {
	if got := BootLabel(-3); got != "boot -3" {
		t.Errorf("BootLabel(-3) = %q, want %q", got, "boot -3")
	}
}
