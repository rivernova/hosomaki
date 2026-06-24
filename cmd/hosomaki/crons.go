// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/rivernova/hosomaki/internal/ai"
	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/prompt"
	"github.com/rivernova/hosomaki/internal/sanitiser"
	"github.com/rivernova/hosomaki/internal/spinner"
	"github.com/rivernova/hosomaki/internal/ui"
	"github.com/spf13/cobra"
)

// crons command logic

func newCronsCmd() *cobra.Command {
	var debug bool

	cmd := &cobra.Command{
		Use:   "crons",
		Short: "Inspect cron jobs and explain what each one does",
		Long: `Reads all classic crontab files on the system (/etc/crontab,
/etc/cron.d/*, and per-user crontabs), explains what each job does in plain
English, and flags anything that looks broken, suspicious, or misconfigured.

v1 scope: classic crontab files only. systemd timers are handled by
hosomaki timers.

hosomaki crons never modifies the system. It is strictly read-only.`,

		Args: cobra.NoArgs,

		RunE: func(_ *cobra.Command, _ []string) error {
			jobs, warnings := collector.Crons()

			san := sanitiser.Default()

			sanitisedJobs := make([]collector.CronJob, len(jobs))
			for i, j := range jobs {
				sanitisedJobs[i] = collector.CronJob{
					Source:   san.Sanitise(j.Source),
					Schedule: j.Schedule,
					User:     san.Sanitise(j.User),
					Command:  san.Sanitise(j.Command),
				}
			}

			jobData := collector.FormatCronsForPrompt(sanitisedJobs)
			env := collector.Env()

			generationPrompt := prompt.Crons(prompt.CronsInput{
				Environment: env,
				Jobs:        jobData,
			})

			fmt.Print(ui.CronsHeader())
			fmt.Print(ui.CronsCollectedSection(len(jobs), warnings))

			return runCrons(generationPrompt, len(jobs), debug)
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr")

	return cmd
}

func cronsPipeline() ai.StreamPipeline[prompt.CronsResult] {
	return ai.NewStreamPipeline(
		provider,
		ai.NewSchema(prompt.SchemaCrons),
		ai.StructValidator[prompt.CronsResult]{
			SemanticCheck: validateCronsResult,
		},
	).WithElementCheck("jobs", ai.ElementCheck(validateCronJob)).
		WithEnum("status", cronStatuses...)
}

func validateCronsResult(r prompt.CronsResult) []string {
	var errs []string
	if strings.TrimSpace(r.Summary) == "" {
		errs = append(errs, "summary must not be empty")
	}
	for i, j := range r.Jobs {
		for _, e := range validateCronJob(j) {
			errs = append(errs, fmt.Sprintf("jobs[%d].%s", i, e))
		}
	}
	return errs
}

var cronStatuses = []string{"ok", "warning", "failed"}

func validateCronJob(j prompt.CronJobEntry) []string {
	var errs []string
	if strings.TrimSpace(j.Source) == "" {
		errs = append(errs, "source must not be empty")
	}
	if strings.TrimSpace(j.Command) == "" {
		errs = append(errs, "command must not be empty")
	}
	if strings.TrimSpace(j.WhatItDoes) == "" {
		errs = append(errs, "what_it_does must not be empty")
	}
	status := strings.TrimSpace(j.Status)
	if status == "" {
		errs = append(errs, "status must not be empty")
	} else if !slices.Contains(cronStatuses, status) {
		errs = append(errs, fmt.Sprintf("status must be 'ok', 'warning', or 'failed', got %q", status))
	}
	if strings.TrimSpace(j.LastRun) == "" {
		errs = append(errs, "last_run must not be empty (use 'unknown' when unavailable)")
	}
	return errs
}

func runCrons(generationPrompt string, collectedCount int, debug bool) error {
	spin := spinner.Start("thinking…")

	pipe := cronsPipeline()
	if debug {
		pipe = pipe.WithDebug(os.Stderr)
	}

	jobCount := 0
	summaryPrinted := false

	result, err := pipe.Run(
		context.Background(),
		generationPrompt,
		ai.StreamOptions{
			OnFirstToken: func() { spin.SetLabel("responding…") },
			OnRepairStart: func(n int) {
				spin.SetLabel(fmt.Sprintf("repairing (attempt %d)…", n))
			},
			OnItem: func(key, raw string) {
				switch key {
				case "summary":
					var s string
					if jsonErr := json.Unmarshal([]byte(raw), &s); jsonErr != nil {
						return
					}
					s = strings.TrimSpace(s)
					if s == "" {
						return
					}
					spin.ClearLine()
					if !summaryPrinted {
						fmt.Print(ui.CronsFindingsHeader())
						summaryPrinted = true
					}
					fmt.Print(ui.RenderCronsSummaryLive(s))

				case "jobs":
					var j prompt.CronJobEntry
					if jsonErr := json.Unmarshal([]byte(raw), &j); jsonErr != nil {
						return
					}
					spin.ClearLine()
					if !summaryPrinted {
						fmt.Print(ui.CronsFindingsHeader())
						summaryPrinted = true
					}
					fmt.Print(ui.RenderCronJobLive(j, jobCount+1))
					jobCount++
				}
			},
		},
	)

	spin.Stop()

	if err != nil && !errors.Is(err, ai.ErrIncomplete) {
		_, ferr := fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if ferr != nil {
			return ferr
		}
		return err
	}
	if errors.Is(err, ai.ErrIncomplete) {
		_, _ = fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	if jobCount == 0 && !summaryPrinted {
		fmt.Print(ui.CronsFindingsHeader())
		fmt.Print(ui.RenderCronsSummaryLive(result.Summary))
	}

	anyIssue := false
	for _, j := range result.Jobs {
		if j.Status == "warning" || j.Status == "failed" {
			anyIssue = true
			break
		}
	}
	if collectedCount > 0 && !anyIssue {
		fmt.Print(ui.CronsCleanResult())
	}

	fmt.Print(ui.Done())
	return nil
}
