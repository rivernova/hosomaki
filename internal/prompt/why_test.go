// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/collector"
)

// unit tests for the why prompt builder

func makeWhyInput(code int) WhyInput {
	return WhyInput{
		Service:  "nginx.service",
		ExitCode: code,
		Logs:     "<ERROR> connection refused\n<INFO> starting up",
		Environment: collector.Environment{
			DistroID:   "ubuntu",
			InitSystem: "systemd",
		},
	}
}

func TestWhy_ContainsSchema(t *testing.T) {
	p := Why(makeWhyInput(1))
	if !strings.Contains(p, SchemaWhy) {
		t.Error("Why() prompt must contain the schema constant")
	}
}

func TestWhy_ContainsServiceName(t *testing.T) {
	p := Why(makeWhyInput(1))
	if !strings.Contains(p, "nginx.service") {
		t.Error("Why() prompt must embed the service name")
	}
}

func TestWhy_ContainsExitCode(t *testing.T) {
	p := Why(makeWhyInput(137))
	if !strings.Contains(p, "137") {
		t.Error("Why() prompt must embed the exit code")
	}
}

func TestWhy_ContainsExitCodeLabel(t *testing.T) {
	p := Why(makeWhyInput(137))
	if !strings.Contains(p, "SIGKILL") {
		t.Error("Why() prompt must embed the label for exit code 137")
	}
}

func TestWhy_ContainsLogs(t *testing.T) {
	p := Why(makeWhyInput(1))
	if !strings.Contains(p, "<ERROR> connection refused") {
		t.Error("Why() prompt must embed the log excerpt verbatim")
	}
}

func TestWhy_ContainsEnvironmentSection(t *testing.T) {
	p := Why(makeWhyInput(1))
	if !strings.Contains(p, "Host environment") {
		t.Error("Why() prompt must contain the environment section")
	}
}

func TestWhy_InstructsPureJSON(t *testing.T) {
	p := Why(makeWhyInput(1))
	if !strings.Contains(p, "Return ONLY a JSON object") {
		t.Error("Why() prompt must instruct the model to return pure JSON")
	}
}

func TestWhy_InstructsNoMarkdown(t *testing.T) {
	p := Why(makeWhyInput(1))
	if !strings.Contains(p, "No markdown") {
		t.Error("Why() prompt must forbid markdown in the output")
	}
}

func TestWhy_InstructsRootCauseFirst(t *testing.T) {
	p := Why(makeWhyInput(1))
	if !strings.Contains(p, "root cause") {
		t.Error("Why() prompt must describe root-cause-first ordering")
	}
}

func TestWhy_InstructsNextSteps(t *testing.T) {
	p := Why(makeWhyInput(1))
	if !strings.Contains(p, `"next_steps"`) {
		t.Error("Why() prompt must reference the next_steps field by name")
	}
}

func TestWhy_ChainOrderingDescribed(t *testing.T) {
	p := Why(makeWhyInput(1))
	if !strings.Contains(p, "root cause →") {
		t.Error("Why() prompt must describe chain ordering: root cause → proximate cause")
	}
}

func TestWhy_EventFieldReferenced(t *testing.T) {
	p := Why(makeWhyInput(1))
	if !strings.Contains(p, `"event"`) {
		t.Error("Why() prompt must reference the 'event' JSON field by name")
	}
}

func TestWhy_SummaryFieldReferenced(t *testing.T) {
	p := Why(makeWhyInput(1))
	if !strings.Contains(p, `"summary"`) {
		t.Error("Why() prompt must reference the 'summary' JSON field by name")
	}
}

func TestExitCodeLabel_GenericError(t *testing.T) {
	if got := exitCodeLabel(1); got != "generic error" {
		t.Errorf("exitCodeLabel(1) = %q, want %q", got, "generic error")
	}
}

func TestExitCodeLabel_SIGKILL(t *testing.T) {
	want := "killed by SIGKILL — likely OOM killer"
	if got := exitCodeLabel(137); got != want {
		t.Errorf("exitCodeLabel(137) = %q, want %q", got, want)
	}
}

func TestExitCodeLabel_SIGTERM(t *testing.T) {
	want := "terminated by SIGTERM"
	if got := exitCodeLabel(143); got != want {
		t.Errorf("exitCodeLabel(143) = %q, want %q", got, want)
	}
}

func TestExitCodeLabel_SIGINT(t *testing.T) {
	want := "terminated by SIGINT (Ctrl-C)"
	if got := exitCodeLabel(130); got != want {
		t.Errorf("exitCodeLabel(130) = %q, want %q", got, want)
	}
}

func TestExitCodeLabel_SignalRange(t *testing.T) {
	label := exitCodeLabel(132)
	if !strings.Contains(label, "signal 4") {
		t.Errorf("exitCodeLabel(132) = %q, expected it to contain 'signal 4'", label)
	}
}

func TestExitCodeLabel_UnknownIsNonEmpty(t *testing.T) {
	if exitCodeLabel(42) == "" {
		t.Error("exitCodeLabel for an unknown code must not be empty")
	}
}

func TestExitCodeLabel_CommandNotFound(t *testing.T) {
	want := "command not found"
	if got := exitCodeLabel(127); got != want {
		t.Errorf("exitCodeLabel(127) = %q, want %q", got, want)
	}
}

func TestExitCodeLabel_PermissionDenied(t *testing.T) {
	want := "permission denied or command not executable"
	if got := exitCodeLabel(126); got != want {
		t.Errorf("exitCodeLabel(126) = %q, want %q", got, want)
	}
}

func TestExitCodeLabel_Segfault(t *testing.T) {
	want := "segmentation fault (SIGSEGV)"
	if got := exitCodeLabel(139); got != want {
		t.Errorf("exitCodeLabel(139) = %q, want %q", got, want)
	}
}

func TestSchemaWhy_ContainsRequiredFields(t *testing.T) {
	for _, field := range []string{"summary", "chain", "event", "detail", "next_steps"} {
		if !strings.Contains(SchemaWhy, field) {
			t.Errorf("SchemaWhy must contain field %q", field)
		}
	}
}
