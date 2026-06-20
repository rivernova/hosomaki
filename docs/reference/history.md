# history

Review past diagnostic results.

## Usage

```bash
hosomaki history [flags]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--limit <n>` | `10` | Show the last N entries |
| `--since <duration>` | — | Show entries newer than duration (e.g. `24h`, `7d`) |
| `--command <name>` | — | Filter by source command (`explain`, `why`, `audit`, `status`, `doctor`) |
| `--clear` | `false` | Clear the history log |

## What gets logged

Hosomaki automatically records the result of every run of:

`explain`, `why`, `audit`, `status`, `doctor`

Each entry stores a timestamp, the source command, and that run's result. Up to 1000 entries are kept, older entries are trimmed automatically.

## Output

`history` summarises the matching entries and surfaces patterns across them — recurring issues, a service that keeps coming up, or a trend over time — rather than just replaying old output verbatim.

::: tip Note on `--clear`
Hosomaki commands are read-only with respect to the system being diagnosed. `--clear` is one of two exceptions that write to Hosomaki's own local state rather than the system itself. It deletes the history log from disk.
:::

## Examples

```bash
# Last 10 entries (default)
hosomaki history

# Only past explain runs
hosomaki history --command explain

# Anything from the last week
hosomaki history --since 7d

# Combine filters
hosomaki history --command doctor --since 24h --limit 5

# Wipe the log
hosomaki history --clear
```

## History file location

The default location follows the XDG Base Directory Specification:

```bash 
$XDG_DATA_HOME/hosomaki/history.json
```

Falls back to `~/.local/share/hosomaki/history.json` if `XDG_DATA_HOME` is not set.