# Architecture

Hosomaki is a Go CLI built on [Cobra](https://github.com/spf13/cobra) and [Viper](https://github.com/spf13/viper). The internal structure enforces a strict separation between data collection, sanitisation, AI interaction, validation, and rendering.

## Package layout

```
hosomaki/
├── cmd/hosomaki/        # Cobra command definitions, one file per command
├── internal/
│   ├── ai/              # Ollama client, streaming pipeline, schema types
│   ├── auditor/         # Audit baseline store (read/write)
│   ├── collector/       # System data collectors (journalctl, ss, proc, etc.)
│   ├── prompt/          # Prompt builders and output schema definitions
│   ├── sanitiser/       # PII/sensitive-data scrubbing layer
│   ├── spinner/         # Terminal spinner
│   ├── ui/              # Output renderers (layout.go, render.go, live.go)
│   └── watcher/         # Real-time journal tail (for `watch` command)
└── main.go
```

## Architectural rules

These rules are enforced at code review and violated by zero current code:

1. **Sanitisation happens in `cmd/`**, before the prompt package is called. The `sanitiser` package is never imported by `internal/prompt`.
2. **`internal/ui/` is split into three files** with explicit responsibilities:
    - `layout.go` — non-AI structural sections (headers, key-value pairs, collected data)
    - `render.go` — AI result renderers (structured findings, summaries)
    - `live.go` — streaming and real-time renderers (`watch` output)
3. **Shared types are reused**, not duplicated across commands.
4. **All commands follow the same pipeline** (see below).

## The AI pipeline

Every command runs through the same five-stage pipeline:

```
collect → sanitise → prompt → validate → repair → render
```

### 1. Collect

The `collector` package gathers raw system data:

- Journal logs via `journalctl`
- Listening sockets via `ss -tlunp`
- Systemd timer state via `systemctl list-timers`
- Crontab files from `/etc/crontab`, `/etc/cron.d/*`, and per-user crontabs
- Mount state from `/proc/mounts` and `df`
- Process information from `/proc/<pid>/`
- System snapshot (uptime, memory, disk) for `status` and `doctor`

### 2. Sanitise

Before any data reaches the prompt, the `sanitiser.Default()` scrubber removes:

- IP addresses (IPv4 and IPv6)
- Hostnames and FQDNs
- Filesystem paths
- UUIDs
- Credentials and tokens
- Usernames

This step is mandatory and cannot be skipped or bypassed.

### 3. Prompt

The `prompt` package builds a tightly constrained prompt from the sanitised data. Each command has its own prompt builder and a companion JSON schema that defines the exact output structure the model must produce.

Prompts explicitly forbid markdown, bullet points, headers, and remediation commands. They specify the exact field names, types, and value constraints the model must return.

### 4. Validate

The model response is parsed against the command's JSON schema. A `StructValidator` then applies semantic checks — for example, verifying that `severity` is exactly `"warning"` or `"info"`, that `summary` is non-empty, and that port strings match expected formats.

### 5. Repair

If validation fails, the pipeline invokes `BuildStructuralRepairPromptWithContext` — a repair prompt that includes the original task context. This ensures the model understands what it was trying to produce, not just that its output was malformed. Technically valid but semantically empty repairs are treated as failures.

### 6. Render

The validated, typed result is passed to `internal/ui/render.go`, which formats it for the terminal using the same layout primitives as the non-AI sections.

## Streaming

For commands like `watch` that process data in real time, `internal/watcher` accumulates incoming journal lines into batches. A batch is submitted to the AI pipeline only when it contains at least one error or warning and a silence window has elapsed (or the batch has hit `--max-lines`). Informational-only batches are discarded silently.

## Further reading

- [AI Pipeline](/guide/pipeline) — deeper walk-through of stream processing and repair
- [Sanitisation](/guide/sanitisation) — what gets scrubbed and how
- [Data Privacy](/guide/privacy) — end-to-end data handling guarantees