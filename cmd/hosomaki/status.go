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
				return runStatusBrief(data, p)
			}
			return runStatusFull(data, p)
		},
	}

	cmd.Flags().BoolVar(&brief, "brief", false, "one-sentence summary instead of a paragraph")
	return cmd
}

func runStatusFull(data ui.SnapshotData, p string) error {
	fmt.Print(ui.StatusHeader())
	fmt.Print(ui.StatusSystemSection(data))
	fmt.Print(ui.StatusInsightsSection(data))

	spin := spinner.Start("thinking…")
	raw, err := provider.GenerateJSON(context.Background(), p, spin.Stop)
	spin.Stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	var result prompt.StatusResult
	if parseErr := ui.ParseJSON(raw, &result); parseErr != nil {
		fmt.Fprintf(os.Stderr, "error: could not parse AI response: %v\n", parseErr)
		fmt.Fprintf(os.Stderr, "raw response:\n%s\n", raw)
		return parseErr
	}

	fmt.Print(ui.RenderStatus(result))
	fmt.Print(ui.RenderStatusSummary(result))
	return nil
}

func runStatusBrief(data ui.SnapshotData, p string) error {
	fmt.Print(ui.StatusHeaderBrief())
	fmt.Print(ui.StatusSystemSectionBrief(data))
	fmt.Print(ui.StatusInsightsSectionBrief(data))

	spin := spinner.Start("thinking…")
	raw, err := provider.GenerateJSON(context.Background(), p, spin.Stop)
	spin.Stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	var result prompt.StatusBriefResult
	if parseErr := ui.ParseJSON(raw, &result); parseErr != nil {
		fmt.Fprintf(os.Stderr, "error: could not parse AI response: %v\n", parseErr)
		fmt.Fprintf(os.Stderr, "raw response:\n%s\n", raw)
		return parseErr
	}

	fmt.Print(ui.RenderStatusBrief(result))
	return nil
}
