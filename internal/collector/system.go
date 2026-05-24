// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package collector gathers raw system data via OS commands.
// It has no knowledge of AI, prompts, or presentation — it only
// runs commands and returns their output as strings.
package collector

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// SystemSnapshot holds the raw output of every system command
// collected at a single point in time.
type SystemSnapshot struct {
	CollectedAt    time.Time
	Uptime         string
	Memory         string
	Disk           string
	FailedServices string
	RecentErrors   string
	TopProcesses   string
}

// Snapshot collects a point-in-time view of the system.
// Commands are run sequentially; a failing command records its error
// message as the field value so the caller always gets a complete struct.
func Snapshot() (*SystemSnapshot, error) {
	s := &SystemSnapshot{
		CollectedAt:    time.Now(),
		Uptime:         run("uptime", "-p"),
		Memory:         run("free", "-h"),
		Disk:           runShell("df -h --output=source,size,used,avail,pcent,target -x tmpfs -x devtmpfs"),
		FailedServices: run("systemctl", "--failed", "--no-legend", "--no-pager"),
		RecentErrors:   runShell("journalctl -p err -n 20 --no-pager --no-hostname -o short-monotonic 2>/dev/null"),
		TopProcesses:   run("ps", "aux", "--sort=-%cpu", "--no-headers"),
	}
	return s, nil
}

// run executes a command and returns its trimmed stdout.
// On error it returns a human-readable description of the failure.
func run(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return fmt.Sprintf("(command failed: %s)", err)
	}
	return strings.TrimSpace(string(out))
}

// runShell executes a shell command string via sh -c.
func runShell(cmd string) string {
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return fmt.Sprintf("(command failed: %s)", err)
	}
	return strings.TrimSpace(string(out))
}
