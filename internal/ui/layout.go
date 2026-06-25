// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rivernova/hosomaki/internal/auditor"
	"github.com/rivernova/hosomaki/internal/collector"
)

// generates the various sections outputs

type SnapshotData struct {
	CollectedAt    time.Time
	Uptime         string
	Memory         string
	Disk           string
	FailedServices string
	RecentErrors   string
}

type ExplainContext struct {
	Source string
	Cmd    string
	Lines  int
	Since  string
	Until  string
}

func StatusHeader() string {
	return Title("status")
}

func StatusHeaderBrief() string {
	return Title("status (brief)")
}

func StatusSystemSection(d SnapshotData) string {
	return Section("system status", systemKV(d))
}

func StatusSystemSectionBrief(d SnapshotData) string {
	return SectionCompact("system", systemCompact(d))
}

func StatusInsightsSection(d SnapshotData) string {
	return Section("local insights", insightBullets(d))
}

func StatusInsightsSectionBrief(d SnapshotData) string {
	return SectionCompact("insights", insightBullets(d))
}

func DoctorHeader() string {
	return Title("doctor")
}

func DoctorHeaderBrief() string {
	return Title("doctor (brief)")
}

func DoctorSystemSection(d SnapshotData) string {
	return Section("system analysis", systemKV(d))
}

func DoctorSystemSectionBrief(d SnapshotData) string {
	return SectionCompact("system", systemCompact(d))
}

func DoctorInsightsSection(d SnapshotData) string {
	return Section("local insights", insightBullets(d))
}

func DoctorInsightsSectionBrief(d SnapshotData) string {
	return SectionCompact("insights", insightBullets(d))
}

func ExplainHeader() string {
	return Title("explain")
}

func ExplainContextSection(c ExplainContext) string {
	return Section("context", explainContextKV(c))
}

func systemKV(d SnapshotData) string {
	var b strings.Builder
	b.WriteString(KeyValue("uptime", formatUptime(d.Uptime)))
	for _, line := range formatMemory(d.Memory) {
		b.WriteString(line)
	}
	for _, line := range formatDisk(d.Disk) {
		b.WriteString(line)
	}
	return b.String()
}

func systemCompact(d SnapshotData) string {
	var parts []string
	if u := formatUptime(d.Uptime); u != "" {
		parts = append(parts, u)
	}
	if lines := formatMemory(d.Memory); len(lines) > 0 {
		parts = append(parts, strings.TrimRight(lines[0], "\n"))
	}
	if lines := formatDisk(d.Disk); len(lines) > 0 {
		parts = append(parts, strings.TrimRight(lines[0], "\n"))
	}
	if len(parts) == 0 {
		return "(no data)\n"
	}
	return strings.Join(parts, " · ") + "\n"
}

func formatUptime(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "(none)"
	}
	s = strings.TrimPrefix(s, "up ")
	s = regexp.MustCompile(`(\d+)\s+days?`).ReplaceAllString(s, "${1}d")
	s = regexp.MustCompile(`(\d+)\s+hours?`).ReplaceAllString(s, "${1}h")
	s = regexp.MustCompile(`(\d+)\s+minutes?`).ReplaceAllString(s, "${1}m")
	s = strings.ReplaceAll(s, ",", "")
	return strings.Join(strings.Fields(s), " ")
}

func formatMemory(raw string) []string {
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 7 && fields[0] == "Mem:" {
			used := cleanUnit(fields[2])
			total := cleanUnit(fields[1])
			avail := cleanUnit(fields[6])
			out = append(out, KeyValue("memory", fmt.Sprintf("%s / %s  (%s free)", used, total, avail)))
		}
		if len(fields) >= 3 && fields[0] == "Swap:" {
			if fields[2] == "0B" || fields[2] == "0" {
				out = append(out, KeyValue("swap", "inactive"))
			} else {
				out = append(out, KeyValue("swap", fmt.Sprintf("%s / %s", cleanUnit(fields[2]), cleanUnit(fields[1]))))
			}
		}
	}
	if len(out) == 0 {
		out = append(out, KeyValue("memory", "(none)"))
	}
	return out
}

func formatDisk(raw string) []string {
	var out []string
	seen := map[string]bool{}
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 || !strings.HasPrefix(fields[0], "/dev/") {
			continue
		}
		dev := fields[0]
		if seen[dev] {
			continue
		}
		seen[dev] = true
		used := cleanUnit(fields[2])
		size := cleanUnit(fields[1])
		pct := fields[4]
		mount := fields[5]
		out = append(out, KeyValue("disk "+mount, fmt.Sprintf("%s / %s  (%s)", used, size, pct)))
	}
	if len(out) == 0 {
		out = append(out, KeyValue("disk", "(none)"))
	}
	return out
}

func cleanUnit(s string) string {
	s = regexp.MustCompile(`(\d)Gi\b`).ReplaceAllString(s, "${1}G")
	s = regexp.MustCompile(`(\d)Mi\b`).ReplaceAllString(s, "${1}M")
	s = regexp.MustCompile(`(\d)Ki\b`).ReplaceAllString(s, "${1}K")
	s = regexp.MustCompile(`(\d+)\.0([GMK])`).ReplaceAllString(s, "${1}${2}")
	return s
}

func insightBullets(d SnapshotData) string {
	var b strings.Builder
	if strings.TrimSpace(d.FailedServices) == "" {
		b.WriteString(BulletOK("no failed services"))
	} else {
		for _, line := range nonEmptyLines(d.FailedServices) {
			b.WriteString(BulletFail(line))
		}
	}
	if strings.TrimSpace(d.RecentErrors) == "" {
		b.WriteString(BulletOK("no recent errors in journal"))
	} else {
		b.WriteString(BulletWarn("recent errors detected in journal"))
	}
	return b.String()
}

func explainContextKV(c ExplainContext) string {
	var b strings.Builder
	b.WriteString(KeyValue("source", orNone(c.Source)))
	if c.Cmd != "" {
		b.WriteString(KeyValue("command", c.Cmd))
	}
	if c.Lines > 0 {
		b.WriteString(KeyValue("lines", fmt.Sprintf("%d", c.Lines)))
	}
	if c.Since != "" {
		b.WriteString(KeyValue("since", c.Since))
	}
	if c.Until != "" {
		b.WriteString(KeyValue("until", c.Until))
	}
	return b.String()
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(none)"
	}
	return strings.TrimSpace(s)
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return out
}

func ExplainEmptyResult() string {
	return Section(
		"result",
		BulletOK("no errors or warnings found in the provided log")+
			BulletOK("the log appears to be informational or debug output only"),
	)
}

func AuditHeader() string { return Title("audit") }

func AuditInitHeader() string { return Title("audit (init)") }

func AuditBaselineSection(b *auditor.AuditBaseline, path string) string {
	var body strings.Builder
	body.WriteString(KeyValue("saved to", path))
	body.WriteString(KeyValue("captured at", b.CreatedAt.Format("2006-01-02 15:04:05")))
	body.WriteString(KeyValue("services", fmt.Sprintf("%d", len(b.Services))))
	body.WriteString(KeyValue("watched files", fmt.Sprintf("%d", len(b.Files))))
	body.WriteString(KeyValue("packages", fmt.Sprintf("%d", len(b.Packages))))
	body.WriteString(KeyValue("users", fmt.Sprintf("%d", len(b.Users))))
	body.WriteString(KeyValue("ports", fmt.Sprintf("%d", len(b.Ports))))
	for _, e := range b.CollectionErrors {
		body.WriteString(BulletWarn(e))
	}
	return Section("baseline", body.String())
}

func AuditDiffSection(d *auditor.AuditDiff, age string) string {
	var body strings.Builder
	body.WriteString(KeyValue("baseline age", age))
	body.WriteString(KeyValue("total changes", fmt.Sprintf("%d", d.TotalChanges())))
	return Section("changes detected", body.String())
}

func AuditLocalChanges(d *auditor.AuditDiff) string {
	var b strings.Builder

	renderList := func(header string, items []string, bullet func(string) string) {
		if len(items) == 0 {
			return
		}
		b.WriteString(sectionHeader(header))
		for _, item := range items {
			b.WriteString(bullet(item))
		}
	}

	renderList("services added", d.ServicesAdded, BulletWarn)
	renderList("services removed", d.ServicesRemoved, BulletWarn)

	renderList("files added", d.FilesAdded, BulletWarn)
	renderList("files removed", d.FilesRemoved, BulletWarn)

	if len(d.FilesModified) > 0 {
		b.WriteString(sectionHeader("files modified"))
		for _, fc := range d.FilesModified {
			b.WriteString(BulletWarn(fc.Path))
			b.WriteString(KeyValue("  size", fmt.Sprintf("%d → %d bytes", fc.OldSize, fc.NewSize)))
		}
	}

	if len(d.PermissionsChanged) > 0 {
		b.WriteString(sectionHeader("permission changes"))
		for _, pc := range d.PermissionsChanged {
			b.WriteString(BulletWarn(pc.Path))
			if pc.OldMode != pc.NewMode {
				b.WriteString(KeyValue("  mode", fmt.Sprintf("%s → %s", pc.OldMode, pc.NewMode)))
			}
			if pc.OldOwner != pc.NewOwner {
				b.WriteString(KeyValue("  owner", fmt.Sprintf("%s → %s", pc.OldOwner, pc.NewOwner)))
			}
			if pc.OldGroup != pc.NewGroup {
				b.WriteString(KeyValue("  group", fmt.Sprintf("%s → %s", pc.OldGroup, pc.NewGroup)))
			}
		}
	}

	renderList("packages installed", d.PackagesAdded, BulletOK)
	renderList("packages removed", d.PackagesRemoved, BulletWarn)

	if len(d.PackagesUpdated) > 0 {
		b.WriteString(sectionHeader("packages updated"))
		for _, pu := range d.PackagesUpdated {
			b.WriteString(BulletOK(fmt.Sprintf("%s  (%s → %s)", pu.Name, pu.OldVersion, pu.NewVersion)))
		}
	}

	renderList("ports opened", d.PortsOpened, BulletWarn)
	renderList("ports closed", d.PortsClosed, BulletWarn)

	renderList("users added", d.UsersAdded, BulletFail)
	renderList("users removed", d.UsersRemoved, BulletWarn)

	return b.String()
}

func AuditNoChanges(age string) string {
	return Section(
		"result",
		BulletOK(fmt.Sprintf("no changes detected since baseline was taken (%s ago)", age)),
	)
}

func WatchHeader(service string) string {
	return Title(fmt.Sprintf("watch — %s", service))
}

func WatchReadyLine(service string, seedLines int) string {
	msg := fmt.Sprintf("tailing %s", service)
	if seedLines > 0 {
		msg += fmt.Sprintf("  (seeded with last %d lines)", seedLines)
	}
	return BulletOK(msg) + "\n"
}

func WatchBatchHeader(t time.Time) string {
	return sectionHeader(fmt.Sprintf("analysis — %s", t.Format("15:04:05")))
}

func WatchShutdownLine() string {
	return "\n" + BulletOK("watch stopped") + "\n"
}

func WhyHeader() string { return Title("why") }

type WhyContext struct {
	Service  string
	ExitCode int
	Lines    int
	Since    string
}

func WhyContextSection(c WhyContext) string {
	var b strings.Builder
	b.WriteString(KeyValue("service", c.Service))
	b.WriteString(KeyValue("exit code", fmt.Sprintf("%d", c.ExitCode)))
	if c.Lines > 0 {
		b.WriteString(KeyValue("lines", fmt.Sprintf("%d", c.Lines)))
	}
	if c.Since != "" {
		b.WriteString(KeyValue("since", c.Since))
	}
	return Section("context", b.String())
}

func PortsHeader() string { return Title("ports") }

func PortsCollectedSection(count int, warnings []string) string {
	var b strings.Builder
	b.WriteString(KeyValue("listening ports", fmt.Sprintf("%d", count)))
	for _, w := range warnings {
		b.WriteString(BulletWarn(w))
	}
	return Section("collected", b.String())
}

func PortsCleanResult() string {
	return Section(
		"result",
		BulletOK("no unexpected ports detected"),
	)
}

func TimersHeader() string { return Title("timers") }

func TimersCollectedSection(count int, warning string) string {
	var b strings.Builder
	b.WriteString(KeyValue("systemd timers", fmt.Sprintf("%d", count)))
	if warning != "" {
		b.WriteString(BulletWarn(warning))
	}
	return Section("collected", b.String())
}

func TimersCleanResult() string {
	return Section("result", BulletOK("all timers are healthy"))
}

func CronsHeader() string { return Title("crons") }

func CronsCollectedSection(count int, warnings []string) string {
	var b strings.Builder
	b.WriteString(KeyValue("cron jobs", fmt.Sprintf("%d", count)))
	for _, w := range warnings {
		b.WriteString(BulletWarn(w))
	}
	return Section("collected", b.String())
}

func CronsCleanResult() string {
	return Section("result", BulletOK("no issues found in cron jobs"))
}

type ExplainPIDContext struct {
	PID  int
	Name string // process name from /proc/<pid>/status. may be empty
}

func ExplainPIDContextSection(c ExplainPIDContext) string {
	var b strings.Builder
	b.WriteString(KeyValue("source", fmt.Sprintf("pid: %d", c.PID)))
	if c.Name != "" {
		b.WriteString(KeyValue("process", c.Name))
	}
	return Section("context", b.String())
}

func MountsHeader() string { return Title("mounts") }

func MountsCollectedSection(total, real, nfs, staleNFS int, warnings []string) string {
	var b strings.Builder
	b.WriteString(KeyValue("active mounts", fmt.Sprintf("%d", total)))
	b.WriteString(KeyValue("real filesystems", fmt.Sprintf("%d", real)))
	if nfs > 0 {
		if staleNFS > 0 {
			b.WriteString(KeyValue("nfs mounts", fmt.Sprintf("%d  (%d stale)", nfs, staleNFS)))
		} else {
			b.WriteString(KeyValue("nfs mounts", fmt.Sprintf("%d", nfs)))
		}
	}
	for _, w := range warnings {
		b.WriteString(BulletWarn(w))
	}
	return Section("collected", b.String())
}

func MountsCleanResult() string {
	return Section("result", BulletOK("all mount points are healthy"))
}

func UpdatesHeader() string {
	return Title("pending updates")
}

func UpdatesNoPending() string {
	return Section("results", "No pending package updates found.\n")
}

func UpdatesNoPendingMsg(msg string) string {
	return Section("results", msg+"\n")
}

func UpdatesPendingList(updates []collector.Update, securityOnly bool) string {
	var b strings.Builder

	label := "collected"
	if securityOnly {
		label = "collected (security-only)"
	}

	b.WriteString(Section(label, fmt.Sprintf("%d package(s)\n", len(updates))))

	for _, u := range updates {
		flags := ""
		if u.Security {
			flags += " [" + styleWarn + "SECURITY" + styleReset + "]"
		}
		if u.RebootRequired {
			flags += " [" + styleWarn + "REBOOT" + styleReset + "]"
		}

		inst := u.Installed
		if inst == "" {
			inst = "?"
		}
		avail := u.Available
		if avail == "" {
			avail = "?"
		}

		_, _ = fmt.Fprintf(&b, "  %s%s  %s → %s\n",
			u.Package, flags, inst, avail)
	}

	return b.String()
}

func UpdatesCleanResult() string {
	return Section("result", BulletOK("no issues found in pending updates"))
}

// history UI

func HistoryHeader() string {
	return Title("diagnostic history")
}

func HistoryNoHistory() string {
	return Section("results", "No diagnostic history found. Run explain, why, audit, status, or doctor to populate it.\n")
}

func HistoryNoMatching(msg string) string {
	return Section("results", msg+"\n")
}

func HistoryEntryCount(n int) string {
	return Section("entries", plural(n, "entry", "entries")+"\n")
}

func HistoryCleared() string {
	return Section("results", "History log cleared.\n")
}

func HistoryCleanResult() string {
	return Section("result", BulletOK("no issues found in diagnostic history"))
}

func FirewallHeader() string { return Title("firewall") }

func FirewallCollectedSection(backend string, ruleCount int, zones []string, warning, readStatus string) string {
	var b strings.Builder
	b.WriteString(KeyValue("backend", backend))
	b.WriteString(KeyValue("read_status", readStatus))
	b.WriteString(KeyValue("rules", fmt.Sprintf("%d", ruleCount)))
	if len(zones) > 0 {
		b.WriteString(KeyValue("zones", strings.Join(zones, ", ")))
	}
	if warning != "" {
		b.WriteString(BulletWarn(warning))
	}
	return Section("collected", b.String())
}

func FirewallCleanResult() string {
	return Section("result", BulletOK("all firewall rules are reasonable"))
}

func FirewallNoRules() string {
	return Section("result", "No firewall rules found.\n")
}

func FirewallReadFailed(warning string) string {
	msg := "Firewall rules could not be read completely."
	if warning != "" {
		msg += " " + warning
	}
	return Section("result", msg+"\n")
}

func FirewallNoBackend() string {
	return Section("result", "No firewall backend detected (tried firewalld, ufw, nftables, iptables).\n")
}
