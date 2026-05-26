// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"
)

func TestExplainWithoutCmd(t *testing.T) {
	p := Explain("some error output", "")
	if strings.Contains(p, "produced by running") {
		t.Error("Explain() with empty cmd should not include command context line")
	}
	if !strings.Contains(p, "some error output") {
		t.Error("Explain() should include the input")
	}
}

func TestExplainWithCmd(t *testing.T) {
	p := Explain("no configuration file provided: not found", "docker compose up")
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
	p := Explain("some error", "   ")
	if strings.Contains(p, "produced by running") {
		t.Error("Explain() with whitespace-only cmd should not include command context line")
	}
}
