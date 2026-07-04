# Changelog

All notable changes to hosomaki are recorded here. Release bodies and signed artifacts live on the [GitHub Releases](https://github.com/rivernova/hosomaki/releases) page.

## [0.5.0] - 2026-07-04

First public pre-release. Local-only Linux intelligence layer with LLM explanations. For now, data never leaves the machine. Pre-1.0.0: the flag and config contract is not yet frozen.

### Added

- Commands: `doctor`, `explain`, `why`, `ports`, `timers`, `crons`, `mounts`, `updates`, `firewall`, `history`, `audit`, `watch`, `status`, `shell-integration`
- Local inference via Ollama (`llama3` by default) through the collect → sanitise → prompt → validate → repair → render pipeline
- Config via file or `HOSOMAKI_` env vars (`ai.model`, `ai.endpoint`, `ai.timeout`, `output.color`, `output.language`)
- `.deb` / `.rpm` packages via nfpm; release checksums signed with cosign (keyless)
- Comprehensive table-driven test suite. CI enforces a 70% aggregate coverage gate and golangci-lint

[0.5.0]: https://github.com/rivernova/hosomaki/releases/tag/v0.5.0