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

func newFirewallCmd() *cobra.Command {
	var crossCheck, debug bool

	cmd := &cobra.Command{
		Use:   "firewall",
		Short: "Explain active firewall rules and flag security concerns",
		Long: `Reads firewall rules from the active backend (firewalld, ufw,
nftables, or iptables) and explains them in plain language.

With --cross-check, cross-references rules against listening ports.

Read-only. Never modifies firewall rules.`,

		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runFirewall(crossCheck, debug)
		},
	}

	cmd.Flags().BoolVar(&crossCheck, "cross-check", false, "cross-reference rules against currently listening ports")
	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr")
	return cmd
}

func runFirewall(crossCheck, debug bool) error {
	spin := spinner.Start("detecting firewall backend…")
	result := collector.FirewallRules()
	spin.Stop()

	fmt.Print(ui.FirewallHeader())

	if result.Backend == collector.BackendNone {
		fmt.Print(ui.FirewallNoBackend())
		if result.Warning != "" {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", result.Warning)
		}
		fmt.Print(ui.Done())
		return nil
	}

	fmt.Print(ui.FirewallCollectedSection(
		string(result.Backend), len(result.Rules), result.Zones, result.Warning, string(result.ReadStatus),
	))

	switch result.ReadStatus {
	case collector.ReadFailed:
		fmt.Print(ui.FirewallReadFailed(result.Warning))
		fmt.Print(ui.Done())
		return nil
	case collector.ReadEmpty:
		fmt.Print(ui.FirewallNoRules())
		fmt.Print(ui.Done())
		return nil
	}

	san := sanitiser.Default()
	sanitisedRules := san.Sanitise(collector.FormatFirewallForPrompt(result))

	var crossCheckData string
	if crossCheck {
		spin = spinner.Start("reading listening ports…")
		ports, portWarns := collector.Ports()
		spin.Stop()
		crossCheckData = collector.FormatPortsForPrompt(ports)
		for _, w := range portWarns {
			_, _ = fmt.Fprintf(os.Stderr, "warning: %s\n", w)
		}
	}

	generationPrompt := prompt.Firewall(prompt.FirewallInput{
		Environment: collector.Env(),
		Rules:       sanitisedRules,
		CrossCheck:  crossCheckData,
	})

	spin = spinner.Start("thinking…")
	pipe := firewallStreamPipeline()
	if debug {
		pipe = pipe.WithDebug(os.Stderr)
	}

	findingCount := 0
	summaryPrinted := false

	firewallResult, err := pipe.Run(context.Background(), generationPrompt, ai.StreamOptions{
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
					fmt.Print(ui.FirewallFindingsHeader())
					summaryPrinted = true
				}
				fmt.Print(ui.RenderFirewallSummaryLive(s))
			case "findings":
				var f prompt.FirewallFinding
				if jsonErr := json.Unmarshal([]byte(raw), &f); jsonErr != nil {
					return
				}
				spin.ClearLine()
				if !summaryPrinted {
					fmt.Print(ui.FirewallFindingsHeader())
					summaryPrinted = true
				}
				fmt.Print(ui.RenderFirewallFindingLive(f, findingCount+1))
				findingCount++
			}
		},
	})
	spin.Stop()

	if err != nil && !errors.Is(err, ai.ErrIncomplete) {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}
	if errors.Is(err, ai.ErrIncomplete) {
		_, _ = fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	if !summaryPrinted {
		fmt.Print(ui.FirewallFindingsHeader())
		fmt.Print(ui.RenderFirewallSummaryLive(firewallResult.Summary))
	}

	if len(firewallResult.Findings) == 0 {
		fmt.Print(ui.FirewallCleanResult())
	} else {
		fmt.Print(ui.RenderFirewallResultSummary(firewallResult))
	}

	fmt.Print(ui.Done())
	return nil
}

func firewallStreamPipeline() ai.StreamPipeline[prompt.FirewallResult] {
	return ai.NewStreamPipeline(
		provider,
		ai.NewSchema(prompt.SchemaFirewall),
		ai.StructValidator[prompt.FirewallResult]{SemanticCheck: validateFirewallResult},
	).WithElementCheck("findings", ai.ElementCheck(validateFirewallFinding))
}

func validateFirewallResult(r prompt.FirewallResult) []string {
	var errs []string
	if strings.TrimSpace(r.Summary) == "" {
		errs = append(errs, "summary must not be empty")
	}
	for i, f := range r.Findings {
		for _, e := range validateFirewallFinding(f) {
			errs = append(errs, fmt.Sprintf("findings[%d].%s", i, e))
		}
	}
	return errs
}

func validateFirewallFinding(f prompt.FirewallFinding) []string {
	var errs []string
	switch strings.TrimSpace(f.Severity) {
	case "critical", "warning", "info":
	default:
		errs = append(errs, fmt.Sprintf("severity must be critical, warning, or info, got %q", f.Severity))
	}
	if strings.TrimSpace(f.Title) == "" {
		errs = append(errs, "title must not be empty")
	}
	if strings.TrimSpace(f.Detail) == "" {
		errs = append(errs, "detail must not be empty")
	}
	return errs
}
