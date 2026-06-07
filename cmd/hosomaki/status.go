// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"context"
	"encoding/json"
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

// status command logic

func newStatusCmd() *cobra.Command {
	var (
		brief bool
		debug bool
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show an AI summary of current system health",
		Long: `Collects a snapshot of the system (uptime, memory, disk, failed services,
recent errors) and asks the AI to summarise what's going on.

All output is validated and repaired automatically before being printed.

  hosomaki status           # paragraph summary with anomaly list
  hosomaki status --brief   # single sentence`,

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
			input := prompt.StatusInput{
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
				return runStatusBrief(data, prompt.Status(input, true), debug)
			}
			return runStatusFull(data, prompt.Status(input, false), debug)
		},
	}

	cmd.Flags().BoolVar(&brief, "brief", false, "one-sentence summary instead of a paragraph")
	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr")
	return cmd
}

func statusBriefPipeline() ai.Pipeline[prompt.StatusBriefResult] {
	return ai.NewPipeline(
		provider,
		ai.NewSchema(prompt.SchemaStatusBrief),
		ai.StructValidator[prompt.StatusBriefResult]{},
	)
}

func statusFullStreamPipeline() ai.StreamPipeline[prompt.StatusResult] {
	return ai.NewStreamPipeline(
		provider,
		ai.NewSchema(prompt.SchemaStatusFull),
		ai.StructValidator[prompt.StatusResult]{},
	)
}

func runStatusFull(data ui.SnapshotData, p string, debug bool) error {
	fmt.Print(ui.StatusHeader())
	fmt.Print(ui.StatusSystemSection(data))
	fmt.Print(ui.StatusInsightsSection(data))

	spin := spinner.Start("thinking…")
	pipe := statusFullStreamPipeline()
	if debug {
		pipe = pipe.WithDebug(os.Stderr)
	}

	anomalyCount := 0
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
				case "overview":
					var s string
					if jsonErr := json.Unmarshal([]byte(raw), &s); jsonErr != nil {
						return
					}
					s = strings.TrimSpace(s)
					if s == "" {
						return
					}
					spin.ClearLine()
					fmt.Print(ui.StatusOverviewHeader())
					fmt.Print(ui.RenderStatusOverviewLive(s))

				case "anomalies":
					var a prompt.StatusAnomaly
					if jsonErr := json.Unmarshal([]byte(raw), &a); jsonErr != nil {
						return
					}
					spin.ClearLine()
					if anomalyCount == 0 {
						fmt.Print(ui.StatusAnomaliesHeader())
					}
					fmt.Print(ui.RenderStatusAnomalyLive(a, anomalyCount+1))
					anomalyCount++
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
		fmt.Print(ui.StatusOverviewHeader())
		fmt.Print(ui.RenderStatusOverviewLive(result.Overview))
		fmt.Print(ui.StatusAnomaliesHeader())
		if len(result.Anomalies) == 0 {
			fmt.Print(ui.BulletOK("no anomalies detected"))
		} else {
			for i, a := range result.Anomalies {
				fmt.Print(ui.RenderStatusAnomalyLive(a, i+1))
			}
		}
	} else {
		if anomalyCount == 0 {
			fmt.Print(ui.StatusAnomaliesHeader())
			fmt.Print(ui.BulletOK("no anomalies detected"))
		}
	}

	fmt.Print(ui.RenderStatusSummary(result))
	fmt.Print(ui.Done())
	return nil
}

func runStatusBrief(data ui.SnapshotData, p string, debug bool) error {
	fmt.Print(ui.StatusHeaderBrief())
	fmt.Print(ui.StatusSystemSectionBrief(data))
	fmt.Print(ui.StatusInsightsSectionBrief(data))

	spin := spinner.Start("thinking…")
	pipe := statusBriefPipeline()
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

	fmt.Print(ui.RenderStatusBrief(result))
	fmt.Print(ui.Done())
	return nil
}
