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
	"github.com/rivernova/hosomaki/internal/stream"
	"github.com/rivernova/hosomaki/internal/ui"
	"github.com/spf13/cobra"
)

// status command implementation

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
			san := sanitiser.Default()
			p := prompt.Status(prompt.StatusInput{
				CollectedAt:    snap.CollectedAt,
				Environment:    snap.Environment,
				Uptime:         snap.Uptime,
				Memory:         snap.Memory,
				Disk:           snap.Disk,
				FailedServices: san.Sanitise(snap.FailedServices),
				RecentErrors:   san.Sanitise(snap.RecentErrors),
				TopProcesses:   san.Sanitise(snap.TopProcesses),
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

func statusBriefPipeline() ai.Pipeline[prompt.StatusBriefResult] {
	return ai.NewPipeline(
		provider,
		ai.NewSchema(prompt.SchemaStatusBrief),
		ai.StructValidator[prompt.StatusBriefResult]{},
	)
}

func runStatusFull(data ui.SnapshotData, p string) error {
	fmt.Print(ui.StatusHeader())
	fmt.Print(ui.StatusSystemSection(data))
	fmt.Print(ui.StatusInsightsSection(data))

	var (
		anomalies            []prompt.StatusAnomaly
		overviewPrinted      bool
		anomalyHeaderPrinted bool
	)

	spin := spinner.Start("thinking…")

	sc := stream.NewArrayItemScanner(func(key, raw string) {
		switch key {
		case "overview":
			var overview string
			if err := json.Unmarshal([]byte(raw), &overview); err != nil {
				return
			}
			overview = strings.TrimSpace(overview)
			if overview == "" {
				return
			}
			spin.Stop()
			fmt.Print(ui.StatusOverviewHeader())
			fmt.Print(ui.RenderStatusOverviewLive(overview))
			overviewPrinted = true

		case "anomalies":
			var a prompt.StatusAnomaly
			if err := json.Unmarshal([]byte(raw), &a); err != nil {
				return
			}
			anomalies = append(anomalies, a)
			if !anomalyHeaderPrinted {
				if !overviewPrinted {
					spin.Stop()
				}
				fmt.Print(ui.StatusAnomaliesHeader())
				anomalyHeaderPrinted = true
			}
			fmt.Print(ui.RenderStatusAnomalyLive(a, len(anomalies)))
		}
	})

	_, err := provider.GenerateStream(context.Background(), p,
		func() { spin.SetLabel("responding…") },
		sc,
	)
	spin.Stop()

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	if !anomalyHeaderPrinted {
		fmt.Print(ui.StatusAnomaliesHeader())
		fmt.Print(ui.BulletOK("no anomalies detected"))
	}

	fmt.Print(ui.RenderStatusSummary(prompt.StatusResult{Anomalies: anomalies}))
	fmt.Print(ui.Done())
	return nil
}

func runStatusBrief(data ui.SnapshotData, p string) error {
	fmt.Print(ui.StatusHeaderBrief())
	fmt.Print(ui.StatusSystemSectionBrief(data))
	fmt.Print(ui.StatusInsightsSectionBrief(data))

	spin := spinner.Start("thinking…")

	result, err := statusBriefPipeline().Run(
		context.Background(),
		p,
		func() { spin.SetLabel("responding…") },
	)
	spin.Stop()

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	fmt.Print(ui.RenderStatusBrief(result))
	fmt.Print(ui.Done())
	return nil
}
