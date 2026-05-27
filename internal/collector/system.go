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

// this file contains logic for collecting a general snapshot of the system state

type SystemSnapshot struct {
	CollectedAt    time.Time
	Environment    Environment
	Uptime         string
	Memory         string
	Disk           string
	FailedServices string
	RecentErrors   string
	TopProcesses   string
	Errors         []string
}

func Snapshot() (*SystemSnapshot, error) {
	s := &SystemSnapshot{
		CollectedAt: time.Now(),
		Environment: Env(),
	}

	set := func(dest *string, val, collectionErr string) {
		*dest = val
		if collectionErr != "" {
			s.Errors = append(s.Errors, collectionErr)
		}
	}

	val, err := run(binUptime, snapshot.uptimeArgs...)
	set(&s.Uptime, val, err)

	val, err = run(binFree, snapshot.memoryArgs...)
	set(&s.Memory, val, err)

	val, err = runShell(snapshot.diskShell)
	set(&s.Disk, val, err)

	val, err = run(binSystemctl, snapshot.failedServicesArgs...)
	set(&s.FailedServices, val, err)

	val, err = runShell(snapshot.recentErrorsShell)
	set(&s.RecentErrors, val, err)

	val, err = run(binPs, snapshot.topProcessesArgs...)
	set(&s.TopProcesses, val, err)

	return s, nil
}

func run(name string, args ...string) (string, string) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", fmt.Sprintf("%s: %s", name, err)
	}
	return strings.TrimSpace(string(out)), ""
}

func runShell(cmd string) (string, string) {
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "", fmt.Sprintf("sh -c %q: %s", cmd, err)
	}
	return strings.TrimSpace(string(out)), ""
}
