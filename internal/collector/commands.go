// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

// this file contains all the commands used to collect the system snapshot data

const (
	binUptime     = "uptime"
	binFree       = "free"
	binSystemctl  = "systemctl"
	binPs         = "ps"
	binJournalctl = "journalctl"
	binDmesg      = "dmesg"
	binDf         = "df"
	binTail       = "tail"
)

var snapshot = struct {
	uptimeArgs         []string
	memoryArgs         []string
	diskShell          string
	failedServicesArgs []string
	recentErrorsShell  string
	topProcessesArgs   []string
}{
	uptimeArgs:         []string{"-p"},
	memoryArgs:         []string{"-h"},
	diskShell:          "df -h --output=source,size,used,avail,pcent,target -x tmpfs -x devtmpfs",
	failedServicesArgs: []string{"--failed", "--no-legend", "--no-pager"},
	recentErrorsShell:  "journalctl -p err -n 50 --no-pager --no-hostname -o short-monotonic 2>/dev/null",
	topProcessesArgs:   []string{"aux", "--sort=-%cpu", "--no-headers"},
}

var journalctl = struct {
	errorLevel []string
	format     []string
}{
	errorLevel: []string{"-p", "err"},
	format:     []string{"--no-pager", "--no-hostname", "-o", "short-monotonic"},
}

const dmesgShell = "dmesg --level=err,warn --notime 2>/dev/null | tail -n %d"
