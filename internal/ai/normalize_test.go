// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"reflect"
	"testing"
)

// unit testing for the normalizing of the LLM response

type normItem struct {
	Severity   string `json:"severity"`
	Title      string `json:"title"`
	Port       string `json:"port"`
	Disruptive bool   `json:"disruptive"`
}

type normDoc struct {
	Summary  string     `json:"summary"`
	Findings []normItem `json:"findings"`
}

func TestCanonicalizeEnum(t *testing.T) {
	allowed := []string{"critical", "warning", "info"}
	cases := []struct {
		in      string
		want    string
		changed bool
	}{
		{"warning", "warning", false},
		{"Warning", "warning", true},
		{"WARNING", "warning", true},
		{"  warning  ", "warning", true},
		{"warn", "warning", true},
		{"crit", "critical", true},
		{"informational", "info", true},
		{"error", "error", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, changed := canonicalizeEnum(c.in, allowed)
		if got != c.want || changed != c.changed {
			t.Errorf("canonicalizeEnum(%q) = (%q,%v), want (%q,%v)", c.in, got, changed, c.want, c.changed)
		}
	}
}

func TestNormalizeJSON_EnumAndTypeAndTrim(t *testing.T) {
	enums := map[string][]string{"severity": {"critical", "warning", "info"}}
	raw := `{"severity":"Warning","title":"  spaced  ","port":8080,"disruptive":"false"}`
	got := normalizeJSON(raw, reflect.TypeFor[normItem](), enums)
	want := `{"disruptive":false,"port":"8080","severity":"warning","title":"spaced"}`
	if got != want {
		t.Fatalf("normalizeJSON:\n got  %s\n want %s", got, want)
	}
}

func TestNormalizeJSON_NestedSliceEnums(t *testing.T) {
	enums := map[string][]string{"severity": {"critical", "warning", "info"}}
	raw := `{"summary":"x","findings":[{"severity":"CRIT","title":"a","port":"22","disruptive":false}]}`
	got := normalizeJSON(raw, reflect.TypeFor[normDoc](), enums)
	want := `{"findings":[{"disruptive":false,"port":"22","severity":"critical","title":"a"}],"summary":"x"}`
	if got != want {
		t.Fatalf("normalizeJSON nested:\n got  %s\n want %s", got, want)
	}
}

func TestNormalizeJSON_EmptyEnumNotInvented(t *testing.T) {
	enums := map[string][]string{"severity": {"critical", "warning", "info"}}
	raw := `{"severity":"","title":"a","port":"1","disruptive":false}`
	got := normalizeJSON(raw, reflect.TypeFor[normItem](), enums)
	if got != `{"disruptive":false,"port":"1","severity":"","title":"a"}` {
		t.Fatalf("empty enum must stay empty (real error for the LLM/drop), got: %s", got)
	}
}

func TestNormalizeJSON_NonStringEnumFieldUntouchedWhenUnregistered(t *testing.T) {
	raw := `{"severity":"Warning","title":"a","port":"1","disruptive":false}`
	got := normalizeJSON(raw, reflect.TypeFor[normItem](), nil)
	if got != `{"disruptive":false,"port":"1","severity":"Warning","title":"a"}` {
		t.Fatalf("without registered enums, severity case must be preserved, got: %s", got)
	}
}

func TestNormalizeJSON_InvalidRawReturnedUnchanged(t *testing.T) {
	raw := `not json`
	if got := normalizeJSON(raw, reflect.TypeFor[normItem](), nil); got != raw {
		t.Fatalf("invalid raw must pass through untouched, got: %s", got)
	}
}
