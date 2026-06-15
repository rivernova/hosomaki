# ports

List all listening TCP and UDP ports with associated process names, flags anything unexpected or potentially concerning.

## Usage

```bash
hosomaki ports [flags]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--debug` | `false` | Print raw model response to stderr |

## What it collects

All listening TCP and UDP sockets, including:

- Protocol (TCP/UDP)
- Bind address and port
- Process name and PID

## Relationship to `audit`

`ports` shows the current state of listening sockets. If you want to track how ports change over time use `hosomaki audit`.

## Examples

```bash
hosomaki ports
hosomaki ports --debug
```

::: tip Elevated privileges
Running as root provides more complete process name information for sockets owned by other users.
:::