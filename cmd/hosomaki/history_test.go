// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"strings"
	"testing"
)

func TestNewHistoryCmd(t *testing.T) {
	cmd := newHistoryCmd()
	if cmd.Use != "history" {
		t.Errorf("expected Use 'history', got %q", cmd.Use)
	}
	if !strings.Contains(cmd.Short, "Review") {
		t.Error("expected Short to contain 'Review'")
	}
}

func TestHistoryFlags(t *testing.T) {
	cmd := newHistoryCmd()
	flags := cmd.Flags()

	if f := flags.Lookup("limit"); f == nil {
		t.Error("expected --limit flag")
	}
	if f := flags.Lookup("since"); f == nil {
		t.Error("expected --since flag")
	}
	if f := flags.Lookup("command"); f == nil {
		t.Error("expected --command flag")
	}
	if f := flags.Lookup("clear"); f == nil {
		t.Error("expected --clear flag")
	}
}

func TestHistoryNoArgs(t *testing.T) {
	cmd := newHistoryCmd()
	if cmd.Args == nil {
		t.Error("expected Args to be set (NoArgs)")
	}
}
