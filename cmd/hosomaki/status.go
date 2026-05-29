// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rivernova/hosomaki/internal/analysis"
	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/insight"
	"github.com/rivernova/hosomaki/internal/output"
	"github.com/rivernova/hosomaki/internal/present"
	"github.com/rivernova/hosomaki/internal/prompt"
	"github.com/rivernova/hosomaki/internal/spinner"
	"github.com/spf13/cobra"
)

// this file contains the "status" command logic.

func newStatusCmd() *cobra.Command {
	var (
		brief     bool
		outputFmt string
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show an AI summary of current system health",
		Long: `Collects a snapshot of the system (uptime, memory, disk, failed services,
recent errors) and asks the AI to summarise what's going on.

  hosomaki status                    # at-a-glance health summary
  hosomaki status --brief            # single sentence
  hosomaki status --output json      # machine-readable JSON`,

		Args: cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			snap, err := collector.Snapshot()
			if err != nil {
				return fmt.Errorf("failed to collect system snapshot: %w", err)
			}

			report := analysis.Analyze(present.AnalysisInput(snap))
			p := prompt.Status(prompt.StatusInput{
				Snapshot: snap,
				Language: appCfg.Output.Language,
				Brief:    brief,
			})

			if outputFmt == "json" {
				return statusJSON(report, p)
			}

			partial := present.StatusReport(report, insight.Status{}, brief)
			_ = currentUI().RenderStatusStream(partial)

			var aiBuf bytes.Buffer
			spin := spinner.Start("thinking…")
			_, genErr := provider.GenerateStream(context.Background(), p, func() {
				spin.Stop()
			}, &aiBuf)
			spin.Stop()

			rawAI := strings.TrimSpace(aiBuf.String())

			st := insight.ParseStatus(rawAI)
			if genErr != nil && st.Raw == "" && len(st.Observations) == 0 {
				st.Raw = "AI analysis unavailable: " + genErr.Error()
			}

			issues := insight.ParseDoctor(rawAI).Issues

			finalRep := present.StatusReportWithAI(report, issues, st, brief)
			currentUI().FinaliseStatus(finalRep)

			return nil
		},
	}

	cmd.Flags().BoolVar(&brief, "brief", false, "one-sentence summary instead of a paragraph")
	cmd.Flags().StringVar(&outputFmt, "output", "", "output format: json")

	return cmd
}

func statusJSON(report analysis.Report, p string) error {
	spin := spinner.Start("thinking…")
	raw, err := provider.Generate(context.Background(), p)
	spin.Stop()

	st := insight.ParseStatus(raw)
	if err != nil && st.Raw == "" && st.Summary == "" {
		st.Summary = "AI summary unavailable: " + err.Error()
	}

	var buf bytes.Buffer
	if encErr := output.WriteStatus(&buf, report, st); encErr != nil {
		return fmt.Errorf("encoding JSON: %w", encErr)
	}
	_, writeErr := os.Stdout.Write(buf.Bytes())
	return writeErr
}
