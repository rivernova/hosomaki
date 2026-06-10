// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// unit tests for the watch command

func TestWatchCmdRegistered(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "watch <service>" {
			found = true
			break
		}
	}
	if !found {
		t.Error("watch command is not registered on the root command")
	}
}

func TestWatchCmdHasLinesFlag(t *testing.T) {
	cmd := newWatchCmd()
	f := cmd.Flags().Lookup("lines")
	if f == nil {
		t.Fatal("watch command is missing the --lines flag")
	}
	if f.DefValue != "20" {
		t.Errorf("--lines default = %q, want '20'", f.DefValue)
	}
}

func TestWatchCmdHasWindowFlag(t *testing.T) {
	cmd := newWatchCmd()
	f := cmd.Flags().Lookup("window")
	if f == nil {
		t.Fatal("watch command is missing the --window flag")
	}
	if f.DefValue != "3s" {
		t.Errorf("--window default = %q, want '3s'", f.DefValue)
	}
}

func TestWatchCmdHasMaxLinesFlag(t *testing.T) {
	cmd := newWatchCmd()
	f := cmd.Flags().Lookup("max-lines")
	if f == nil {
		t.Fatal("watch command is missing the --max-lines flag")
	}
	if f.DefValue != "50" {
		t.Errorf("--max-lines default = %q, want '50'", f.DefValue)
	}
}

func TestWatchCmdHasDebugFlag(t *testing.T) {
	cmd := newWatchCmd()
	f := cmd.Flags().Lookup("debug")
	if f == nil {
		t.Fatal("watch command is missing the --debug flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--debug default = %q, want 'false'", f.DefValue)
	}
}

func TestWatchCmdHasNoWatchFlag(t *testing.T) {
	cmd := newWatchCmd()
	if f := cmd.Flags().Lookup("watch"); f != nil {
		t.Error("watch command must not register a --watch flag")
	}
}

func TestWatchCmdRequiresExactlyOneArg(t *testing.T) {
	cmd := newWatchCmd()

	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("watch command must require at least one positional argument")
	}
	if err := cmd.Args(cmd, []string{"nginx", "extra"}); err == nil {
		t.Error("watch command must reject more than one positional argument")
	}
	if err := cmd.Args(cmd, []string{"nginx"}); err != nil {
		t.Errorf("watch command must accept exactly one positional argument, got error: %v", err)
	}
}

func TestWatchCmdShortDescription(t *testing.T) {
	if newWatchCmd().Short == "" {
		t.Error("watch command must have a non-empty Short description")
	}
}

func TestWatchCmdLongContainsKeyPhrases(t *testing.T) {
	long := newWatchCmd().Long
	for _, phrase := range []string{
		"Ctrl-C",
		"silence window",
		"never modifies",
		"read-only",
		"sanitisation",
	} {
		if !strings.Contains(long, phrase) {
			t.Errorf("watch Long help text is missing expected phrase %q", phrase)
		}
	}
}

func TestWatchCmdHelp_DoesNotPanic(t *testing.T) {
	cmd := newWatchCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("help panicked: %v", r)
		}
	}()
	_ = cmd.Help()
}

func TestWatchCmdDefaultWindowIsPositive(t *testing.T) {
	cmd := newWatchCmd()
	f := cmd.Flags().Lookup("window")
	if f == nil {
		t.Fatal("missing --window flag")
	}
	d, err := time.ParseDuration(f.DefValue)
	if err != nil {
		t.Fatalf("--window default is not a valid duration: %q", f.DefValue)
	}
	if d <= 0 {
		t.Errorf("--window default must be positive, got %v", d)
	}
}

func TestWatchCmdDefaultMaxLinesIsPositive(t *testing.T) {
	cmd := newWatchCmd()
	f := cmd.Flags().Lookup("max-lines")
	if f == nil {
		t.Fatal("missing --max-lines flag")
	}
	if f.DefValue == "0" || f.DefValue == "" {
		t.Errorf("--max-lines default must be positive, got %q", f.DefValue)
	}
}
