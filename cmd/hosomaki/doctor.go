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
	"github.com/rivernova/hosomaki/internal/stream"
	"github.com/rivernova/hosomaki/internal/ui"
	"github.com/spf13/cobra"
)

// doctor command implementation

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
			data := ui.SnapshotData{
				CollectedAt:    snap.CollectedAt,
				Uptime:         snap.Uptime,
				Memory:         snap.Memory,
				Disk:           snap.Disk,
				FailedServices: snap.FailedServices,
				RecentErrors:   snap.RecentErrors,
			}
			san := sanitiser.Default()
			p := prompt.Doctor(prompt.DoctorInput{
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
				return runDoctorBrief(data, p)
			}
			return runDoctorFull(data, p)
		},
	}

	cmd.Flags().BoolVar(&brief, "brief", false, "one sentence per issue instead of a full diagnosis")
	return cmd
}

func doctorBriefPipeline() ai.Pipeline[prompt.DoctorBriefResult] {
	return ai.NewPipeline(
		provider,
		ai.NewSchema(prompt.SchemaDoctorBrief),
		ai.StructValidator[prompt.DoctorBriefResult]{},
	)
}

func runDoctorFull(data ui.SnapshotData, p string) error {
	fmt.Print(ui.DoctorHeader())
	fmt.Print(ui.DoctorSystemSection(data))
	fmt.Print(ui.DoctorInsightsSection(data))

	var (
		issues              []prompt.DoctorIssue
		actions             []prompt.DoctorAction
		issueHeaderPrinted  bool
		actionHeaderPrinted bool
	)

	spin := spinner.Start("diagnosing…")

	sc := stream.NewArrayItemScanner(func(key, raw string) {
		switch key {
		case "issues":
			var iss prompt.DoctorIssue
			if err := json.Unmarshal([]byte(raw), &iss); err != nil {
				return
			}
			issues = append(issues, iss)
			if !issueHeaderPrinted {
				spin.Stop()
				fmt.Print(ui.DoctorIssuesHeader())
				issueHeaderPrinted = true
			}
			fmt.Print(ui.RenderDoctorIssueLive(iss, len(issues)))

		case "actions":
			var act prompt.DoctorAction
			if err := json.Unmarshal([]byte(raw), &act); err != nil {
				return
			}
			actions = append(actions, act)
			if !actionHeaderPrinted {
				fmt.Print(ui.DoctorActionsHeader())
				actionHeaderPrinted = true
			}
			fmt.Print(ui.RenderDoctorActionLive(act, len(actions)))
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

	if !issueHeaderPrinted {
		fmt.Print(ui.DoctorIssuesHeader())
		fmt.Print(ui.BulletOK("no issues detected"))
	}
	if !actionHeaderPrinted {
		fmt.Print(ui.DoctorActionsHeader())
		fmt.Print(ui.BulletOK("no actions required"))
	}

	fmt.Print(ui.RenderDoctorSummary(prompt.DoctorResult{Issues: issues, Actions: actions}))
	fmt.Print(ui.Done())
	return nil
}

func runDoctorBrief(data ui.SnapshotData, p string) error {
	fmt.Print(ui.DoctorHeaderBrief())
	fmt.Print(ui.DoctorSystemSectionBrief(data))
	fmt.Print(ui.DoctorInsightsSectionBrief(data))

	spin := spinner.Start("diagnosing…")

	result, err := doctorBriefPipeline().Run(
		context.Background(),
		p,
		func() { spin.SetLabel("responding…") },
	)
	spin.Stop()

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	fmt.Print(ui.RenderDoctorBrief(result))
	fmt.Print(ui.RenderDoctorSummary(result))
	fmt.Print(ui.Done())
	return nil
}
