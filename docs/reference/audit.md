# audit

Surface changes since a baseline snapshot.

## Usage

```bash
hosomaki audit [flags]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--init` | `false` | Take a new baseline and save it to disk |
| `--baseline <path>` | `~/.local/share/hosomaki/audit-baseline.json` | Path to baseline file |
| `--dirs <path,...>` | — | Additional directories to track (comma-separated) |
| `--debug` | `false` | Print raw model response to stderr |

## Workflow

### 1. Take a baseline

```bash
hosomaki audit --init
```
Baseline is saved to `~/.local/share/hosomaki/audit-baseline.json` by default.

### 2. Diff against it

```bash
hosomaki audit
```

Compares the current system state against the stored baseline and flags anything significant.

## Examples

```bash
# Standard workflow
hosomaki audit --init
# ... time passes ...
hosomaki audit

# Track additional directories
hosomaki audit --init --dirs /etc,/usr/local/bin
hosomaki audit --dirs /etc,/usr/local/bin

# Use a custom baseline path
hosomaki audit --init --baseline /tmp/pre-deploy-baseline.json
hosomaki audit --baseline /tmp/pre-deploy-baseline.json
```

::: tip Pre/post deployment
`audit` is well-suited for pre/post deployment comparison. Take a baseline before deploying, then diff immediately after to surface unexpected changes.
:::

## Baseline file location

The default location follows the XDG Base Directory Specification:

```
$XDG_DATA_HOME/hosomaki/audit-baseline.json
```

Falls back to `~/.local/share/hosomaki/audit-baseline.json` if `XDG_DATA_HOME` is not set.