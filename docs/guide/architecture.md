# Architecture

Hosomaki is CLI built on [Cobra](https://github.com/spf13/cobra) and [Viper](https://github.com/spf13/viper). The internal structure enforces a strict separation between data collection, sanitisation, AI interaction, validation, and rendering.

## Architectural rules

These rules are enforced at code review and violated by zero current code:

1. **Sanitisation** happens before the prompt package is called.
2. **All commands follow the same pipeline** 

## The AI pipeline

Every command runs through the same five-stage pipeline:

```
collect → sanitise → prompt → validate → repair → render
```

### 1. Collect

The `collector` package gathers raw system data:

- Journal logs.
- Listening sockets.
- Systemd timer state.
- Crontab files, and per-user crontabs.
- Mount state.
- Process information.
- System snapshot.

### 2. Sanitise

Before any data reaches the prompts, hosomaki removes:

- IP addresses (IPv4 and IPv6)
- Hostnames and FQDNs
- Filesystem paths
- UUIDs
- Credentials and tokens
- Usernames

This step is mandatory and cannot be skipped or bypassed.

### 3. Prompt

Each command has its own prompt builder and a companion JSON schema that defines the exact output structure the model must produce.

Prompts specify the exact field names, types, and value constraints the model must return.

### 4. Validate

The model response is parsed against the command's JSON schema.

### 5. Repair

If validation fails, the pipeline invokes  a repair prompt that includes the original task context. This ensures the model understands what it was trying to produce, not just that its output was malformed. Technically valid but semantically empty repairs are treated as failures.

### 6. Render

The validated, typed result is passed to renderer.

## Streaming

For commands that process data in real time, there is a watcher that accumulates incoming data into batches. A batch is submitted to the pipeline only when it contains at least one error or warning and a silence window has elapsed. Informational-only batches are discarded silently.

## Further reading

- [AI Pipeline](/guide/pipeline) — deeper walk-through of stream processing and repair
- [Sanitisation](/guide/sanitisation) — what gets scrubbed and how
- [Data Privacy](/guide/privacy) — end-to-end data handling guarantees