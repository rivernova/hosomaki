# ports

List all listening TCP and UDP ports with associated process names, and ask the AI to flag anything unexpected or potentially concerning.

## Usage

```bash
hosomaki ports [flags]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--debug` | `false` | Print raw model response to stderr |

## What it collects

Runs `ss -tlunp` to collect all listening TCP and UDP sockets, including:

- Protocol (TCP/UDP)
- Bind address and port
- Process name and PID

## Output

The AI analysis produces:

- **Summary** — overall picture of listening ports and general exposure level (max 40 words)
- **Findings** — one entry per distinct concern, each with:
    - `severity` — `warning` (investigate promptly) or `info` (worth noting)
    - `port` — protocol and address, copied verbatim from collected data
    - `title` — concise label
    - `detail` — 2–4 sentences describing what is unusual and why

The AI is explicitly instructed not to suggest specific remediation commands.

## Relationship to `audit`

`ports` shows the current state of listening sockets. To track how ports change over time — ports opened or closed since a baseline — use `hosomaki audit`.

## Examples

```bash
hosomaki ports
hosomaki ports --debug
```

::: tip Elevated privileges
Running as root provides more complete process name information for sockets owned by other users.
:::