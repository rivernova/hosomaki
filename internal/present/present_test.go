// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package present

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/analysis"
	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/insight"
	"github.com/rivernova/hosomaki/internal/render"
)

// unit tests for the present package.

func TestAnalysisInputPrefersFullKernel(t *testing.T) {
	in := AnalysisInput(&collector.SystemSnapshot{
		Environment: collector.Environment{
			Kernel:     "6.8.12",
			KernelFull: "Linux 6.8.12 x86_64",
		},
		Memory: "Mem: 1Gi",
	})
	if in.Kernel != "Linux 6.8.12 x86_64" {
		t.Errorf("expected full kernel, got %q", in.Kernel)
	}
	if in.Memory != "Mem: 1Gi" {
		t.Errorf("memory not carried through: %q", in.Memory)
	}
}

func TestAnalysisInputFallsBackToShortKernel(t *testing.T) {
	in := AnalysisInput(&collector.SystemSnapshot{
		Environment: collector.Environment{Kernel: "6.8.12"},
	})
	if in.Kernel != "6.8.12" {
		t.Errorf("expected short kernel fallback, got %q", in.Kernel)
	}
}

func TestAnalysisInputNilSnapshot(t *testing.T) {
	in := AnalysisInput(nil)
	if in.Kernel != "" || in.Memory != "" {
		t.Error("nil snapshot should yield zero Input")
	}
}

func TestRstatusMappings(t *testing.T) {
	cases := []struct {
		in   analysis.Level
		want render.Status
	}{
		{analysis.Neutral, render.Neutral},
		{analysis.OK, render.OK},
		{analysis.Info, render.Info},
		{analysis.Warn, render.Warn},
		{analysis.Crit, render.Crit},
	}
	for _, c := range cases {
		if got := rstatus(c.in); got != c.want {
			t.Errorf("rstatus(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestSeverityToStatusMappings(t *testing.T) {
	wantMap := map[string]render.Status{
		"crit": render.Crit, "warn": render.Warn,
		"ok": render.OK, "info": render.Info,
	}
	cases := []struct{ in, want string }{
		{"critical", "crit"},
		{"high", "crit"},
		{"medium", "warn"},
		{"low", "ok"},
		{"", "info"},
		{"nonsense", "info"},
	}
	for _, c := range cases {
		got := severityToStatus(c.in)
		if got != wantMap[c.want] {
			t.Errorf("severityToStatus(%q) = %v, want %v", c.in, got, wantMap[c.want])
		}
	}
}

func TestSourceDisplayName(t *testing.T) {
	cases := []struct{ source, want string }{
		{"service:nginx", "nginx"},
		{"service:postgresql", "postgresql"},
		{"file:error.log", "error.log"},
		{"dmesg", "kernel"},
		{"pipe", "system"},
		{"inline", "input"},
		{"summary", "summary"},
		{"", "system"},
		{"custom-source", "custom-source"},
	}
	for _, c := range cases {
		got := sourceDisplayName(c.source)
		if got != c.want {
			t.Errorf("sourceDisplayName(%q) = %q, want %q", c.source, got, c.want)
		}
	}
}

func TestDoctorReportMapsComponents(t *testing.T) {
	rep := analysis.Report{FailedCount: 1, Anomalies: 1}
	ai := insight.Analysis{
		Components: []insight.Component{{
			Source:     "service:NetworkManager",
			Severity:   "high",
			Pattern:    "dhcp timeout observed on interface eth0",
			Cause:      "conflict with wpa_supplicant binding to the same interface",
			Suggestion: "systemctl restart NetworkManager",
		}},
	}

	out := DoctorReport(rep, ai, false)

	if out.Title != "hosomaki doctor" {
		t.Errorf("unexpected title %q", out.Title)
	}
	if len(out.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(out.Components))
	}
	c := out.Components[0]
	if c.Source != "service:NetworkManager" {
		t.Errorf("source not carried: %q", c.Source)
	}
	if c.DisplayName != "NetworkManager" {
		t.Errorf("display name wrong: %q, want NetworkManager", c.DisplayName)
	}
	if c.Suggestion == "" {
		t.Error("suggestion should be present for doctor")
	}
	if len(c.Details) < 2 {
		t.Fatalf("expected at least 2 details (pattern+cause), got %d", len(c.Details))
	}
	if c.Details[0].Key != "detected pattern" {
		t.Errorf("first detail key = %q, want detected pattern", c.Details[0].Key)
	}
	if c.Details[1].Key != "probable cause" {
		t.Errorf("second detail key = %q, want probable cause", c.Details[1].Key)
	}
	if len(out.Summary) == 0 {
		t.Error("summary tallies should not be empty")
	}
}

func TestDoctorReportBriefOmitsProcessLines(t *testing.T) {
	out := DoctorReport(analysis.Report{}, insight.Analysis{}, true)
	if len(out.ProcessLines) != 0 {
		t.Error("brief doctor should omit process lines")
	}
}

func TestDoctorReportFullHasProcessLines(t *testing.T) {
	out := DoctorReport(analysis.Report{}, insight.Analysis{}, false)
	if len(out.ProcessLines) == 0 {
		t.Error("full doctor should include process lines")
	}
}

func TestDoctorReportRawFallback(t *testing.T) {
	out := DoctorReport(analysis.Report{}, insight.Analysis{Raw: "model said something"}, false)
	if out.RawInsight != "model said something" {
		t.Errorf("raw insight not carried: %q", out.RawInsight)
	}
}

func TestDoctorSummaryHealthy(t *testing.T) {
	out := DoctorReport(analysis.Report{}, insight.Analysis{}, false)
	if len(out.Summary) != 1 || out.Summary[0].Text != "healthy system" {
		t.Errorf("expected 'healthy system' summary for clean report, got %v", out.Summary)
	}
}

func TestStatusReportWithAnalysisStructure(t *testing.T) {
	rep := analysis.Report{FailedCount: 1}
	rep.Metrics = []analysis.Metric{{Label: "cpu", Value: "10%", Level: analysis.OK}}
	rep.Findings = []analysis.Finding{{Level: analysis.Warn, Text: "cups degraded"}}

	ai := insight.Analysis{
		Components: []insight.Component{{
			Source:   "pipe",
			Severity: "medium",
			Pattern:  "cups service degraded due to missing printer backend",
			Cause:    "the backend driver was removed in a recent update",
		}},
	}
	out := StatusReportWithAnalysis(rep, ai, false)

	if out.Title != "hosomaki status" {
		t.Errorf("unexpected title: %q", out.Title)
	}
	if len(out.Metrics) == 0 {
		t.Error("metrics should be populated")
	}
	if len(out.Services) == 0 {
		t.Error("services should come from findings")
	}
	if len(out.Components) == 0 {
		t.Error("components should come from AI analysis")
	}
	if len(out.Summary) == 0 {
		t.Error("summary should be populated")
	}
}

func TestStatusComponentNoSuggestion(t *testing.T) {
	ai := insight.Analysis{
		Components: []insight.Component{{
			Source:     "pipe",
			Pattern:    "high memory usage",
			Cause:      "multiple browser tabs consuming swap",
			Suggestion: "kill some processes",
		}},
	}
	out := StatusReportWithAnalysis(analysis.Report{}, ai, false)
	if len(out.Components) == 0 {
		t.Fatal("expected 1 component")
	}
	if out.Components[0].Suggestion != "" {
		t.Error("status components must not carry a suggestion")
	}
	// Cause must be preserved for status.
	hasCause := false
	for _, d := range out.Components[0].Details {
		if d.Key == "probable cause" {
			hasCause = true
		}
	}
	if !hasCause {
		t.Error("status components must carry cause detail")
	}
}

func TestStatusReportFallbackSummary(t *testing.T) {
	out := StatusReport(analysis.Report{FailedCount: 2}, false)
	if len(out.Summary) == 0 {
		t.Error("fallback summary should always be non-empty")
	}
}

func TestExplainReportStructuredComponents(t *testing.T) {
	ai := insight.Analysis{
		Components: []insight.Component{{
			Source:  "service:nginx",
			Pattern: "nginx failed to bind to port 80 — address already in use",
			Cause:   "a previous nginx process was not cleanly terminated and still holds the socket",
		}},
	}
	out := ExplainReport(render.InputInfo{Origin: "service", Detail: "nginx"}, "", ai)
	if len(out.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(out.Components))
	}
	c := out.Components[0]
	if c.Source != "service:nginx" {
		t.Errorf("source not carried: %q", c.Source)
	}
	if c.DisplayName != "nginx" {
		t.Errorf("display name wrong: %q, want nginx", c.DisplayName)
	}
}

func TestExplainReportProseFallback(t *testing.T) {
	ai := insight.Analysis{Raw: "something went wrong"}
	out := ExplainReport(render.InputInfo{}, "", ai)
	if out.RawText != "something went wrong" {
		t.Errorf("raw text not carried: %q", out.RawText)
	}
}

func TestExplainComponentNoSuggestion(t *testing.T) {
	ai := insight.Analysis{
		Components: []insight.Component{{
			Source:     "dmesg",
			Pattern:    "oom killer fired",
			Cause:      "memory exhausted",
			Suggestion: "this must be stripped for explain",
		}},
	}
	out := ExplainReport(render.InputInfo{}, "", ai)
	if len(out.Components) == 0 {
		t.Fatal("expected 1 component")
	}
	if out.Components[0].Suggestion != "" {
		t.Error("explain components must not carry a suggestion")
	}
	if out.Components[0].DisplayName != "kernel" {
		t.Errorf("display name wrong: %q, want kernel", out.Components[0].DisplayName)
	}
}

func TestContextLine(t *testing.T) {
	if ContextLine("") != "" {
		t.Error("empty command should produce empty context line")
	}
	got := ContextLine("docker compose up")
	if !strings.HasSuffix(got, "docker compose up") {
		t.Errorf("unexpected context line: %q", got)
	}
}

func TestPluralSingular(t *testing.T) {
	if plural(1, "service degraded", "services degraded") != "1 service degraded" {
		t.Error("plural() singular case wrong")
	}
	if plural(3, "service degraded", "services degraded") != "3 services degraded" {
		t.Error("plural() plural case wrong")
	}
}

func TestIsDisruptive(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"systemctl restart nginx", false},
		{"WARNING: potentially disruptive — backup first", true},
		{"This action is irreversible — proceed with caution", true},
		{"journalctl -xe | tail -50", false},
		{"data loss may occur if you proceed", true},
	}
	for _, c := range cases {
		got := isDisruptive(c.text)
		if got != c.want {
			t.Errorf("isDisruptive(%q) = %v, want %v", c.text, got, c.want)
		}
	}
}
