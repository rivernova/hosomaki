// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rivernova/hosomaki/internal/ai"
	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/prompt"
	"github.com/rivernova/hosomaki/internal/sanitiser"
	"github.com/rivernova/hosomaki/internal/spinner"
	"github.com/rivernova/hosomaki/internal/store"
	"github.com/rivernova/hosomaki/internal/ui"
	"github.com/spf13/cobra"
)

// why command logic

const (
	exitCodeMin = 1
	exitCodeMax = 255
)

func newWhyCmd() *cobra.Command {
	var (
		service string
		lines   int
		since   string
		debug   bool
	)

	cmd := &cobra.Command{
		Use:   "why <exit-code>",
		Short: "Explain why a service exited with a nonzero exit code",
		Long: `Pulls surrounding journal context for a systemd service and explains
the failure chain — what happened, why it failed, and what to do about it.

The exit code is a required positional argument. --service is a required flag.
hosomaki why never modifies the system; it is strictly read-only.

Exit codes outside 1–255 and exit code 0 are rejected before any journal
collection takes place.

Examples:
  hosomaki why 1   --service nginx
  hosomaki why 137 --service myapp --lines 100
  hosomaki why 127 --service postgresql --since "10 min ago"`,

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			code, err := parseExitCode(args[0])
			if err != nil {
				return err
			}

			service = strings.TrimSpace(service)

			rawLogs, err := collector.WhyLogs(service, collector.LogOptions{
				Lines: lines,
				Since: since,
			})
			if err != nil {
				return fmt.Errorf("why: %w", err)
			}

			sanitised := sanitiser.Default().Sanitise(rawLogs)

			env := collector.Env()
			generationPrompt := prompt.Why(prompt.WhyInput{
				Service:     service,
				ExitCode:    code,
				Logs:        sanitised,
				Environment: env,
			})

			fmt.Print(ui.WhyHeader())
			fmt.Print(ui.WhyContextSection(ui.WhyContext{
				Service:  service,
				ExitCode: code,
				Lines:    lines,
				Since:    since,
			}))

			return runWhy(generationPrompt, debug)
		},
	}

	cmd.Flags().StringVarP(&service, "service", "s", "",
		"systemd service to pull context from (required)")
	cmd.Flags().IntVarP(&lines, "lines", "n", 0,
		"number of journal lines to collect (default 50)")
	cmd.Flags().StringVar(&since, "since", "",
		`collect logs since this time (journalctl format, e.g. "10 min ago")`)
	cmd.Flags().BoolVar(&debug, "debug", false,
		"print raw model response to stderr")

	if err := cmd.MarkFlagRequired("service"); err != nil {
		panic(fmt.Sprintf("why: MarkFlagRequired: %v", err))
	}

	return cmd
}

func parseExitCode(raw string) (int, error) {
	code, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf(
			"invalid exit code %q: must be an integer between %d and %d",
			raw, exitCodeMin, exitCodeMax,
		)
	}
	if code == 0 {
		return 0, fmt.Errorf("exit code 0 means success — there is nothing to explain")
	}
	if code < exitCodeMin || code > exitCodeMax {
		return 0, fmt.Errorf(
			"exit code %d is out of range: valid exit codes are %d–%d",
			code, exitCodeMin, exitCodeMax,
		)
	}
	return code, nil
}

func whyPipeline() ai.Pipeline[prompt.WhyResult] {
	return ai.NewPipeline(
		provider,
		ai.NewSchema(prompt.SchemaWhy),
		ai.StructValidator[prompt.WhyResult]{
			SemanticCheck: validateWhyResult,
		},
	)
}

func validateWhyResult(r prompt.WhyResult) []string {
	var errs []string

	if strings.TrimSpace(r.Summary) == "" {
		errs = append(errs, "summary must not be empty")
	}
	if len(r.Chain) == 0 {
		errs = append(errs, "chain must contain at least one step")
	}
	for i, step := range r.Chain {
		if strings.TrimSpace(step.Event) == "" {
			errs = append(errs, fmt.Sprintf("chain[%d].event must not be empty", i))
		}
		if strings.TrimSpace(step.Detail) == "" {
			errs = append(errs, fmt.Sprintf("chain[%d].detail must not be empty", i))
		}
	}
	if len(r.NextSteps) == 0 {
		errs = append(errs, "next_steps must contain at least one remediation step")
	}

	return errs
}

func runWhy(generationPrompt string, debug bool) error {
	spin := spinner.Start("thinking…")

	pipe := whyPipeline()
	if debug {
		pipe = pipe.WithDebug(os.Stderr)
	}

	result, err := pipe.Run(
		context.Background(),
		generationPrompt,
		ai.RunOptions{
			OnFirstToken:  func() { spin.SetLabel("responding…") },
			OnRepairStart: func(n int) { spin.SetLabel(fmt.Sprintf("repairing (attempt %d)…", n)) },
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

	renderWhyResult(result, debug)
	return nil
}

func renderWhyResult(result prompt.WhyResult, debug bool) {
	fmt.Print(ui.WhySummaryHeader())
	fmt.Print(ui.RenderWhySummaryLive(result.Summary))

	if len(result.Chain) > 0 {
		fmt.Print(ui.WhyChainHeader())
		for i, step := range result.Chain {
			fmt.Print(ui.RenderWhyStepLive(step, i+1))
		}
	}

	if len(result.NextSteps) > 0 {
		fmt.Print(ui.WhyNextStepsHeader())
		for i, step := range result.NextSteps {
			fmt.Print(ui.RenderWhyNextStepLive(step, i+1))
		}
	}

	fmt.Print(ui.RenderWhySummary(result))
	if err := store.Record("why", result); err != nil && debug {
		_, _ = fmt.Fprintf(os.Stderr, "history: record why: %v\n", err)
	}
	fmt.Print(ui.Done())
}
