// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"strings"
	"testing"
)

// unit testing for explain

func TestResolveInputMessageArgument(t *testing.T) {
	got, err := resolveInput(resolveParams{
		args: []string{" kernel:", "OOM", "killer "},
	})
	if err != nil {
		t.Fatalf("resolveInput() error = %v", err)
	}

	want := "kernel: OOM killer"
	if got != want {
		t.Fatalf("resolveInput() = %q, want %q", got, want)
	}
}

func TestResolveInputEmptyMessageArgument(t *testing.T) {
	_, err := resolveInput(resolveParams{
		args: []string{" ", "\t"},
	})
	if err == nil {
		t.Fatal("resolveInput() error = nil, want non-empty message error")
	}

	if !strings.Contains(err.Error(), "message was empty") {
		t.Fatalf("resolveInput() error = %q, want message was empty", err)
	}
}
