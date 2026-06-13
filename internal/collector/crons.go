// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package collector

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// cron collection logic for the crons command

type CronJob struct {
	Source   string
	Schedule string
	User     string
	Command  string
}

func Crons() ([]CronJob, []string) {
	var jobs []CronJob
	var warnings []string

	appendJobs := func(got []CronJob) { jobs = append(jobs, got...) }
	appendWarn := func(w string) {
		if w != "" {
			warnings = append(warnings, w)
		}
	}

	got, warn := parseCrontabFile("/etc/crontab", true)
	appendWarn(warn)
	appendJobs(got)

	cronDJobs, cronDWarnings := collectCronD()
	jobs = append(jobs, cronDJobs...)
	warnings = append(warnings, cronDWarnings...)

	userJobs, userWarnings := collectUserCrontabs()
	jobs = append(jobs, userJobs...)
	warnings = append(warnings, userWarnings...)

	return jobs, warnings
}

func collectCronD() ([]CronJob, []string) {
	entries, err := os.ReadDir("/etc/cron.d")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, []string{fmt.Sprintf("/etc/cron.d: %v", err)}
	}

	var jobs []CronJob
	var warnings []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, "~") ||
			strings.HasSuffix(name, ".dpkg-old") ||
			strings.HasSuffix(name, ".dpkg-dist") ||
			strings.HasSuffix(name, ".rpmsave") ||
			strings.HasSuffix(name, ".rpmnew") {
			continue
		}
		path := filepath.Join("/etc/cron.d", name)
		got, warn := parseCrontabFile(path, true)
		if warn != "" {
			warnings = append(warnings, warn)
		}
		jobs = append(jobs, got...)
	}

	return jobs, warnings
}

func collectUserCrontabs() ([]CronJob, []string) {
	users, warn := loginShellUsers()
	var warnings []string
	if warn != "" {
		warnings = append(warnings, warn)
	}

	var jobs []CronJob

	for _, user := range users {
		out, err := exec.Command(binCrontab, "-l", "-u", user).Output()
		if err != nil {
			continue
		}
		source := "user:" + user
		got := parseUserCrontabLines(source, strings.Split(string(out), "\n"))
		jobs = append(jobs, got...)
	}

	return jobs, warnings
}

func loginShellUsers() ([]string, string) {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, fmt.Sprintf("cannot read /etc/passwd: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	nologinSuffixes := []string{"nologin", "false", "sync", "shutdown", "halt"}

	var users []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}
		username := parts[0]
		shell := parts[6]
		skip := false
		for _, suffix := range nologinSuffixes {
			if strings.HasSuffix(shell, suffix) {
				skip = true
				break
			}
		}
		if !skip && username != "" {
			users = append(users, username)
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return users, fmt.Sprintf("error scanning /etc/passwd: %v", scanErr)
	}

	return users, ""
}

func parseCrontabFile(path string, systemFormat bool) ([]CronJob, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ""
		}
		return nil, fmt.Sprintf("%s: %v", path, err)
	}

	lines := strings.Split(string(data), "\n")
	if systemFormat {
		return parseSystemCrontabLines(path, lines), ""
	}
	return parseUserCrontabLines(path, lines), ""
}

func parseSystemCrontabLines(source string, lines []string) []CronJob {
	var jobs []CronJob
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if isVariableAssignment(line) {
			continue
		}

		if strings.HasPrefix(line, "@") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				jobs = append(jobs, CronJob{
					Source:   source,
					Schedule: parts[0],
					User:     parts[1],
					Command:  strings.Join(parts[2:], " "),
				})
			}
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}
		jobs = append(jobs, CronJob{
			Source:   source,
			Schedule: strings.Join(fields[:5], " "),
			User:     fields[5],
			Command:  strings.Join(fields[6:], " "),
		})
	}
	return jobs
}

func parseUserCrontabLines(source string, lines []string) []CronJob {
	var jobs []CronJob
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if isVariableAssignment(line) {
			continue
		}

		if strings.HasPrefix(line, "@") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
				jobs = append(jobs, CronJob{
					Source:   source,
					Schedule: parts[0],
					Command:  strings.TrimSpace(parts[1]),
				})
			}
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		jobs = append(jobs, CronJob{
			Source:   source,
			Schedule: strings.Join(fields[:5], " "),
			Command:  strings.Join(fields[5:], " "),
		})
	}
	return jobs
}

func isVariableAssignment(line string) bool {
	idx := strings.IndexByte(line, '=')
	if idx <= 0 {
		return false
	}
	lvalue := line[:idx]
	return !strings.ContainsAny(lvalue, " \t")
}

func FormatCronsForPrompt(jobs []CronJob) string {
	if len(jobs) == 0 {
		return "(no cron jobs found)"
	}
	var b strings.Builder
	for _, j := range jobs {
		_, _ = fmt.Fprintf(&b, "source:   %s\n", j.Source)
		_, _ = fmt.Fprintf(&b, "schedule: %s\n", j.Schedule)
		if j.User != "" {
			_, _ = fmt.Fprintf(&b, "user:     %s\n", j.User)
		}
		_, _ = fmt.Fprintf(&b, "command:  %s\n", j.Command)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
