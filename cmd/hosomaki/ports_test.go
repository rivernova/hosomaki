// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/prompt"
)

// unit tests for the ports command

func TestPortsCmd_HasDebugFlag(t *testing.T) {
	cmd := newPortsCmd()
	f := cmd.Flags().Lookup("debug")
	if f == nil {
		t.Fatal("ports command is missing the --debug flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--debug default = %q, want %q", f.DefValue, "false")
	}
}

func TestPortsCmd_NoArgs(t *testing.T) {
	cmd := newPortsCmd()
	if cmd.Args == nil {
		t.Fatal("ports command must define an Args validator")
	}
	if err := cmd.Args(cmd, []string{"unexpected"}); err == nil {
		t.Error("ports command must reject positional arguments")
	}
}

func TestPortsCmd_AcceptsNoPositionalArgs(t *testing.T) {
	cmd := newPortsCmd()
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("ports command must accept zero positional arguments, got error: %v", err)
	}
}

func TestValidatePortsResult_ValidResult(t *testing.T) {
	r := prompt.PortsResult{
		Summary: "Three ports are listening; one warrants attention.",
		Findings: []prompt.PortsFinding{
			{
				Severity: "warning",
				Port:     "tcp 0.0.0.0:3306",
				Title:    "MySQL exposed on all interfaces",
				Detail:   "MySQL is bound to all interfaces, making it reachable from the network.",
			},
		},
	}
	errs := validatePortsResult(r)
	if len(errs) != 0 {
		t.Errorf("validatePortsResult() returned unexpected errors: %v", errs)
	}
}

func TestValidatePortsResult_EmptySummary(t *testing.T) {
	r := prompt.PortsResult{
		Summary:  "",
		Findings: []prompt.PortsFinding{},
	}
	errs := validatePortsResult(r)
	if len(errs) == 0 {
		t.Error("validatePortsResult() must reject empty summary")
	}
	if !strings.Contains(strings.Join(errs, " "), "summary") {
		t.Errorf("error must mention 'summary', got: %v", errs)
	}
}

func TestValidatePortsResult_EmptyFindings(t *testing.T) {
	r := prompt.PortsResult{
		Summary:  "All ports look normal.",
		Findings: []prompt.PortsFinding{},
	}
	errs := validatePortsResult(r)
	if len(errs) != 0 {
		t.Errorf("validatePortsResult() must accept empty findings array, got: %v", errs)
	}
}

func TestValidatePortsResult_EmptySeverityProducesOneError(t *testing.T) {
	r := prompt.PortsResult{
		Summary: "One issue found.",
		Findings: []prompt.PortsFinding{
			{
				Severity: "",
				Port:     "tcp 0.0.0.0:22",
				Title:    "Something",
				Detail:   "Some detail here.",
			},
		},
	}
	errs := validatePortsResult(r)
	severityErrs := 0
	for _, e := range errs {
		if strings.Contains(e, "severity") {
			severityErrs++
		}
	}
	if severityErrs != 1 {
		t.Errorf("empty severity must produce exactly 1 severity error, got %d: %v", severityErrs, errs)
	}
}

func TestValidatePortsResult_InvalidSeverity(t *testing.T) {
	r := prompt.PortsResult{
		Summary: "One issue found.",
		Findings: []prompt.PortsFinding{
			{
				Severity: "critical", // only "warning" and "info" are valid
				Port:     "tcp 0.0.0.0:22",
				Title:    "Something",
				Detail:   "Some detail here about the issue.",
			},
		},
	}
	errs := validatePortsResult(r)
	if len(errs) == 0 {
		t.Error("validatePortsResult() must reject invalid severity value")
	}
	if !strings.Contains(strings.Join(errs, " "), "severity") {
		t.Errorf("error must mention 'severity', got: %v", errs)
	}
}

func TestValidatePortsResult_MissingPort(t *testing.T) {
	r := prompt.PortsResult{
		Summary: "One issue found.",
		Findings: []prompt.PortsFinding{
			{
				Severity: "warning",
				Port:     "",
				Title:    "Something",
				Detail:   "Some detail here.",
			},
		},
	}
	errs := validatePortsResult(r)
	if len(errs) == 0 {
		t.Error("validatePortsResult() must reject finding with empty port")
	}
	if !strings.Contains(strings.Join(errs, " "), "port") {
		t.Errorf("error must mention 'port', got: %v", errs)
	}
}

func TestValidatePortsResult_MissingTitle(t *testing.T) {
	r := prompt.PortsResult{
		Summary: "One issue found.",
		Findings: []prompt.PortsFinding{
			{
				Severity: "info",
				Port:     "tcp 0.0.0.0:8080",
				Title:    "",
				Detail:   "Some detail here.",
			},
		},
	}
	errs := validatePortsResult(r)
	if len(errs) == 0 {
		t.Error("validatePortsResult() must reject finding with empty title")
	}
}

func TestValidatePortsResult_MissingDetail(t *testing.T) {
	r := prompt.PortsResult{
		Summary: "One issue found.",
		Findings: []prompt.PortsFinding{
			{
				Severity: "warning",
				Port:     "tcp 0.0.0.0:8080",
				Title:    "Something",
				Detail:   "",
			},
		},
	}
	errs := validatePortsResult(r)
	if len(errs) == 0 {
		t.Error("validatePortsResult() must reject finding with empty detail")
	}
}

func TestValidatePortsResult_BothSeverityValuesAccepted(t *testing.T) {
	for _, sev := range []string{"warning", "info"} {
		r := prompt.PortsResult{
			Summary: "Some ports found.",
			Findings: []prompt.PortsFinding{
				{
					Severity: sev,
					Port:     "tcp 0.0.0.0:9000",
					Title:    "A title",
					Detail:   "Some detail here about this finding.",
				},
			},
		}
		errs := validatePortsResult(r)
		if len(errs) != 0 {
			t.Errorf("validatePortsResult() must accept severity %q, got errors: %v", sev, errs)
		}
	}
}
