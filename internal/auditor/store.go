// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package auditor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

//  manages the on-disk lifecycle

var ErrNoBaseline = errors.New("no audit baseline found — run `hosomaki audit --init` to create one")

var ErrBaselineVersion = errors.New("audit baseline was created by an incompatible version of hosomaki")

func DefaultPath() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("audit store: cannot determine home directory: %w", err)
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "hosomaki", "audit-baseline.json"), nil
}

func Load(path string) (*AuditBaseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoBaseline
		}
		return nil, fmt.Errorf("audit store: read %q: %w", path, err)
	}

	var b AuditBaseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("audit store: parse %q: %w", path, err)
	}

	if b.Version != baselineVersion {
		return nil, fmt.Errorf("%w (stored=%d, expected=%d)", ErrBaselineVersion, b.Version, baselineVersion)
	}

	return &b, nil
}

func Save(path string, b *AuditBaseline) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("audit store: create directory for %q: %w", path, err)
	}

	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("audit store: marshal baseline: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".audit-baseline-*.tmp")
	if err != nil {
		return fmt.Errorf("audit store: create temp file in %q: %w", dir, err)
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
		return fmt.Errorf("audit store: write temp file %q: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("audit store: close temp file %q: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("audit store: rename %q → %q: %w", tmpName, path, err)
	}

	success = true
	return nil
}
