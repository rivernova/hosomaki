# Data Privacy

Hosomaki is built around a single principle: **your data stays on your machine**.

There are no analytics, no telemetry, no crash reports, and no usage metrics.

## Data flow

```
System
  │
  ▼
collector. raw data
  │
  ▼
sanitiser. strips IPs, paths, credentials, hostnames, UUIDs
  │
  ▼
prompt builder. constructs constrained prompt from sanitised data
  │
  ▼
Ollama. local model
  │
  ▼
validator / repair
  │
  ▼
terminal output
```

No stage after the sanitiser has access to raw system data.

## Audit baseline

The audit baseline (`hosomaki audit --init`) stores a snapshot of file hashes, package versions, listening ports, and systemd unit states. This file is written to `~/.local/share/hosomaki/audit-baseline.json` by default and never transmitted anywhere.

## History log

`explain`, `why`, `audit`, `status`, and `doctor` each record their result to a local history log (`~/.local/share/hosomaki/history.json` by default) so `hosomaki history` can surface them later. Only the model's final, validated output is stored.

## Future multi-provider support

The roadmap includes optional support for cloud-based model providers (Phase 4). This will be an explicit opt-in — the local-only default will never change. When cloud providers are enabled, the same sanitisation layer will apply before any data leaves the machine, and the configuration will clearly surface which provider is active.