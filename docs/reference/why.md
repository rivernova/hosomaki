# why

Given an exit code and a service, reconstruct the full failure chain.

## Usage

```bash
hosomaki why <exit-code> --service <name> [flags]
```

## Arguments

| Argument | Description |
|---|---|
| `<exit-code>` | Process exit code to explain |

## Flags

| Flag | Default | Description |
|---|---|---|
| `--service <name>` | — | Service whose journal to read |
| `--lines <N>` | `50` | Number of journal lines to include |
| `--since <time>` | — | Start of time range |

## Examples

```bash
# Why did nginx exit with code 1?
hosomaki why 1 --service nginx

# OOM kill (exit code 137 = SIGKILL)
hosomaki why 137 --service myapp --lines 100

# Scope to a time window
hosomaki why 1 --service nginx --since "10 min ago"
```