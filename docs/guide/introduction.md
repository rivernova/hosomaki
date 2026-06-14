# Introduction

Hosomaki is a Linux system diagnostics CLI that reads your system and helps you understand what is happening in plain language.

It uses a local language model via [Ollama](https://ollama.com) and **never sends anything off your machine**. No cloud. No telemetry. No account. Your data stays on your system.

## What it does

When something goes wrong on a Linux system — a service crashing, a mount hanging, a port appearing that shouldn't be there — the usual path is `journalctl`, `ss`, `systemctl`, and several minutes of reading. Hosomaki shortens that loop by collecting the relevant system state and handing it to a local model with a tightly constrained prompt, then returning a structured, readable analysis.

Every command follows the same pipeline:

```
collect → sanitise → prompt → validate → repair → render
```

The sanitisation step is non-negotiable. It strips IP addresses, hostnames, paths, UUIDs, and credentials before anything enters the model context.

## Core commands

| Command | What it does |
|---|---|
| `explain` | Explain errors from a service, boot, log file, pipe, or running process |
| `status` | Quick summary of current system health |
| `doctor` | Full diagnosis with concrete suggested actions |
| `audit` | Surface changes since a baseline snapshot |
| `watch` | Tail a service and explain errors in real time |
| `why` | Reconstruct the failure chain for a given exit code and service |
| `ports` | List listening ports and flag anything unexpected |
| `timers` | Inspect systemd timers and flag failures or overdue schedules |
| `crons` | Read all crontabs and explain what each job does |
| `mounts` | Inspect active mounts, detect stale NFS, and flag disks nearing capacity |
| `shell-integration` | Install a shell wrapper that explains failed commands automatically |

## Design principles

**Read-only.** Every command collects data and surfaces insights. None of them modify the system.

**Local first.** The AI model runs on your machine via Ollama. Internet connectivity is never required.

**Sanitise first, always.** Sensitive material is stripped before the model ever sees it. This happens in the `cmd/` layer, before the prompt package is called, and cannot be bypassed.

**Structured output.** Every command produces validated, typed JSON from the model. Invalid or semantically empty responses are repaired or rejected — not silently passed through.

## Next steps

- [Installation](/guide/installation) — install Hosomaki and Ollama
- [Quick Start](/guide/quickstart) — run your first command in under two minutes
- [Architecture](/guide/architecture) — understand the pipeline in depth