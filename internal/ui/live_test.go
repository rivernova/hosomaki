// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/prompt"
)

// unit tests for streaming render functions

func TestRenderUpdatesFindingLive_SecurityShowsDetail(t *testing.T) {
	f := prompt.UpdateFinding{
		Package:   "code",
		Available: "1.125.1-1",
		Category:  "security",
		Detail:    "Fixes a remote code execution vulnerability.",
	}
	got := RenderUpdatesFindingLive(f, 0)
	if !strings.Contains(got, "Fixes a remote code execution vulnerability.") {
		t.Errorf("RenderUpdatesFindingLive() for security finding missing detail: %q", got)
	}
}

func TestRenderUpdatesFindingLive_MajorShowsDetail(t *testing.T) {
	f := prompt.UpdateFinding{
		Package:   "docker-ce",
		Available: "3:29.6.0-1",
		Category:  "major",
		Detail:    "Major version bump; review breaking changes before upgrading.",
	}
	got := RenderUpdatesFindingLive(f, 0)
	if !strings.Contains(got, "Major version bump") {
		t.Errorf("RenderUpdatesFindingLive() for major finding missing detail: %q", got)
	}
}

func TestRenderUpdatesFindingLive_MinorOmitsDetailEvenIfPresent(t *testing.T) {
	f := prompt.UpdateFinding{
		Package:   "libuuid",
		Available: "2.41.5-1",
		Category:  "minor",
		Detail:    "Routine patch release.",
	}
	got := RenderUpdatesFindingLive(f, 0)
	if strings.Contains(got, "Routine patch release.") {
		t.Errorf("RenderUpdatesFindingLive() for minor finding should omit detail, got: %q", got)
	}
}

func TestRenderUpdatesFindingLive_UnknownOmitsDetail(t *testing.T) {
	f := prompt.UpdateFinding{
		Package:   "liblastlog2",
		Available: "2.41.5-1",
		Category:  "unknown",
		Detail:    "Could not determine category.",
	}
	got := RenderUpdatesFindingLive(f, 0)
	if strings.Contains(got, "Could not determine category.") {
		t.Errorf("RenderUpdatesFindingLive() for unknown finding should omit detail, got: %q", got)
	}
}

func TestRenderUpdatesFindingLive_EmptyDetailAddsNothing(t *testing.T) {
	f := prompt.UpdateFinding{
		Package:   "code",
		Available: "1.125.1-1",
		Category:  "security",
		Detail:    "",
	}
	got := RenderUpdatesFindingLive(f, 0)
	if strings.Count(got, "\n") != 2 {
		t.Errorf("RenderUpdatesFindingLive() with empty detail should have no detail line, got: %q", got)
	}
}

func TestRenderUpdatesFindingLive_EmptyPackageReturnsEmpty(t *testing.T) {
	f := prompt.UpdateFinding{Package: "", Detail: "should not appear"}
	if got := RenderUpdatesFindingLive(f, 0); got != "" {
		t.Errorf("RenderUpdatesFindingLive() with empty package = %q, want empty string", got)
	}
}

func TestRenderUpdatesFindingLive_NoAvailableVersionStillTerminatesLine(t *testing.T) {
	f := prompt.UpdateFinding{Package: "mystery-pkg", Available: "", Category: "unknown"}
	got := RenderUpdatesFindingLive(f, 0)
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("RenderUpdatesFindingLive() with no version info should still end in a newline, got: %q", got)
	}
}
