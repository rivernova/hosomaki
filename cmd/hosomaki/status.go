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

			p := prompt.Status(prompt.StatusInput{
				CollectedAt:    snap.CollectedAt,
				Uptime:         snap.Uptime,
				Memory:         snap.Memory,
				Disk:           snap.Disk,
				FailedServices: snap.FailedServices,
				RecentErrors:   snap.RecentErrors,
				TopProcesses:   snap.TopProcesses,
			}, brief)

			spin := spinner.Start("thinking…")
			_, err = provider.GenerateStream(context.Background(), p,
				func() { spin.Stop() },
				os.Stdout,
			)
			if err != nil {
				spin.Stop()
				return err
			}
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().BoolVar(&brief, "brief", false, "one-sentence summary instead of a paragraph")
	return cmd
}
