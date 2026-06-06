// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"
	"time"
)

// unit tests for JSON sentinel presence in prompts and prompt style instructions

func TestStatusFullPromptStyle(t *testing.T) {
	p := Status(StatusInput{CollectedAt: time.Now()}, false)
	if !strings.Contains(p, "five to eight sentences") {
		t.Error("Status() full prompt should instruct five to eight sentences")
	}
	if !strings.Contains(p, "Do not suggest fixes") {
		t.Error("Status() full prompt should forbid suggesting fixes")
	}
}

func TestStatusBriefPromptStyle(t *testing.T) {
	p := Status(StatusInput{CollectedAt: time.Now()}, true)
	if !strings.Contains(p, "ONE sentence") {
		t.Error("Status() brief prompt should instruct exactly one sentence")
	}
	if strings.Contains(p, "five to eight sentences") {
		t.Error("Status() brief prompt should not contain full-mode instruction")
	}
}
