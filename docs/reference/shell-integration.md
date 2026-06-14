# shell-integration

Print a shell function that wraps commands and explains failures automatically.

## Usage

```bash
hosomaki shell-integration [flags]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--shell <name>` | auto-detect | Shell to generate snippet for (`bash`, `zsh`, `fish`) |

## How it works

`shell-integration` prints a shell function called `explain`. When you prefix any command with `explain`, it:

1. Runs the command and captures its output
2. If the command exits with a non-zero code, pipes the captured output to `hosomaki explain --cmd <cmdline>`
3. If the command succeeds, prints the output normally

The function is entirely client-side — it doesn't modify how Hosomaki works.

## Installation

### Bash

```bash
hosomaki shell-integration --shell bash >> ~/.bashrc && source ~/.bashrc
```

### Zsh

```bash
hosomaki shell-integration --shell zsh >> ~/.zshrc && source ~/.zshrc
```

### Fish

```bash
hosomaki shell-integration --shell fish >> ~/.config/fish/config.fish
```

## Usage after installation

```bash
explain make build
explain systemctl start myapp
explain go test ./...
explain apt-get upgrade
```

If the command fails, Hosomaki automatically explains what went wrong.

## Examples

```bash
# Print the bash snippet (without installing)
hosomaki shell-integration --shell bash

# Install to bash
hosomaki shell-integration --shell bash >> ~/.bashrc && source ~/.bashrc

# Use it
explain make build
```

::: tip Shell detection
If `--shell` is omitted, Hosomaki detects the current shell from the `$SHELL` environment variable.
:::