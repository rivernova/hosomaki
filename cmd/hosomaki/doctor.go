// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/rivernova/hosomaki/internal/ai"
	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/prompt"
	"github.com/rivernova/hosomaki/internal/sanitiser"
	"github.com/rivernova/hosomaki/internal/spinner"
	"github.com/rivernova/hosomaki/internal/store"
	"github.com/rivernova/hosomaki/internal/ui"
	"github.com/spf13/cobra"
)

// doctor command logic

func newDoctorCmd() *cobra.Command {
	var (
		brief bool
		debug bool
	)

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Full system diagnosis with concrete suggested actions",
		Long: `Collects a snapshot of the system (uptime, memory, disk, failed services,
recent errors) and asks the AI to diagnose what is wrong and what to do about it.

Unlike ` + "`hosomaki status`" + `, which only describes what it sees, doctor goes further:
for each detected issue it explains the likely cause and proposes specific actions
you can take — commands to run, files to inspect, configuration values to change.

If a suggested action is potentially disruptive or irreversible, the output will
say so explicitly before describing it. Doctor never modifies the system itself.

All output is validated and repaired automatically before being printed.

  hosomaki doctor           # full diagnosis with suggested actions
  hosomaki doctor --brief   # one sentence per issue`,

		Args: cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			snap, err := collector.Snapshot()
			if err != nil {
				return fmt.Errorf("failed to collect system snapshot: %w", err)
			}
			data := ui.SnapshotData{
				CollectedAt:    snap.CollectedAt,
				Uptime:         snap.Uptime,
				Memory:         snap.Memory,
				Disk:           snap.Disk,
				FailedServices: snap.FailedServices,
				RecentErrors:   snap.RecentErrors,
			}
			san := sanitiser.Default()
			input := prompt.DoctorInput{
				CollectedAt:    snap.CollectedAt,
				Environment:    snap.Environment,
				Uptime:         snap.Uptime,
				Memory:         snap.Memory,
				Disk:           snap.Disk,
				FailedServices: san.Sanitise(snap.FailedServices),
				RecentErrors:   san.Sanitise(snap.RecentErrors),
				TopProcesses:   san.Sanitise(snap.TopProcesses),
			}

			if brief {
				return runDoctorBrief(data, prompt.Doctor(input, true), debug)
			}
			return runDoctorFull(data, prompt.Doctor(input, false), debug)
		},
	}

	cmd.Flags().BoolVar(&brief, "brief", false, "one sentence per issue instead of a full diagnosis")
	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr")
	return cmd
}

func doctorFullStreamPipeline() ai.StreamPipeline[prompt.DoctorResult] {
	return ai.NewStreamPipeline(
		provider,
		ai.NewSchema(prompt.SchemaDoctorFull),
		ai.StructValidator[prompt.DoctorResult]{},
	)
}

func doctorBriefPipeline() ai.Pipeline[prompt.DoctorBriefResult] {
	return ai.NewPipeline(
		provider,
		ai.NewSchema(prompt.SchemaDoctorBrief),
		ai.StructValidator[prompt.DoctorBriefResult]{},
	)
}

func runDoctorFull(data ui.SnapshotData, p string, debug bool) error {
	fmt.Print(ui.DoctorHeader())
	fmt.Print(ui.DoctorSystemSection(data))
	fmt.Print(ui.DoctorInsightsSection(data))

	spin := spinner.Start("diagnosing…")
	pipe := doctorFullStreamPipeline()
	if debug {
		pipe = pipe.WithDebug(os.Stderr)
	}

	issueCount := 0
	actionCount := 0
	wasRepaired := false

	result, err := pipe.Run(
		context.Background(),
		p,
		ai.StreamOptions{
			OnFirstToken: func() { spin.SetLabel("responding…") },
			OnRepairStart: func(n int) {
				wasRepaired = true
				spin.SetLabel(fmt.Sprintf("repairing (attempt %d)…", n))
			},
			OnItem: func(key, raw string) {
				switch key {
				case "issues":
					var iss prompt.DoctorIssue
					if jsonErr := json.Unmarshal([]byte(raw), &iss); jsonErr != nil {
						return
					}
					spin.ClearLine()
					if issueCount == 0 {
						fmt.Print(ui.DoctorIssuesHeader())
					}
					fmt.Print(ui.RenderDoctorIssueLive(iss, issueCount+1))
					issueCount++

				case "actions":
					var act prompt.DoctorAction
					if jsonErr := json.Unmarshal([]byte(raw), &act); jsonErr != nil {
						return
					}
					spin.ClearLine()
					if actionCount == 0 {
						fmt.Print(ui.DoctorActionsHeader())
					}
					fmt.Print(ui.RenderDoctorActionLive(act, actionCount+1))
					actionCount++
				}
			},
		},
	)

	spin.Stop()

	if err != nil {
		_, ferr := fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if ferr != nil {
			return ferr
		}
		return err
	}

	// if a repair happened, items may not match the validated
	if wasRepaired {
		fmt.Print(ui.DoctorIssuesHeader())
		if len(result.Issues) == 0 {
			fmt.Print(ui.BulletOK("no issues detected"))
		} else {
			for i, iss := range result.Issues {
				fmt.Print(ui.RenderDoctorIssueLive(iss, i+1))
			}
		}
		fmt.Print(ui.DoctorActionsHeader())
		if len(result.Actions) == 0 {
			fmt.Print(ui.BulletOK("no actions required"))
		} else {
			for i, act := range result.Actions {
				fmt.Print(ui.RenderDoctorActionLive(act, i+1))
			}
		}
	} else {
		if issueCount == 0 {
			fmt.Print(ui.DoctorIssuesHeader())
			fmt.Print(ui.BulletOK("no issues detected"))
		}
		if actionCount == 0 {
			fmt.Print(ui.DoctorActionsHeader())
			fmt.Print(ui.BulletOK("no actions required"))
		}
	}

	fmt.Print(ui.RenderDoctorSummary(result))
	if err := store.Record("doctor", result); err != nil && debug {
		_, _ = fmt.Fprintf(os.Stderr, "history: record doctor: %v\n", err)
	}
	fmt.Print(ui.Done())
	return nil
}

func runDoctorBrief(data ui.SnapshotData, p string, debug bool) error {
	fmt.Print(ui.DoctorHeaderBrief())
	fmt.Print(ui.DoctorSystemSectionBrief(data))
	fmt.Print(ui.DoctorInsightsSectionBrief(data))

	spin := spinner.Start("diagnosing…")
	pipe := doctorBriefPipeline()
	if debug {
		pipe = pipe.WithDebug(os.Stderr)
	}

	result, err := pipe.Run(
		context.Background(),
		p,
		ai.RunOptions{
			OnFirstToken:  func() { spin.SetLabel("responding…") },
			OnRepairStart: func(n int) { spin.SetLabel(fmt.Sprintf("repairing (attempt %d)…", n)) },
		},
	)
	spin.Stop()

	if err != nil {
		_, ferr := fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if ferr != nil {
			return ferr
		}
		return err
	}

	fmt.Print(ui.RenderDoctorBrief(result))
	if err := store.Record("doctor", result); err != nil && debug {
		_, _ = fmt.Fprintf(os.Stderr, "history: record doctor: %v\n", err)
	}
	fmt.Print(ui.Done())
	return nil
}
