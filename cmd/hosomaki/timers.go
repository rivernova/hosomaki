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
	"strings"

	"github.com/rivernova/hosomaki/internal/ai"
	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/prompt"
	"github.com/rivernova/hosomaki/internal/sanitiser"
	"github.com/rivernova/hosomaki/internal/spinner"
	"github.com/rivernova/hosomaki/internal/ui"
	"github.com/spf13/cobra"
)

// timers command logic

func newTimersCmd() *cobra.Command {
	var debug bool

	cmd := &cobra.Command{
		Use:   "timers",
		Short: "Inspect all systemd timers and flag failures or overdue schedules",
		Long: `Collects all active and inactive systemd timers, shows when each one
last ran and when it will next run, and asks the AI to flag any that have
failed, never run, or appear to be overdue.

Timers with no recorded last run are reported as last_run: "never".

hosomaki timers never modifies the system. It is strictly read-only.`,

		Args: cobra.NoArgs,

		RunE: func(_ *cobra.Command, _ []string) error {
			entries, warning := collector.Timers()

			san := sanitiser.Default()

			sanitisedEntries := make([]collector.TimerEntry, len(entries))
			for i, e := range entries {
				sanitisedEntries[i] = collector.TimerEntry{
					Unit:        san.Sanitise(e.Unit),
					Activates:   san.Sanitise(e.Activates),
					Next:        e.Next,
					Last:        e.Last,
					LastResult:  e.LastResult,
					ActiveState: e.ActiveState,
				}
			}

			timerData := collector.FormatTimersForPrompt(sanitisedEntries)
			env := collector.Env()

			generationPrompt := prompt.Timers(prompt.TimersInput{
				Environment: env,
				Timers:      timerData,
			})

			fmt.Print(ui.TimersHeader())
			fmt.Print(ui.TimersCollectedSection(len(entries), warning))

			return runTimers(generationPrompt, len(entries), debug)
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr")

	return cmd
}

func timersPipeline() ai.StreamPipeline[prompt.TimersResult] {
	return ai.NewStreamPipeline(
		provider,
		ai.NewSchema(prompt.SchemaTimers),
		ai.StructValidator[prompt.TimersResult]{
			SemanticCheck: validateTimersResult,
		},
	).WithElementCheck("timers", ai.ElementCheck(validateTimer))
}

func validateTimersResult(r prompt.TimersResult) []string {
	var errs []string
	if strings.TrimSpace(r.Summary) == "" {
		errs = append(errs, "summary must not be empty")
	}
	for i, e := range r.Timers {
		for _, msg := range validateTimer(e) {
			errs = append(errs, fmt.Sprintf("timers[%d].%s", i, msg))
		}
	}
	return errs
}

func validateTimer(e prompt.TimerEntry) []string {
	var errs []string
	if strings.TrimSpace(e.Name) == "" {
		errs = append(errs, "name must not be empty")
	}
	status := strings.TrimSpace(e.Status)
	if status == "" {
		errs = append(errs, "status must not be empty")
	} else if status != "ok" && status != "warning" && status != "failed" {
		errs = append(errs, fmt.Sprintf("status must be 'ok', 'warning', or 'failed', got %q", status))
	}
	if strings.TrimSpace(e.LastRun) == "" {
		errs = append(errs, "last_run must not be empty")
	}
	if strings.TrimSpace(e.NextRun) == "" {
		errs = append(errs, "next_run must not be empty")
	}
	return errs
}

func runTimers(generationPrompt string, collectedCount int, debug bool) error {
	spin := spinner.Start("thinking…")

	pipe := timersPipeline()
	if debug {
		pipe = pipe.WithDebug(os.Stderr)
	}

	timerCount := 0
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
						fmt.Print(ui.TimersFindingsHeader())
						summaryPrinted = true
					}
					fmt.Print(ui.RenderTimersSummaryLive(s))

				case "timers":
					var e prompt.TimerEntry
					if jsonErr := json.Unmarshal([]byte(raw), &e); jsonErr != nil {
						return
					}
					spin.ClearLine()
					if !summaryPrinted {
						fmt.Print(ui.TimersFindingsHeader())
						summaryPrinted = true
					}
					fmt.Print(ui.RenderTimerLive(e, timerCount+1))
					timerCount++
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

	if timerCount == 0 && !summaryPrinted {
		fmt.Print(ui.TimersFindingsHeader())
		fmt.Print(ui.RenderTimersSummaryLive(result.Summary))
	}

	anyIssue := false
	for _, e := range result.Timers {
		if e.Status == "warning" || e.Status == "failed" {
			anyIssue = true
			break
		}
	}
	if collectedCount > 0 && !anyIssue {
		fmt.Print(ui.TimersCleanResult())
	}

	fmt.Print(ui.Done())
	return nil
}
