# Quick Start

Get up and running in under two minutes.

## 1. Check system health

```bash
hosomaki status
```

Returns a snapshot of uptime, memory, disk, failed services, and recent errors, followed by an AI-generated summary.

```bash
hosomaki status --brief   # one-sentence summary
```

## 2. Explain a failing service

```bash
hosomaki explain --service nginx
```

Hosomaki reads the last 50 lines of the nginx journal, sanitises them, and returns a structured explanation with root cause analysis and suggested investigation steps.

```bash
# Read more lines
hosomaki explain --service nginx --lines 100

# Scope to a time range
hosomaki explain --service nginx --since "1 hour ago"

# Correlate two services
hosomaki explain --context nginx,postgresql
```

## 3. Full diagnosis

```bash
hosomaki doctor
```

Deeper than `status` — surfaces concrete suggested actions, not just observations.

## 4. Watch a service in real time

```bash
hosomaki watch nginx
```

Tails the journal and explains error batches as they arrive. Press `Ctrl-C` to stop.

## 5. Audit for changes

```bash
# Take a baseline
hosomaki audit --init

# Later, diff against it
hosomaki audit
```

## 6. Shell integration

Install the `explain` wrapper so failed commands are automatically explained:

```bash
hosomaki shell-integration --shell bash >> ~/.bashrc && source ~/.bashrc

# Now any command can be prefixed with 'explain'
explain make build
explain systemctl start myapp
```

## Common patterns

```bash
# Pipe arbitrary log output
journalctl -p err -n 50 | hosomaki explain

# Explain a kernel message inline
hosomaki explain "kernel: OOM killer activated on process 1234"

# Explain what exited with code 137
hosomaki why 137 --service myapp

# Check listening ports
hosomaki ports

# Inspect systemd timers
hosomaki timers

# Read all crontabs
hosomaki crons

# Mount health
hosomaki mounts
```