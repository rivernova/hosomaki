// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import "testing"

func TestFilterErrorLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"whitespace only", "   \n\t\n", ""},
		{"no matching keywords", "started service\nlistening on socket\nready", ""},
		{"keeps error line", "starting up\nconnection failed: timeout\nshutting down", "connection failed: timeout"},
		{"case insensitive", "FATAL: disk full", "FATAL: disk full"},
		{"keeps only matching lines", "ok line\nsegfault in worker\nok again\npanic: nil deref", "segfault in worker\npanic: nil deref"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterErrorLines(tt.in); got != tt.want {
				t.Errorf("filterErrorLines(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
