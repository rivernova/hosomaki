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

// this file contains the shell-integration command, which provides users with a convenient way to set up automatic explanation

var shellSnippets = map[string]snippet{
	"bash": {
		rcFile: "~/.bashrc",
		code: `# hosomaki shell integration
explain() {
  local out
  out=$("$@" 2>&1)
  local code=$?
  if [ $code -ne 0 ]; then
    echo "$out"
    echo
    echo "$out" | hosomaki explain --cmd "$*"
  else
    echo "$out"
  fi
  return $code
}`,
	},
	"zsh": {
		rcFile: "~/.zshrc",
		code: `# hosomaki shell integration
explain() {
  local out
  out=$("$@" 2>&1)
  local code=$?
  if [[ $code -ne 0 ]]; then
    echo "$out"
    echo
    echo "$out" | hosomaki explain --cmd "$*"
  else
    echo "$out"
  fi
  return $code
}`,
	},
	"fish": {
		rcFile: "~/.config/fish/config.fish",
		code: `# hosomaki shell integration
function explain
  set out (command $argv 2>&1)
  set code $status
  if test $code -ne 0
    echo $out
    echo
    echo $out | hosomaki explain --cmd (string join ' ' $argv)
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
		Short: "Install shell wrapper for automatic error explanation",
		Long: `Installs a small shell wrapper. Any command prefixed with explain
will be analysed automatically if it fails.

  hosomaki shell-integration --shell bash >> ~/.bashrc && source ~/.bashrc
  hosomaki shell-integration --shell zsh  >> ~/.zshrc  && source ~/.zshrc
  hosomaki shell-integration --shell fish >> ~/.config/fish/config.fish

Then just prefix any command with explain:

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

			if !stdoutIsTerminal() {
				fmt.Println(s.code)
				return nil
			}

			ui := currentUI()
			ui.Title("hosomaki shell-integration")

			ui.Section("what it does")
			ui.Blank()
			ui.Detail("", "wraps any command with explain to auto-analyse failures")
			ui.Detail("", "on failure: runs the command, shows output, then asks hosomaki to explain it")

			ui.Section("install")
			ui.Blank()
			ui.Metric("shell", shell, 0)
			ui.Metric("config", s.rcFile, 0)
			ui.Blank()
			ui.Process(fmt.Sprintf("hosomaki shell-integration --shell %s >> %s && source %s", shell, s.rcFile, s.rcFile))

			ui.Section("usage")
			ui.Blank()
			ui.Process("explain sudo systemctl start nginx")
			ui.Process("explain make build")
			ui.Process("explain docker compose up")

			ui.Done()
			return nil
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "", "target shell: bash, zsh, or fish (default: auto-detect)")
	return cmd
}

func detectShell() string {
	base := filepath.Base(os.Getenv("SHELL"))
	if _, ok := shellSnippets[base]; ok {
		return base
	}
	return "bash"
}

func stdoutIsTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
