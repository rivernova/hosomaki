# firewall

Explain active firewall rules in plain language and cross-reference them against active network listeners.

## Usage

```bash
hosomaki firewall [flags]
```

## Flags

| Flag | Default | Description |
|---|---------|---|
| `--cross-check` | `false` | Cross-reference active rules against currently listening system ports |

## How it works

Hosomaki inspects the system's active firewall configuration using a top-down hierarchy. It detects and interprets rules from the following backends based on what is active:

1. `firewalld` or `ufw` (Application-layer abstractions)

2. `nftables` (Modern packet-filtering engine)

3. `iptables` (Legacy packet-filtering engine)

## Output

`firewall` translates complex, backend-specific rule syntax into simple plain language.

## Cross-Check Behavior

Hosomaki maps the firewall rules against currently open sockets to flag anomalies like orphaned rules or unprotected services.

::: danger Safety & False Negatives
Hosomaki enforces a strict fail-closed reporting policy. If a ruleset cannot be completely read, the command will immediately abort with an error rather than risking a security-relevant false negative or reporting partial results as complete.
:::

## Examples

```bash
# explain active firewall rules
hosomaki firewall

# find gaps or dead rules
hosomaki firewall --cross-check
```

## Constraints

**Privileges:** Requires appropriate system privileges to reliably query backends like `nftables` or `iptables-save`. Without them, the command will safely abort.