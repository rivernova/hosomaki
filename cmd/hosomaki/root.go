// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"fmt"
	"os"

	"github.com/rivernova/hosomaki/internal/ai"
	"github.com/rivernova/hosomaki/internal/ai/ollama"
	"github.com/rivernova/hosomaki/internal/config"
	"github.com/spf13/cobra"
)

// this file contains the root command and global state

var (
	cfgFile string
	version string
)

var provider ai.Provider

func Execute(v string) {
	version = v
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "hosomaki",
	Short: "Local intelligence layer for Linux",
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

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/hosomaki/config.yaml)")

	rootCmd.AddCommand(
		newExplainCmd(),
		newStatusCmd(),
		newShellIntegrationCmd(),
	)
}
