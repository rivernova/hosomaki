// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package insight

import (
	"testing"
)

// unit testing for insight parsing

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
		t.Errorf("source = %q, want %q", c.Source, "pipe")
	}
	if c.Pattern == "" {
		t.Error("pattern should not be empty")
	}
	if c.Cause == "" {
		t.Error("cause should not be empty")
	}
	if c.Severity != "high" {
		t.Errorf("severity = %q, want %q", c.Severity, "high")
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

func TestParseStatusStripssuggestion(t *testing.T) {
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

func TestParseKeyValueBlocksFallback(t *testing.T) {
	raw := `pattern: the sshd service rejected all incoming connections
cause: the host key file /etc/ssh/ssh_host_rsa_key has incorrect permissions

pattern: systemd-journald hit its rate limit
cause: a runaway process was producing thousands of log entries per second`

	a := parseAnalysis(raw)
	if len(a.Components) != 2 {
		t.Fatalf("expected 2 components from key-value fallback, got %d", len(a.Components))
	}
	if a.Components[0].Pattern == "" {
		t.Error("first component pattern should not be empty")
	}
	if a.Components[1].Cause == "" {
		t.Error("second component cause should not be empty")
	}
}

func TestHealthyContentSuppressed(t *testing.T) {
	raw := `<analysis>
  <component>
    <source>pipe</source>
    <pattern>no issues found</pattern>
    <cause>the system is healthy and operating normally</cause>
  </component>
</analysis>`

	a := ParseDoctor(raw)
	if len(a.Components) != 0 {
		t.Errorf("healthy content should be suppressed, got %d components", len(a.Components))
	}
}

func TestIsHealthyContent(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"no issues detected", true},
		{"system is healthy", true},
		{"no errors found in the journal", true},
		{"nginx service is failing to bind", false},
		{"oom killer was activated", false},
		{"all systems operational", true},
	}
	for _, c := range cases {
		got := isHealthyContent(c.text)
		if got != c.want {
			t.Errorf("isHealthyContent(%q) = %v, want %v", c.text, got, c.want)
		}
	}
}
