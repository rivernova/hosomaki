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

// this file contains the implementation of the "doctor" command

func newDoctorCmd() *cobra.Command {
	var brief bool

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

  hosomaki doctor           # full diagnosis with suggested actions
  hosomaki doctor --brief   # one sentence per issue`,

		Args: cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			snap, err := collector.Snapshot()
			if err != nil {
				return fmt.Errorf("failed to collect system snapshot: %w", err)
			}

			p := prompt.Doctor(prompt.DoctorInput{
				CollectedAt:    snap.CollectedAt,
				Environment:    snap.Environment,
				Uptime:         snap.Uptime,
				Memory:         snap.Memory,
				Disk:           snap.Disk,
				FailedServices: snap.FailedServices,
				RecentErrors:   snap.RecentErrors,
				TopProcesses:   snap.TopProcesses,
			}, brief)

			spin := spinner.Start("diagnosing…")
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

	cmd.Flags().BoolVar(&brief, "brief", false, "one sentence per issue instead of a full diagnosis")
	return cmd
}
