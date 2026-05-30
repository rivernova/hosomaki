// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package analysis

import (
	"fmt"
	"strconv"
	"strings"
)

// this file contains the core analysis logic that turns raw input strings into structured reports of metrics and findings

type Level int

const (
	Neutral Level = iota
	OK
	Info
	Warn
	Crit
)

type Input struct {
	Kernel         string
	Uptime         string
	Memory         string
	Disk           string
	FailedServices string
	RecentErrors   string
}

type Metric struct {
	Label string
	Value string
	Level Level
}

type Finding struct {
	Level Level
	Text  string
}

type Report struct {
	Metrics  []Metric
	Findings []Finding

	FailedCount int
	Anomalies   int
}

const (
	memWarnPct  = 75
	memCritPct  = 90
	diskWarnPct = 85
	diskCritPct = 95
	svcCritN    = 3
)

func Analyze(in Input) Report {
	var r Report

	r.analyzeKernel(in.Kernel)
	r.analyzeUptime(in.Uptime)
	r.analyzeMemory(in.Memory)
	r.analyzeDisk(in.Disk)
	r.analyzeServices(in.FailedServices)
	r.analyzeErrors(in.RecentErrors)

	for _, f := range r.Findings {
		if f.Level == Warn || f.Level == Crit {
			r.Anomalies++
		}
	}
	return r
}

func (r *Report) add(m Metric, findings ...Finding) {
	r.Metrics = append(r.Metrics, m)
	r.Findings = append(r.Findings, findings...)
}

func (r *Report) analyzeKernel(kernel string) {
	kernel = strings.TrimSpace(kernel)
	if kernel == "" {
		return
	}
	r.Metrics = append(r.Metrics, Metric{"kernel", kernel, Neutral})
}

func (r *Report) analyzeUptime(uptime string) {
	uptime = strings.TrimSpace(uptime)
	if uptime == "" {
		return
	}
	r.Metrics = append(r.Metrics, Metric{"uptime", uptime, OK})
}

func (r *Report) analyzeMemory(free string) {
	used, total, pct, ok := parseMemory(free)
	if !ok {
		return
	}
	lvl := OK
	finding := Finding{OK, "memory within range"}
	switch {
	case pct >= memCritPct:
		lvl = Crit
		finding = Finding{Crit, fmt.Sprintf("memory critically high — %d%% in use", pct)}
	case pct >= memWarnPct:
		lvl = Warn
		finding = Finding{Warn, fmt.Sprintf("memory pressure — %d%% in use", pct)}
	}
	r.add(Metric{"memory", fmt.Sprintf("%s / %s (%d%%)", used, total, pct), lvl}, finding)
}

func (r *Report) analyzeDisk(df string) {
	worst, mount, ok := parseDiskMax(df)
	if !ok {
		return
	}
	lvl := OK
	finding := Finding{OK, "disk usage within range"}
	switch {
	case worst >= diskCritPct:
		lvl = Crit
		finding = Finding{Crit, fmt.Sprintf("disk almost full — %s at %d%%", mount, worst)}
	case worst >= diskWarnPct:
		lvl = Warn
		finding = Finding{Warn, fmt.Sprintf("disk filling up — %s at %d%%", mount, worst)}
	}
	r.add(Metric{"disk", fmt.Sprintf("%d%% %s", worst, mount), lvl}, finding)
}

func (r *Report) analyzeServices(text string) {
	names := failedServiceNames(text)
	r.FailedCount = len(names)

	switch {
	case len(names) == 0:
		r.add(
			Metric{"services", "all healthy", OK},
			Finding{OK, "no failed services"},
		)
	default:
		lvl := Warn
		if len(names) >= svcCritN {
			lvl = Crit
		}
		noun := "service"
		if len(names) > 1 {
			noun = "services"
		}
		r.add(
			Metric{"services", fmt.Sprintf("%d failed", len(names)), lvl},
			Finding{lvl, fmt.Sprintf("%d failed %s: %s", len(names), noun, strings.Join(names, ", "))},
		)
	}
}

func (r *Report) analyzeErrors(block string) {
	if strings.TrimSpace(block) == "" {
		r.add(
			Metric{"journal", "no recent errors", OK},
			Finding{OK, "no recent errors in the journal"},
		)
		return
	}
	n := countNonEmptyLines(block)
	r.add(
		Metric{"journal", fmt.Sprintf("%d recent errors", n), Info},
		Finding{Info, fmt.Sprintf("%d recent error lines in the journal", n)},
	)
}

func parseMemory(free string) (used, total string, pct int, ok bool) {
	for _, line := range strings.Split(free, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(fields[0]), "mem") {
			continue
		}
		total, used = fields[1], fields[2]
		tb, terr := parseHumanSize(total)
		ub, uerr := parseHumanSize(used)
		if terr != nil || uerr != nil || tb <= 0 {
			return "", "", 0, false
		}
		return used, total, int(ub/tb*100 + 0.5), true
	}
	return "", "", 0, false
}

func parseDiskMax(df string) (pct int, mount string, ok bool) {
	worst := -1
	for _, line := range strings.Split(df, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		for i, f := range fields {
			if !strings.HasSuffix(f, "%") {
				continue
			}
			v, err := strconv.Atoi(strings.TrimSuffix(f, "%"))
			if err != nil {
				continue
			}
			m := "/"
			if i+1 < len(fields) {
				m = strings.Join(fields[i+1:], " ")
			}
			if v > worst {
				worst, mount = v, m
			}
		}
	}
	if worst < 0 {
		return 0, "", false
	}
	return worst, mount, true
}

func failedServiceNames(text string) []string {
	var names []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		idx := 0
		if fields[0] == "●" || fields[0] == "*" {
			idx = 1
		}
		if idx < len(fields) {
			names = append(names, fields[idx])
		}
	}
	return names
}

func countNonEmptyLines(s string) int {
	n := 0
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

func parseHumanSize(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}

	mult := 1.0
	u := s
	if strings.HasSuffix(u, "i") || strings.HasSuffix(u, "I") {
		u = u[:len(u)-1]
	}
	if len(u) > 0 {
		switch u[len(u)-1] {
		case 'B', 'b':
			u = u[:len(u)-1]
		case 'K', 'k':
			mult, u = 1<<10, u[:len(u)-1]
		case 'M', 'm':
			mult, u = 1<<20, u[:len(u)-1]
		case 'G', 'g':
			mult, u = 1<<30, u[:len(u)-1]
		case 'T', 't':
			mult, u = 1<<40, u[:len(u)-1]
		case 'P', 'p':
			mult, u = 1<<50, u[:len(u)-1]
		}
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(u), 64)
	if err != nil {
		return 0, fmt.Errorf("parse size %q: %w", s, err)
	}
	return val * mult, nil
}
