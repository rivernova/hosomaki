# Command Reference

All Hosomaki commands are read-only. None of them modify the system.

## Commands at a glance

| Command | What it does |
|---|---|
| [`explain`](/reference/explain) | Explain errors from a service, boot, log file, pipe, inline text, or a running process |
| [`status`](/reference/status) | Quick summary of current system health |
| [`doctor`](/reference/doctor) | Full diagnosis with concrete suggested actions |
| [`audit`](/reference/audit) | Surface changes since a baseline snapshot |
| [`watch`](/reference/watch) | Tail a service journal and explain errors in real time |
| [`why`](/reference/why) | Reconstruct the failure chain for a given exit code and service |
| [`ports`](/reference/ports) | List listening ports and flag anything unexpected |
| [`timers`](/reference/timers) | Inspect systemd timers and flag failures or overdue schedules |
| [`crons`](/reference/crons) | Read all crontabs and explain what each job does |
| [`mounts`](/reference/mounts) | Inspect active mounts, detect stale NFS, and flag disks nearing capacity |
| [`shell-integration`](/reference/shell-integration) | Install a shell wrapper that explains failed commands automatically |

## Common flags

All commands accept:

| Flag | Description |
|---|---|
| `--debug` | Print raw model response to stderr before parsing |
| `--help` | Show usage and flags for the command |

Run `hosomaki <command> --help` for the full flag reference of any command.

## Output format

Commands write to stdout. The output structure is:

```
── <command> ──────────────────────────────────────────────────

  context
  ─────────────────────────────────────────────────────────
  key          value

  <ai-generated sections>
```

The `--debug` flag appends the raw model JSON to stderr.