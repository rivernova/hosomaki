// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/ai"
	"github.com/rivernova/hosomaki/internal/auditor"
	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/ui"
)

// each llm stream command runs start to finish against
// a fake LLM through four cases — a good answer, one needing repair, a cut-off one,
// and a multi-element streamed one

type commandCase struct {
	name           string
	run            func(prompt string) error
	goodStream     string
	repairStream   string
	repairResponse string
	streamedStream string
	markerOne      string
	markerTwo      string
	toleratesCut   bool
}

func runHosomakiCommandCapture(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	runErr := func() error {
		defer func() {
			os.Stdout = orig
			_ = w.Close()
		}()
		return fn()
	}()

	return <-done, runErr
}

func withFakeProvider(t *testing.T, f *fakeProvider) {
	t.Helper()
	orig := provider
	provider = f
	t.Cleanup(func() { provider = orig })
}

func commandCases() []commandCase {
	return []commandCase{
		{
			name:           "explain",
			run:            func(p string) error { return runExplain(p, false) },
			goodStream:     `{"issues":[{"what":"MARKER_A","why":"cause"}]}`,
			repairStream:   `{"issues":[{"what":"MARKER_A"}]}`,
			repairResponse: `{"what":"MARKER_A","why":"cause"}`,
			streamedStream: `{"issues":[{"what":"MARKER_A","why":"a"},{"what":"MARKER_B","why":"b"}]}`,
			markerOne:      "MARKER_A",
			markerTwo:      "MARKER_B",
			toleratesCut:   false,
		},
		{
			name:           "ports",
			run:            func(p string) error { return runPorts(p, false) },
			goodStream:     `{"summary":"s","findings":[{"severity":"warning","port":"22","title":"MARKER_A","detail":"d"}]}`,
			repairStream:   `{"summary":"s","findings":[{"severity":"warning","title":"MARKER_A"}]}`,
			repairResponse: `{"severity":"warning","port":"22","title":"MARKER_A","detail":"d"}`,
			streamedStream: `{"summary":"s","findings":[{"severity":"warning","port":"22","title":"MARKER_A","detail":"d"},{"severity":"info","port":"80","title":"MARKER_B","detail":"d"}]}`,
			markerOne:      "MARKER_A",
			markerTwo:      "MARKER_B",
			toleratesCut:   true,
		},
		{
			name:           "mounts",
			run:            func(p string) error { return runMounts(p, false) },
			goodStream:     `{"summary":"s","findings":[{"severity":"warning","mount_point":"/data","title":"MARKER_A","detail":"d"}]}`,
			repairStream:   `{"summary":"s","findings":[{"severity":"warning","title":"MARKER_A"}]}`,
			repairResponse: `{"severity":"warning","mount_point":"/data","title":"MARKER_A","detail":"d"}`,
			streamedStream: `{"summary":"s","findings":[{"severity":"warning","mount_point":"/data","title":"MARKER_A","detail":"d"},{"severity":"info","mount_point":"/srv","title":"MARKER_B","detail":"d"}]}`,
			markerOne:      "MARKER_A",
			markerTwo:      "MARKER_B",
			toleratesCut:   true,
		},
		{
			name:           "timers",
			run:            func(p string) error { return runTimers(p, 1, false) },
			goodStream:     `{"summary":"s","timers":[{"name":"MARKER_A","schedule":"daily","last_run":"x","next_run":"y","status":"ok","detail":"d"}]}`,
			repairStream:   `{"summary":"s","timers":[{"name":"MARKER_A","schedule":"daily"}]}`,
			repairResponse: `{"name":"MARKER_A","schedule":"daily","last_run":"x","next_run":"y","status":"ok","detail":"d"}`,
			streamedStream: `{"summary":"s","timers":[{"name":"MARKER_A","schedule":"daily","last_run":"x","next_run":"y","status":"ok","detail":"d"},{"name":"MARKER_B","schedule":"weekly","last_run":"x","next_run":"y","status":"ok","detail":"d"}]}`,
			markerOne:      "MARKER_A",
			markerTwo:      "MARKER_B",
			toleratesCut:   true,
		},
		{
			name:           "crons",
			run:            func(p string) error { return runCrons(p, 1, false) },
			goodStream:     `{"summary":"s","jobs":[{"source":"MARKER_A","schedule":"daily","command":"c","what_it_does":"w","last_run":"x","status":"ok","detail":"d"}]}`,
			repairStream:   `{"summary":"s","jobs":[{"source":"MARKER_A","schedule":"daily"}]}`,
			repairResponse: `{"source":"MARKER_A","schedule":"daily","command":"c","what_it_does":"w","last_run":"x","status":"ok","detail":"d"}`,
			streamedStream: `{"summary":"s","jobs":[{"source":"MARKER_A","schedule":"daily","command":"c","what_it_does":"w","last_run":"x","status":"ok","detail":"d"},{"source":"MARKER_B","schedule":"weekly","command":"c","what_it_does":"w","last_run":"x","status":"ok","detail":"d"}]}`,
			markerOne:      "MARKER_A",
			markerTwo:      "MARKER_B",
			toleratesCut:   true,
		},
		{
			name:           "doctor-full",
			run:            func(p string) error { return runDoctorFull(ui.SnapshotData{}, p, false) },
			goodStream:     `{"issues":[{"severity":"warning","title":"MARKER_A","detail":"d"}],"actions":[{"description":"act","disruptive":false}]}`,
			repairStream:   `{"issues":[{"severity":"warning","title":"MARKER_A"}],"actions":[{"description":"act","disruptive":false}]}`,
			repairResponse: `{"severity":"warning","title":"MARKER_A","detail":"d"}`,
			streamedStream: `{"issues":[{"severity":"warning","title":"MARKER_A","detail":"d"},{"severity":"info","title":"MARKER_B","detail":"d"}],"actions":[{"description":"act","disruptive":false}]}`,
			markerOne:      "MARKER_A",
			markerTwo:      "MARKER_B",
			toleratesCut:   true,
		},
		{
			name:           "status-full",
			run:            func(p string) error { return runStatusFull(ui.SnapshotData{}, p, false) },
			goodStream:     `{"overview":"ok","anomalies":[{"severity":"warning","title":"MARKER_A","detail":"d"}]}`,
			repairStream:   `{"overview":"ok","anomalies":[{"severity":"warning","title":"MARKER_A"}]}`,
			repairResponse: `{"severity":"warning","title":"MARKER_A","detail":"d"}`,
			streamedStream: `{"overview":"ok","anomalies":[{"severity":"warning","title":"MARKER_A","detail":"d"},{"severity":"info","title":"MARKER_B","detail":"d"}]}`,
			markerOne:      "MARKER_A",
			markerTwo:      "MARKER_B",
			toleratesCut:   true,
		},
		{
			name: "audit",
			run: func(_ string) error {
				return runAuditAI(context.Background(), &auditor.AuditDiff{}, "1h", collector.Environment{}, false)
			},
			goodStream:     `{"summary":"s","findings":[{"severity":"warning","category":"service","title":"MARKER_A","detail":"d"}]}`,
			repairStream:   `{"summary":"s","findings":[{"severity":"warning","category":"service","title":"MARKER_A"}]}`,
			repairResponse: `{"severity":"warning","category":"service","title":"MARKER_A","detail":"d"}`,
			streamedStream: `{"summary":"s","findings":[{"severity":"warning","category":"service","title":"MARKER_A","detail":"d"},{"severity":"info","category":"file","title":"MARKER_B","detail":"d"}]}`,
			markerOne:      "MARKER_A",
			markerTwo:      "MARKER_B",
			toleratesCut:   true,
		},
		{
			name: "firewall",
			run: func(_ string) error {
				return runFirewallLlm(collector.FirewallResult{
					Backend:    collector.BackendIptables,
					ReadStatus: collector.ReadOK,
					Rules:      []collector.FirewallRule{{Backend: collector.BackendIptables, Action: "ACCEPT", Port: "22"}},
				}, false, false)
			},
			goodStream:     `{"summary":"s","findings":[{"severity":"warning","rule":"r","port":"22","title":"MARKER_A","detail":"d"}]}`,
			repairStream:   `{"summary":"s","findings":[{"severity":"warning","rule":"r","port":"22","title":"MARKER_A"}]}`,
			repairResponse: `{"severity":"warning","rule":"r","port":"22","title":"MARKER_A","detail":"d"}`,
			streamedStream: `{"summary":"s","findings":[{"severity":"warning","rule":"r","port":"22","title":"MARKER_A","detail":"d"},{"severity":"info","rule":"r2","port":"80","title":"MARKER_B","detail":"d"}]}`,
			markerOne:      "MARKER_A",
			markerTwo:      "MARKER_B",
			toleratesCut:   true,
		},
	}
}

func TestCommandPipeline_GoodAnswer(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	for _, tc := range commandCases() {
		t.Run(tc.name, func(t *testing.T) {
			withFakeProvider(t, &fakeProvider{stream: tc.goodStream})
			out, err := runHosomakiCommandCapture(t, func() error { return tc.run("prompt") })
			if err != nil {
				t.Fatalf("good answer should succeed, got: %v", err)
			}
			if !strings.Contains(out, tc.markerOne) {
				t.Fatalf("rendered output missing %q:\n%s", tc.markerOne, out)
			}
		})
	}
}

func TestCommandPipeline_BrokenTriggersRepair(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	for _, tc := range commandCases() {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeProvider{stream: tc.repairStream, repairs: []string{tc.repairResponse}}
			withFakeProvider(t, f)
			out, err := runHosomakiCommandCapture(t, func() error { return tc.run("prompt") })
			if err != nil && !errors.Is(err, ai.ErrIncomplete) {
				t.Fatalf("repairable answer should not hard-fail, got: %v", err)
			}
			if f.jsonCall == 0 {
				t.Fatal("expected a repair call, provider repair was never invoked")
			}
			if !strings.Contains(out, tc.markerOne) {
				t.Fatalf("repaired output missing %q:\n%s", tc.markerOne, out)
			}
		})
	}
}

func TestCommandPipeline_CutOffAnswer(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	for _, tc := range commandCases() {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeProvider{stream: "{}", repairs: []string{"{}"}}
			withFakeProvider(t, f)
			_, err := runHosomakiCommandCapture(t, func() error { return tc.run("prompt") })
			if tc.toleratesCut {
				if err != nil {
					t.Fatalf("%s should degrade gracefully on cut-off, got: %v", tc.name, err)
				}
				return
			}
			if !errors.Is(err, ai.ErrIncomplete) {
				t.Fatalf("%s should surface ErrIncomplete on cut-off, got: %v", tc.name, err)
			}
		})
	}
}

func TestCommandPipeline_StreamedMultipleItems(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	for _, tc := range commandCases() {
		t.Run(tc.name, func(t *testing.T) {
			withFakeProvider(t, &fakeProvider{stream: tc.streamedStream})
			out, err := runHosomakiCommandCapture(t, func() error { return tc.run("prompt") })
			if err != nil {
				t.Fatalf("streamed answer should succeed, got: %v", err)
			}
			if !strings.Contains(out, tc.markerOne) || !strings.Contains(out, tc.markerTwo) {
				t.Fatalf("streamed output missing one of %q/%q:\n%s", tc.markerOne, tc.markerTwo, out)
			}
		})
	}
}

type nonStreamCase struct {
	name   string
	run    func(prompt string) error
	good   string
	broken string
	fixed  string
	marker string
}

func nonStreamCases() []nonStreamCase {
	return []nonStreamCase{
		{
			name:   "why",
			run:    func(p string) error { return runWhy(p, false) },
			good:   `{"summary":"MARKER_W","chain":[{"event":"e","detail":"d"}],"next_steps":["x"]}`,
			broken: `{"chain":[{"event":"e","detail":"d"}],"next_steps":["x"]}`,
			fixed:  `{"summary":"MARKER_W","chain":[{"event":"e","detail":"d"}],"next_steps":["x"]}`,
			marker: "MARKER_W",
		},
		{
			name:   "doctor-brief",
			run:    func(p string) error { return runDoctorBrief(ui.SnapshotData{}, p, false) },
			good:   `{"summary":"MARKER_D"}`,
			broken: `{}`,
			fixed:  `{"summary":"MARKER_D"}`,
			marker: "MARKER_D",
		},
		{
			name:   "status-brief",
			run:    func(p string) error { return runStatusBrief(ui.SnapshotData{}, p, false) },
			good:   `{"summary":"MARKER_S"}`,
			broken: `{}`,
			fixed:  `{"summary":"MARKER_S"}`,
			marker: "MARKER_S",
		},
	}
}

func TestCommandPipeline_NonStreamGoodAnswer(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	for _, tc := range nonStreamCases() {
		t.Run(tc.name, func(t *testing.T) {
			withFakeProvider(t, &fakeProvider{repairs: []string{tc.good}})
			out, err := runHosomakiCommandCapture(t, func() error { return tc.run("prompt") })
			if err != nil {
				t.Fatalf("good answer should succeed, got: %v", err)
			}
			if !strings.Contains(out, tc.marker) {
				t.Fatalf("rendered output missing %q:\n%s", tc.marker, out)
			}
		})
	}
}

func TestCommandPipeline_NonStreamBrokenTriggersRepair(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	for _, tc := range nonStreamCases() {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeProvider{repairs: []string{tc.broken, tc.fixed}}
			withFakeProvider(t, f)
			out, err := runHosomakiCommandCapture(t, func() error { return tc.run("prompt") })
			if err != nil {
				t.Fatalf("repairable answer should succeed, got: %v", err)
			}
			if f.jsonCall < 2 {
				t.Fatalf("expected an initial call plus a repair, got %d calls", f.jsonCall)
			}
			if !strings.Contains(out, tc.marker) {
				t.Fatalf("repaired output missing %q:\n%s", tc.marker, out)
			}
		})
	}
}

func TestCommandPipeline_NonStreamUnrepairableFails(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	for _, tc := range nonStreamCases() {
		t.Run(tc.name, func(t *testing.T) {
			withFakeProvider(t, &fakeProvider{repairs: []string{tc.broken}})
			_, err := runHosomakiCommandCapture(t, func() error { return tc.run("prompt") })
			if err == nil {
				t.Fatalf("%s should hard-fail when repair never succeeds", tc.name)
			}
		})
	}
}
