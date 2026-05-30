// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/rivernova/hosomaki/internal/analysis"
	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/insight"
	"github.com/rivernova/hosomaki/internal/output"
	"github.com/rivernova/hosomaki/internal/present"
	"github.com/rivernova/hosomaki/internal/prompt"
	"github.com/rivernova/hosomaki/internal/spinner"
	"github.com/spf13/cobra"
)

// this file contains the "doctor" command logic.

func newDoctorCmd() *cobra.Command {
	var (
		brief     bool
		outputFmt string
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

  hosomaki doctor                    # full diagnosis with suggested actions
  hosomaki doctor --brief            # one sentence per issue
  hosomaki doctor --output json      # machine-readable JSON`,

		Args: cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			snap, err := collector.Snapshot()
			if err != nil {
				return fmt.Errorf("failed to collect system snapshot: %w", err)
			}

			report := analysis.Analyze(present.AnalysisInput(snap))
			p := prompt.Doctor(prompt.DoctorInput{
				Snapshot: snap,
				Language: appCfg.Output.Language,
				Brief:    brief,
			})

			if outputFmt == "json" {
				return doctorJSON(report, p)
			}

			pre := present.DoctorReport(report, insight.Analysis{}, brief)
			_ = currentUI().RenderDoctorStream(pre)

			spin := spinner.Start("thinking…")
			raw, genErr := provider.GenerateStream(context.Background(), p, func() {
				spin.Writing("writing…")
			}, nil)
			spin.Stop()

			doc := insight.ParseDoctor(raw)
			if genErr != nil && doc.Raw == "" && len(doc.Components) == 0 {
				doc.Raw = "AI analysis unavailable: " + genErr.Error()
			}

			finalRep := present.DoctorReport(report, doc, brief)
			currentUI().FinaliseDoctor(finalRep)

			return nil
		},
	}

	cmd.Flags().BoolVar(&brief, "brief", false, "one sentence per issue instead of a full diagnosis")
	cmd.Flags().StringVar(&outputFmt, "output", "", "output format: json")

	return cmd
}

func doctorJSON(report analysis.Report, p string) error {
	spin := spinner.Start("thinking…")
	raw, err := provider.GenerateStream(context.Background(), p, func() {
		spin.Writing("writing…")
	}, nil)
	spin.Stop()

	doc := insight.ParseDoctor(raw)
	if err != nil && doc.Raw == "" && len(doc.Components) == 0 {
		doc.Raw = "AI analysis unavailable: " + err.Error()
	}

	var buf bytes.Buffer
	if encErr := output.WriteDoctor(&buf, report, doc); encErr != nil {
		return fmt.Errorf("encoding JSON: %w", encErr)
	}
	_, writeErr := os.Stdout.Write(buf.Bytes())
	return writeErr
}
