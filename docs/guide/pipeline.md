# AI Pipeline

This page describes the streaming pipeline that underpins every Hosomaki command.

## Overview

```
StreamPipeline[T]
  │
  ├── provider (Ollama client)
  ├── schema   (JSON Schema for output type T)
  └── validator (StructValidator[T])
        ├── JSON parse
        ├── Schema check
        └── SemanticCheck func
```

`ai.NewStreamPipeline` constructs a typed pipeline parameterised on the result type `T` — for example, `prompt.PortsResult` or `prompt.ExplainResult`.

## Validation

After the model streams a complete response, the pipeline:

1. Parses the raw string as JSON
2. Validates it against the registered schema
3. Calls `SemanticCheck(result T) []string` — returns a list of error strings, empty on success

If any step fails, the pipeline moves to repair.

## Repair

The repair stage calls `BuildStructuralRepairPromptWithContext`, which bundles:

- The **original task prompt** (what the model was asked to do)
- The **malformed output** (what the model actually returned)
- The **validation errors** (what was wrong)
- The **schema** (what is expected)

This context-preserving repair is essential. A repair prompt without the original task cannot produce a semantically correct response — it can only produce something that parses correctly. Hosomaki treats a repair that passes JSON validation but produces an empty or placeholder result as a failure.

## Debug mode

Every command accepts `--debug`, which prints the raw model response to stderr before parsing. This is useful for diagnosing unexpected validation failures.

```bash
hosomaki ports --debug
hosomaki explain --service nginx --debug
```

## Stream pipeline in `watch`

The `watch` command uses the same pipeline but drives it from an event loop in `internal/watcher`:

```
journal tail
  │
  ├── accumulate lines into buffer
  │
  ├── on silence window expiry (or max-lines reached):
  │     ├── filter: contains error/warning? → yes → flush to pipeline
  │     │                                    → no  → discard silently
  │     └── flush: sanitise → prompt → pipeline → render
  │
  └── on Ctrl-C: drain buffer, cancel context, print shutdown line
```

The silence window (default `5s`) and max-lines (default `30`) are tunable per invocation.