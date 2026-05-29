// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

// all commands are defined here and referenced by name
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

// variable parts (line counts, service names, boot indices) are handled
// at call sites in logs.go using these as the stable base
var snapshot = struct {
	uptimeArgs         []string
	memoryArgs         []string
	diskShell          string // needs shell: multiple --output flags and -x exclusions
	failedServicesArgs []string
	recentErrorsShell  string // needs shell: 2>/dev/null redirection
	topProcessesArgs   []string
}{
	uptimeArgs:         []string{"-p"},
	memoryArgs:         []string{"-h"},
	diskShell:          "df -h --output=source,size,used,avail,pcent,target -x tmpfs -x devtmpfs",
	failedServicesArgs: []string{"--failed", "--no-legend", "--no-pager"},
	recentErrorsShell:  "journalctl -p err -n 20 --no-pager --no-hostname -o short-monotonic 2>/dev/null",
	topProcessesArgs:   []string{"aux", "--sort=-%cpu", "--no-headers"},
}

// variable flags (-u, -b, -n) are appended at call sites
var journalctl = struct {
	errorLevel []string
	format     []string
}{
	errorLevel: []string{"-p", "err"},
	format:     []string{"--no-pager", "--no-hostname", "-o", "short-monotonic"},
}

const dmesgShell = "dmesg --level=err,warn --notime 2>/dev/null | tail -n %d"
