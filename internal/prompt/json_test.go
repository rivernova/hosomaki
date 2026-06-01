// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prompt

import (
	"strings"
	"testing"
	"time"

	"github.com/rivernova/hosomaki/internal/collector"
)

func TestDoctorFullPromptContainsJSONSentinel(t *testing.T) {
	p := Doctor(DoctorInput{CollectedAt: time.Now()}, false)
	if !strings.Contains(p, "---JSON---") {
		t.Error("Doctor() full prompt must contain ---JSON--- sentinel")
	}
	if !strings.Contains(p, "---END---") {
		t.Error("Doctor() full prompt must contain ---END--- sentinel")
	}
}

func TestDoctorBriefPromptContainsJSONSentinel(t *testing.T) {
	p := Doctor(DoctorInput{CollectedAt: time.Now()}, true)
	if !strings.Contains(p, "---JSON---") {
		t.Error("Doctor() brief prompt must contain ---JSON--- sentinel")
	}
}

func TestStatusFullPromptContainsJSONSentinel(t *testing.T) {
	p := Status(StatusInput{CollectedAt: time.Now()}, false)
	if !strings.Contains(p, "---JSON---") {
		t.Error("Status() full prompt must contain ---JSON--- sentinel")
	}
	if !strings.Contains(p, "---END---") {
		t.Error("Status() full prompt must contain ---END--- sentinel")
	}
}

func TestStatusBriefPromptContainsJSONSentinel(t *testing.T) {
	p := Status(StatusInput{CollectedAt: time.Now()}, true)
	if !strings.Contains(p, "---JSON---") {
		t.Error("Status() brief prompt must contain ---JSON--- sentinel")
	}
}

func TestExplainPromptContainsJSONSentinel(t *testing.T) {
	p := Explain("some error", "", collector.Environment{})
	if !strings.Contains(p, "---JSON---") {
		t.Error("Explain() prompt must contain ---JSON--- sentinel")
	}
	if !strings.Contains(p, "---END---") {
		t.Error("Explain() prompt must contain ---END--- sentinel")
	}
}

func TestDoctorJSONSentinelAppearsAfterAnalysis(t *testing.T) {
	p := Doctor(DoctorInput{CollectedAt: time.Now()}, false)
	jsonIdx := strings.Index(p, "---JSON---")
	snapshotIdx := strings.Index(p, "System snapshot:")
	if jsonIdx < snapshotIdx {
		t.Error("Doctor() ---JSON--- sentinel must appear after the system snapshot, not before")
	}
}

func TestStatusJSONSentinelAppearsAfterSnapshot(t *testing.T) {
	p := Status(StatusInput{CollectedAt: time.Now()}, false)
	jsonIdx := strings.Index(p, "---JSON---")
	snapshotIdx := strings.Index(p, "System snapshot:")
	if jsonIdx < snapshotIdx {
		t.Error("Status() ---JSON--- sentinel must appear after the system snapshot, not before")
	}
}

func TestStatusFullPromptStyle(t *testing.T) {
	p := Status(StatusInput{CollectedAt: time.Now()}, false)
	if !strings.Contains(p, "five to eight sentences") {
		t.Error("Status() full prompt should instruct five to eight sentences")
	}
	if !strings.Contains(p, "Do not suggest fixes") {
		t.Error("Status() full prompt should forbid suggesting fixes")
	}
}

func TestStatusBriefPromptStyle(t *testing.T) {
	p := Status(StatusInput{CollectedAt: time.Now()}, true)
	if !strings.Contains(p, "ONE sentence") {
		t.Error("Status() brief prompt should instruct exactly one sentence")
	}
	if strings.Contains(p, "five to eight sentences") {
		t.Error("Status() brief prompt should not contain full-mode instruction")
	}
}
