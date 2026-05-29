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
	defaultDmesgLines   = 60
	defaultFileLines    = 100
)

type execResult struct {
	stdout string
	stderr string
	err    error
}

func runCmd(name string, args ...string) execResult {
	cmd := exec.Command(name, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return execResult{
		stdout: strings.TrimSpace(outBuf.String()),
		stderr: strings.TrimSpace(errBuf.String()),
		err:    err,
	}
}

func isPermissionError(stderr string) bool {
	lower := strings.ToLower(stderr)
	markers := []string{
		"permission denied",
		"operation not permitted",
		"access denied",
		"not authorized",
		"failed to open",
		"failed to access",
		"unauthorized",
		"polkit",
	}
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

func permissionErr(source string) error {
	return fmt.Errorf(
		"permission denied reading %s\n"+
			"journalctl restricts access to logs for other users or boot sessions.\n"+
			"Run with elevated privileges: sudo hosomaki explain %s",
		source, source,
	)
}

func isJournalContent(out string) bool {
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return false
	}
	sentinels := []string{
		"-- No entries --",
		"-- no entries --",
		"No journal files were found.",
	}
	for _, s := range sentinels {
		if strings.Contains(trimmed, s) && !looksLikeLogLine(trimmed) {
			return false
		}
	}
	return true
}

func looksLikeLogLine(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "[") {
			return true
		}
		if strings.Contains(line, "]:") {
			return true
		}
	}
	return false
}

func ServiceLogs(service string, opts LogOptions) (string, error) {
	n := lines(opts.Lines, defaultServiceLines)

	args := []string{"-u", service, "-n", strconv.Itoa(n)}
	args = append(args, journalctl.errorLevel...)
	args = append(args, journalctl.format...)
	res := runCmd(binJournalctl, args...)
	if isPermissionError(res.stderr) {
		return "", permissionErr("--service " + service)
	}
	if res.err == nil && isJournalContent(res.stdout) {
		return res.stdout, nil
	}

	args = []string{"-u", service, "-n", strconv.Itoa(n)}
	args = append(args, journalctl.format...)
	res = runCmd(binJournalctl, args...)
	if isPermissionError(res.stderr) {
		return "", permissionErr("--service " + service)
	}
	if res.err != nil || !isJournalContent(res.stdout) {
		return "", fmt.Errorf("no logs found for service %q — is the service name correct and has it run recently?", service)
	}
	return res.stdout, nil
}

func BootLogs(bootIndex int, opts LogOptions) (string, error) {
	n := lines(opts.Lines, defaultBootLines)

	args := []string{"-b", strconv.Itoa(bootIndex), "-n", strconv.Itoa(n)}
	args = append(args, journalctl.errorLevel...)
	args = append(args, journalctl.format...)
	res := runCmd(binJournalctl, args...)
	if isPermissionError(res.stderr) {
		return "", permissionErr("--boot " + strconv.Itoa(bootIndex))
	}
	if res.err == nil && isJournalContent(res.stdout) {
		return res.stdout, nil
	}

	args = []string{"-b", strconv.Itoa(bootIndex), "-n", strconv.Itoa(n)}
	args = append(args, journalctl.format...)
	res = runCmd(binJournalctl, args...)
	if isPermissionError(res.stderr) {
		return "", permissionErr("--boot " + strconv.Itoa(bootIndex))
	}
	if res.err != nil || !isJournalContent(res.stdout) {
		return "", fmt.Errorf("no logs found for boot %d — the boot index may be out of range", bootIndex)
	}
	return res.stdout, nil
}

func DmesgLogs(opts LogOptions) (string, error) {
	n := lines(opts.Lines, defaultDmesgLines)

	raw, _ := runShell("dmesg 2>&1 | head -n 3")
	if isPermissionError(raw) {
		return "", fmt.Errorf(
			"permission denied reading dmesg (kernel.dmesg_restrict=1)\n" +
				"Run with elevated privileges: sudo hosomaki explain --dmesg",
		)
	}

	val, collErr := runShell(fmt.Sprintf(dmesgShell, n))
	if collErr == "" && val != "" {
		return val, nil
	}

	val, collErr = runShell(fmt.Sprintf(
		"dmesg 2>/dev/null | grep -iE '(error|err|warn|fail|panic|oops|segfault|oom|crit|bug)' | tail -n %d", n,
	))
	if collErr == "" && val != "" {
		return val, nil
	}

	val, collErr = runShell(fmt.Sprintf("dmesg 2>/dev/null | tail -n %d", n))
	if collErr != "" || val == "" {
		return "", fmt.Errorf("dmesg produced no output — the kernel ring buffer may be empty")
	}
	return val, nil
}

func FileLogs(path string, opts LogOptions) (string, error) {
	n := lines(opts.Lines, defaultFileLines)

	if _, err := os.Stat(path); err != nil {
		if os.IsPermission(err) {
			return "", fmt.Errorf(
				"permission denied reading %q\n"+
					"Run with elevated privileges: sudo hosomaki explain --file %s",
				path, path,
			)
		}
		return "", fmt.Errorf("cannot read log file %q: %w", path, err)
	}

	res := runCmd(binTail, "-n", strconv.Itoa(n), path)
	if isPermissionError(res.stderr) {
		return "", fmt.Errorf(
			"permission denied reading %q\n"+
				"Run with elevated privileges: sudo hosomaki explain --file %s",
			path, path,
		)
	}
	if res.err != nil || res.stdout == "" {
		return "", fmt.Errorf("log file %q is empty or unreadable", path)
	}

	if filtered := filterErrorLines(res.stdout); filtered != "" {
		return filtered, nil
	}
	return res.stdout, nil
}

func filterErrorLines(text string) string {
	keywords := []string{
		"error", "err", "warn", "fatal", "crit",
		"fail", "panic", "exception", "denied", "refused",
		"timeout", "traceback", "segfault", "oom",
	}
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
