# Daemon Mode

::: tip Roadmap
Daemon mode is planned for **Phase 3** of the Hosomaki roadmap. This page reflects the intended design and will be updated as implementation progresses.
:::

## Overview

The Hosomaki daemon will run as a background service, enabling:

- Persistent journal watching across multiple services simultaneously
- A local API for the native UI layer to consume
- Scheduled analysis runs (e.g. daily digest of system health)
- Proactive alerting when health thresholds are breached

## Design principles

- The daemon will not require root. It will use the same read-only collection strategy as the CLI.
- The same sanitise → prompt → validate → repair pipeline used by the CLI will be reused unchanged.

## Next steps

See [Configuration](/daemon/configuration) for the planned config schema.