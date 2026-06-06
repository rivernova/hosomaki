// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"testing"
)

// unit tests for status prompt generation, focused on ensuring the JSON sentinel is present in all prompts and prompt styles

func TestLimitLinesKeepsShortInput(t *testing.T) {
	input := "line 1\nline 2"

	got := limitLines(input, maxTopProcessLines)
	if got != input {
		t.Fatalf("limitLines() = %q, want %q", got, input)
	}
}

func TestLimitLinesKeepsEmptyInput(t *testing.T) {
	if got := limitLines("", maxTopProcessLines); got != "" {
		t.Fatalf("limitLines() = %q, want empty string", got)
	}
}
