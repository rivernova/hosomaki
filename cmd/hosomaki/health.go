// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check AI provider connectivity",
		Long: `Verifies that the configured AI provider is reachable before running AI commands.

Use this to confirm Ollama is running and responding at the configured endpoint.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if provider == nil {
				return fmt.Errorf("provider not initialized")
			}
			if err := provider.HealthCheck(cmd.Context()); err != nil {
				return err
			}
			cmd.Println("AI provider is healthy")
			return nil
		},
	}
}
