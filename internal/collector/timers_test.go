// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// unit tests for the timers collector

func TestParseKeyValues_Basic(t *testing.T) {
	out := "Id=logrotate.timer\nActiveState=active\nTriggers=logrotate.service\n"
	kv := parseKeyValues(out)
	if kv["Id"] != "logrotate.timer" {
		t.Errorf("Id = %q, want %q", kv["Id"], "logrotate.timer")
	}
	if kv["ActiveState"] != "active" {
		t.Errorf("ActiveState = %q, want %q", kv["ActiveState"], "active")
	}
}

func TestParseKeyValues_ValueContainsEquals(t *testing.T) {
	out := "Description=foo=bar\n"
	kv := parseKeyValues(out)
	if kv["Description"] != "foo=bar" {
		t.Errorf("Description = %q, want %q", kv["Description"], "foo=bar")
	}
}

func TestParseKeyValues_EmptyValue(t *testing.T) {
	out := "Result=\n"
	kv := parseKeyValues(out)
	if kv["Result"] != "" {
		t.Errorf("Result = %q, want empty string for empty value", kv["Result"])
	}
}

func TestParseKeyValues_EmptyInput(t *testing.T) {
	kv := parseKeyValues("")
	if len(kv) != 0 {
		t.Errorf("parseKeyValues('') = %v, want empty map", kv)
	}
}

func TestFirstOf_SingleToken(t *testing.T) {
	got := firstOf("logrotate.service")
	if got != "logrotate.service" {
		t.Errorf("firstOf() = %q, want %q", got, "logrotate.service")
	}
}

func TestFirstOf_MultipleTokens(t *testing.T) {
	got := firstOf("logrotate.service backup.service")
	if got != "logrotate.service" {
		t.Errorf("firstOf() = %q, want %q", got, "logrotate.service")
	}
}

func TestFirstOf_Empty(t *testing.T) {
	got := firstOf("")
	if got != "" {
		t.Errorf("firstOf('') = %q, want empty string", got)
	}
}

func TestParseUSec_ValidTimestamp(t *testing.T) {
	ts := parseUSec("1000000")
	if ts.IsZero() {
		t.Error("parseUSec('1000000') returned zero time, want non-zero")
	}
	if ts.Unix() != 1 {
		t.Errorf("parseUSec('1000000').Unix() = %d, want 1", ts.Unix())
	}
}

func TestParseUSec_Zero(t *testing.T) {
	ts := parseUSec("0")
	if !ts.IsZero() {
		t.Error("parseUSec('0') should return zero time")
	}
}

func TestParseUSec_Empty(t *testing.T) {
	ts := parseUSec("")
	if !ts.IsZero() {
		t.Error("parseUSec('') should return zero time")
	}
}

func TestParseUSec_InvalidCharacters(t *testing.T) {
	ts := parseUSec("not-a-number")
	if !ts.IsZero() {
		t.Error("parseUSec('not-a-number') should return zero time")
	}
}

func TestUSecToHuman_NeverWhenBothZero(t *testing.T) {
	got := uSecToHuman("0", "0")
	if got != "never" {
		t.Errorf("uSecToHuman('0','0') = %q, want %q", got, "never")
	}
}

func TestUSecToHuman_NeverWhenBothEmpty(t *testing.T) {
	got := uSecToHuman("", "")
	if got != "never" {
		t.Errorf("uSecToHuman('','') = %q, want %q", got, "never")
	}
}

func TestUSecToHuman_ValidRealtimeTakesPriority(t *testing.T) {
	usec := "1718064000000000"
	got := uSecToHuman(usec, "0")
	if got == "never" {
		t.Errorf("uSecToHuman(%q, '0') = 'never', want a formatted timestamp", usec)
	}
	if !strings.Contains(got, "2024") {
		t.Errorf("uSecToHuman(%q) = %q, want a timestamp in 2024", usec, got)
	}
}

func TestUSecToHuman_FallsBackToMonotonic(t *testing.T) {
	usec := "1718064000000000"
	got := uSecToHuman("0", usec)
	if got == "never" {
		t.Errorf("uSecToHuman('0', %q) = 'never', want a formatted timestamp", usec)
	}
}

func TestParseUSec_RoundTrip(t *testing.T) {
	known := time.Date(2024, 6, 11, 0, 0, 0, 0, time.UTC)
	usec := uint64(known.Unix()) * 1_000_000
	ts := parseUSec(fmt.Sprint(usec))
	if ts.IsZero() {
		t.Fatal("parseUSec returned zero for a valid timestamp")
	}
	if ts.UTC().Year() != 2024 || ts.UTC().Month() != 6 || ts.UTC().Day() != 11 {
		t.Errorf("parseUSec round-trip = %v, want 2024-06-11", ts.UTC())
	}
}

func TestFormatTimersForPrompt_Empty(t *testing.T) {
	out := FormatTimersForPrompt(nil)
	if !strings.Contains(out, "no systemd timers found") {
		t.Errorf("FormatTimersForPrompt(nil) = %q, want 'no systemd timers found'", out)
	}
}

func TestFormatTimersForPrompt_NeverLastRun(t *testing.T) {
	entries := []TimerEntry{
		{Unit: "test.timer", Activates: "test.service", Last: "never", Next: "never", ActiveState: "inactive"},
	}
	out := FormatTimersForPrompt(entries)
	if !strings.Contains(out, "last_run:") {
		t.Errorf("FormatTimersForPrompt() missing 'last_run:' field, got:\n%s", out)
	}
	if !strings.Contains(out, "never") {
		t.Errorf("FormatTimersForPrompt() missing 'never' value, got:\n%s", out)
	}
	if !strings.Contains(out, "test.timer") {
		t.Errorf("FormatTimersForPrompt() missing unit name, got:\n%s", out)
	}
}

func TestFormatTimersForPrompt_AllFieldsPresent(t *testing.T) {
	entries := []TimerEntry{
		{
			Unit:        "logrotate.timer",
			Activates:   "logrotate.service",
			Last:        "2024-06-10 00:00:01 UTC",
			Next:        "2024-06-11 00:00:00 UTC",
			ActiveState: "active",
			LastResult:  "success",
		},
	}
	out := FormatTimersForPrompt(entries)
	for _, want := range []string{"unit:", "activates:", "last_run:", "next_run:", "active_state:", "last_result:"} {
		if !strings.Contains(out, want) {
			t.Errorf("FormatTimersForPrompt() missing field %q, got:\n%s", want, out)
		}
	}
}

func TestFormatTimersForPrompt_OmitsLastResultWhenEmpty(t *testing.T) {
	entries := []TimerEntry{
		{Unit: "test.timer", Activates: "test.service", Last: "never", Next: "never", LastResult: ""},
	}
	out := FormatTimersForPrompt(entries)
	if strings.Contains(out, "last_result:") {
		t.Error("FormatTimersForPrompt() should omit last_result when empty")
	}
}

func TestFormatTimersForPrompt_MultipleEntries(t *testing.T) {
	entries := []TimerEntry{
		{Unit: "a.timer", Activates: "a.service", Last: "never", Next: "never"},
		{Unit: "b.timer", Activates: "b.service", Last: "never", Next: "never"},
	}
	out := FormatTimersForPrompt(entries)
	if !strings.Contains(out, "a.timer") || !strings.Contains(out, "b.timer") {
		t.Errorf("FormatTimersForPrompt() missing timer units, got:\n%s", out)
	}
}
