# watch

Tail a service journal and explain new errors as they appear in real time.

## Usage

```bash
hosomaki watch <service> [flags]
```

## Arguments

| Argument | Description |
|---|---|
| `<service>` | Systemd service name to tail (required) |

## Flags

| Flag | Default | Description |
|---|---|---|
| `--lines <N>` | `10` | Lines to seed from the journal on startup (0 to skip) |
| `--window <duration>` | `5s` | Silence window after an error/warning before flushing the batch |
| `--max-lines <N>` | `30` | Flush the batch when it reaches this many lines regardless of the window |
| `--debug` | `false` | Print raw model response to stderr |

## How it works

On startup, `watch` seeds the view with the last `--lines` lines from the journal (after sanitisation), then enters tail mode.

Incoming lines are accumulated into a batch. The batch is flushed to the AI pipeline when:

- A silence window has elapsed **and** the batch contains at least one error or warning line
- **Or** the batch has reached `--max-lines`

Batches containing only informational lines are discarded silently.

Press `Ctrl-C` to stop. `watch` drains any pending batch, cancels the journal tail, and prints a clean shutdown line.

## Examples

```bash
# Watch nginx
hosomaki watch nginx

# Seed with the last 20 lines
hosomaki watch nginx --lines 20

# Skip seeding entirely
hosomaki watch nginx --lines 0

# Tune batching: shorter window, more lines per batch
hosomaki watch nginx --window 10s --max-lines 50

# Skip seeding, long window for low-volume services
hosomaki watch myapp --lines 0 --window 30s
```

::: tip Low-volume services
For services that log infrequently, increase `--window` to avoid submitting single-line batches. `--window 30s` or `--window 60s` works well for cron-driven services.
:::