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

func newUpdatesCmd() *cobra.Command {
	var (
		securityOnly bool
		debug        bool
	)

	cmd := &cobra.Command{
		Use:   "updates",
		Short: "Explain pending package updates before applying them",
		Long: `Lists pending package updates and explains what each one changes -
flagging security fixes, major version bumps, and updates that require a
reboot.

Read-only. Does not apply updates.

Examples:
  hosomaki updates
  hosomaki updates --security-only`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			env := collector.Env()

			spin := spinner.Start("checking for pending updates\u2026")
			pending, err := collector.Updates(env)
			spin.Stop()
			if err != nil {
				return fmt.Errorf("updates: %w", err)
			}

			fmt.Print(ui.UpdatesHeader())

			if len(pending) == 0 {
				fmt.Print(ui.UpdatesNoPending())
				fmt.Print(ui.Done())
				return nil
			}

			// Filter by --security-only
			filtered := pending
			if securityOnly {
				var sec []collector.Update
				for _, p := range pending {
					if p.Security {
						sec = append(sec, p)
					}
				}
				filtered = sec
			}

			if len(filtered) == 0 {
				msg := "No pending updates"
				if securityOnly {
					msg = "No security-related pending updates"
				}
				fmt.Print(ui.UpdatesNoPendingMsg(msg))
				fmt.Print(ui.Done())
				return nil
			}

			fmt.Print(ui.UpdatesPendingList(filtered, securityOnly))

			// Sanitise before passing to prompt
			san := sanitiser.Default()
			sanitisedText := san.Sanitise(collector.FormatUpdatesForPrompt(filtered))

			
			p := prompt.Updates(prompt.UpdatesInput{
				Environment:  env,
				Updates:      sanitisedText,
				SecurityOnly: securityOnly,
			})

			spin = spinner.Start("thinking\u2026")
			pipe := updatesStreamPipeline()
			if debug {
				pipe = pipe.WithDebug(os.Stderr)
			}

			summaryPrinted := false

			result, err := pipe.Run(
				context.Background(),
				p,
				ai.StreamOptions{
					OnFirstToken: func() { spin.SetLabel("responding\u2026") },
					OnRepairStart: func(n int) {
						spin.SetLabel(fmt.Sprintf("repairing (attempt %d)\u2026", n))
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
								fmt.Print(ui.UpdatesFindingsHeader())
								summaryPrinted = true
							}
							fmt.Print(ui.RenderUpdatesSummaryLive(s))

						case "updates":
							var u prompt.UpdateFinding
							if jsonErr := json.Unmarshal([]byte(raw), &u); jsonErr != nil {
								return
							}
							spin.ClearLine()
							if !summaryPrinted {
								fmt.Print(ui.UpdatesFindingsHeader())
								summaryPrinted = true
							}
							fmt.Print(ui.RenderUpdatesFindingLive(u, 0))
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

			if !summaryPrinted {
				fmt.Print(ui.UpdatesFindingsHeader())
				fmt.Print(ui.RenderUpdatesSummaryLive(result.Summary))
				for _, u := range result.Updates {
					fmt.Print(ui.RenderUpdatesFindingLive(u, 0))
				}
			}

			if len(result.Updates) == 0 {
				fmt.Print(ui.UpdatesCleanResult())
			} else {
				fmt.Print(ui.RenderUpdatesResultSummary(result))
			}
			fmt.Print(ui.Done())
			return nil
		},
	}

	cmd.Flags().BoolVar(&securityOnly, "security-only", false, "show only security-related updates")
	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr")

	return cmd
}

func updatesStreamPipeline() ai.StreamPipeline[prompt.UpdatesResult] {
	return ai.NewStreamPipeline(
		provider,
		ai.NewSchema(prompt.SchemaUpdates),
		ai.StructValidator[prompt.UpdatesResult]{
			SemanticCheck: validateUpdatesResult,
		},
	)
}

func validateUpdatesResult(r prompt.UpdatesResult) []string {
	var errs []string
	if strings.TrimSpace(r.Summary) == "" {
		errs = append(errs, "summary must not be empty")
	}
	for i, u := range r.Updates {
		if strings.TrimSpace(u.Package) == "" {
			errs = append(errs, fmt.Sprintf("updates[%d].package must not be empty", i))
		}
		cat := u.Category
		if cat != "security" && cat != "major" && cat != "minor" && cat != "unknown" {
			errs = append(errs, fmt.Sprintf("updates[%d].category must be 'security'/'major'/'minor'/'unknown', got %q", i, cat))
		}
	}
	return errs
}
