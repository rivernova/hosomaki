// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// this file contains the implementation of the "shell-integration" command

var shellSnippets = map[string]snippet{
	"bash": {
		rcFile: "~/.bashrc",
		code: `# hosomaki shell integration
# Wraps a command so that on failure its output is explained automatically.
# Usage: explain <command> [args...]
explain() {
  local out
  out=$("$@" 2>&1)
  local code=$?
  if [ $code -ne 0 ]; then
    echo "$out"
    echo
    echo "$out" | hosomaki explain
  else
    echo "$out"
  fi
  return $code
}`,
	},
	"zsh": {
		rcFile: "~/.zshrc",
		code: `# hosomaki shell integration
# Wraps a command so that on failure its output is explained automatically.
# Usage: explain <command> [args...]
explain() {
  local out
  out=$("$@" 2>&1)
  local code=$?
  if [[ $code -ne 0 ]]; then
    echo "$out"
    echo
    echo "$out" | hosomaki explain
  else
    echo "$out"
  fi
  return $code
}`,
	},
	"fish": {
		rcFile: "~/.config/fish/config.fish",
		code: `# hosomaki shell integration
# Wraps a command so that on failure its output is explained automatically.
# Usage: explain <command> [args...]
function explain
  set out (command $argv 2>&1)
  set code $status
  if test $code -ne 0
    echo $out
    echo
    echo $out | hosomaki explain
  else
    echo $out
  end
  return $code
end`,
	},
}

type snippet struct {
	rcFile string
	code   string
}

func newShellIntegrationCmd() *cobra.Command {
	var shell string

	cmd := &cobra.Command{
		Use:   "shell-integration",
		Short: "Print shell integration snippet for automatic error explanation",
		Long: `Prints a shell function that wraps commands and explains failures automatically.

Supported shells: bash, zsh, fish

Add the snippet to your shell config:
  hosomaki shell-integration --shell bash >> ~/.bashrc && source ~/.bashrc
  hosomaki shell-integration --shell zsh  >> ~/.zshrc  && source ~/.zshrc
  hosomaki shell-integration --shell fish >> ~/.config/fish/config.fish

Then use it:
  explain sudo systemctl start nginx
  explain make build
  explain docker compose up`,

		Args: cobra.NoArgs,

		RunE: func(cmd *cobra.Command, args []string) error {
			if shell == "" {
				shell = detectShell()
			}

			s, ok := shellSnippets[shell]
			if !ok {
				return fmt.Errorf(
					"unsupported shell %q — supported: bash, zsh, fish\n"+
						"Detect your shell with: echo $SHELL", shell,
				)
			}

			fmt.Println(s.code)

			fmt.Fprintf(os.Stderr, "\n# Add to %s with:\n", s.rcFile)
			fmt.Fprintf(os.Stderr, "#   hosomaki shell-integration --shell %s >> %s\n", shell, s.rcFile)

			return nil
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "", "target shell: bash, zsh, or fish (default: auto-detect)")
	return cmd
}

// returns the name of the current shell by inspecting $SHELL
func detectShell() string {
	base := filepath.Base(os.Getenv("SHELL"))
	if _, ok := shellSnippets[base]; ok {
		return base
	}
	return "bash"
}
