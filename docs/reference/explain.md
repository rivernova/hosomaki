# explain

Explain errors from a variety of sources. Hosomaki adapts to whatever you provide — a service name, a log file, a piped stream, an inline message, or a running process PID.

## Usage

```bash
hosomaki explain [flags] [message]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--service <name>` | — | Read logs from a systemd service journal |
| `--boot [N]` | — | Read logs from the current boot (or boot index N, e.g. `-1` for previous) |
| `--dmesg` | — | Read the kernel ring buffer |
| `--file <path>` | — | Read from a log file |
| `--context <svc,...>` | — | Correlate logs from multiple services (minimum 2) |
| `--pid <N>` | — | Inspect a running process by PID |
| `--diff <from>:<to>` | — | Compare two boots (e.g. `-1` or `-2:-1`) |
| `--lines <N>` | `50` | Number of lines to read |
| `--since <time>` | — | Start of time range (journalctl-compatible) |
| `--until <time>` | — | End of time range |
| `--cmd <cmdline>` | — | Original command (used by shell-integration) |
| `--debug` | `false` | Print raw model response to stderr |

## Input modes

`explain` accepts input in several mutually exclusive ways:

### Service journal

```bash
hosomaki explain --service nginx
hosomaki explain --service nginx --lines 100
hosomaki explain --service nginx --since "1 hour ago"
hosomaki explain --service nginx --since "2024-01-15 14:00" --until "2024-01-15 15:00"
```

### Boot logs

```bash
hosomaki explain --boot          # current boot
hosomaki explain --boot -1       # previous boot
hosomaki explain --boot -2       # two boots ago
```

### Kernel messages

```bash
hosomaki explain --dmesg
hosomaki explain --dmesg --since "30 minutes ago"
```

### Log file

```bash
hosomaki explain --file /var/log/syslog
hosomaki explain --file /var/log/nginx/error.log --lines 200
```

### Multi-service correlation

```bash
hosomaki explain --context nginx,postgresql
hosomaki explain --context app,redis,nginx
```

Requires at least two services.

### Process inspection

```bash
hosomaki explain --pid 1234
```

Reads `/proc/<pid>/` — cmdline, status, fd count, open files, memory maps — and explains what the process is doing and whether anything looks unusual.

### Boot diff

```bash
hosomaki explain --diff -1        # current boot vs. previous
hosomaki explain --diff -2:-1     # compare any two boots
```

### Pipe

```bash
journalctl -p err -n 50 | hosomaki explain
cat /var/log/app.log | hosomaki explain
```

### Inline message

```bash
hosomaki explain "kernel: OOM killer activated on process nginx"
hosomaki explain "FAILED: unit myapp.service entered failed state"
```

## Examples

```bash
# Explain the last 100 lines of the nginx journal
hosomaki explain --service nginx --lines 100

# Explain errors since 2 hours ago
hosomaki explain --service nginx --since "2 hours ago"

# Pipe dmesg output directly
dmesg | tail -n 30 | hosomaki explain

# Inspect a running process
hosomaki explain --pid $(pgrep nginx | head -1)

# Compare current boot vs previous
hosomaki explain --diff -1
```