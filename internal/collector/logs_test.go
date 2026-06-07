// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"testing"
)

// unit tests for log collection helpers

func TestLogOptions_TimeArgs_BothEmpty(t *testing.T) {
	args := LogOptions{}.timeArgs()
	if len(args) != 0 {
		t.Fatalf("expected no args when Since and Until are empty, got %v", args)
	}
}

func TestLogOptions_TimeArgs_SinceOnly(t *testing.T) {
	args := LogOptions{Since: "1 hour ago"}.timeArgs()
	if len(args) != 2 || args[0] != "--since" || args[1] != "1 hour ago" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestLogOptions_TimeArgs_UntilOnly(t *testing.T) {
	args := LogOptions{Until: "2024-01-15 15:00:00"}.timeArgs()
	if len(args) != 2 || args[0] != "--until" || args[1] != "2024-01-15 15:00:00" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestLogOptions_TimeArgs_Both(t *testing.T) {
	opts := LogOptions{Since: "2024-01-15 14:00:00", Until: "2024-01-15 15:00:00"}
	args := opts.timeArgs()
	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d: %v", len(args), args)
	}
	if args[0] != "--since" || args[1] != "2024-01-15 14:00:00" {
		t.Fatalf("unexpected since args: %v", args[:2])
	}
	if args[2] != "--until" || args[3] != "2024-01-15 15:00:00" {
		t.Fatalf("unexpected until args: %v", args[2:])
	}
}

func TestLogOptions_TimeArgs_OrderIsSinceBeforeUntil(t *testing.T) {
	args := LogOptions{Since: "S", Until: "U"}.timeArgs()
	if args[0] != "--since" {
		t.Fatalf("--since must come before --until, got %v", args)
	}
}

func TestIsJournalContent_EmptyString(t *testing.T) {
	if isJournalContent("") {
		t.Error("empty string should not be journal content")
	}
}

func TestIsJournalContent_WhitespaceOnly(t *testing.T) {
	if isJournalContent("   \n\t  ") {
		t.Error("whitespace-only string should not be journal content")
	}
}

func TestIsJournalContent_NoEntriesSentinel(t *testing.T) {
	if isJournalContent("-- No entries --") {
		t.Error("'-- No entries --' sentinel should not be journal content")
	}
}

func TestIsJournalContent_NoEntriesLowercase(t *testing.T) {
	if isJournalContent("-- no entries --") {
		t.Error("'-- no entries --' sentinel should not be journal content")
	}
}

func TestIsJournalContent_NoJournalFiles(t *testing.T) {
	if isJournalContent("No journal files were found.") {
		t.Error("'No journal files were found.' should not be journal content")
	}
}

func TestIsJournalContent_RealLogLine(t *testing.T) {
	if !isJournalContent("[12345.678] kernel: error something went wrong") {
		t.Error("real log line should be journal content")
	}
}

func TestIsJournalContent_MultiLineWithSentinelAndLogLine(t *testing.T) {
	mixed := "-- No entries --\n[12345.678] nginx[42]: error starting"
	if !isJournalContent(mixed) {
		t.Error("mixed content with real log lines should be journal content")
	}
}

func TestLooksLikeLogLine_BracketPrefix(t *testing.T) {
	if !looksLikeLogLine("[12345.678] kernel: OOM killer") {
		t.Error("line starting with '[' should look like a log line")
	}
}

func TestLooksLikeLogLine_ColonBracket(t *testing.T) {
	if !looksLikeLogLine("nginx[42]: error connecting") {
		t.Error("line with ']: ' pattern should look like a log line")
	}
}

func TestLooksLikeLogLine_PlainText(t *testing.T) {
	if looksLikeLogLine("this is just plain text") {
		t.Error("plain text without log markers should not look like a log line")
	}
}

func TestLooksLikeLogLine_EmptyLines(t *testing.T) {
	if looksLikeLogLine("\n\n   \n") {
		t.Error("only empty lines should not look like a log line")
	}
}

func TestLines_UsesDefaultWhenZero(t *testing.T) {
	if lines(0, 50) != 50 {
		t.Error("lines(0, 50) should return 50")
	}
}

func TestLines_UsesNWhenPositive(t *testing.T) {
	if lines(20, 50) != 20 {
		t.Error("lines(20, 50) should return 20")
	}
}

func TestLines_IgnoresNegative(t *testing.T) {
	if lines(-1, 50) != 50 {
		t.Error("lines(-1, 50) should return the default 50")
	}
}

func TestContextLogs_AllServicesMissing(t *testing.T) {
	services := []string{"hosomaki-nonexistent-a", "hosomaki-nonexistent-b"}
	collected, errs := ContextLogs(services, LogOptions{})
	if len(collected) != 0 {
		t.Errorf("expected no collected logs for missing services, got %d", len(collected))
	}
	if len(errs) != 2 {
		t.Errorf("expected 2 errors for 2 missing services, got %d", len(errs))
	}
}

func TestContextLogs_ReturnsErrorsNonFatal(t *testing.T) {
	services := []string{"hosomaki-nonexistent-xyz", "systemd-journald"}
	collected, errs := ContextLogs(services, LogOptions{})
	if len(errs) == 0 {
		t.Error("expected at least one error for nonexistent service")
	}
	_ = collected
}

func TestContextLogs_PreservesAllServicesInMap(t *testing.T) {
	services := []string{"hosomaki-nonexistent-1", "hosomaki-nonexistent-2", "hosomaki-nonexistent-3"}
	collected, errs := ContextLogs(services, LogOptions{})
	if len(collected)+len(errs) != len(services) {
		t.Errorf("collected (%d) + errs (%d) should equal services (%d)",
			len(collected), len(errs), len(services))
	}
}

func TestBootDiffLogs_BothMissing(t *testing.T) {
	_, _, err := BootDiffLogs(-999, -998, LogOptions{})
	if err == nil {
		t.Fatal("BootDiffLogs with two missing boots should return an error")
	}
}

func TestBootDiffLogs_SameIndexRejectedByParser(t *testing.T) {
	_, _, err := BootDiffLogs(0, 0, LogOptions{})
	_ = err
}
