// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package insight

import (
	"testing"
)

// unit tests for strict XML-only insight parsing and normalisation.

func TestParseAnalysisWellFormedXML(t *testing.T) {
	raw := `<analysis>
  <component>
    <source>pipe</source>
    <pattern>The nginx web server failed to bind to port 80 because the port was already occupied by another process.</pattern>
    <cause>A previous nginx instance was not cleanly terminated and its socket file remains open, preventing a new bind.</cause>
    <severity>high</severity>
    <suggestion>Run: ss -tlnp | grep :80 to identify the conflicting process, then: kill -9 &lt;pid&gt; and: systemctl start nginx</suggestion>
  </component>
</analysis>`

	a := ParseDoctor(raw)
	if len(a.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(a.Components))
	}
	c := a.Components[0]
	if c.Source != "pipe" {
		t.Errorf("source = %q, want pipe", c.Source)
	}
	if c.Pattern == "" {
		t.Error("pattern should not be empty")
	}
	if c.Cause == "" {
		t.Error("cause should not be empty")
	}
	if c.Severity != "high" {
		t.Errorf("severity = %q, want high", c.Severity)
	}
	if c.Suggestion == "" {
		t.Error("suggestion should not be empty for doctor parse")
	}
}

func TestParseAnalysisMultipleComponents(t *testing.T) {
	raw := `<analysis>
  <component>
    <source>service:nginx</source>
    <pattern>nginx failed to start</pattern>
    <cause>port conflict</cause>
    <severity>high</severity>
    <suggestion>kill conflicting process</suggestion>
  </component>
  <component>
    <source>service:postgresql</source>
    <pattern>database connection refused</pattern>
    <cause>max_connections limit exceeded</cause>
    <severity>medium</severity>
    <suggestion>increase max_connections in postgresql.conf</suggestion>
  </component>
</analysis>`

	a := ParseDoctor(raw)
	if len(a.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(a.Components))
	}
	if a.Components[0].Source != "service:nginx" {
		t.Errorf("first source = %q", a.Components[0].Source)
	}
	if a.Components[1].Source != "service:postgresql" {
		t.Errorf("second source = %q", a.Components[1].Source)
	}
}

func TestParseAnalysisEmptyAnalysisTag(t *testing.T) {
	a := ParseDoctor(`<analysis></analysis>`)
	if len(a.Components) != 0 {
		t.Errorf("expected 0 components for empty analysis, got %d", len(a.Components))
	}
	if a.Raw != "" {
		t.Errorf("expected empty raw for empty analysis, got %q", a.Raw)
	}
}

func TestParseAnalysisRawFallback(t *testing.T) {
	a := ParseDoctor("something went completely wrong and no xml appeared")
	if len(a.Components) != 0 {
		t.Errorf("expected 0 components for raw fallback, got %d", len(a.Components))
	}
	if a.Raw == "" {
		t.Error("raw should be populated when XML parsing fails")
	}
}

func TestParseAnalysisProseOnlyReturnedAsRaw(t *testing.T) {
	raw := `pattern: the sshd service rejected all incoming connections
cause: the host key file /etc/ssh/ssh_host_rsa_key has incorrect permissions`

	a := parseXMLAnalysis(raw)
	if len(a.Components) != 0 {
		t.Errorf("prose-only input must not produce components, got %d", len(a.Components))
	}
	if a.Raw == "" {
		t.Error("raw should be populated for prose-only input")
	}
}

func TestParseExplainStripsSeverityAndSuggestion(t *testing.T) {
	raw := `<analysis>
  <component>
    <source>dmesg</source>
    <pattern>kernel oom killer activated</pattern>
    <cause>memory exhausted by chrome processes</cause>
    <severity>critical</severity>
    <suggestion>add more RAM or set memory limits</suggestion>
  </component>
</analysis>`

	a := ParseExplain(raw)
	if len(a.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(a.Components))
	}
	c := a.Components[0]
	if c.Severity != "" {
		t.Errorf("explain must strip severity, got %q", c.Severity)
	}
	if c.Suggestion != "" {
		t.Errorf("explain must strip suggestion, got %q", c.Suggestion)
	}
	if c.Pattern == "" {
		t.Error("pattern must be preserved")
	}
	if c.Cause == "" {
		t.Error("cause must be preserved")
	}
}

func TestParseStatusStripsSuggestionKeepsCause(t *testing.T) {
	raw := `<analysis>
  <component>
    <source>pipe</source>
    <pattern>disk usage at 94 percent on root partition</pattern>
    <cause>log files accumulated over several months without rotation</cause>
    <severity>high</severity>
    <suggestion>run journalctl --vacuum-size=500M to reclaim space</suggestion>
  </component>
</analysis>`

	a := ParseStatus(raw)
	if len(a.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(a.Components))
	}
	c := a.Components[0]
	if c.Suggestion != "" {
		t.Errorf("status must strip suggestion, got %q", c.Suggestion)
	}
	if c.Cause == "" {
		t.Error("status must preserve cause")
	}
	if c.Severity == "" {
		t.Error("status must preserve severity")
	}
}

func TestParseAnalysisDefaultsSourceToPipe(t *testing.T) {
	raw := `<analysis>
  <component>
    <pattern>some pattern</pattern>
    <cause>some cause</cause>
  </component>
</analysis>`

	a := ParseDoctor(raw)
	if len(a.Components) == 0 {
		t.Fatal("expected at least 1 component")
	}
	if a.Components[0].Source != "pipe" {
		t.Errorf("missing source should default to pipe, got %q", a.Components[0].Source)
	}
}

func TestNormaliseSeverityMappings(t *testing.T) {
	cases := []struct{ in, want string }{
		{"low", "ok"},
		{"minor", "ok"},
		{"medium", "warn"},
		{"warn", "warn"},
		{"warning", "warn"},
		{"high", "crit"},
		{"critical", "crit"},
		{"crit", "crit"},
		{"fatal", "crit"},
		{"", "info"},
		{"unknown", "info"},
	}
	for _, c := range cases {
		got := NormaliseSeverity(c.in)
		if got != c.want {
			t.Errorf("NormaliseSeverity(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseSummaryComponentPreserved(t *testing.T) {
	raw := `<analysis>
  <component>
    <source>pipe</source>
    <pattern>nginx failed to bind to port 80</pattern>
    <cause>port already in use by another process</cause>
    <severity>high</severity>
    <suggestion>run ss -tlnp | grep :80 to find the conflicting process and terminate it</suggestion>
  </component>
  <component>
    <source>summary</source>
    <pattern>One critical service failure detected requiring immediate attention</pattern>
    <cause>Port conflict is preventing nginx from starting</cause>
    <severity>high</severity>
    <suggestion>Resolve the port 80 conflict as described in the preceding component</suggestion>
  </component>
</analysis>`

	a := ParseDoctor(raw)
	if len(a.Components) != 2 {
		t.Fatalf("expected 2 components including summary, got %d", len(a.Components))
	}
	last := a.Components[len(a.Components)-1]
	if last.Source != "summary" {
		t.Errorf("last component source = %q, want summary", last.Source)
	}
	if last.Pattern == "" {
		t.Error("summary component pattern must not be empty")
	}
}

func TestParseHealthySummaryComponentKept(t *testing.T) {
	// A summary component with healthy content MUST be preserved.
	// The parser must not filter based on content.
	raw := `<analysis>
  <component>
    <source>summary</source>
    <pattern>No issues detected. All services are operating normally.</pattern>
    <cause>All monitored metrics are within normal operating ranges.</cause>
    <severity>low</severity>
    <suggestion>No action required.</suggestion>
  </component>
</analysis>`

	a := ParseDoctor(raw)
	if len(a.Components) != 1 {
		t.Errorf("healthy summary component must be preserved, got %d components", len(a.Components))
	}
	if a.Components[0].Source != "summary" {
		t.Errorf("source = %q, want summary", a.Components[0].Source)
	}
}
