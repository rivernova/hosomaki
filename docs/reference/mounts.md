# mounts

Inspect active mounts, detect stale NFS, and flag disks approaching capacity.

## Usage

```bash
hosomaki mounts [flags]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--debug` | `false` | Print raw model response to stderr |

## What it collects

- Active mount entries from `/proc/mounts`
- Disk usage via `df`
- NFS mount reachability

## Output

The AI analysis flags:

- Filesystems approaching capacity (typically >80% used)
- Stale or unresponsive NFS mounts
- Unusual or unexpected mount points
- Filesystems mounted read-only that are expected to be read-write

## Examples

```bash
hosomaki mounts
hosomaki mounts --debug
```