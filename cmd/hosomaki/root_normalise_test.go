// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"reflect"
	"testing"
)

// unit testing for root command normalisation logic

func TestNormaliseNegativeIntFlag(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		flag  string
		want  []string
	}{
		{
			name:  "negative index with space",
			input: []string{"hosomaki", "explain", "--boot", "-1"},
			flag:  "--boot",
			want:  []string{"hosomaki", "explain", "--boot=-1"},
		},
		{
			name:  "negative two-digit index",
			input: []string{"hosomaki", "explain", "--boot", "-10"},
			flag:  "--boot",
			want:  []string{"hosomaki", "explain", "--boot=-10"},
		},
		{
			name:  "positive index unchanged",
			input: []string{"hosomaki", "explain", "--boot", "0"},
			flag:  "--boot",
			want:  []string{"hosomaki", "explain", "--boot", "0"},
		},
		{
			name:  "flag alone unchanged (NoOptDefVal handles it)",
			input: []string{"hosomaki", "explain", "--boot"},
			flag:  "--boot",
			want:  []string{"hosomaki", "explain", "--boot"},
		},
		{
			name:  "already equals form unchanged",
			input: []string{"hosomaki", "explain", "--boot=-1"},
			flag:  "--boot",
			want:  []string{"hosomaki", "explain", "--boot=-1"},
		},
		{
			name:  "other flags not touched",
			input: []string{"hosomaki", "explain", "--service", "nginx"},
			flag:  "--boot",
			want:  []string{"hosomaki", "explain", "--service", "nginx"},
		},
		{
			name:  "non-numeric negative not rewritten",
			input: []string{"hosomaki", "explain", "--boot", "-service"},
			flag:  "--boot",
			want:  []string{"hosomaki", "explain", "--boot", "-service"},
		},
		{
			name:  "flag with other args around it",
			input: []string{"hosomaki", "explain", "--config", "cfg.yaml", "--boot", "-2", "--dmesg"},
			flag:  "--boot",
			want:  []string{"hosomaki", "explain", "--config", "cfg.yaml", "--boot=-2", "--dmesg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normaliseNegativeIntFlag(tt.input, tt.flag)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normaliseNegativeIntFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNegativeInt(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"-1", true},
		{"-10", true},
		{"-99", true},
		{"0", false},
		{"1", false},
		{"-", false},
		{"-a", false},
		{"-1a", false},
		{"--boot", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isNegativeInt(tt.input)
			if got != tt.want {
				t.Errorf("isNegativeInt(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
