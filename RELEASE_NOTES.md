## hosomaki v0.5.0

First public pre-release. Local-only Linux layer intelligence with plain languahe explanations.

Expect it to be usable, not yet stable.

### What it does

Runs system diagnostics locally and explains them using a model served by [Ollama](https://ollama.com) (`llama3` by default). Every command runs through the same pipeline: collect → sanitise → prompt → validate → repair → render. Raw system data is sanitised before it ever reaches the model.

### Commands

- `doctor` — overall system health check
- `explain` — explain arbitrary log lines or error output
- `why <exit-code>` — explain a process exit code
- `ports` — explain listening ports and what owns them
- `timers` — explain systemd timers
- `crons` — explain cron jobs
- `mounts` — explain mounted filesystems
- `updates` — explain pending package updates
- `firewall` — explain active firewall rules
- `history` — local log of past diagnostics
- `audit` — system audit
- `watch <service>` — stream and explain a service live
- `status` — runtime and status info
- `shell-integration` — shell hook setup

### Notable

- Config via file or env vars (`ai.model`, `ai.endpoint`, `ai.timeout`, `output.color`, `output.language`)

### Install

`.deb` / `.rpm` packages or build from source. Requires Ollama running locally with a model pulled. See the installation docs.