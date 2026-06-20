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
	"time"

	"github.com/rivernova/hosomaki/internal/ai"
	"github.com/rivernova/hosomaki/internal/collector"
	"github.com/rivernova/hosomaki/internal/historian"
	"github.com/rivernova/hosomaki/internal/prompt"
	"github.com/rivernova/hosomaki/internal/sanitiser"
	"github.com/rivernova/hosomaki/internal/spinner"
	"github.com/rivernova/hosomaki/internal/ui"
	"github.com/spf13/cobra"
)

func newHistoryCmd() *cobra.Command {
	var (
		limit    int
		since    string
		filterCmd string
		clear    bool
		debug    bool
	)

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Review past diagnostic results",
		Long: `Shows a log of past results from explain, why, audit, status, and doctor
commands so you can revisit previous insights without re-running the model.

Results are stored automatically in ~/.local/share/hosomaki/history.json.

Examples:
  hosomaki history
  hosomaki history --command explain
  hosomaki history --since 7d
  hosomaki history --clear`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			env := collector.Env()

			path, err := historian.DefaultPath()
			if err != nil {
				return err
			}

			// --clear just deletes the log
			if clear {
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("history: %w", err)
				}
				fmt.Print(ui.HistoryCleared())
				fmt.Print(ui.Done())
				return nil
			}

			log, err := historian.Load(path)
			if err != nil {
				fmt.Print(ui.HistoryHeader())
				fmt.Print(ui.HistoryNoHistory())
				fmt.Print(ui.Done())
				return nil
			}

			// Filter
			entries := log.Entries
			if filterCmd != "" {
				var filtered []historian.HistoryEntry
				for _, e := range entries {
					if e.Command == filterCmd {
						filtered = append(filtered, e)
					}
				}
				entries = filtered
			}
			if since != "" {
				dur, err := time.ParseDuration(since)
				if err != nil {
					d, err2 := time.ParseDuration(since + "h")
					if err2 != nil {
						return fmt.Errorf("history: invalid duration %q (try 24h, 7d, 30m)", since)
					}
					dur = d
				}
				cutoff := time.Now().Add(-dur)
				var filtered []historian.HistoryEntry
				for _, e := range entries {
					if e.Timestamp.After(cutoff) {
						filtered = append(filtered, e)
					}
				}
				entries = filtered
			}
			if limit > 0 && len(entries) > limit {
				entries = entries[len(entries)-limit:]
			}

			fmt.Print(ui.HistoryHeader())

			if len(entries) == 0 {
				msg := "No matching history entries found"
				fmt.Print(ui.HistoryNoMatching(msg))
				fmt.Print(ui.Done())
				return nil
			}

			fmt.Print(ui.HistoryEntryCount(len(entries)))

			// Build filter description
			filterDesc := buildFilterDesc(filterCmd, since, limit)

			// Sanitise before prompt
			san := sanitiser.Default()
			sanitised := san.Sanitise(formatHistoryForPrompt(entries))

			p := prompt.History(prompt.HistoryInput{
				Environment: env,
				History:     sanitised,
				FilterDesc:  filterDesc,
			})

			spin := spinner.Start("thinking\u2026")
			pipe := historyStreamPipeline()
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
								fmt.Print(ui.HistoryFindingsHeader())
								summaryPrinted = true
							}
							fmt.Print(ui.RenderHistorySummaryLive(s))

						case "entries":
							var e prompt.HistoryEntry
							if jsonErr := json.Unmarshal([]byte(raw), &e); jsonErr != nil {
								return
							}
							spin.ClearLine()
							if !summaryPrinted {
								fmt.Print(ui.HistoryFindingsHeader())
								summaryPrinted = true
							}
							fmt.Print(ui.RenderHistoryEntryLive(e, 0))
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
				fmt.Print(ui.HistoryFindingsHeader())
				fmt.Print(ui.RenderHistorySummaryLive(result.Summary))
				for _, e := range result.Entries {
					fmt.Print(ui.RenderHistoryEntryLive(e, 0))
				}
			}

			if len(result.Entries) == 0 {
				fmt.Print(ui.HistoryCleanResult())
			} else {
				fmt.Print(ui.RenderHistoryResultSummary(result))
			}
			fmt.Print(ui.Done())
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "show the last N entries (default: 10)")
	cmd.Flags().StringVar(&since, "since", "", "show entries newer than duration (e.g. 24h, 7d)")
	cmd.Flags().StringVar(&filterCmd, "command", "", "filter by source command (explain, why, audit, status, doctor)")
	cmd.Flags().BoolVar(&clear, "clear", false, "clear the history log")
	cmd.Flags().BoolVar(&debug, "debug", false, "print raw model response to stderr")

	return cmd
}

func historyStreamPipeline() ai.StreamPipeline[prompt.HistoryResult] {
	return ai.NewStreamPipeline(
		provider,
		ai.NewSchema(prompt.SchemaHistory),
		ai.StructValidator[prompt.HistoryResult]{
			SemanticCheck: validateHistoryResult,
		},
	)
}

func validateHistoryResult(r prompt.HistoryResult) []string {
	var errs []string
	if strings.TrimSpace(r.Summary) == "" {
		errs = append(errs, "summary must not be empty")
	}
	for i, e := range r.Entries {
		if strings.TrimSpace(e.Timestamp) == "" {
			errs = append(errs, fmt.Sprintf("entries[%d].timestamp must not be empty", i))
		}
		if strings.TrimSpace(e.Command) == "" {
			errs = append(errs, fmt.Sprintf("entries[%d].command must not be empty", i))
		}
	}
	return errs
}

func formatHistoryForPrompt(entries []historian.HistoryEntry) string {
	if len(entries) == 0 {
		return "(no history entries)"
	}
	var b strings.Builder
	for i, e := range entries {
		ts := e.Timestamp.Format(time.RFC3339)
		summary := extractSummary(e)
		_, _ = fmt.Fprintf(&b, "%d. [%s] %s: %s\n", i+1, ts, e.Command, summary)
	}
	return strings.TrimSpace(b.String())
}

func extractSummary(e historian.HistoryEntry) string {
	switch e.Command {
	case "explain":
		var r struct {
			Issues []struct {
				What string `json:"what"`
			} `json:"issues"`
		}
		if err := json.Unmarshal(e.Result, &r); err == nil && len(r.Issues) > 0 {
			return strings.TrimSpace(r.Issues[0].What)
		}
	case "why", "audit":
		var r struct {
			Summary string `json:"summary"`
		}
		if err := json.Unmarshal(e.Result, &r); err == nil && r.Summary != "" {
			return strings.TrimSpace(r.Summary)
		}
	case "status":
		// Try full status result first (Overview)
		var full struct {
			Overview string `json:"overview"`
		}
		if err := json.Unmarshal(e.Result, &full); err == nil && full.Overview != "" {
			return strings.TrimSpace(full.Overview)
		}
		// Fallback to brief (Summary)
		var brief struct {
			Summary string `json:"summary"`
		}
		if err := json.Unmarshal(e.Result, &brief); err == nil && brief.Summary != "" {
			return strings.TrimSpace(brief.Summary)
		}
	case "doctor":
		// Try full doctor result first (Issues/Actions)
		var raw map[string]any
		if err := json.Unmarshal(e.Result, &raw); err == nil {
			if issues, ok := raw["issues"]; ok && issues != nil {
				issuesArr, _ := issues.([]any)
				if len(issuesArr) > 0 {
					if first, ok := issuesArr[0].(map[string]any); ok {
						if title, ok := first["title"].(string); ok && title != "" {
							return strings.TrimSpace(title)
						}
					}
				}
				return "no issues detected"
			}
		}
		// Fallback to brief (Summary)
		var brief struct {
			Summary string `json:"summary"`
		}
		if err := json.Unmarshal(e.Result, &brief); err == nil && brief.Summary != "" {
			return strings.TrimSpace(brief.Summary)
		}
	}
	// Fallback: show raw truncated
	raw := string(e.Result)
	if len(raw) > 80 {
		raw = raw[:80] + "..."
	}
	return raw
}

func buildFilterDesc(cmd, since string, limit int) string {
	var parts []string
	if limit > 0 {
		parts = append(parts, fmt.Sprintf("last %d", limit))
	} else {
		parts = append(parts, "all")
	}
	if cmd != "" {
		parts = append(parts, cmd)
	}
	parts = append(parts, "entries")
	if since != "" {
		parts = append(parts, "from last "+since)
	}
	return strings.Join(parts, " ")
}
