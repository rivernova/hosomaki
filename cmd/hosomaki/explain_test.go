// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit testing for explain command setup

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

func TestResolveInputSinceWithoutJournalSource(t *testing.T) {
	cases := []struct {
		name   string
		params resolveParams
	}{
		{
			name:   "since with dmesg",
			params: resolveParams{dmesg: true, opts: collector.LogOptions{Since: "1 hour ago"}},
		},
		{
			name:   "since with file",
			params: resolveParams{file: "/var/log/syslog", opts: collector.LogOptions{Since: "1 hour ago"}},
		},
		{
			name:   "since with positional arg only",
			params: resolveParams{args: []string{"some error"}, opts: collector.LogOptions{Since: "1 hour ago"}},
		},
		{
			name:   "until with dmesg",
			params: resolveParams{dmesg: true, opts: collector.LogOptions{Until: "now"}},
		},
		{
			name:   "since and until with no source",
			params: resolveParams{opts: collector.LogOptions{Since: "1 hour ago", Until: "now"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := resolveInput(tc.params)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "--since") && !strings.Contains(err.Error(), "--until") {
				t.Fatalf("error should mention --since or --until, got: %q", err.Error())
			}
		})
	}
}

func TestResolveInputSinceUnquotedValueDetected(t *testing.T) {
	_, err := resolveInput(resolveParams{
		service: "nginx",
		opts:    collector.LogOptions{Since: "1"},
		args:    []string{"hour", "ago"},
	})
	if err == nil {
		t.Fatal("expected error for unquoted --since value, got nil")
	}
	if !strings.Contains(err.Error(), "quote") {
		t.Fatalf("error should hint at quoting, got: %q", err.Error())
	}
}

func TestResolveInputUntilUnquotedValueDetected(t *testing.T) {
	_, err := resolveInput(resolveParams{
		bootChanged: true,
		boot:        "0",
		opts:        collector.LogOptions{Until: "30"},
		args:        []string{"min", "ago"},
	})
	if err == nil {
		t.Fatal("expected error for unquoted --until value, got nil")
	}
	if !strings.Contains(err.Error(), "quote") {
		t.Fatalf("error should hint at quoting, got: %q", err.Error())
	}
}

func TestResolveInputSinceWithService(t *testing.T) {
	_, err := resolveInput(resolveParams{
		service: "nonexistent-service-xyz",
		opts:    collector.LogOptions{Since: "1 hour ago"},
	})
	if err != nil && strings.Contains(err.Error(), "--since") {
		t.Fatalf("should not get --since validation error for --service source, got: %q", err.Error())
	}
}

func TestResolveInputUntilWithBoot(t *testing.T) {
	_, err := resolveInput(resolveParams{
		bootChanged: true,
		boot:        "0",
		opts:        collector.LogOptions{Until: "now"},
	})
	if err != nil && strings.Contains(err.Error(), "--until") {
		t.Fatalf("should not get --until validation error for --boot source, got: %q", err.Error())
	}
}

func TestExplainCmdHasSinceFlag(t *testing.T) {
	cmd := newExplainCmd()
	f := cmd.Flags().Lookup("since")
	if f == nil {
		t.Fatal("explain command is missing the --since flag")
	}
	if f.DefValue != "" {
		t.Errorf("--since default = %q, want empty string", f.DefValue)
	}
}

func TestExplainCmdHasUntilFlag(t *testing.T) {
	cmd := newExplainCmd()
	f := cmd.Flags().Lookup("until")
	if f == nil {
		t.Fatal("explain command is missing the --until flag")
	}
	if f.DefValue != "" {
		t.Errorf("--until default = %q, want empty string", f.DefValue)
	}
}
