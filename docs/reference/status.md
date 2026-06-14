# status

Quick summary of current system health.

## Usage

```bash
hosomaki status [flags]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--brief` | `false` | One-sentence summary instead of full output |
| `--debug` | `false` | Print raw model response to stderr |

## What it collects

- System uptime
- Memory usage (total / used / available)
- Disk usage per mounted filesystem
- Failed systemd services
- Recent errors from the system journal

## Examples

```bash
hosomaki status
hosomaki status --brief
```

## Output

`status` renders a system section with collected metrics, followed by an AI-generated insights section identifying anything that warrants attention.

`--brief` condenses everything into a single sentence suitable for scripting or status bars.