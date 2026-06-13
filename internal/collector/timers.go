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

// timer collection logic for the timers command

type TimerEntry struct {
	Unit        string
	Activates   string
	Next        string
	Last        string
	LastResult  string
	ActiveState string
}

func Timers() ([]TimerEntry, string) {
	units, warn := listTimerUnits()
	if warn != "" {
		return nil, warn
	}
	if len(units) == 0 {
		return nil, ""
	}

	entries := make([]TimerEntry, 0, len(units))
	for _, unit := range units {
		e := showTimer(unit)
		entries = append(entries, e)
	}
	return entries, ""
}

func listTimerUnits() ([]string, string) {
	out, errMsg := run(
		binSystemctl,
		"list-units", "--type=timer", "--all", "--no-pager", "--no-legend", "--plain",
	)
	if errMsg != "" {
		return nil, fmt.Sprintf("systemctl list-units: %s", errMsg)
	}
	var units []string
	for _, line := range nonEmptyLines(out) {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		name = strings.TrimLeft(name, "●• \t")
		if strings.HasSuffix(name, ".timer") {
			units = append(units, name)
		}
	}
	return units, ""
}

func showTimer(unit string) TimerEntry {
	props := []string{
		"Id",
		"ActiveState",
		"Triggers",
		"NextElapseUSecRealtime",
		"NextElapseUSecMonotonic",
		"LastTriggerUSec",
		"Result",
	}

	args := append(
		[]string{"show", "--no-pager", "--property=" + strings.Join(props, ",")},
		unit,
	)
	out, err := exec.Command(binSystemctl, args...).Output()
	if err != nil {
		return TimerEntry{Unit: unit, Next: "never", Last: "never"}
	}

	kv := parseKeyValues(string(out))

	id := kv["Id"]
	if id == "" {
		id = unit
	}

	activates := firstOf(kv["Triggers"])
	activeState := kv["ActiveState"]
	lastResult := kv["Result"]

	next := uSecToHuman(kv["NextElapseUSecRealtime"], kv["NextElapseUSecMonotonic"])
	last := uSecToHuman(kv["LastTriggerUSec"], "")

	if activates != "" && lastResult == "" {
		lastResult = serviceResult(activates)
	}

	return TimerEntry{
		Unit:        id,
		Activates:   activates,
		Next:        next,
		Last:        last,
		LastResult:  lastResult,
		ActiveState: activeState,
	}
}

func parseKeyValues(out string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])
		if k != "" {
			m[k] = v
		}
	}
	return m
}

func firstOf(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func uSecToHuman(realtime, monotonic string) string {
	if ts := parseUSec(realtime); !ts.IsZero() {
		return ts.Local().Format("2006-01-02 15:04:05 MST")
	}
	if ts := parseUSec(monotonic); !ts.IsZero() {
		return ts.Local().Format("2006-01-02 15:04:05 MST")
	}
	return "never"
}

func parseUSec(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return time.Time{}
	}
	var usec uint64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return time.Time{}
		}
		usec = usec*10 + uint64(ch-'0')
	}
	if usec == 0 {
		return time.Time{}
	}
	sec := int64(usec / 1_000_000)
	nsec := int64((usec % 1_000_000) * 1_000)
	return time.Unix(sec, nsec)
}

func serviceResult(svc string) string {
	if svc == "" {
		return ""
	}
	out, err := exec.Command(
		binSystemctl, "show", "--no-pager", "--property=Result", svc,
	).Output()
	if err != nil {
		return ""
	}
	kv := parseKeyValues(string(out))
	return kv["Result"]
}

func FormatTimersForPrompt(entries []TimerEntry) string {
	if len(entries) == 0 {
		return "(no systemd timers found)"
	}
	var b strings.Builder
	for _, e := range entries {
		activates := e.Activates
		if activates == "" {
			activates = "(unknown)"
		}
		last := e.Last
		if last == "" {
			last = "never"
		}
		next := e.Next
		if next == "" {
			next = "never"
		}
		_, _ = fmt.Fprintf(&b, "unit:         %s\n", e.Unit)
		_, _ = fmt.Fprintf(&b, "activates:    %s\n", activates)
		_, _ = fmt.Fprintf(&b, "active_state: %s\n", e.ActiveState)
		_, _ = fmt.Fprintf(&b, "last_run:     %s\n", last)
		_, _ = fmt.Fprintf(&b, "next_run:     %s\n", next)
		if e.LastResult != "" {
			_, _ = fmt.Fprintf(&b, "last_result:  %s\n", e.LastResult)
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
