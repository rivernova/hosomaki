// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/insight"
	"github.com/rivernova/hosomaki/internal/output"
	"github.com/rivernova/hosomaki/internal/present"
	"github.com/rivernova/hosomaki/internal/prompt"
	"github.com/rivernova/hosomaki/internal/render"
	"github.com/rivernova/hosomaki/internal/spinner"
	"github.com/spf13/cobra"
)

// this file contains the "explain" command

func newExplainCmd() *cobra.Command {
	var (
		service   string
		bootStr   string
		dmesg     bool
		file      string
		lines     int
		cmd_      string
		outputFmt string
	)

	cmd := &cobra.Command{
		Use:   "explain [message]",
		Short: "Explain log output or an error message in plain language",
		Long: `Explain analyses log output and tells you what happened and what to do.

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

			params := resolveParams{
				args:        args,
				service:     service,
				boot:        bootStr,
				bootChanged: bootChanged,
				dmesg:       dmesg,
				file:        file,
				opts:        opts,
			}

			inputText, err := resolveInput(params)
			if err != nil {
				return err
			}

			inputInfo := buildInputInfo(params, inputText)

			env := collector.Env()
			p := prompt.Explain(inputText, cmd_, env, appCfg.Output.Language)

			if outputFmt == "json" {
				return explainJSON(inputText, cmd_, p)
			}

			initialRep := render.ExplainReport{
				Title:     "hosomaki explain",
				InputInfo: inputInfo,
				Context:   present.ContextLine(cmd_),
			}
			processLines := []string{
				"analizing behavior…",
				"correlating logs…",
				"detecting patterns…",
			}
			_ = currentUI().RenderExplainStream(initialRep, processLines)

			var aiBuf bytes.Buffer
			spin := spinner.Start("thinking…")
			_, genErr := provider.GenerateStream(context.Background(), p, func() {
				spin.Stop()
			}, &aiBuf)
			spin.Stop()

			rawAI := strings.TrimSpace(aiBuf.String())

			doc := insight.ParseDoctor(rawAI)
			if genErr != nil && doc.Raw == "" && len(doc.Issues) == 0 {
				doc.Raw = "AI analysis unavailable: " + genErr.Error()
			}

			finalRep := present.ExplainReportFromIssues(inputInfo, cmd_, doc.Issues, doc.Raw)
			currentUI().FinaliseExplain(finalRep)

			return nil
		},
	}

	cmd.Flags().StringVarP(&service, "service", "s", "", "explain recent errors for a systemd service (e.g. nginx, sshd)")
	cmd.Flags().StringVar(&bootStr, "boot", "0", "explain errors from a specific boot (0=current, -1=previous, …)")
	cmd.Flags().BoolVar(&dmesg, "dmesg", false, "explain kernel errors and warnings from dmesg")
	cmd.Flags().StringVarP(&file, "file", "f", "", "explain errors from a log file")
	cmd.Flags().IntVarP(&lines, "lines", "n", 0, "number of log lines to read (default varies by source)")
	cmd.Flags().StringVar(&cmd_, "cmd", "", "the command that produced this output (set automatically by shell integration)")
	cmd.Flags().StringVar(&outputFmt, "output", "", "output format: json")
	cmd.Flags().Lookup("boot").NoOptDefVal = "0"

	return cmd
}

func buildInputInfo(p resolveParams, inputText string) render.InputInfo {
	switch {
	case p.service != "":
		return render.InputInfo{Origin: "service", Detail: p.service, Lines: countLines(inputText)}
	case p.bootChanged:
		return render.InputInfo{Origin: "boot", Detail: "boot " + p.boot, Lines: countLines(inputText)}
	case p.dmesg:
		return render.InputInfo{Origin: "dmesg", Lines: countLines(inputText)}
	case p.file != "":
		return render.InputInfo{Origin: "file", Detail: p.file, Lines: countLines(inputText)}
	case len(p.args) > 0:
		return render.InputInfo{Origin: "text", Lines: countLines(inputText)}
	default:
		return render.InputInfo{Origin: "pipe", Lines: countLines(inputText)}
	}
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Split(strings.TrimRight(s, "\n"), "\n"))
}

func explainJSON(input, command, p string) error {
	var buf bytes.Buffer
	spin := spinner.Start("thinking…")
	_, err := provider.GenerateStream(context.Background(), p, func() { spin.Stop() }, &buf)
	spin.Stop()

	rawText := strings.TrimSpace(buf.String())
	if err != nil && rawText == "" {
		rawText = "AI explanation unavailable: " + err.Error()
	}

	var out bytes.Buffer
	if encErr := output.WriteExplain(&out, input, command, rawText); encErr != nil {
		return fmt.Errorf("encoding JSON: %w", encErr)
	}
	_, writeErr := os.Stdout.Write(out.Bytes())
	return writeErr
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
