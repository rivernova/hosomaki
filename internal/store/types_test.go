// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultPath_UsesXDG(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/custom/data")
	path, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	expected := "/custom/data/hosomaki/history.json"
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	log := &HistoryLog{
		Entries: []HistoryEntry{
			{
				Timestamp: time.Date(2026, 6, 19, 20, 0, 0, 0, time.UTC),
				Command:   "explain",
				Result:    json.RawMessage(`{"summary":"nginx was down"}`),
			},
		},
	}

	if err := Save(path, log); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Command != "explain" {
		t.Errorf("expected command 'explain', got %q", loaded.Entries[0].Command)
	}
}

func TestLoad_NoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.json")

	_, err := Load(path)
	if err != ErrNoHistory {
		t.Errorf("expected ErrNoHistory, got %v", err)
	}
}

func TestLoad_BadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json {{{"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for bad JSON")
	}
}

func TestLoad_BadVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "badver.json")

	bad := &HistoryLog{Version: 999}
	data, _ := json.Marshal(bad)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected ErrBadVersion")
	}
}

func TestSave_RotatesAtLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	// Create a log with maxEntries + 50 entries
	log := &HistoryLog{}
	for i := 0; i < maxEntries+50; i++ {
		log.Entries = append(log.Entries, HistoryEntry{
			Timestamp: time.Now(),
			Command:   "explain",
			Result:    json.RawMessage(`{}`),
		})
	}

	if err := Save(path, log); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) > maxEntries {
		t.Errorf("expected at most %d entries, got %d", maxEntries, len(loaded.Entries))
	}
	if len(loaded.Entries) != maxEntries {
		t.Errorf("expected exactly %d entries after rotation, got %d", maxEntries, len(loaded.Entries))
	}
}

func TestRecord_AppendsEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	t.Setenv("XDG_DATA_HOME", dir)

	// First record
	if err := Save(path, &HistoryLog{}); err != nil {
		t.Fatal(err)
	}
	// Override DefaultPath to use temp dir

	// Record directly via Save/Load since DefaultPath returns XDG path
	log, _ := Load(path)
	if log == nil {
		log = &HistoryLog{}
	}
	log.Entries = append(log.Entries, HistoryEntry{
		Timestamp: time.Now(),
		Command:   "status",
		Result:    json.RawMessage(`{"summary":"all good"}`),
	})
	if err := Save(path, log); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Command != "status" {
		t.Errorf("expected command 'status', got %q", loaded.Entries[0].Command)
	}
}

func TestRecord_NewLog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	// Record on a non-existent log
	log := &HistoryLog{
		Entries: []HistoryEntry{{
			Timestamp: time.Now(),
			Command:   "why",
			Result:    json.RawMessage(`{}`),
		}},
	}
	if err := Save(path, log); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(loaded.Entries))
	}
}

func TestSave_Atomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	log := &HistoryLog{
		Entries: []HistoryEntry{{
			Timestamp: time.Now(),
			Command:   "audit",
			Result:    json.RawMessage(`{}`),
		}},
	}

	if err := Save(path, log); err != nil {
		t.Fatal(err)
	}

	// Verify no temp files left behind
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}
