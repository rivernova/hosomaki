// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/prompt"
)

// unit tests for the why command

func TestParseExitCode_ValidMinimum(t *testing.T) {
	code, err := parseExitCode("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

func TestParseExitCode_ValidMaximum(t *testing.T) {
	code, err := parseExitCode("255")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 255 {
		t.Fatalf("expected 255, got %d", code)
	}
}

func TestParseExitCode_ValidMidRange(t *testing.T) {
	code, err := parseExitCode("137")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 137 {
		t.Fatalf("expected 137, got %d", code)
	}
}

func TestParseExitCode_ZeroRejected(t *testing.T) {
	_, err := parseExitCode("0")
	if err == nil {
		t.Fatal("expected error for exit code 0, got nil")
	}
	if !strings.Contains(err.Error(), "success") {
		t.Errorf("error for code 0 should mention 'success', got: %v", err)
	}
}

func TestParseExitCode_NegativeRejected(t *testing.T) {
	_, err := parseExitCode("-1")
	if err == nil {
		t.Fatal("expected error for negative exit code, got nil")
	}
}

func TestParseExitCode_256Rejected(t *testing.T) {
	_, err := parseExitCode("256")
	if err == nil {
		t.Fatal("expected error for exit code 256, got nil")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Errorf("error for 256 should mention 'out of range', got: %v", err)
	}
}

func TestParseExitCode_NonIntegerRejected(t *testing.T) {
	_, err := parseExitCode("crash")
	if err == nil {
		t.Fatal("expected error for non-integer input, got nil")
	}
	if !strings.Contains(err.Error(), "invalid exit code") {
		t.Errorf("error should mention 'invalid exit code', got: %v", err)
	}
}

func TestParseExitCode_FloatRejected(t *testing.T) {
	_, err := parseExitCode("1.5")
	if err == nil {
		t.Fatal("expected error for float input, got nil")
	}
}

func TestParseExitCode_EmptyStringRejected(t *testing.T) {
	_, err := parseExitCode("")
	if err == nil {
		t.Fatal("expected error for empty string, got nil")
	}
}

func TestParseExitCode_WhitespaceIsTrimmed(t *testing.T) {
	code, err := parseExitCode("  42  ")
	if err != nil {
		t.Fatalf("unexpected error for whitespace-padded input: %v", err)
	}
	if code != 42 {
		t.Fatalf("expected 42, got %d", code)
	}
}

func validWhyResult() prompt.WhyResult {
	return prompt.WhyResult{
		Summary: "nginx failed because the configuration file was missing.",
		Chain: []prompt.WhyStep{
			{
				Event:  "configuration file absent",
				Detail: "The service attempted to load the config on startup.",
			},
			{
				Event:  "process exited with code 1",
				Detail: "nginx exited immediately after reporting the missing file.",
			},
		},
		NextSteps: []string{"Restore the configuration file from the last known-good backup."},
	}
}

func TestValidateWhyResult_ValidResult(t *testing.T) {
	if errs := validateWhyResult(validWhyResult()); len(errs) != 0 {
		t.Fatalf("expected no errors for valid result, got: %v", errs)
	}
}

func TestValidateWhyResult_EmptySummary(t *testing.T) {
	r := validWhyResult()
	r.Summary = ""
	errs := validateWhyResult(r)
	if len(errs) == 0 {
		t.Fatal("expected error for empty summary")
	}
	if !strings.Contains(errs[0], "summary") {
		t.Errorf("error should mention 'summary', got: %v", errs[0])
	}
}

func TestValidateWhyResult_WhitespaceSummaryRejected(t *testing.T) {
	r := validWhyResult()
	r.Summary = "   "
	if errs := validateWhyResult(r); len(errs) == 0 {
		t.Fatal("expected error for whitespace-only summary")
	}
}

func TestValidateWhyResult_EmptyChainRejected(t *testing.T) {
	r := validWhyResult()
	r.Chain = nil
	errs := validateWhyResult(r)
	if len(errs) == 0 {
		t.Fatal("expected error for empty chain")
	}
	if !strings.Contains(errs[0], "chain") {
		t.Errorf("error should mention 'chain', got: %v", errs[0])
	}
}

func TestValidateWhyResult_ChainStepMissingEvent(t *testing.T) {
	r := validWhyResult()
	r.Chain[0].Event = ""
	if errs := validateWhyResult(r); len(errs) == 0 {
		t.Fatal("expected error for chain step with empty event")
	}
}

func TestValidateWhyResult_ChainStepWhitespaceDetailRejected(t *testing.T) {
	r := validWhyResult()
	r.Chain[0].Detail = "   "
	if errs := validateWhyResult(r); len(errs) == 0 {
		t.Fatal("expected error for chain step with whitespace-only detail")
	}
}

func TestValidateWhyResult_EmptyNextStepsRejected(t *testing.T) {
	r := validWhyResult()
	r.NextSteps = nil
	errs := validateWhyResult(r)
	if len(errs) == 0 {
		t.Fatal("expected error for empty next_steps")
	}
	if !strings.Contains(errs[0], "next_steps") {
		t.Errorf("error should mention 'next_steps', got: %v", errs[0])
	}
}

func TestValidateWhyResult_AllEmptyReportsMultipleErrors(t *testing.T) {
	errs := validateWhyResult(prompt.WhyResult{})
	if len(errs) < 3 {
		t.Fatalf("expected at least 3 errors for empty result, got %d: %v", len(errs), errs)
	}
}
