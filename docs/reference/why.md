# why

Given an exit code and a service, reconstruct the full failure chain.

## Usage

```bash
hosomaki why <exit-code> --service <name> [flags]
```

## Arguments

| Argument | Description |
|---|---|
| `<exit-code>` | Process exit code to explain (required) |

## Flags

| Flag | Default | Description |
|---|---|---|
| `--service <name>` | — | Service whose journal to read (required) |
| `--lines <N>` | `50` | Number of journal lines to include |
| `--since <time>` | — | Start of time range (journalctl-compatible) |
| `--debug` | `false` | Print raw model response to stderr |

## Examples

```bash
# Why did nginx exit with code 1?
hosomaki why 1 --service nginx

# OOM kill (exit code 137 = SIGKILL)
hosomaki why 137 --service myapp --lines 100

# Scope to a time window
hosomaki why 1 --service nginx --since "10 min ago"
```

## Common exit codes

| Code | Signal | Typical cause |
|---|---|---|
| `1` | — | General error |
| `2` | — | Misuse of shell built-ins |
| `126` | — | Permission denied |
| `127` | — | Command not found |
| `130` | SIGINT | Ctrl-C |
| `137` | SIGKILL | OOM kill or `kill -9` |
| `143` | SIGTERM | Clean shutdown requested |