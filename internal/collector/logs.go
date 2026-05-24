// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"fmt"
	"os"
	"strings"
)

// LogOptions controls how many lines a log collector fetches.
type LogOptions struct {
	Lines int
}

const (
	defaultServiceLines = 50
	defaultBootLines    = 80
	defaultDmesgLines   = 50
	defaultFileLines    = 100
)

// ServiceLogs returns recent error-level journal entries for a systemd service.
func ServiceLogs(service string, opts LogOptions) (string, error) {
	n := opts.Lines
	if n <= 0 {
		n = defaultServiceLines
	}
	out := runShell(fmt.Sprintf(
		"journalctl -u %s -p err -n %d --no-pager --no-hostname -o short-monotonic 2>/dev/null",
		shellQuote(service), n,
	))
	if out == "" {
		return "", fmt.Errorf("no error logs found for service %q — is the service name correct?", service)
	}
	return out, nil
}

// BootLogs returns error-level journal entries from the given boot index.
// Index 0 is the current boot; -1 is the previous one.
func BootLogs(bootIndex int, opts LogOptions) (string, error) {
	n := opts.Lines
	if n <= 0 {
		n = defaultBootLines
	}
	out := runShell(fmt.Sprintf(
		"journalctl -b %d -p err -n %d --no-pager --no-hostname -o short-monotonic 2>/dev/null",
		bootIndex, n,
	))
	if out == "" {
		return "", fmt.Errorf("no error logs found for boot %d", bootIndex)
	}
	return out, nil
}

// DmesgLogs returns recent kernel error and warning messages.
func DmesgLogs(opts LogOptions) (string, error) {
	n := opts.Lines
	if n <= 0 {
		n = defaultDmesgLines
	}
	out := runShell(fmt.Sprintf(
		"dmesg --level=err,warn --notime 2>/dev/null | tail -n %d",
		n,
	))
	if out == "" {
		return "", fmt.Errorf("no kernel errors or warnings found in dmesg")
	}
	return out, nil
}

// FileLogs reads the tail of a log file, returning only lines that look like
// errors. If no error-like lines are found it returns the raw tail instead.
func FileLogs(path string, opts LogOptions) (string, error) {
	n := opts.Lines
	if n <= 0 {
		n = defaultFileLines
	}

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("cannot read log file %q: %w", path, err)
	}

	raw := runShell(fmt.Sprintf("tail -n %d %s 2>/dev/null", n, shellQuote(path)))
	if raw == "" {
		return "", fmt.Errorf("log file %q is empty or unreadable", path)
	}

	if filtered := filterErrorLines(raw); filtered != "" {
		return filtered, nil
	}
	return raw, nil
}

// filterErrorLines returns only lines that contain common error keywords.
func filterErrorLines(text string) string {
	keywords := []string{"error", "err", "warn", "fatal", "crit", "fail", "panic", "exception"}
	var kept []string
	for _, line := range strings.Split(text, "\n") {
		lower := strings.ToLower(line)
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				kept = append(kept, line)
				break
			}
		}
	}
	return strings.Join(kept, "\n")
}

// shellQuote wraps a string in single quotes, escaping any existing single
// quotes. Use this before interpolating user input into shell commands.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
