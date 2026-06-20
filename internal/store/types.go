// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// data types for the history store log

const logVersion = 1
const maxEntries = 1000

var (
	ErrNoHistory  = errors.New("no history found — run a diagnostic command first")
	ErrBadVersion = errors.New("history log version mismatch")
)

type HistoryEntry struct {
	Timestamp time.Time       `json:"timestamp"`
	Command   string          `json:"command"` // "explain", "why", "audit", "status", "doctor"
	Result    json.RawMessage `json:"result"`
}

type HistoryLog struct {
	Version int            `json:"version"`
	Entries []HistoryEntry `json:"entries"`
}

func DefaultPath() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("store: cannot determine home directory: %w", err)
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "hosomaki", "history.json"), nil
}

func Load(path string) (*HistoryLog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoHistory
		}
		return nil, fmt.Errorf("store: read %q: %w", path, err)
	}

	var log HistoryLog
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, fmt.Errorf("store: parse %q: %w", path, err)
	}

	if log.Version != logVersion {
		return nil, fmt.Errorf("%w (stored=%d, expected=%d)", ErrBadVersion, log.Version, logVersion)
	}

	return &log, nil
}

func Save(path string, log *HistoryLog) error {
	log.Version = logVersion

	if len(log.Entries) > maxEntries {
		log.Entries = log.Entries[len(log.Entries)-maxEntries:]
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("store: create directory for %q: %w", path, err)
	}

	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return fmt.Errorf("store: marshal log: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".history-*.tmp")
	if err != nil {
		return fmt.Errorf("store: create temp file in %q: %w", dir, err)
	}
	tmpName := tmp.Name()

	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("store: write temp file %q: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("store: close temp file %q: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("store: rename %q → %q: %w", tmpName, path, err)
	}

	success = true
	return nil
}

func Record(cmd string, result any) error {
	path, err := DefaultPath()
	if err != nil {
		return err
	}

	log, err := Load(path)
	if err != nil && !errors.Is(err, ErrNoHistory) {
		return err
	}
	if log == nil {
		log = &HistoryLog{}
	}

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("store: marshal result for %q: %w", cmd, err)
	}

	log.Entries = append(log.Entries, HistoryEntry{
		Timestamp: time.Now(),
		Command:   cmd,
		Result:    data,
	})

	return Save(path, log)
}
