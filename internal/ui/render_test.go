// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/prompt"
)

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

func TestParseExplainJSONCanonicalArrayShape(t *testing.T) {
	raw := `{"issues":[{"what":"disk is full","why":"logs accumulated"},{"what":"sshd failed","why":"port in use"}]}`
	var r prompt.ExplainResult
	if err := ParseExplainJSON(raw, &r); err != nil {
		t.Fatalf("ParseExplainJSON() error = %v", err)
	}
	if len(r.Issues) != 2 {
		t.Fatalf("Issues len = %d, want 2", len(r.Issues))
	}
	if r.Issues[0].What != "disk is full" {
		t.Errorf("Issues[0].What = %q, want disk is full", r.Issues[0].What)
	}
	if r.Issues[1].Why != "port in use" {
		t.Errorf("Issues[1].Why = %q, want port in use", r.Issues[1].Why)
	}
}

func TestParseExplainJSONFallbackSingleEntry(t *testing.T) {
	raw := `{"what":"disk is full","why":"logs accumulated"}`
	var r prompt.ExplainResult
	if err := ParseExplainJSON(raw, &r); err != nil {
		t.Fatalf("ParseExplainJSON() error = %v", err)
	}
	if len(r.Issues) != 1 {
		t.Fatalf("Issues len = %d, want 1", len(r.Issues))
	}
	if r.Issues[0].What != "disk is full" {
		t.Errorf("Issues[0].What = %q, want disk is full", r.Issues[0].What)
	}
}

func TestParseExplainJSONArrayValues(t *testing.T) {
	raw := `{"issues":[{"what":["Kernel logging stopped.","Signal 15 sent."],"why":["System shutting down.","SIGTERM by systemd."]}]}`
	var r prompt.ExplainResult
	if err := ParseExplainJSON(raw, &r); err != nil {
		t.Fatalf("ParseExplainJSON() error = %v", err)
	}
	if len(r.Issues) != 1 {
		t.Fatalf("Issues len = %d, want 1", len(r.Issues))
	}
	if !strings.Contains(r.Issues[0].What, "Kernel logging stopped") {
		t.Errorf("Issues[0].What = %q, expected array items joined", r.Issues[0].What)
	}
	if !strings.Contains(r.Issues[0].Why, "SIGTERM") {
		t.Errorf("Issues[0].Why = %q, expected array items joined", r.Issues[0].Why)
	}
}

func TestParseExplainJSONAliasKeys(t *testing.T) {
	raw := `{"issues":[{"what_is_happening":"nginx crashed","why_it_is_happening":"OOM killed it"}]}`
	var r prompt.ExplainResult
	if err := ParseExplainJSON(raw, &r); err != nil {
		t.Fatalf("ParseExplainJSON() error = %v", err)
	}
	if len(r.Issues) != 1 {
		t.Fatalf("Issues len = %d, want 1", len(r.Issues))
	}
	if r.Issues[0].What != "nginx crashed" {
		t.Errorf("Issues[0].What = %q, want nginx crashed", r.Issues[0].What)
	}
}

func TestParseExplainJSONMarkdownFenceAndPreamble(t *testing.T) {
	raw := "Here is the analysis:\n```json\n{\"issues\":[{\"what\":\"disk full\",\"why\":\"logs grew\"}]}\n```"
	var r prompt.ExplainResult
	if err := ParseExplainJSON(raw, &r); err != nil {
		t.Fatalf("ParseExplainJSON() error = %v", err)
	}
	if len(r.Issues) != 1 || r.Issues[0].What != "disk full" {
		t.Errorf("Issues = %v, want one entry with What=disk full", r.Issues)
	}
}

func TestParseExplainJSONNoJSONReturnsError(t *testing.T) {
	var r prompt.ExplainResult
	if err := ParseExplainJSON("just prose, no JSON", &r); err == nil {
		t.Error("ParseExplainJSON() should return error when no JSON object found")
	}
}

func TestRenderExplainSingleIssueNoNumber(t *testing.T) {
	out := RenderExplain(prompt.ExplainResult{
		Issues: []prompt.ExplainEntry{
			{What: "The OOM killer terminated nginx.", Why: "Memory exhausted by a PHP leak."},
		},
	})
	if !strings.Contains(out, "what is happening") {
		t.Error("RenderExplain() single issue should show plain 'what is happening' title")
	}
	if !strings.Contains(out, "why it is happening") {
		t.Error("RenderExplain() single issue should show plain 'why it is happening' title")
	}
	if strings.Contains(out, "issue 1") {
		t.Error("RenderExplain() single issue should not show issue number")
	}
	if !strings.Contains(out, "OOM killer") {
		t.Error("RenderExplain() should include what text")
	}
}

func TestRenderExplainMultipleIssuesNumbered(t *testing.T) {
	out := RenderExplain(prompt.ExplainResult{
		Issues: []prompt.ExplainEntry{
			{What: "sshd failed to start.", Why: "Port 22 already bound."},
			{What: "nginx exited.", Why: "Config syntax error."},
		},
	})
	if !strings.Contains(out, "issue 1") {
		t.Error("RenderExplain() multiple issues should show 'issue 1' heading")
	}
	if !strings.Contains(out, "issue 2") {
		t.Error("RenderExplain() multiple issues should show 'issue 2' heading")
	}
	if !strings.Contains(out, "sshd failed") {
		t.Error("RenderExplain() should include first issue what text")
	}
	if !strings.Contains(out, "nginx exited") {
		t.Error("RenderExplain() should include second issue what text")
	}
}

func TestRenderExplainEmptyResult(t *testing.T) {
	out := RenderExplain(prompt.ExplainResult{})
	if !strings.Contains(out, "(no information)") {
		t.Error("RenderExplain() with empty result should show '(no information)'")
	}
}
