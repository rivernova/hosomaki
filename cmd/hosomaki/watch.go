// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rivernova/hosomaki/internal/ai"
	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/prompt"
	"github.com/rivernova/hosomaki/internal/sanitiser"
	"github.com/rivernova/hosomaki/internal/spinner"
	"github.com/rivernova/hosomaki/internal/ui"
	"github.com/rivernova/hosomaki/internal/watcher"
	"github.com/spf13/cobra"
)

// watch command logic

func newWatchCmd() *cobra.Command {
	var (
		lines    int
		window   time.Duration
		maxLines int
		debug    bool
	)

	cmd := &cobra.Command{
		Use:   "watch <service>",
		Short: "Tail a service journal and explain new errors as they appear",
		Long: `Tails the systemd journal for a service in real time and explains
new errors and warnings as they appear, using the same sanitisation,
validation, and repair pipeline as the other commands.

The command runs until you press Ctrl-C.

On startup it seeds the view with the last N lines from the journal
(--lines), then enters tail mode. Incoming lines are accumulated into
a batch. When a silence window elapses after an error or warning, or
the batch reaches the maximum line count, the batch is sent to the AI
for explanation.

Only batches containing at least one error or warning are submitted to
the AI. Informational-only batches are discarded silently.

hosomaki watch never modifies the system. It is read-only.`,

		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			service := strings.TrimSpace(args[0])
			if service == "" {
				return fmt.Errorf("service name must not be empty")
			}

			seedLines := lines
			if seedLines < 0 {
				seedLines = 0
			}

			env := collector.Env()

			pipe := watchStreamPipeline()
			if debug {
				pipe = pipe.WithDebug(os.Stderr)
			}

			cfg := watcher.Config{
				Service:   service,
				SeedLines: seedLines,
				Buffer: watcher.BufferConfig{
					SilenceWindow: window,
					MaxLines:      maxLines,
				},

				Sanitise: sanitiser.DefaultPerLine().Sanitise,
				OnFlush:  makeFlushFunc(service, env, pipe),
				OnLine: nil,
			}

			w, err := watcher.New(cfg)
			if err != nil {
				return fmt.Errorf("watch: %w", err)
			}

			fmt.Print(ui.WatchHeader(service))
			fmt.Print(ui.WatchReadyLine(service, seedLines))

			// Install signal handler so Ctrl-C cancels the context cleanly
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			runErr := w.Run(ctx)

			fmt.Print(ui.WatchShutdownLine())

			if ctx.Err() != nil {
				return nil
			}
			return runErr
		},
	}

	cmd.Flags().IntVarP(&lines, "lines", "n", 20, "number of historical lines to seed on startup (0 to disable)")
	cmd.Flags().DurationVar(&window, "window", 3*time.Second, "silence window before flushing a non-full batch to the AI")
	cmd.Flags().IntVar(&maxLines, "max-lines", 50, "maximum batch size before forcing a flush to the AI")
	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr")

	return cmd
}

var batchCollapser = sanitiser.New(sanitiser.CollapseRepeats{})

func makeFlushFunc(service string, env collector.Environment, pipe ai.StreamPipeline[prompt.ExplainResult]) watcher.FlushFunc {
	return func(ctx context.Context, batch string) error {
		collapsed := batchCollapser.Sanitise(batch)

		p := prompt.Watch(prompt.WatchInput{
			Service:     service,
			Batch:       collapsed,
			Environment: env,
		})

		spin := spinner.Start("analysing…")

		issueCount := 0
		headerPrinted := false
		wasRepaired := false
		batchTime := time.Now()

		result, err := pipe.Run(ctx, p, ai.StreamOptions{
			OnFirstToken: func() { spin.SetLabel("responding…") },
			OnRepairStart: func(n int) {
				wasRepaired = true
				spin.SetLabel(fmt.Sprintf("repairing (attempt %d)…", n))
			},
			OnItem: func(key, raw string) {
				if key != "issues" {
					return
				}
				var entry prompt.ExplainEntry
				if jsonErr := json.Unmarshal([]byte(raw), &entry); jsonErr != nil {
					return
				}
				if strings.TrimSpace(entry.What) == "" && strings.TrimSpace(entry.Why) == "" {
					return
				}
				spin.ClearLine()
				if !headerPrinted {
					fmt.Print(ui.WatchBatchHeader(batchTime))
					headerPrinted = true
				}
				fmt.Print(ui.RenderExplainEntryLive(entry, issueCount+1, issueCount > 0))
				issueCount++
			},
		})

		spin.Stop()

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "watch: analysis error: %v\n", err)
			return nil
		}

		if wasRepaired && len(result.Issues) > 0 {
			if !headerPrinted {
				fmt.Print(ui.WatchBatchHeader(batchTime))
			}
			for i, entry := range result.Issues {
				fmt.Print(ui.RenderExplainEntryLive(entry, i+1, len(result.Issues) > 1))
			}
		}

		return nil
	}
}

func watchStreamPipeline() ai.StreamPipeline[prompt.ExplainResult] {
	return ai.NewStreamPipeline(
		provider,
		ai.NewSchema(prompt.SchemaExplain),
		ai.StructValidator[prompt.ExplainResult]{},
	)
}
