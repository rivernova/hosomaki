All Hosomaki commands are read-only with respect to the system they're diagnosing, none of them modify system state. Two commands write to Hosomaki's own local state directory: `audit --init` creates the baseline snapshot, and `history --clear` deletes the history log.

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
| [`updates`](/reference/updates) | Explain pending package updates before applying them |
| [`history`](/reference/history) | Review past diagnostic results |
| [`shell-integration`](/reference/shell-integration) | Install a shell wrapper that explains failed commands automatically |

## Common flags

All commands accept:

| Flag | Description |
|---|---|
| `--debug` | Print raw model response |

Run `hosomaki <command> --help` for the full flag reference of any command.