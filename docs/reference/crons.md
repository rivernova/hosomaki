# crons

Read all classic crontab files and explain what each job does in plain language.

## Usage

```bash
hosomaki crons [flags]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--debug` | `false` | Print raw model response to stderr |

## What it collects

Reads all classic crontab files on the system:

- `/etc/crontab`
- `/etc/cron.d/*`
- Per-user crontabs (via `crontab -l -u <user>`)

Each job is collected with its schedule, user, command, and source file. Shell variable assignments and comments are skipped.

## Output

The AI analysis produces a plain-language explanation of each job — what it does, when it runs, and whether anything looks broken, suspicious, or misconfigured.

## Scope

`crons` covers **classic crontab files only**. Systemd timers are handled by `hosomaki timers`.

## Examples

```bash
hosomaki crons
hosomaki crons --debug
```

::: tip Per-user crontabs
Reading per-user crontabs requires root privileges or the ability to run `crontab -l -u <user>`. Running without root will surface system crontabs only.
:::