# crons

Read all classic crontab files and explain what each job does in plain language.

## Usage

```bash
hosomaki crons [flags]
```

## What it collects

Reads all classic crontab files on the system.

Shell variable assignments and comments are skipped.

## Output

Hosomaki will explain each job. What it does, when it runs, and whether anything looks broken, suspicious, or misconfigured.

## Examples

```bash
hosomaki crons
hosomaki crons --debug
```

::: tip Per-user crontabs
Reading per-user crontabs requires root privileges. Running without root will surface system crontabs only.
:::