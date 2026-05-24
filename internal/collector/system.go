// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type SystemSnapshot struct {
	CollectedAt    time.Time
	Uptime         string
	Memory         string
	Disk           string
	FailedServices string
	RecentErrors   string
	TopProcesses   string
}

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

func run(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return fmt.Sprintf("(command failed: %s)", err)
	}
	return strings.TrimSpace(string(out))
}

func runShell(cmd string) string {
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return fmt.Sprintf("(command failed: %s)", err)
	}
	return strings.TrimSpace(string(out))
}
