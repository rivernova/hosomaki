// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"os/exec"
	"strings"
	"time"
)

// SystemSnapshot holds raw data collected from the system.
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
// Each field runs a shell command and stores its raw output.
// Errors are soft — a failed command stores an empty string,
// so a missing tool doesn't break the whole snapshot.
func Snapshot() (*SystemSnapshot, error) {
	s := &SystemSnapshot{
		CollectedAt:    time.Now(),
		Uptime:         run("uptime", "-p"),
		Memory:         run("free", "-h"),
		Disk:           run("df", "-h", "--output=source,size,used,avail,pcent,target", "-x", "tmpfs", "-x", "devtmpfs"),
		FailedServices: run("systemctl", "--failed", "--no-legend", "--no-pager"),
		RecentErrors:   runShell("journalctl -p err -n 20 --no-pager --no-hostname -o short-monotonic 2>/dev/null"),
		TopProcesses:   run("ps", "aux", "--sort=-%cpu", "--no-headers"),
	}
	return s, nil
}

// run executes a command and returns trimmed stdout. Errors are silenced.
func run(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// runShell runs a shell command string via sh -c.
func runShell(cmd string) string {
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
