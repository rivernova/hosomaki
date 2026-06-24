// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"context"
	"encoding/json"
	"errors"
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

// mounts command logic

func newMountsCmd() *cobra.Command {
	var debug bool

	cmd := &cobra.Command{
		Use:   "mounts",
		Short: "Inspect active mounts, detect stale NFS, and flag disks approaching capacity",
		Long: `Reads /proc/mounts and df to inspect all active mount points. For NFS
mounts it performs a non-blocking staleness check with a hard timeout to avoid
hanging on unresponsive servers. Disk usage thresholds are flagged by the AI.

Data sources used (all read-only):
  /proc/mounts   mount table
  df -Pl         disk usage for local, non-NFS filesystems
  stat           NFS responsiveness probe (2 second timeout, per mount)

hosomaki mounts never modifies the system.`,

		Args: cobra.NoArgs,

		RunE: func(_ *cobra.Command, _ []string) error {
			result := collector.Mounts()

			total := len(result.Entries)
			mountKinds, nfs, stale := countMountKinds(result.Entries)

			san := sanitiser.Default()
			sanitisedMounts := san.Sanitise(collector.FormatMountsForPrompt(result.Entries))

			env := collector.Env()
			generationPrompt := prompt.Mounts(prompt.MountsInput{
				Environment: env,
				Mounts:      sanitisedMounts,
			})

			fmt.Print(ui.MountsHeader())
			fmt.Print(ui.MountsCollectedSection(total, mountKinds, nfs, stale, result.Warnings))

			return runMounts(generationPrompt, debug)
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr")
	return cmd
}

func countMountKinds(entries []collector.MountEntry) (real, nfs, staleNFS int) {
	for _, e := range entries {
		if collector.IsPseudoFS(e.FSType) {
			continue
		}
		real++
		if collector.IsNFS(e.FSType) {
			nfs++
			if e.NFSStale {
				staleNFS++
			}
		}
	}
	return
}

func mountsPipeline() ai.StreamPipeline[prompt.MountsResult] {
	return ai.NewStreamPipeline(
		provider,
		ai.NewSchema(prompt.SchemaMounts),
		ai.StructValidator[prompt.MountsResult]{
			SemanticCheck: validateMountsResult,
		},
	).WithElementCheck("findings", ai.ElementCheck(validateMountsFinding))
}

func validateMountsResult(r prompt.MountsResult) []string {
	var errs []string
	if strings.TrimSpace(r.Summary) == "" {
		errs = append(errs, "summary must not be empty")
	}
	for i, f := range r.Findings {
		for _, e := range validateMountsFinding(f) {
			errs = append(errs, fmt.Sprintf("findings[%d].%s", i, e))
		}
	}
	return errs
}

func validateMountsFinding(f prompt.MountFinding) []string {
	var errs []string
	sev := strings.TrimSpace(f.Severity)
	if sev == "" {
		errs = append(errs, "severity must not be empty")
	} else if sev != "critical" && sev != "warning" && sev != "info" {
		errs = append(errs, fmt.Sprintf(
			"severity must be 'critical', 'warning', or 'info', got %q", sev,
		))
	}
	if strings.TrimSpace(f.MountPoint) == "" {
		errs = append(errs, "mount_point must not be empty")
	}
	if strings.TrimSpace(f.Title) == "" {
		errs = append(errs, "title must not be empty")
	}
	if strings.TrimSpace(f.Detail) == "" {
		errs = append(errs, "detail must not be empty")
	}
	return errs
}

func runMounts(generationPrompt string, debug bool) error {
	spin := spinner.Start("thinking…")

	pipe := mountsPipeline()
	if debug {
		pipe = pipe.WithDebug(os.Stderr)
	}

	findingCount := 0
	summaryPrinted := false

	result, err := pipe.Run(
		context.Background(),
		generationPrompt,
		ai.StreamOptions{
			OnFirstToken: func() { spin.SetLabel("responding…") },
			OnRepairStart: func(n int) {
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
						fmt.Print(ui.MountsFindingsHeader())
						summaryPrinted = true
					}
					fmt.Print(ui.RenderMountsSummaryLive(s))

				case "findings":
					var f prompt.MountFinding
					if jsonErr := json.Unmarshal([]byte(raw), &f); jsonErr != nil {
						return
					}
					spin.ClearLine()
					if !summaryPrinted {
						fmt.Print(ui.MountsFindingsHeader())
						summaryPrinted = true
					}
					fmt.Print(ui.RenderMountsFindingLive(f, findingCount+1))
					findingCount++
				}
			},
		},
	)

	spin.Stop()

	if err != nil && !errors.Is(err, ai.ErrIncomplete) {
		_, ferr := fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if ferr != nil {
			return ferr
		}
		return err
	}
	if errors.Is(err, ai.ErrIncomplete) {
		_, _ = fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	if !summaryPrinted {
		fmt.Print(ui.MountsFindingsHeader())
		fmt.Print(ui.RenderMountsSummaryLive(result.Summary))
	}

	if len(result.Findings) == 0 {
		fmt.Print(ui.MountsCleanResult())
	} else {
		fmt.Print(ui.RenderMountsResultSummary(result))
	}

	fmt.Print(ui.Done())
	return nil
}
