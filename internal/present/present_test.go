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

func TestSevStatusMappings(t *testing.T) {
	cases := []struct{ in, want string }{
		{"critical", "crit"},
		{"warning", "warn"},
		{"healthy", "ok"},
		{"info", "info"},
		{"nonsense", "info"},
	}
	for _, c := range cases {
		got := sevStatus(c.in)
		wantStatus := map[string]render.Status{
			"crit": render.Crit, "warn": render.Warn,
			"ok": render.OK, "info": render.Info,
		}[c.want]
		if got != wantStatus {
			t.Errorf("sevStatus(%q) = %v, want %v", c.in, got, wantStatus)
		}
	}
}

func TestDoctorReportMapsIssues(t *testing.T) {
	rep := analysis.Report{FailedCount: 1, Anomalies: 1}
	ai := insight.Doctor{
		Summary: "one service down",
		Issues: []insight.Issue{{
			Subject:  "NetworkManager",
			Severity: "warn",
			Pattern:  "dhcp timeout",
			Cause:    "conflict with wpa_supplicant",
			Details:  []string{"recurring every boot"},
			Actions:  []insight.Action{{Description: "restart it", Command: "systemctl restart NetworkManager", Disruptive: true}},
		}},
	}

	out := DoctorReport(rep, ai, false)

	if out.Title != "hosomaki doctor" {
		t.Errorf("unexpected title %q", out.Title)
	}
	if len(out.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(out.Issues))
	}
	iss := out.Issues[0]
	if iss.Subject != "NetworkManager" {
		t.Errorf("subject not carried: %q", iss.Subject)
	}
	// pattern + cause + one plain detail = 3 detail rows
	if len(iss.Details) != 3 {
		t.Fatalf("expected 3 detail rows, got %d", len(iss.Details))
	}
	if iss.Details[0].Key != "detected pattern" || iss.Details[1].Key != "probable cause" {
		t.Error("structured details mislabelled")
	}
	if len(iss.Actions) != 1 || !iss.Actions[0].Disruptive {
		t.Error("action mapping lost the disruptive flag")
	}
	if len(out.Summary) == 0 {
		t.Error("summary tallies should not be empty")
	}
}

func TestDoctorReportBriefOmitsProcessLines(t *testing.T) {
	out := DoctorReport(analysis.Report{}, insight.Doctor{Summary: "ok"}, true)
	if len(out.ProcessLines) != 0 {
		t.Error("brief doctor should omit process lines")
	}
}

func TestDoctorReportFullHasProcessLines(t *testing.T) {
	out := DoctorReport(analysis.Report{}, insight.Doctor{Summary: "ok"}, false)
	if len(out.ProcessLines) == 0 {
		t.Error("full doctor should include process lines")
	}
}

func TestDoctorReportRawFallback(t *testing.T) {
	out := DoctorReport(analysis.Report{}, insight.Doctor{Raw: "model said something"}, false)
	if out.RawInsight != "model said something" {
		t.Errorf("raw insight not carried: %q", out.RawInsight)
	}
}

func TestDoctorSummaryHealthy(t *testing.T) {
	out := DoctorReport(analysis.Report{}, insight.Doctor{}, false)
	if len(out.Summary) != 1 || out.Summary[0].Text != "system healthy" {
		t.Error("expected 'system healthy' summary for clean report")
	}
}

func TestStatusReportStructure(t *testing.T) {
	rep := analysis.Report{FailedCount: 1}
	rep.Metrics = []analysis.Metric{{Label: "cpu", Value: "10%", Level: analysis.OK}}
	rep.Findings = []analysis.Finding{{Level: analysis.Warn, Text: "cups degraded"}}

	ai := insight.Status{
		Observations: []insight.Observation{{Level: "warn", Text: "1 service with warnings"}},
	}
	out := StatusReport(rep, ai)

	if out.Title != "hosomaki status" {
		t.Errorf("unexpected title: %q", out.Title)
	}
	if len(out.Metrics) == 0 {
		t.Error("metrics should be populated")
	}
	if len(out.Services) == 0 {
		t.Error("services should come from findings")
	}
	if len(out.Summary) == 0 {
		t.Error("summary should be populated from AI observations")
	}
	if out.Summary[0].Text != "1 service with warnings" {
		t.Errorf("summary text not carried from AI: %q", out.Summary[0].Text)
	}
}

func TestStatusReportFallbackSummary(t *testing.T) {
	// No AI observations → fall back to deterministic counts.
	out := StatusReport(analysis.Report{FailedCount: 2}, insight.Status{})
	if len(out.Summary) == 0 {
		t.Error("fallback summary should always be non-empty")
	}
}

func TestExplainReportStructuredIssues(t *testing.T) {
	ai := insight.Doctor{
		Issues: []insight.Issue{{Subject: "NetworkManager", Severity: "warn", Pattern: "dhcp timeout"}},
	}
	out := ExplainReport("", ai)
	if len(out.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(out.Issues))
	}
	if out.Issues[0].Subject != "NetworkManager" {
		t.Errorf("issue subject not carried: %q", out.Issues[0].Subject)
	}
}

func TestExplainReportProseFallback(t *testing.T) {
	ai := insight.Doctor{Raw: "something went wrong"}
	out := ExplainReport("", ai)
	if out.RawText != "something went wrong" {
		t.Errorf("raw text not carried: %q", out.RawText)
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
