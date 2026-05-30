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

// this file contains the root command and shared logic

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

func renderHelp(v string) {
	u := currentUI()
	u.Title("hosomaki  " + v)
	u.Blank()
	u.Detail("", "local intelligence layer for Linux — powered by a local AI model via Ollama")

	u.Section("explain")
	u.Blank()
	u.Detail("", "understands what's going on — adapts to whatever you throw at it")
	u.Blank()
	u.Process("journalctl -p err -n 20 | hosomaki explain")
	u.Process("hosomaki explain --service nginx")
	u.Process("hosomaki explain --boot -1")
	u.Process("hosomaki explain --dmesg")
	u.Process("hosomaki explain --file /var/log/nginx/error.log")
	u.Process(`hosomaki explain "kernel: OOM killer activated"`)

	u.Section("status")
	u.Blank()
	u.Detail("", "quick health snapshot — uptime, memory, disk, failed services, recent errors")
	u.Blank()
	u.Process("hosomaki status")
	u.Process("hosomaki status --brief")

	u.Section("doctor")
	u.Blank()
	u.Detail("", "full diagnosis with concrete suggested actions — never modifies the system")
	u.Blank()
	u.Process("hosomaki doctor")
	u.Process("hosomaki doctor --brief")

	u.Section("shell-integration")
	u.Blank()
	u.Detail("", "prefix any command with explain to auto-analyse failures")
	u.Blank()
	u.Process("hosomaki shell-integration --shell bash >> ~/.bashrc && source ~/.bashrc")
	u.Process("hosomaki shell-integration --shell zsh  >> ~/.zshrc  && source ~/.zshrc")
	u.Process("hosomaki shell-integration --shell fish >> ~/.config/fish/config.fish")

	u.Done()
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
	Run: func(cmd *cobra.Command, args []string) {
		renderHelp(version)
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Init(cfgFile)
		if err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}
		appCfg = cfg

		switch cfg.AI.Provider {
		case "ollama", "":
			provider = ollama.New(
				cfg.AI.Endpoint,
				cfg.AI.Model,
				cfg.AI.Timeout,
				cfg.AI.Temperature,
				cfg.AI.NumPredict,
			)
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
