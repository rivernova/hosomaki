// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"context"
	"fmt"
	"os"
	"unicode"

	"github.com/rivernova/hosomaki/internal/ai"
	"github.com/rivernova/hosomaki/internal/ai/ollama"
	"github.com/rivernova/hosomaki/internal/config"
	"github.com/spf13/cobra"
)

// root command and global state

var (
	cfgFile string
	version string
)

var provider ai.Provider

func Execute(v string) {
	version = v
	rootCmd.Version = version

	os.Args = normaliseNegativeIntFlag(os.Args, "--boot")
	os.Args = normaliseNegativeIntFlag(os.Args, "--diff")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func normaliseNegativeIntFlag(args []string, flag string) []string {
	out := make([]string, 0, len(args))
	skip := false
	for i, arg := range args {
		if skip {
			skip = false
			continue
		}
		if arg == flag && i+1 < len(args) && isNegativeInt(args[i+1]) {
			out = append(out, flag+"="+args[i+1])
			skip = true
			continue
		}
		out = append(out, arg)
	}
	return out
}

func isNegativeInt(s string) bool {
	if len(s) < 2 || s[0] != '-' {
		return false
	}
	for _, r := range s[1:] {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

type healthChecker interface {
	Ping(ctx context.Context) error
}

var rootCmd = &cobra.Command{
	Use:          "hosomaki",
	SilenceUsage: true,
	Short:        "Local intelligence layer for Linux",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Init(cfgFile)
		if err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}

		switch cfg.AI.Provider {
		case "ollama", "":
			provider = ollama.New(cfg.AI.Endpoint, cfg.AI.Model, cfg.AI.Timeout)
		default:
			return fmt.Errorf("unknown AI provider %q — supported: ollama", cfg.AI.Provider)
		}

		if hc, ok := provider.(healthChecker); ok {
			if err := hc.Ping(cmd.Context()); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/hosomaki/config.yaml)")

	rootCmd.AddCommand(
		newAuditCmd(),
		newWatchCmd(),
		newDoctorCmd(),
		newExplainCmd(),
		newStatusCmd(),
		newShellIntegrationCmd(),
	)
}
