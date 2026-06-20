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
	"time"

	"github.com/rivernova/hosomaki/internal/ai"
	"github.com/rivernova/hosomaki/internal/auditor"
	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/prompt"
	"github.com/rivernova/hosomaki/internal/sanitiser"
	"github.com/rivernova/hosomaki/internal/spinner"
	"github.com/rivernova/hosomaki/internal/store"
	"github.com/rivernova/hosomaki/internal/ui"
	"github.com/spf13/cobra"
)

// audit command logic

func newAuditCmd() *cobra.Command {
	var (
		initBaseline bool
		baselinePath string
		dirs         string
		debug        bool
	)

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Surface system changes since the last baseline snapshot",
		Long: `Compares the current system state against a stored baseline to surface
changes: files modified, services added or removed, permission changes,
new or removed listening ports, package updates, and user account changes.

On the first run, create a baseline with --init:
  hosomaki audit --init

Subsequent runs diff the current system against that baseline:
  hosomaki audit

To reset the baseline to the current state:
  hosomaki audit --init

The baseline is stored at ~/.local/share/hosomaki/audit-baseline.json
(respects $XDG_DATA_HOME). A custom path can be set with --baseline.

hosomaki audit never modifies the system. It is read-only.`,

		Args: cobra.NoArgs,

		RunE: func(cmd *cobra.Command, _ []string) error {
			bPath, err := resolveBaselinePath(baselinePath)
			if err != nil {
				return err
			}

			watchDirs := parseDirs(dirs)
			env := collector.Env()

			if initBaseline {
				return runAuditInit(bPath, watchDirs, env)
			}
			return runAuditDiff(cmd.Context(), bPath, watchDirs, env, debug)
		},
	}

	cmd.Flags().BoolVar(&initBaseline, "init", false, "create (or reset) the baseline snapshot")
	cmd.Flags().StringVar(&baselinePath, "baseline", "", "path to the baseline file (default: ~/.local/share/hosomaki/audit-baseline.json)")
	cmd.Flags().StringVar(&dirs, "dirs", "", "comma-separated directories to track for file/permission changes (default: /etc,/usr/local/bin,/usr/local/sbin)")
	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr")

	return cmd
}

func runAuditInit(path string, dirs []string, env collector.Environment) error {
	fmt.Print(ui.AuditInitHeader())

	spin := spinner.Start("collecting baseline…")
	b := auditor.Collect(auditor.CollectOptions{
		WatchDirs:   dirs,
		Environment: env,
	})
	spin.Stop()

	if err := auditor.Save(path, b); err != nil {
		return fmt.Errorf("audit init: %w", err)
	}

	fmt.Print(ui.AuditBaselineSection(b, path))
	fmt.Print(ui.Done())
	return nil
}

func runAuditDiff(ctx context.Context, path string, dirs []string, env collector.Environment, debug bool) error {
	fmt.Print(ui.AuditHeader())

	baseline, err := auditor.Load(path)
	if err != nil {
		if errors.Is(err, auditor.ErrNoBaseline) {
			return fmt.Errorf("%w", err)
		}
		return fmt.Errorf("audit: load baseline: %w", err)
	}

	spin := spinner.Start("collecting current state…")
	current := auditor.Collect(auditor.CollectOptions{
		WatchDirs:   dirs,
		Environment: env,
	})
	spin.Stop()

	diff := auditor.Diff(baseline, current)

	for _, cerr := range current.CollectionErrors {
		_, _ = fmt.Fprintf(os.Stderr, "warning: %s\n", cerr)
	}

	age := humanDuration(diff.BaselineAge)

	fmt.Print(ui.AuditDiffSection(diff, age))

	if diff.IsEmpty() {
		fmt.Print(ui.AuditNoChanges(age))
		fmt.Print(ui.Done())
		return nil
	}

	fmt.Print(ui.AuditLocalChanges(diff))

	san := sanitiser.Default()
	sanitised := sanitiseDiff(diff, san)

	return runAuditAI(ctx, sanitised, age, env, debug)
}

func sanitiseDiff(d *auditor.AuditDiff, san *sanitiser.Sanitiser) *auditor.AuditDiff {
	s := func(v string) string { return san.Sanitise(v) }
	sl := func(ss []string) []string {
		if len(ss) == 0 {
			return ss
		}
		out := make([]string, len(ss))
		for i, v := range ss {
			out[i] = s(v)
		}
		return out
	}

	out := &auditor.AuditDiff{
		BaselineAge: d.BaselineAge,

		ServicesAdded:   sl(d.ServicesAdded),
		ServicesRemoved: sl(d.ServicesRemoved),

		FilesAdded:   sl(d.FilesAdded),
		FilesRemoved: sl(d.FilesRemoved),

		PackagesAdded:   sl(d.PackagesAdded),
		PackagesRemoved: sl(d.PackagesRemoved),

		PortsOpened: sl(d.PortsOpened),
		PortsClosed: sl(d.PortsClosed),

		UsersAdded:   sl(d.UsersAdded),
		UsersRemoved: sl(d.UsersRemoved),
	}

	out.FilesModified = make([]auditor.FileChange, len(d.FilesModified))
	for i, fc := range d.FilesModified {
		out.FilesModified[i] = auditor.FileChange{
			Path:     s(fc.Path),
			OldMtime: fc.OldMtime,
			NewMtime: fc.NewMtime,
			OldSize:  fc.OldSize,
			NewSize:  fc.NewSize,
		}
	}

	out.PermissionsChanged = make([]auditor.PermChange, len(d.PermissionsChanged))
	for i, pc := range d.PermissionsChanged {
		out.PermissionsChanged[i] = auditor.PermChange{
			Path:     s(pc.Path),
			OldMode:  pc.OldMode,
			NewMode:  pc.NewMode,
			OldOwner: s(pc.OldOwner),
			NewOwner: s(pc.NewOwner),
			OldGroup: s(pc.OldGroup),
			NewGroup: s(pc.NewGroup),
		}
	}

	out.PackagesUpdated = make([]auditor.PackageChange, len(d.PackagesUpdated))
	for i, pu := range d.PackagesUpdated {
		out.PackagesUpdated[i] = auditor.PackageChange{
			Name:       s(pu.Name),
			OldVersion: s(pu.OldVersion),
			NewVersion: s(pu.NewVersion),
		}
	}

	return out
}

func runAuditAI(ctx context.Context, diff *auditor.AuditDiff, age string, env collector.Environment, debug bool) error {
	p := prompt.Audit(prompt.AuditInput{
		Environment: env,
		Diff:        diff,
		BaselineAge: age,
	})

	spin := spinner.Start("thinking…")

	pipe := auditStreamPipeline()
	if debug {
		pipe = pipe.WithDebug(os.Stderr)
	}

	findingCount := 0
	summaryPrinted := false
	wasRepaired := false

	result, err := pipe.Run(ctx, p, ai.StreamOptions{
		OnFirstToken: func() { spin.SetLabel("responding…") },
		OnRepairStart: func(n int) {
			wasRepaired = true
			spin.SetLabel(fmt.Sprintf("repairing (attempt %d)…", n))
		},
		OnItem: func(key, raw string) {
			switch key {
			case "summary":
				var str string
				if jsonErr := json.Unmarshal([]byte(raw), &str); jsonErr != nil {
					return
				}
				str = strings.TrimSpace(str)
				if str == "" {
					return
				}
				spin.ClearLine()
				if !summaryPrinted {
					fmt.Print(ui.AuditFindingsHeader())
					summaryPrinted = true
				}
				fmt.Print(ui.RenderAuditSummaryLive(str))

			case "findings":
				var f prompt.AuditFinding
				if jsonErr := json.Unmarshal([]byte(raw), &f); jsonErr != nil {
					return
				}
				spin.ClearLine()
				if !summaryPrinted {
					fmt.Print(ui.AuditFindingsHeader())
					summaryPrinted = true
				}
				fmt.Print(ui.RenderAuditFindingLive(f, findingCount+1))
				findingCount++
			}
		},
	})

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
		fmt.Print(ui.RenderAuditSummaryLive(result.Summary))
		for i, f := range result.Findings {
			fmt.Print(ui.RenderAuditFindingLive(f, i+1))
		}
	} else if findingCount == 0 && !summaryPrinted {
		fmt.Print(ui.AuditFindingsHeader())
		fmt.Print(ui.RenderAuditSummaryLive(result.Summary))
	}

	fmt.Print(ui.RenderAuditResultSummary(result))
	if err := store.Record("audit", result); err != nil && debug {
		_, _ = fmt.Fprintf(os.Stderr, "history: record audit: %v\n", err)
	}
	fmt.Print(ui.Done())
	return nil
}

func auditStreamPipeline() ai.StreamPipeline[prompt.AuditResult] {
	return ai.NewStreamPipeline(
		provider,
		ai.NewSchema(prompt.SchemaAudit),
		ai.StructValidator[prompt.AuditResult]{},
	)
}

func resolveBaselinePath(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	path, err := auditor.DefaultPath()
	if err != nil {
		return "", fmt.Errorf("audit: resolve baseline path: %w", err)
	}
	return path, nil
}

func parseDirs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func humanDuration(d time.Duration) string {
	d = d.Truncate(time.Minute)
	if d <= 0 {
		return "just now"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 && days == 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if len(parts) == 0 {
		return "< 1m"
	}
	return strings.Join(parts, " ")
}
