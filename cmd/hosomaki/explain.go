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

	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/prompt"
	"github.com/rivernova/hosomaki/internal/spinner"
	"github.com/rivernova/hosomaki/internal/stream"
	"github.com/rivernova/hosomaki/internal/ui"
	"github.com/spf13/cobra"
)

// explain command implementation

func newExplainCmd() *cobra.Command {
	var (
		service string
		bootStr string
		dmesg   bool
		file    string
		lines   int
		cmd_    string
		debug   bool
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

The --cmd flag provides the originating command as context (set automatically
by the shell integration):
  echo "$out" | hosomaki explain --cmd "docker compose up"`,

		Args: cobra.ArbitraryArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			opts := collector.LogOptions{Lines: lines}
			bootChanged := cmd.Flags().Changed("boot")

			input, err := resolveInput(resolveParams{
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

			ctx := ui.ExplainContext{
				Source: source,
				Cmd:    strings.TrimSpace(cmd_),
				Lines:  lines,
			}

			env := collector.Env()
			p := prompt.Explain(input, cmd_, env)

			return runExplain(ctx, p, debug)
		},
	}

	cmd.Flags().StringVarP(&service, "service", "s", "", "explain recent errors for a systemd service (e.g. nginx, sshd)")
	cmd.Flags().StringVar(&bootStr, "boot", "0", "explain errors from a specific boot (0=current, -1=previous, …)")
	cmd.Flags().BoolVar(&dmesg, "dmesg", false, "explain kernel errors and warnings from dmesg")
	cmd.Flags().StringVarP(&file, "file", "f", "", "explain errors from a log file")
	cmd.Flags().IntVarP(&lines, "lines", "n", 0, "number of log lines to read (default varies by source)")
	cmd.Flags().StringVar(&cmd_, "cmd", "", "the command that produced this output (set automatically by shell integration)")
	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr before rendering")
	cmd.Flags().Lookup("boot").NoOptDefVal = "0"

	return cmd
}

func runExplain(ctx ui.ExplainContext, p string, debug bool) error {
	fmt.Print(ui.ExplainHeader())
	fmt.Print(ui.ExplainContextSection(ctx))

	var (
		entries     []prompt.ExplainEntry
		spinStopped bool
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
		entries = append(entries, entry)

		if !spinStopped {
			spin.Stop()
			spinStopped = true
		}

		fmt.Print(ui.RenderExplainEntryLive(entry, len(entries), true))
	})

	_, err := provider.GenerateStream(context.Background(), p,
		func() { spin.SetLabel("responding…") },
		sc,
	)
	spin.Stop()

	if debug {
		fmt.Fprintf(os.Stderr, "\n--- raw model response ---\n%s\n--- end ---\n\n", sc.Raw())
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	if len(entries) == 0 {
		var result prompt.ExplainResult
		if parseErr := ui.ParseExplainJSON(sc.Raw(), &result); parseErr == nil && len(result.Issues) > 0 {
			for i, entry := range result.Issues {
				fmt.Print(ui.RenderExplainEntryLive(entry, i+1, len(result.Issues) > 1))
			}
		} else {
			fmt.Print(ui.Section("what is happening", "(no information)"))
			fmt.Print(ui.Section("why it is happening", "(no information)"))
		}
	}

	fmt.Print(ui.Done())
	return nil
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
