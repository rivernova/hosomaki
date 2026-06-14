# timers

Inspect all systemd timers and flag failures or overdue schedules.

## Usage

```bash
hosomaki timers [flags]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--debug` | `false` | Print raw model response to stderr |

## What it collects

Runs `systemctl list-timers --all` to collect all active and inactive systemd timers, including:

- Timer unit name and the service it activates
- Last run time and result
- Next scheduled run
- Active state

Timers with no recorded last run are reported as `last_run: never`.

## Output

The AI analysis flags timers that have:

- Failed on last run
- Never run at all
- Missed their expected run window (appear overdue)
- Been inactive for an unexpectedly long time

## Relationship to `crons`

`timers` covers **systemd timers**. Classic crontab files (`/etc/crontab`, `/etc/cron.d/*`, per-user crontabs) are handled by `hosomaki crons`.

## Examples

```bash
hosomaki timers
hosomaki timers --debug
```