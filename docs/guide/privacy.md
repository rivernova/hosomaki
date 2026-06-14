# Data Privacy

Hosomaki is built around a single principle: **your data stays on your machine**.

## What runs where

| Component | Location |
|---|---|
| Hosomaki CLI | Your machine |
| Language model (Ollama) | Your machine |
| System data collection | Your machine |
| Sanitisation | Your machine, before prompting |
| AI inference | Your machine |

Nothing is sent to Anthropic, OpenAI, or any other external service. There are no analytics, no telemetry, no crash reports, and no usage metrics.

## Data flow

```
System (journalctl, ss, proc...)
  │
  ▼
collector — raw data (never leaves this stage as-is)
  │
  ▼
sanitiser — strips IPs, paths, credentials, hostnames, UUIDs
  │
  ▼
prompt builder — constructs constrained prompt from sanitised data
  │
  ▼
Ollama (local) — inference on localhost
  │
  ▼
validator / repair — structured result verified before display
  │
  ▼
terminal output
```

No stage after the sanitiser has access to raw system data. The Ollama API call is made to `localhost` (or your configured Ollama URL) and never traverses the internet.

## Audit baseline

The audit baseline (`hosomaki audit --init`) stores a snapshot of file hashes, package versions, listening ports, and systemd unit states. This file is written to `~/.local/share/hosomaki/audit-baseline.json` by default and never transmitted anywhere.

## Future multi-provider support

The roadmap includes optional support for cloud-based model providers (Phase 4). This will be an explicit opt-in — the local-only default will never change. When cloud providers are enabled, the same sanitisation layer will apply before any data leaves the machine, and the configuration will clearly surface which provider is active.