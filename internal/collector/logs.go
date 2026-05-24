// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// this file contains all log collection logic

type LogOptions struct {
	Lines int
}

const (
	defaultServiceLines = 50
	defaultBootLines    = 80
	defaultDmesgLines   = 50
	defaultFileLines    = 100
)

func ServiceLogs(service string, opts LogOptions) (string, error) {
	n := lines(opts.Lines, defaultServiceLines)
	args := append(
		[]string{"-u", service, "-n", strconv.Itoa(n)},
		append(journalctl.errorLevel, journalctl.format...)...,
	)
	out, err := exec.Command(binJournalctl, args...).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return "", fmt.Errorf("no error logs found for service %q — is the service name correct?", service)
	}
	return strings.TrimSpace(string(out)), nil
}

func BootLogs(bootIndex int, opts LogOptions) (string, error) {
	n := lines(opts.Lines, defaultBootLines)
	args := append(
		[]string{"-b", strconv.Itoa(bootIndex), "-n", strconv.Itoa(n)},
		append(journalctl.errorLevel, journalctl.format...)...,
	)
	out, err := exec.Command(binJournalctl, args...).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return "", fmt.Errorf("no error logs found for boot %d", bootIndex)
	}
	return strings.TrimSpace(string(out)), nil
}

func DmesgLogs(opts LogOptions) (string, error) {
	n := lines(opts.Lines, defaultDmesgLines)
	val, collErr := runShell(fmt.Sprintf(dmesgShell, n))
	if collErr != "" || val == "" {
		return "", fmt.Errorf("no kernel errors or warnings found in dmesg")
	}
	return val, nil
}

func FileLogs(path string, opts LogOptions) (string, error) {
	n := lines(opts.Lines, defaultFileLines)

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("cannot read log file %q: %w", path, err)
	}

	out, err := exec.Command(binTail, "-n", strconv.Itoa(n), path).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return "", fmt.Errorf("log file %q is empty or unreadable", path)
	}

	raw := strings.TrimSpace(string(out))
	if filtered := filterErrorLines(raw); filtered != "" {
		return filtered, nil
	}
	return raw, nil
}

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

func lines(n, defaultN int) int {
	if n > 0 {
		return n
	}
	return defaultN
}
