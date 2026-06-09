// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package auditor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// unit tests for baseline persistence

func TestSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	orig := &AuditBaseline{
		Version:   baselineVersion,
		CreatedAt: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Services:  []string{"nginx.service", "ssh.service"},
		Packages:  []string{"curl 7.68.0"},
		Users:     []string{"root"},
	}

	if err := Save(path, orig); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got.Version != orig.Version {
		t.Errorf("Version = %d, want %d", got.Version, orig.Version)
	}
	if !got.CreatedAt.Equal(orig.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, orig.CreatedAt)
	}
	if len(got.Services) != len(orig.Services) {
		t.Errorf("Services len = %d, want %d", len(got.Services), len(orig.Services))
	}
}

func TestSave_CreatesParentDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "baseline.json")

	b := &AuditBaseline{Version: baselineVersion, CreatedAt: time.Now()}
	if err := Save(path, b); err != nil {
		t.Fatalf("Save() should create parent dirs, got error: %v", err)
	}

	if _, statErr := os.Stat(path); statErr != nil {
		t.Errorf("file not found after Save: %v", statErr)
	}
}

func TestSave_Atomic_ExistingFileNotCorruptedOnError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	first := &AuditBaseline{Version: baselineVersion, CreatedAt: time.Now(), Users: []string{"alice"}}
	if err := Save(path, first); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	second := &AuditBaseline{Version: baselineVersion, CreatedAt: time.Now(), Users: []string{"bob"}}
	if err := Save(path, second); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("expected 1 file in dir after atomic save, got %v", names)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load after second Save: %v", err)
	}
	if len(got.Users) != 1 || got.Users[0] != "bob" {
		t.Errorf("Users = %v, want [bob]", got.Users)
	}
}

func TestLoad_ErrNoBaseline_WhenFileMissing(t *testing.T) {
	_, err := Load("/nonexistent/path/baseline.json")
	if !errors.Is(err, ErrNoBaseline) {
		t.Errorf("Load(missing) = %v, want ErrNoBaseline", err)
	}
}

func TestLoad_ErrBaselineVersion_WhenVersionMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	b := &AuditBaseline{Version: baselineVersion + 99, CreatedAt: time.Now()}
	if err := Save(path, b); err != nil {
		t.Fatalf("Save: %v", err)
	}

	_, err := Load(path)
	if !errors.Is(err, ErrBaselineVersion) {
		t.Errorf("Load(wrong version) = %v, want ErrBaselineVersion", err)
	}
}

func TestLoad_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	if err := os.WriteFile(path, []byte("not json {{{"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load(corrupt) should return error, got nil")
	}
	if errors.Is(err, ErrNoBaseline) || errors.Is(err, ErrBaselineVersion) {
		t.Errorf("Load(corrupt) returned wrong sentinel: %v", err)
	}
}

func TestDefaultPath_ReturnsNonEmpty(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error: %v", err)
	}
	if path == "" {
		t.Error("DefaultPath() returned empty string")
	}
}

func TestDefaultPath_RespectsXDGDataHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)

	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error: %v", err)
	}

	expected := filepath.Join(dir, "hosomaki", "audit-baseline.json")
	if path != expected {
		t.Errorf("DefaultPath() = %q, want %q", path, expected)
	}
}

func TestDefaultPath_FallsBackToHomeDir(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")

	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".local", "share", "hosomaki", "audit-baseline.json")
	if path != expected {
		t.Errorf("DefaultPath() = %q, want %q", path, expected)
	}
}
