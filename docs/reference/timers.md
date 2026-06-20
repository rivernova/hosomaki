# timers

Inspect all systemd timers and flag failures or overdue schedules.

## Usage

```bash
hosomaki timers [flags]
```

## What it collects

Collects all active and inactive systemd timers, including:

- Timer unit name and the service it activates
- Last run time and result
- Next scheduled run
- Active state

Timers with no recorded last run are reported as `last_run: never`.

## Output

Hosomaki will flag timers that have:

- Failed on last run
- Never run at all
- Missed their expected run window
- Been inactive for an unexpectedly long time

## Relationship to `crons`

`timers` covers **systemd timers**. Classic crontab files are handled by `hosomaki crons`.

## Examples

```bash
hosomaki timers
hosomaki timers --debug
```