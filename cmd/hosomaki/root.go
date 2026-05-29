// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"fmt"
	"os"
	"unicode"

	"github.com/rivernova/hosomaki/internal/ai"
	"github.com/rivernova/hosomaki/internal/ai/ollama"
	"github.com/rivernova/hosomaki/internal/config"
	"github.com/rivernova/hosomaki/internal/render"
	"github.com/spf13/cobra"
)

// this file contains the root command and shared setup logic for hosomaki

var (
	cfgFile string
	version string
)

var (
	provider ai.Provider
	appCfg   config.Config
	ui       *render.Renderer
)

func currentUI() *render.Renderer {
	if ui == nil {
		ui = render.New(os.Stdout)
	}
	return ui
}

func buildRenderer(cfg config.Config, w *os.File) *render.Renderer {
	if !cfg.Output.Color {
		return render.New(w, render.WithColor(false))
	}
	return render.New(w)
}

func Execute(v string) {
	version = v
	rootCmd.Version = version
	os.Args = normaliseNegativeIntFlag(os.Args, "--boot")

	if err := rootCmd.Execute(); err != nil {
		errUI := render.New(os.Stderr)
		if ui != nil {
			errUI = buildRenderer(appCfg, os.Stderr)
		}
		errUI.Error(err)
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

var rootCmd = &cobra.Command{
	Use:           "hosomaki",
	Short:         "Local intelligence layer for Linux",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Init(cfgFile)
		if err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}
		appCfg = cfg

		switch cfg.AI.Provider {
		case "ollama", "":
			provider = ollama.New(cfg.AI.Endpoint, cfg.AI.Model, cfg.AI.Timeout)
		default:
			return fmt.Errorf("unknown AI provider %q — supported: ollama", cfg.AI.Provider)
		}

		ui = buildRenderer(cfg, os.Stdout)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default: ~/.config/hosomaki/config.yaml)")

	rootCmd.AddCommand(
		newDoctorCmd(),
		newExplainCmd(),
		newStatusCmd(),
		newShellIntegrationCmd(),
	)
}
