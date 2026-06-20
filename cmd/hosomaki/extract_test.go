package hosomaki

import (
	"encoding/json"
	"testing"

	"github.com/rivernova/hosomaki/internal/historian"
)

func TestExtractSummaryExplain(t *testing.T) {
	result, _ := json.Marshal(map[string]any{
		"issues": []map[string]any{
			{"what": "High CPU usage from kernel process", "why": "Memory pressure"},
		},
	})
	e := historian.HistoryEntry{Command: "explain", Result: result}
	got := extractSummary(e)
	want := "High CPU usage from kernel process"
	if got != want {
		t.Errorf("extractSummary(explain) = %q, want %q", got, want)
	}
}

func TestExtractSummaryWhy(t *testing.T) {
	result, _ := json.Marshal(map[string]any{
		"summary": "Service failed due to OOM kill",
	})
	e := historian.HistoryEntry{Command: "why", Result: result}
	got := extractSummary(e)
	want := "Service failed due to OOM kill"
	if got != want {
		t.Errorf("extractSummary(why) = %q, want %q", got, want)
	}
}

func TestExtractSummaryAudit(t *testing.T) {
	result, _ := json.Marshal(map[string]any{
		"summary": "3 config files changed since last audit",
		"findings": []map[string]any{
			{"title": "SSH config modified"},
		},
	})
	e := historian.HistoryEntry{Command: "audit", Result: result}
	got := extractSummary(e)
	want := "3 config files changed since last audit"
	if got != want {
		t.Errorf("extractSummary(audit) = %q, want %q", got, want)
	}
}

func TestExtractSummaryStatusFull(t *testing.T) {
	result, _ := json.Marshal(map[string]any{
		"overview": "System healthy, 2 anomalies detected",
		"anomalies": []map[string]any{
			{"title": "High disk I/O"},
		},
	})
	e := historian.HistoryEntry{Command: "status", Result: result}
	got := extractSummary(e)
	want := "System healthy, 2 anomalies detected"
	if got != want {
		t.Errorf("extractSummary(status full) = %q, want %q", got, want)
	}
}

func TestExtractSummaryStatusBrief(t *testing.T) {
	result, _ := json.Marshal(map[string]any{
		"summary": "System running smoothly",
	})
	e := historian.HistoryEntry{Command: "status", Result: result}
	got := extractSummary(e)
	want := "System running smoothly"
	if got != want {
		t.Errorf("extractSummary(status brief) = %q, want %q", got, want)
	}
}

func TestExtractSummaryDoctorFull(t *testing.T) {
	result, _ := json.Marshal(map[string]any{
		"issues": []map[string]any{
			{"title": "Nginx not running", "severity": "error"},
		},
		"actions": []map[string]any{
			{"description": "Start nginx", "disruptive": false},
		},
	})
	e := historian.HistoryEntry{Command: "doctor", Result: result}
	got := extractSummary(e)
	want := "Nginx not running"
	if got != want {
		t.Errorf("extractSummary(doctor full) = %q, want %q", got, want)
	}
}

func TestExtractSummaryDoctorBrief(t *testing.T) {
	result, _ := json.Marshal(map[string]any{
		"summary": "System needs attention",
	})
	e := historian.HistoryEntry{Command: "doctor", Result: result}
	got := extractSummary(e)
	want := "System needs attention"
	if got != want {
		t.Errorf("extractSummary(doctor brief) = %q, want %q", got, want)
	}
}

func TestExtractSummaryFallback(t *testing.T) {
	result, _ := json.Marshal(map[string]any{
		"weird": "shape",
	})
	e := historian.HistoryEntry{Command: "unknown", Result: result}
	got := extractSummary(e)
	if got == "" {
		t.Error("extractSummary(unknown) returned empty string, expected truncated fallback")
	}
}

func TestExtractSummaryEmptyResult(t *testing.T) {
	e := historian.HistoryEntry{Command: "explain", Result: json.RawMessage("{}")}
	got := extractSummary(e)
	if got == "" {
		t.Error("extractSummary({}) returned empty string, expected truncated fallback")
	}
}
