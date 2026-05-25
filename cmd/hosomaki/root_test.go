// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecuteExposesVersionFlag(t *testing.T) {
	var out bytes.Buffer

	rootCmd.SetOut(&out)
	rootCmd.SetArgs([]string{"--version"})
	defer rootCmd.SetOut(nil)
	defer rootCmd.SetArgs(nil)

	Execute("test-version")

	got := out.String()
	if !strings.Contains(got, "test-version") {
		t.Fatalf("expected version output to contain %q, got %q", "test-version", got)
	}
}
