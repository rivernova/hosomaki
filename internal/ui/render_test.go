// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/prompt"
)

// unit tests for rendering functions

func TestParseJSONPlain(t *testing.T) {
	var v map[string]string
	if err := ParseJSON(`{"key":"value"}`, &v); err != nil {
		t.Fatalf("ParseJSON() error = %v", err)
	}
	if v["key"] != "value" {
		t.Errorf("ParseJSON() key = %q, want value", v["key"])
	}
}

func TestParseJSONStripsMarkdownFence(t *testing.T) {
	var v map[string]int
	if err := ParseJSON("```json\n{\"n\":3}\n```", &v); err != nil {
		t.Fatalf("ParseJSON() with fence error = %v", err)
	}
	if v["n"] != 3 {
		t.Errorf("ParseJSON() n = %d, want 3", v["n"])
	}
}

func TestParseJSONStripsPlainFence(t *testing.T) {
	var v map[string]int
	if err := ParseJSON("```\n{\"n\":7}\n```", &v); err != nil {
		t.Fatalf("ParseJSON() with plain fence error = %v", err)
	}
	if v["n"] != 7 {
		t.Errorf("ParseJSON() n = %d, want 7", v["n"])
	}
}

func TestParseJSONIgnoresPreamble(t *testing.T) {
	raw := "Here is the analysis you requested:\n{\"what\":\"disk full\",\"why\":\"logs grew\"}"
	var v map[string]string
	if err := ParseJSON(raw, &v); err != nil {
		t.Fatalf("ParseJSON() with preamble error = %v", err)
	}
	if v["what"] != "disk full" {
		t.Errorf("ParseJSON() what = %q, want disk full", v["what"])
	}
}

func TestParseJSONIgnoresEpilogue(t *testing.T) {
	raw := `{"what":"OOM killer","why":"memory leak"}` + "\nI hope this helps."
	var v map[string]string
	if err := ParseJSON(raw, &v); err != nil {
		t.Fatalf("ParseJSON() with epilogue error = %v", err)
	}
	if v["why"] != "memory leak" {
		t.Errorf("ParseJSON() why = %q, want memory leak", v["why"])
	}
}

func TestParseJSONStringWithBraces(t *testing.T) {
	raw := `{"what":"process {nginx} crashed","why":"segfault in {worker}"}`
	var v map[string]string
	if err := ParseJSON(raw, &v); err != nil {
		t.Fatalf("ParseJSON() with braces-in-string error = %v", err)
	}
	if v["what"] != "process {nginx} crashed" {
		t.Errorf("ParseJSON() what = %q", v["what"])
	}
}

func TestParseJSONNestedObject(t *testing.T) {
	raw := `{"outer":{"inner":"value"}}`
	var v map[string]interface{}
	if err := ParseJSON(raw, &v); err != nil {
		t.Fatalf("ParseJSON() nested error = %v", err)
	}
}

func TestParseJSONEmptyStringReturnsError(t *testing.T) {
	var v map[string]string
	if err := ParseJSON("", &v); err == nil {
		t.Error("ParseJSON() with empty string should return an error")
	}
}

func TestParseJSONProseOnlyReturnsError(t *testing.T) {
	var v map[string]string
	if err := ParseJSON("this is just prose, no JSON at all", &v); err == nil {
		t.Error("ParseJSON() with prose only should return an error")
	}
}

func TestExtractJSONObjectSimple(t *testing.T) {
	s, ok := extractJSONObject(`{"a":"b"}`)
	if !ok {
		t.Fatal("extractJSONObject() ok = false, want true")
	}
	if s != `{"a":"b"}` {
		t.Errorf("extractJSONObject() = %q", s)
	}
}

func TestExtractJSONObjectWithPreamble(t *testing.T) {
	s, ok := extractJSONObject(`Some text before {"a":"b"} and after`)
	if !ok {
		t.Fatal("extractJSONObject() ok = false")
	}
	if s != `{"a":"b"}` {
		t.Errorf("extractJSONObject() = %q, want {\"a\":\"b\"}", s)
	}
}

func TestExtractJSONObjectNested(t *testing.T) {
	s, ok := extractJSONObject(`{"a":{"b":"c"}}`)
	if !ok {
		t.Fatal("extractJSONObject() ok = false")
	}
	if s != `{"a":{"b":"c"}}` {
		t.Errorf("extractJSONObject() = %q", s)
	}
}

func TestExtractJSONObjectBracesInString(t *testing.T) {
	input := `{"msg":"has {braces} inside"}`
	s, ok := extractJSONObject(input)
	if !ok {
		t.Fatal("extractJSONObject() ok = false")
	}
	if s != input {
		t.Errorf("extractJSONObject() = %q, want %q", s, input)
	}
}

func TestExtractJSONObjectNotFound(t *testing.T) {
	_, ok := extractJSONObject("no json here at all")
	if ok {
		t.Error("extractJSONObject() ok = true, want false")
	}
}

func TestExtractJSONObjectUnclosed(t *testing.T) {
	_, ok := extractJSONObject(`{"unclosed": "object"`)
	if ok {
		t.Error("extractJSONObject() ok = true for unclosed object, want false")
	}
}

func TestParseExplainJSONCanonicalKeys(t *testing.T) {
	raw := `{"what":"disk is full","why":"logs accumulated"}`
	var r prompt.ExplainResult
	if err := ParseExplainJSON(raw, &r); err != nil {
		t.Fatalf("ParseExplainJSON() error = %v", err)
	}
	if r.What != "disk is full" {
		t.Errorf("What = %q, want disk is full", r.What)
	}
	if r.Why != "logs accumulated" {
		t.Errorf("Why = %q, want logs accumulated", r.Why)
	}
}

func TestParseExplainJSONArrayValues(t *testing.T) {
	raw := `{
		"what": ["Kernel logging stopped.", "Signal 15 sent."],
		"why":  ["System is shutting down.", "SIGTERM triggered by systemd."]
	}`
	var r prompt.ExplainResult
	if err := ParseExplainJSON(raw, &r); err != nil {
		t.Fatalf("ParseExplainJSON() error = %v", err)
	}
	if !strings.Contains(r.What, "Kernel logging stopped") {
		t.Errorf("What = %q, expected to contain array items joined", r.What)
	}
	if !strings.Contains(r.Why, "SIGTERM") {
		t.Errorf("Why = %q, expected to contain array items joined", r.Why)
	}
}

func TestParseExplainJSONAliasKeys(t *testing.T) {
	raw := `{"what_is_happening":"nginx crashed","why_it_is_happening":"OOM killed it"}`
	var r prompt.ExplainResult
	if err := ParseExplainJSON(raw, &r); err != nil {
		t.Fatalf("ParseExplainJSON() error = %v", err)
	}
	if r.What != "nginx crashed" {
		t.Errorf("What = %q, want nginx crashed", r.What)
	}
	if r.Why != "OOM killed it" {
		t.Errorf("Why = %q, want OOM killed it", r.Why)
	}
}

func TestParseExplainJSONMixedArrayAndString(t *testing.T) {
	raw := `{"what":["Event A.","Event B."],"why":"Root cause here."}`
	var r prompt.ExplainResult
	if err := ParseExplainJSON(raw, &r); err != nil {
		t.Fatalf("ParseExplainJSON() error = %v", err)
	}
	if !strings.Contains(r.What, "Event A") {
		t.Errorf("What = %q, expected Event A", r.What)
	}
	if r.Why != "Root cause here." {
		t.Errorf("Why = %q, want Root cause here.", r.Why)
	}
}

func TestParseExplainJSONNoJSONReturnsError(t *testing.T) {
	var r prompt.ExplainResult
	if err := ParseExplainJSON("just prose, no JSON", &r); err == nil {
		t.Error("ParseExplainJSON() should return error when no JSON object found")
	}
}

func TestRenderDoctorEmptyResult(t *testing.T) {
	out := RenderDoctor(prompt.DoctorResult{})
	if !strings.Contains(out, "no issues detected") {
		t.Error("RenderDoctor() with empty result should show 'no issues detected'")
	}
	if !strings.Contains(out, "no actions required") {
		t.Error("RenderDoctor() with empty result should show 'no actions required'")
	}
}

func TestRenderDoctorFailedIssueUsesBulletFail(t *testing.T) {
	out := RenderDoctor(prompt.DoctorResult{
		Issues:  []prompt.DoctorIssue{{Severity: "failed", Summary: "nginx.service has failed"}},
		Actions: []prompt.DoctorAction{{Description: "Run systemctl restart nginx"}},
	})
	if !strings.Contains(out, "✗") {
		t.Error("RenderDoctor() failed issue should use fail bullet (✗)")
	}
	if !strings.Contains(out, "nginx.service has failed") {
		t.Error("RenderDoctor() should include the issue summary")
	}
}

func TestRenderDoctorWarningIssueUsesBulletWarn(t *testing.T) {
	out := RenderDoctor(prompt.DoctorResult{
		Issues:  []prompt.DoctorIssue{{Severity: "warning", Summary: "disk usage above 80%"}},
		Actions: []prompt.DoctorAction{{Description: "Run du -sh /var to find large directories"}},
	})
	if !strings.Contains(out, "!") {
		t.Error("RenderDoctor() warning issue should use warn bullet (!)")
	}
}

func TestRenderDoctorDisruptiveActionFlagged(t *testing.T) {
	out := RenderDoctor(prompt.DoctorResult{
		Issues:  []prompt.DoctorIssue{{Severity: "failed", Summary: "postgresql.service failed"}},
		Actions: []prompt.DoctorAction{{Description: "Run pg_resetwal", Disruptive: true}},
	})
	if !strings.Contains(out, "[disruptive]") {
		t.Error("RenderDoctor() disruptive action should be flagged with [disruptive]")
	}
}

func TestRenderDoctorNonDisruptiveActionNoFlag(t *testing.T) {
	out := RenderDoctor(prompt.DoctorResult{
		Issues:  []prompt.DoctorIssue{{Severity: "warning", Summary: "high memory pressure"}},
		Actions: []prompt.DoctorAction{{Description: "Run free -h", Disruptive: false}},
	})
	if strings.Contains(out, "[disruptive]") {
		t.Error("RenderDoctor() non-disruptive action must not be flagged")
	}
}

func TestRenderDoctorSectionsPresent(t *testing.T) {
	out := RenderDoctor(prompt.DoctorResult{
		Issues:  []prompt.DoctorIssue{{Severity: "warning", Summary: "something"}},
		Actions: []prompt.DoctorAction{{Description: "do something"}},
	})
	if !strings.Contains(out, "issues") {
		t.Error("RenderDoctor() output should contain 'issues' section")
	}
	if !strings.Contains(out, "suggested actions") {
		t.Error("RenderDoctor() output should contain 'suggested actions' section")
	}
}

func TestRenderDoctorSummaryCountsIssues(t *testing.T) {
	result := prompt.DoctorResult{
		Issues: []prompt.DoctorIssue{
			{Severity: "failed", Summary: "a"},
			{Severity: "warning", Summary: "b"},
			{Severity: "warning", Summary: "c"},
		},
		Actions: []prompt.DoctorAction{
			{Description: "x"},
			{Description: "y", Disruptive: true},
		},
	}
	out := RenderDoctorSummary(result)
	if !strings.Contains(out, "3 issues found") {
		t.Errorf("RenderDoctorSummary() should show 3 issues found, got:\n%s", out)
	}
	if !strings.Contains(out, "2 actions suggested") {
		t.Errorf("RenderDoctorSummary() should show 2 actions suggested, got:\n%s", out)
	}
	if !strings.Contains(out, "1 action flagged as disruptive") {
		t.Errorf("RenderDoctorSummary() should show 1 action flagged as disruptive, got:\n%s", out)
	}
}

func TestRenderDoctorSummaryNoDisruptiveWhenNone(t *testing.T) {
	result := prompt.DoctorResult{
		Issues:  []prompt.DoctorIssue{{Severity: "warning", Summary: "x"}},
		Actions: []prompt.DoctorAction{{Description: "y", Disruptive: false}},
	}
	out := RenderDoctorSummary(result)
	if strings.Contains(out, "disruptive") {
		t.Error("RenderDoctorSummary() should not mention disruptive when none flagged")
	}
}

func TestRenderStatusOverviewPresent(t *testing.T) {
	out := RenderStatus(prompt.StatusResult{
		Overview:  "The system is running well. Memory is fine. Disk is fine.",
		Anomalies: nil,
	})
	if !strings.Contains(out, "system overview") {
		t.Error("RenderStatus() should contain 'system overview' section")
	}
	if !strings.Contains(out, "running well") {
		t.Error("RenderStatus() should include the overview text")
	}
}

func TestRenderStatusAnomaliesSection(t *testing.T) {
	out := RenderStatus(prompt.StatusResult{
		Overview: "Fine.",
		Anomalies: []prompt.StatusAnomaly{
			{Severity: "failed", Summary: "sshd.service failed"},
			{Severity: "warning", Summary: "disk at 91%"},
		},
	})
	if !strings.Contains(out, "anomalies") {
		t.Error("RenderStatus() should contain 'anomalies' section")
	}
	if !strings.Contains(out, "sshd.service failed") {
		t.Error("RenderStatus() should include the anomaly summary")
	}
	if !strings.Contains(out, "disk at 91%") {
		t.Error("RenderStatus() should include the warning summary")
	}
}

func TestRenderStatusNoAnomalies(t *testing.T) {
	out := RenderStatus(prompt.StatusResult{Overview: "All good.", Anomalies: nil})
	if !strings.Contains(out, "no anomalies detected") {
		t.Error("RenderStatus() with no anomalies should show 'no anomalies detected'")
	}
}

func TestRenderStatusBriefShowsSummary(t *testing.T) {
	out := RenderStatusBrief(prompt.StatusBriefResult{Summary: "System is healthy."})
	if !strings.Contains(out, "System is healthy.") {
		t.Error("RenderStatusBrief() should include the summary text")
	}
}

func TestRenderExplainBothSections(t *testing.T) {
	out := RenderExplain(prompt.ExplainResult{
		What: "The OOM killer terminated nginx.",
		Why:  "Available memory was exhausted by a memory leak in a PHP worker.",
	})
	if !strings.Contains(out, "what is happening") {
		t.Error("RenderExplain() should contain 'what is happening' section")
	}
	if !strings.Contains(out, "why it is happening") {
		t.Error("RenderExplain() should contain 'why it is happening' section")
	}
	if !strings.Contains(out, "OOM killer") {
		t.Error("RenderExplain() should include the what text")
	}
	if !strings.Contains(out, "memory leak") {
		t.Error("RenderExplain() should include the why text")
	}
}

func TestRenderExplainEmptyResult(t *testing.T) {
	out := RenderExplain(prompt.ExplainResult{})
	if !strings.Contains(out, "(no information)") {
		t.Error("RenderExplain() with empty result should show '(no information)'")
	}
}
