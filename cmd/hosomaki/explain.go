// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
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

// explain command implementation

func newExplainCmd() *cobra.Command {
	var (
		service  string
		bootStr  string
		dmesg    bool
		file     string
		lines    int
		cmd_     string
		debug    bool
		noStream bool
	)

	cmd := &cobra.Command{
		Use:   "explain [message]",
		Short: "Explain log output or an error message in plain language",
		Long: `Explain analyses log output and tells you what happened and why.

Without flags, it reads from stdin (pipe) or accepts a message as an argument:
  journalctl -p err -n 20 | hosomaki explain
  dmesg | tail -50         | hosomaki explain
  hosomaki explain "kernel: OOM killer activated on process nginx"

With flags, it collects the logs for you — no copy-pasting needed:
  hosomaki explain --service nginx
  hosomaki explain --service postgresql --lines 100
  hosomaki explain --boot
  hosomaki explain --boot -1
  hosomaki explain --dmesg
  hosomaki explain --file /var/log/nginx/error.log

All input is sanitised locally before being sent to the LLM — sensitive
identifiers (paths, IPs, hex digests, package versions) are replaced with
placeholders to prevent LLM safety filters from blocking the analysis.

By default explain streams each issue to the terminal as soon as the model
produces it. Pass --no-stream to use the validate/repair pipeline instead,
which is slower but guarantees a structurally valid response or a clear
error message.`,

		Args: cobra.ArbitraryArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			opts := collector.LogOptions{Lines: lines}
			bootChanged := cmd.Flags().Changed("boot")

			rawInput, err := resolveInput(resolveParams{
				args:        args,
				service:     service,
				boot:        bootStr,
				bootChanged: bootChanged,
				dmesg:       dmesg,
				file:        file,
				opts:        opts,
			})
			if err != nil {
				return err
			}

			source := resolveSourceLabel(resolveParams{
				args:        args,
				service:     service,
				boot:        bootStr,
				bootChanged: bootChanged,
				dmesg:       dmesg,
				file:        file,
			})

			explainCtx := ui.ExplainContext{
				Source: source,
				Cmd:    strings.TrimSpace(cmd_),
				Lines:  lines,
			}

			env := collector.Env()
			sanitised := sanitiser.Default().Sanitise(rawInput)
			generationPrompt := prompt.Explain(sanitised, cmd_, env)

			if noStream {
				return runExplainPipeline(explainCtx, generationPrompt, debug)
			}
			return runExplainStream(explainCtx, generationPrompt)
		},
	}

	cmd.Flags().StringVarP(&service, "service", "s", "", "explain recent errors for a systemd service (e.g. nginx, sshd)")
	cmd.Flags().StringVar(&bootStr, "boot", "0", "explain errors from a specific boot (0=current, -1=previous, …)")
	cmd.Flags().BoolVar(&dmesg, "dmesg", false, "explain kernel errors and warnings from dmesg")
	cmd.Flags().StringVarP(&file, "file", "f", "", "explain errors from a log file")
	cmd.Flags().IntVarP(&lines, "lines", "n", 0, "number of log lines to read (default varies by source)")
	cmd.Flags().StringVar(&cmd_, "cmd", "", "the command that produced this output (set automatically by shell integration)")
	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr before rendering")
	cmd.Flags().BoolVar(&noStream, "no-stream", false, "disable live streaming; use validate/repair pipeline instead")
	cmd.Flags().Lookup("boot").NoOptDefVal = "0"

	return cmd
}

func runExplainStream(ctx ui.ExplainContext, p string) error {
	fmt.Print(ui.ExplainHeader())
	fmt.Print(ui.ExplainContextSection(ctx))

	var (
		count         int
		headerPrinted bool
	)

	spin := spinner.Start("thinking…")

	sc := stream.NewArrayItemScanner(func(key, raw string) {
		if key != "issues" {
			return
		}
		var entry prompt.ExplainEntry
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			return
		}
		count++
		if !headerPrinted {
			spin.Stop()
			headerPrinted = true
		}
		fmt.Print(ui.RenderExplainEntryLive(entry, count, true))
	})

	_, err := provider.GenerateStream(
		context.Background(),
		p,
		func() { spin.SetLabel("responding…") },
		sc,
	)
	spin.Stop()

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	if count == 0 {
		fmt.Print(ui.ExplainEmptyResult())
	}

	fmt.Print(ui.Done())
	return nil
}

func runExplainPipeline(ctx ui.ExplainContext, p string, debug bool) error {
	fmt.Print(ui.ExplainHeader())
	fmt.Print(ui.ExplainContextSection(ctx))

	spin := spinner.Start("thinking…")
	pipe := explainPipeline()
	if debug {
		pipe = pipe.WithDebug(os.Stderr)
	}

	result, err := pipe.Run(
		context.Background(),
		p,
		func() { spin.SetLabel("responding…") },
	)
	spin.Stop()

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	if len(result.Issues) == 0 {
		fmt.Print(ui.ExplainEmptyResult())
	} else {
		multi := len(result.Issues) > 1
		for i, entry := range result.Issues {
			fmt.Print(ui.RenderExplainEntryLive(entry, i+1, multi))
		}
	}

	fmt.Print(ui.Done())
	return nil
}

func explainPipeline() ai.Pipeline[prompt.ExplainResult] {
	return ai.NewPipeline(
		provider,
		ai.NewSchema(prompt.SchemaExplain),
		ai.StructValidator[prompt.ExplainResult]{},
	)
}

func resolveSourceLabel(p resolveParams) string {
	switch {
	case p.service != "":
		return "service: " + p.service
	case p.bootChanged:
		index, err := strconv.Atoi(p.boot)
		if err != nil {
			return "boot: " + p.boot
		}
		if index == 0 {
			return "boot: current"
		}
		return fmt.Sprintf("boot: %d", index)
	case p.dmesg:
		return "dmesg"
	case p.file != "":
		return "file: " + p.file
	case len(p.args) > 0:
		return "argument"
	default:
		return "stdin"
	}
}

type resolveParams struct {
	args        []string
	service     string
	boot        string
	bootChanged bool
	dmesg       bool
	file        string
	opts        collector.LogOptions
}

func resolveInput(p resolveParams) (string, error) {
	sources := 0
	if p.service != "" {
		sources++
	}
	if p.bootChanged {
		sources++
	}
	if p.dmesg {
		sources++
	}
	if p.file != "" {
		sources++
	}
	if sources > 1 {
		return "", fmt.Errorf("only one of --service, --boot, --dmesg, --file may be used at a time")
	}

	switch {
	case p.service != "":
		return collector.ServiceLogs(p.service, p.opts)
	case p.bootChanged:
		bootIndex, err := strconv.Atoi(p.boot)
		if err != nil {
			return "", fmt.Errorf("invalid boot index %q: must be an integer", p.boot)
		}
		return collector.BootLogs(bootIndex, p.opts)
	case p.dmesg:
		return collector.DmesgLogs(p.opts)
	case p.file != "":
		return collector.FileLogs(p.file, p.opts)
	case len(p.args) > 0:
		input := strings.TrimSpace(strings.Join(p.args, " "))
		if input == "" {
			return "", fmt.Errorf("message was empty — provide a non-empty message")
		}
		return input, nil
	case isStdinPiped():
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
		input := strings.TrimSpace(string(raw))
		if input == "" {
			return "", fmt.Errorf("stdin was empty — pipe some log output or provide a message as argument")
		}
		return input, nil
	default:
		return "", fmt.Errorf(
			"no input provided\n\n" +
				"  Pipe logs:         journalctl -p err -n 20 | hosomaki explain\n" +
				"  By service:        hosomaki explain --service nginx\n" +
				"  By boot:           hosomaki explain --boot\n" +
				"  Kernel messages:   hosomaki explain --dmesg\n" +
				"  From a file:       hosomaki explain --file /var/log/syslog\n" +
				"  Quick message:     hosomaki explain \"error text here\"",
		)
	}
}

func isStdinPiped() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}
