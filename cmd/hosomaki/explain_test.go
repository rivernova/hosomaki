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

func TestExplainCmdHasContextFlag(t *testing.T) {
	cmd := newExplainCmd()
	f := cmd.Flags().Lookup("context")
	if f == nil {
		t.Fatal("explain command is missing the --context flag")
	}
	if f.DefValue != "" {
		t.Errorf("--context default = %q, want empty string", f.DefValue)
	}
}

func TestResolveInputContextRequiresAtLeastTwoServices(t *testing.T) {
	_, err := resolveInput(resolveParams{
		contexts: []string{"nginx"},
	})
	if err == nil {
		t.Fatal("expected error for single-service --context, got nil")
	}
	if !strings.Contains(err.Error(), "at least 2") {
		t.Fatalf("error should mention 'at least 2', got: %q", err.Error())
	}
}

func TestResolveInputContextMutuallyExclusiveWithService(t *testing.T) {
	_, err := resolveInput(resolveParams{
		service:  "nginx",
		contexts: []string{"mongodb", "rabbitmq"},
	})
	if err == nil {
		t.Fatal("expected error when --context and --service are combined, got nil")
	}
	if !strings.Contains(err.Error(), "--context") || !strings.Contains(err.Error(), "--service") {
		t.Fatalf("error should mention both --context and --service, got: %q", err.Error())
	}
}

func TestResolveInputContextMutuallyExclusiveWithDmesg(t *testing.T) {
	_, err := resolveInput(resolveParams{
		dmesg:    true,
		contexts: []string{"nginx", "mongodb"},
	})
	if err == nil {
		t.Fatal("expected error when --context and --dmesg are combined, got nil")
	}
}

func TestResolveInputSinceAllowedWithContext(t *testing.T) {
	_, err := resolveInput(resolveParams{
		contexts: []string{"nonexistent-a", "nonexistent-b"},
		opts:     collector.LogOptions{Since: "1 hour ago"},
	})
	if err != nil && strings.Contains(err.Error(), "--since") {
		t.Fatalf("should not get --since validation error for --context source, got: %q", err.Error())
	}
}

func TestResolveInputUntilAllowedWithContext(t *testing.T) {
	_, err := resolveInput(resolveParams{
		contexts: []string{"nonexistent-a", "nonexistent-b"},
		opts:     collector.LogOptions{Until: "now"},
	})
	if err != nil && strings.Contains(err.Error(), "--until") {
		t.Fatalf("should not get --until validation error for --context source, got: %q", err.Error())
	}
}

func TestResolveInputSinceAndUntilAllowedWithContext(t *testing.T) {
	_, err := resolveInput(resolveParams{
		contexts: []string{"nonexistent-a", "nonexistent-b"},
		opts:     collector.LogOptions{Since: "2024-01-15 14:00:00", Until: "2024-01-15 15:00:00"},
	})
	if err != nil && (strings.Contains(err.Error(), "--since") || strings.Contains(err.Error(), "--until")) {
		t.Fatalf("should not get time validation error for --context source, got: %q", err.Error())
	}
}

func TestResolveInputSinceUnquotedValueDetectedWithContext(t *testing.T) {
	_, err := resolveInput(resolveParams{
		contexts: []string{"nginx", "mongodb"},
		opts:     collector.LogOptions{Since: "30"},
		args:     []string{"min", "ago"},
	})
	if err == nil {
		t.Fatal("expected error for unquoted --since value with --context, got nil")
	}
	if !strings.Contains(err.Error(), "quote") {
		t.Fatalf("error should hint at quoting, got: %q", err.Error())
	}
}

func TestResolveSourceLabelContext(t *testing.T) {
	label := resolveSourceLabel(resolveParams{
		contexts: []string{"nginx", "mongodb", "rabbitmq"},
	})
	want := "context: nginx, mongodb, rabbitmq"
	if label != want {
		t.Errorf("resolveSourceLabel() = %q, want %q", label, want)
	}
}

func TestResolveSourceLabelContextTwoServices(t *testing.T) {
	label := resolveSourceLabel(resolveParams{
		contexts: []string{"nginx", "postgresql"},
	})
	want := "context: nginx, postgresql"
	if label != want {
		t.Errorf("resolveSourceLabel() = %q, want %q", label, want)
	}
}

func TestParseBootDiffSingleIndex(t *testing.T) {
	d, err := parseBootDiff("-1")
	if err != nil {
		t.Fatalf("parseBootDiff(-1) error = %v", err)
	}
	if d.from != -1 || d.to != 0 {
		t.Errorf("parseBootDiff(-1) = {%d, %d}, want {-1, 0}", d.from, d.to)
	}
}

func TestParseBootDiffColonForm(t *testing.T) {
	d, err := parseBootDiff("-2:-1")
	if err != nil {
		t.Fatalf("parseBootDiff(-2:-1) error = %v", err)
	}
	if d.from != -2 || d.to != -1 {
		t.Errorf("parseBootDiff(-2:-1) = {%d, %d}, want {-2, -1}", d.from, d.to)
	}
}

func TestParseBootDiffEmpty(t *testing.T) {
	d, err := parseBootDiff("")
	if err != nil {
		t.Fatalf("parseBootDiff('') error = %v", err)
	}
	if d != nil {
		t.Errorf("parseBootDiff('') should return nil, got %+v", d)
	}
}

func TestParseBootDiffSameIndexRejected(t *testing.T) {
	_, err := parseBootDiff("0")
	if err == nil {
		t.Fatal("parseBootDiff('0') should error: from and to are identical")
	}
}

func TestParseBootDiffSameIndexColonRejected(t *testing.T) {
	_, err := parseBootDiff("-1:-1")
	if err == nil {
		t.Fatal("parseBootDiff('-1:-1') should error: from and to are identical")
	}
}

func TestParseBootDiffInvalidString(t *testing.T) {
	_, err := parseBootDiff("abc")
	if err == nil {
		t.Fatal("parseBootDiff('abc') should error")
	}
}

func TestParseBootDiffInvalidColonForm(t *testing.T) {
	_, err := parseBootDiff("-1:abc")
	if err == nil {
		t.Fatal("parseBootDiff('-1:abc') should error")
	}
}

func TestExplainCmdHasDiffFlag(t *testing.T) {
	cmd := newExplainCmd()
	f := cmd.Flags().Lookup("diff")
	if f == nil {
		t.Fatal("explain command is missing the --diff flag")
	}
	if f.DefValue != "" {
		t.Errorf("--diff default = %q, want empty string", f.DefValue)
	}
}

func TestResolveInputDiffMutuallyExclusiveWithService(t *testing.T) {
	d := &bootDiff{from: -1, to: 0}
	_, err := resolveInput(resolveParams{
		service: "nginx",
		diff:    d,
	})
	if err == nil {
		t.Fatal("expected error when --diff and --service are combined, got nil")
	}
	if !strings.Contains(err.Error(), "--diff") || !strings.Contains(err.Error(), "--service") {
		t.Fatalf("error should mention both flags, got: %q", err.Error())
	}
}

func TestResolveInputDiffMutuallyExclusiveWithDmesg(t *testing.T) {
	d := &bootDiff{from: -1, to: 0}
	_, err := resolveInput(resolveParams{
		dmesg: true,
		diff:  d,
	})
	if err == nil {
		t.Fatal("expected error when --diff and --dmesg are combined, got nil")
	}
}

func TestResolveInputDiffSinceRejected(t *testing.T) {
	d := &bootDiff{from: -1, to: 0}
	_, err := resolveInput(resolveParams{
		diff: d,
		opts: collector.LogOptions{Since: "1 hour ago"},
	})
	if err == nil {
		t.Fatal("expected error when --diff is combined with --since, got nil")
	}
	if !strings.Contains(err.Error(), "--since") {
		t.Fatalf("error should mention --since, got: %q", err.Error())
	}
}

func TestResolveSourceLabelDiff(t *testing.T) {
	label := resolveSourceLabel(resolveParams{
		diff: &bootDiff{from: -1, to: 0},
	})
	if !strings.Contains(label, "diff") {
		t.Errorf("resolveSourceLabel for diff should contain 'diff', got %q", label)
	}
	if !strings.Contains(label, "-1") || !strings.Contains(label, "0") {
		t.Errorf("resolveSourceLabel for diff should contain both boot indices, got %q", label)
	}
}

func TestExplainCmdHasPIDFlag(t *testing.T) {
	cmd := newExplainCmd()
	f := cmd.Flags().Lookup("pid")
	if f == nil {
		t.Fatal("explain command is missing the --pid flag")
	}
	if f.DefValue != "0" {
		t.Errorf("--pid default = %q, want \"0\"", f.DefValue)
	}
}

func TestResolveInputPIDMutuallyExclusiveWithService(t *testing.T) {
	_, err := resolveInput(resolveParams{
		service: "nginx",
		pid:     1234,
	})
	if err == nil {
		t.Fatal("expected error when --pid and --service are combined, got nil")
	}
	if !strings.Contains(err.Error(), "--pid") || !strings.Contains(err.Error(), "--service") {
		t.Fatalf("error should mention both --pid and --service, got: %q", err.Error())
	}
}

func TestResolveInputPIDMutuallyExclusiveWithDmesg(t *testing.T) {
	_, err := resolveInput(resolveParams{
		dmesg: true,
		pid:   1234,
	})
	if err == nil {
		t.Fatal("expected error when --pid and --dmesg are combined, got nil")
	}
}

func TestResolveInputPIDMutuallyExclusiveWithFile(t *testing.T) {
	_, err := resolveInput(resolveParams{
		file: "/var/log/syslog",
		pid:  1234,
	})
	if err == nil {
		t.Fatal("expected error when --pid and --file are combined, got nil")
	}
}

func TestResolveInputPIDMutuallyExclusiveWithBoot(t *testing.T) {
	_, err := resolveInput(resolveParams{
		bootChanged: true,
		boot:        "0",
		pid:         1234,
	})
	if err == nil {
		t.Fatal("expected error when --pid and --boot are combined, got nil")
	}
}

func TestResolveInputPIDMutuallyExclusiveWithDiff(t *testing.T) {
	d := &bootDiff{from: -1, to: 0}
	_, err := resolveInput(resolveParams{
		diff: d,
		pid:  1234,
	})
	if err == nil {
		t.Fatal("expected error when --pid and --diff are combined, got nil")
	}
}

func TestResolveInputPIDNonExistent(t *testing.T) {
	_, err := resolveInput(resolveParams{pid: 999999999})
	if err == nil {
		t.Fatal("expected error for non-existent PID, got nil")
	}
	if !strings.Contains(err.Error(), "999999999") {
		t.Fatalf("error should mention the PID, got: %q", err.Error())
	}
}

func TestResolveInputPIDZeroIsIgnored(t *testing.T) {
	_, err := resolveInput(resolveParams{pid: 0, args: []string{"some error message"}})
	if err != nil {
		// An error here is acceptable only if it is NOT about PID 0.
		if strings.Contains(err.Error(), "PID 0") {
			t.Fatalf("pid=0 should not be treated as an active --pid source, got: %q", err.Error())
		}
	}
}

func TestResolveSourceLabelPID(t *testing.T) {
	label := resolveSourceLabel(resolveParams{pid: 42})
	want := "pid: 42"
	if label != want {
		t.Errorf("resolveSourceLabel() with pid = %q, want %q", label, want)
	}
}
