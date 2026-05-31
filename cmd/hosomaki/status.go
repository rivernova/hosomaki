// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"context"
	"fmt"
	"os"

	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/prompt"
	"github.com/rivernova/hosomaki/internal/spinner"
	"github.com/rivernova/hosomaki/internal/ui"
	"github.com/spf13/cobra"
)

// this file contains the implementation of the "status" command

func newStatusCmd() *cobra.Command {
	var brief bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show an AI summary of current system health",
		Long: `Collects a snapshot of the system (uptime, memory, disk, failed services,
recent errors) and asks the AI to summarise what's going on.

  hosomaki status           # paragraph summary
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

			p := prompt.Status(prompt.StatusInput{
				CollectedAt:    snap.CollectedAt,
				Environment:    snap.Environment,
				Uptime:         snap.Uptime,
				Memory:         snap.Memory,
				Disk:           snap.Disk,
				FailedServices: snap.FailedServices,
				RecentErrors:   snap.RecentErrors,
				TopProcesses:   snap.TopProcesses,
			}, brief)

			if brief {
				printStatusBrief(data, p)
			} else {
				printStatusFull(data, p)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&brief, "brief", false, "one-sentence summary instead of a paragraph")
	return cmd
}

// printStatusFull renders the full-mode status layout:
//
//	Status
//	──────────────────────────────────────────────
//
//	System status
//	──────────────────────────────────────────────
//	<key/value metrics>
//
//	Local insights
//	──────────────────────────────────────────────
//	<✓ ! ✗ bullets>
//
//	AI analysis
//	──────────────────────────────────────────────
//	<AI streaming output — untouched>
//
//	Summary
//	──────────────────────────────────────────────
//	<summary lines>
func printStatusFull(data ui.SnapshotData, p string) {
	fmt.Print(ui.StatusHeader())
	fmt.Print(ui.StatusSystemSection(data))
	fmt.Print(ui.StatusInsightsSection(data))
	fmt.Print(ui.StatusAIHeader())

	spin := spinner.Start("thinking…")
	_, err := provider.GenerateStream(context.Background(), p,
		func() { spin.Stop() },
		os.Stdout,
	)
	if err != nil {
		spin.Stop()
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Println()

	fmt.Print(ui.StatusSummary(data))
}

// printStatusBrief renders the brief-mode status layout:
//
//	Status (brief)
//	──────────────────────────────────────────────
//
//	System
//	<compact metrics>
//
//	Insights
//	<compact bullets>
//
//	AI
//	<AI streaming output — untouched>
//
//	Summary
//	<compact summary>
func printStatusBrief(data ui.SnapshotData, p string) {
	fmt.Print(ui.StatusHeaderBrief())
	fmt.Print(ui.StatusSystemSectionBrief(data))
	fmt.Print(ui.StatusInsightsSectionBrief(data))
	fmt.Print(ui.StatusAIHeaderBrief())

	spin := spinner.Start("thinking…")
	_, err := provider.GenerateStream(context.Background(), p,
		func() { spin.Stop() },
		os.Stdout,
	)
	if err != nil {
		spin.Stop()
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Println()

	fmt.Print(ui.StatusSummaryBrief(data))
}
