# mounts

Inspect active mounts, detect stale NFS, and flag disks approaching capacity.

## Usage

```bash
hosomaki mounts [flags]
```

## Output

Hosomaki will flag:

- Filesystems approaching capacity (typically >80% used)
- Stale or unresponsive NFS mounts
- Unusual or unexpected mount points
- Filesystems mounted read-only that are expected to be read-write

## Examples

```bash
hosomaki mounts
hosomaki mounts --debug
```