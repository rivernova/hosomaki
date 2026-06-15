# Native UI Roadmap

::: tip Roadmap
The native UI layer is planned for **Phase 6** of the Hosomaki roadmap. This page reflects the intended direction.
:::

## Vision

A native desktop UI that consumes the Hosomaki daemon API and surfaces system diagnostics in a structured, interactive interface, without requiring the terminal.

## Goals

- Real-time service health dashboard
- Interactive exploration of Hosomaki's findings
- Historical trend views
- Notification integration for threshold breaches

## Non-goals

- The UI will not introduce cloud connectivity. The local-only, no-telemetry guarantee applies to all Hosomaki layers.

## Dependencies

The native UI requires:

1. **Phase 3 — Daemon mode** — persistent background service and local API
2. **Phase 4 — Memory layer** — historical data storage for trend views
3. **Phase 5 — Multi-provider AI** (optional) — for users who want cloud model support. This is still being designed.

The CLI layer (Phases 1) is complete and stable independently of the UI roadmap.