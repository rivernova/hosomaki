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

// ports command logic

func newPortsCmd() *cobra.Command {
	var debug bool

	cmd := &cobra.Command{
		Use:   "ports",
		Short: "List listening ports with process names and flag anything unexpected",
		Long: `Collects all currently listening TCP and UDP ports with their associated
process names, then asks the AI to identify anything unexpected or potentially
concerning.

This command shows the current state of listening sockets. For tracking how ports
change over time (ports opened or closed since a baseline), use hosomaki audit.

hosomaki ports never modifies the system. It is strictly read-only.`,

		Args: cobra.NoArgs,

		RunE: func(_ *cobra.Command, _ []string) error {
			entries, warnings := collector.Ports()

			san := sanitiser.Default()
			rawFormatted := collector.FormatPortsForPrompt(entries)
			sanitisedPorts := san.Sanitise(rawFormatted)

			env := collector.Env()
			generationPrompt := prompt.Ports(prompt.PortsInput{
				Environment: env,
				Ports:       sanitisedPorts,
			})

			fmt.Print(ui.PortsHeader())
			fmt.Print(ui.PortsCollectedSection(len(entries), warnings))

			return runPorts(generationPrompt, debug)
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr")

	return cmd
}

func portsPipeline() ai.StreamPipeline[prompt.PortsResult] {
	return ai.NewStreamPipeline(
		provider,
		ai.NewSchema(prompt.SchemaPorts),
		ai.StructValidator[prompt.PortsResult]{
			SemanticCheck: validatePortsResult,
		},
	)
}

func validatePortsResult(r prompt.PortsResult) []string {
	var errs []string
	if strings.TrimSpace(r.Summary) == "" {
		errs = append(errs, "summary must not be empty")
	}
	for i, f := range r.Findings {
		sev := strings.TrimSpace(f.Severity)
		if sev == "" {
			errs = append(errs, fmt.Sprintf("findings[%d].severity must not be empty", i))
		} else if sev != "warning" && sev != "info" {
			errs = append(errs, fmt.Sprintf("findings[%d].severity must be 'warning' or 'info', got %q", i, sev))
		}
		if strings.TrimSpace(f.Port) == "" {
			errs = append(errs, fmt.Sprintf("findings[%d].port must not be empty", i))
		}
		if strings.TrimSpace(f.Title) == "" {
			errs = append(errs, fmt.Sprintf("findings[%d].title must not be empty", i))
		}
		if strings.TrimSpace(f.Detail) == "" {
			errs = append(errs, fmt.Sprintf("findings[%d].detail must not be empty", i))
		}
	}
	return errs
}

func runPorts(generationPrompt string, debug bool) error {
	spin := spinner.Start("thinking…")

	pipe := portsPipeline()
	if debug {
		pipe = pipe.WithDebug(os.Stderr)
	}

	findingCount := 0
	summaryPrinted := false
	wasRepaired := false

	result, err := pipe.Run(
		context.Background(),
		generationPrompt,
		ai.StreamOptions{
			OnFirstToken: func() { spin.SetLabel("responding…") },
			OnRepairStart: func(n int) {
				wasRepaired = true
				spin.SetLabel(fmt.Sprintf("repairing (attempt %d)…", n))
			},
			OnItem: func(key, raw string) {
				switch key {
				case "summary":
					var s string
					if jsonErr := json.Unmarshal([]byte(raw), &s); jsonErr != nil {
						return
					}
					s = strings.TrimSpace(s)
					if s == "" {
						return
					}
					spin.ClearLine()
					if !summaryPrinted {
						fmt.Print(ui.AuditFindingsHeader())
						summaryPrinted = true
					}
					fmt.Print(ui.RenderPortsSummaryLive(s))

				case "findings":
					var f prompt.PortsFinding
					if jsonErr := json.Unmarshal([]byte(raw), &f); jsonErr != nil {
						return
					}
					spin.ClearLine()
					if !summaryPrinted {
						fmt.Print(ui.AuditFindingsHeader())
						summaryPrinted = true
					}
					fmt.Print(ui.RenderPortsFindingLive(f, findingCount+1))
					findingCount++
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

	if wasRepaired {
		if !summaryPrinted {
			fmt.Print(ui.AuditFindingsHeader())
		}
		fmt.Print(ui.RenderPortsSummaryLive(result.Summary))
		for i, f := range result.Findings {
			fmt.Print(ui.RenderPortsFindingLive(f, i+1))
		}
	} else if findingCount == 0 && !summaryPrinted {
		fmt.Print(ui.AuditFindingsHeader())
		fmt.Print(ui.RenderPortsSummaryLive(result.Summary))
	}

	if len(result.Findings) == 0 {
		fmt.Print(ui.PortsCleanResult())
	} else {
		fmt.Print(ui.RenderPortsResultSummary(result))
	}

	fmt.Print(ui.Done())
	return nil
}
