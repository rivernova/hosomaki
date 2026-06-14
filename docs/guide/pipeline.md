# Pipeline

This page describes the streaming pipeline that underpins every Hosomaki command.

## Validation

After the model streams a complete response, the pipeline:

1. Parses the raw string as JSON
2. Validates it against the registered schema
3. Semantically checks and returns a list of error strings, empty on success

If any step fails, the pipeline moves to repair.

## Repair

The repair stage calls the repairer, which bundles:

- The **original task prompt** 
- The **malformed output**
- The **validation errors** 
- The **schema**

This context-preserving repair is essential. A repair prompt without the original task cannot produce a semantically correct response.

## Debug mode

Every command accepts `--debug`, which prints the raw model response. This is useful for diagnosing unexpected validation failures.

```bash
hosomaki ports --debug
hosomaki explain --service nginx --debug
```

## Stream pipeline in live commands

The live commands use the same pipeline but drives it from an event loop in the watcher:

```
  tail
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

The silence window and max-lines are tunable per invocation.