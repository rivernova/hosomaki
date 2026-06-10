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

			// Build the AI pipeline once. StreamPipeline is a value type;
			// WithDebug returns a new value, so we resolve the final pipeline
			// here rather than inside the flush closure to avoid repeated
			// re-wrapping on every batch.
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
				// Sanitise is applied per line at ingest time, before buffering.
				// By the time OnFlush is called the batch is already clean.
				Sanitise: sanitiser.Default().Sanitise,
				OnFlush:  makeFlushFunc(service, env, pipe),
				// OnLine is intentionally nil — raw lines are not echoed to the
				// terminal. The operator sees AI explanations only, keeping the
				// output readable during noisy log bursts.
				OnLine: nil,
			}

			w, err := watcher.New(cfg)
			if err != nil {
				return fmt.Errorf("watch: %w", err)
			}

			fmt.Print(ui.WatchHeader(service))
			fmt.Print(ui.WatchReadyLine(service, seedLines))

			// Install signal handler so Ctrl-C cancels the context cleanly.
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			runErr := w.Run(ctx)

			fmt.Print(ui.WatchShutdownLine())

			// context.Canceled is a clean shutdown (Ctrl-C), not an error to surface.
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

// makeFlushFunc returns the OnFlush callback used by the watcher.
// It receives a pre-built, fully configured pipeline so the closure
// has no mutable state — the same pipeline value is reused for every batch.
//
// The batch arriving here is already sanitised (sanitiser.Default() is
// applied per line at ingest time). We do not re-sanitise.
//
// AI errors within a batch are non-fatal: they are logged to stderr and
// the watcher continues. Killing the tail loop because one batch failed
// would be worse than skipping a single explanation.
func makeFlushFunc(service string, env collector.Environment, pipe ai.StreamPipeline[prompt.WatchResult]) watcher.FlushFunc {
	return func(ctx context.Context, batch string) error {
		p := prompt.Watch(prompt.WatchInput{
			Service:     service,
			Batch:       batch,
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
				var issue prompt.WatchIssue
				if jsonErr := json.Unmarshal([]byte(raw), &issue); jsonErr != nil {
					return
				}
				if strings.TrimSpace(issue.What) == "" && strings.TrimSpace(issue.Why) == "" {
					return
				}
				spin.ClearLine()
				if !headerPrinted {
					fmt.Print(ui.WatchBatchHeader(batchTime))
					headerPrinted = true
				}
				fmt.Print(ui.RenderWatchIssueLive(issue, issueCount+1, issueCount > 0))
				issueCount++
			},
		})

		spin.Stop()

		if err != nil {
			// Non-fatal: log and continue tailing.
			_, _ = fmt.Fprintf(os.Stderr, "watch: analysis error: %v\n", err)
			return nil
		}

		// If repair happened the streamed output may be incomplete;
		// re-render from the fully validated result.
		if wasRepaired && len(result.Issues) > 0 {
			if !headerPrinted {
				fmt.Print(ui.WatchBatchHeader(batchTime))
			}
			for i, issue := range result.Issues {
				fmt.Print(ui.RenderWatchIssueLive(issue, i+1, len(result.Issues) > 1))
			}
		}

		// Empty result: the model found no issues. Stay quiet — same behaviour
		// as explain when it returns {"issues":[]}.

		return nil
	}
}

// watchStreamPipeline constructs the streaming AI pipeline for the watch command.
// WatchResult is structurally identical to ExplainResult; they share the same
// schema and validator path.
func watchStreamPipeline() ai.StreamPipeline[prompt.WatchResult] {
	return ai.NewStreamPipeline(
		provider,
		ai.NewSchema(prompt.SchemaWatch),
		ai.StructValidator[prompt.WatchResult]{},
	)
}
