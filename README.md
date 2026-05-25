# Hosomaki

**Your Linux system has a story to tell. Hosomaki is the moment it finally finds its voice.**

<p align="center">
  <img src="assets/hosomaki_logo.svg" alt="Hosomaki" width="350"/>
</p>

> Local intelligence layer for Linux — with and without AI.

Hosomaki reads your system and helps you understand what's happening in plain language.  
It works in two modes:

- **AI mode**: uses a local model to explain issues, correlate events and summarise system behaviour.  
- **Insight mode**: a deterministic, AI‑free analysis path based on rules, heuristics, diffs and structured inspection.

No cloud. No telemetry. Your system, your data, your choice.


## Commands

### `explain`

 To understand what's going on. It adapts to whatever you throw at it.

```bash
# Pipe any log output directly
journalctl -p err -n 20 | hosomaki explain
dmesg | tail -50         | hosomaki explain

# By systemd service. hosomaki will fetch the logs
hosomaki explain --service nginx
hosomaki explain --service postgresql --lines 100

# Errors from the last boot. This one is useful after a crash
hosomaki explain --boot
hosomaki explain --boot -1        # the boot before that

# Kernel messages
hosomaki explain --dmesg

# Any log file
hosomaki explain --file /var/log/nginx/error.log
hosomaki explain --file /var/log/syslog

# Quick line with copy-paste
hosomaki explain "kernel: OOM killer activated on process nginx"
```

### `status`

Quick health snapshot. Collects data and summarises everything.

```bash
hosomaki status           # paragraph summary
hosomaki status --brief   # single sentence
```

### `shell-integration`

Installs a small wrapper so any command prefixed with explain will be analysed if it fails.

```bash
hosomaki shell-integration --shell bash >> ~/.bashrc && source ~/.bashrc
hosomaki shell-integration --shell zsh  >> ~/.zshrc  && source ~/.zshrc
hosomaki shell-integration --shell fish >> ~/.config/fish/config.fish
```

Then just prefix any command with `explain`:

```bash
explain sudo systemctl start nginx
explain make build
explain docker compose up
```

If the command fails, hosomaki explains the error automatically.

### Coming soon

These are planned features.

**`hosomaki doctor`**  
Full system diagnosis with concrete suggested actions. Instead of just describing what it sees like `status` does, it tells you what to actually do about it.

**`hosomaki predict`**  
Spots potential failures before they happen by analysing patterns across logs, services, and system behaviour over time.

**`hosomaki audit`**  
Surfaces invisible system changes like files modified, services added or removed, permission changes, new processes, package updates.

**`hosomaki trace <process>`**  
Intelligent tracing of a running process like syscalls, resource usage, relevant events, with a plain-language explanation of what it's doing and whether anything looks wrong.

---

**`hosomaki explain --since / --until`**  
Explanations restricted to a specific time window.

**`hosomaki explain --context`**  
Pull logs from multiple related services at once, for example nginx + alloy + prometheus, to understand cascading failures.

**`hosomaki explain --diff`**  
Compare logs from two boots side by side and explain what changed.

---

**`hosomaki watch <service>`**  
Real‑time log tailing.

**`hosomaki compare --service <name> --boot -1`**  
Compare a service's behaviour between the current boot and the previous one.

**`hosomaki why <exit-code> --service <name>`**  
Given a nonzero exit code, pull surrounding context and explain the failure chain.

**`hosomaki summarise --since <time>`**  
Digest of everything that went wrong in a time window, grouped by severity levels.

---

**`hosomaki history`**  
Local log of past explanations, so the user can revisit insights without re-running the model and saving some time. This can be in `~/.local/share/hosomaki/`.

**`hosomaki alias`**  
Save long invocations under short names, for example `hosomaki alias nginx-errors "explain --service nginx --lines 100"`. This is telling hosomaki: “Create a new mini‑command called *nginx-errors* that internally runs `hosomaki explain --service nginx --lines 100`.”

**`--output json`**  
Instead of explaining things in plain language, it returns clean JSON object that the user can pipe into another tool.

---

**Memory layer for semantic search and long-term context (RAG)**  
Hosomaki will optionally persist system snapshots, log explanations, and error patterns in a local vector database.

When enabled, the memory layer powers:

- `hosomaki history` —> semantic search over past explanations. `hosomaki history "OOM nginx"` surfaces all previous incidents involving memory pressure on nginx, even if the original logs used different wording.
- `hosomaki predict` —> the current system state is compared against historical snapshots. If similar past states were followed by a failure, hosomaki warns you before it happens.
- `hosomaki explain` —> when the same error pattern has appeared before, the model receives the previous explanation and whether the suggested fix worked, producing progressively better answers for recurring issues.
- `hosomaki audit` —> behavioural drift detection across snapshots over time.

Storage: **pgvector** implementation for those who want it, and a lightweight local fallback using SQLite with embeddings generated by Ollama's `/api/embeddings` endpoint. The memory layer is opt-in and can be disabled entirely.

---

**Daemon mode for continuous monitoring and instant insights**  
Hosomaki will include an optional background service that continuously monitors logs, services, resource usage, and system events.  

This is not a user-facing command:
- Real-time anomaly detection  
- Proactive warnings  
- Continuous predictions  
- Instant `status` and `doctor` responses  

---

**`hosomaki ports`**  
List listening ports with process names. It will flag anything unusual or unexpected.

**`hosomaki crons`**  
Parse system + user crontabs, explain what each job does, and when it last ran or failed.

**`hosomaki mounts`**  
Check mount health, detect stale NFS mounts, full disks approaching thresholds, and slow mountpoints.

**`hosomaki timers`**  
Systemd timer inspection. This would be the modern equivalent of cron analysis.

---

**`hosomaki env-check`**  
Scan `/proc/<pid>/environ` for common misconfigurations.

---

**`hosomaki insight`**  
Toggle a deterministic, AI‑free analysis mode. When enabled, all commands run without invoking a model. This will perform log filtering, correlation, anomaly detection, diffing and heuristics using rule‑based logic only.  
This one is useful for servers without local models.  
Can be activated or deactivated at any time:

hosomaki insight on     # without AI
hosomaki insight off    # with AI

---

**Multi‑provider AI support**  
Hosomaki's AI layer will become fully pluggable:

- **OpenAI** (API key, optional, never sending logs unless explicitly allowed)
- **Anthropic** (Claude models)
- **Other local or remote providers** via a unified interface

Fallback to `hosomaki insight`.

---

## Requirements

- Linux (systemd-based distro recommended)
- Go 1.22+
- [Ollama](https://ollama.com) running locally with a model pulled

> Ollama is currently the only supported provider. Support for OpenAI, Anthropic,
> and any OpenAI-compatible endpoint (LM Studio, llama.cpp, etc.) is planned.
> An opt-in cloud analysis path is also planned, logs are scrubbed locally before
> anything leaves the machine. `hosomaki insight` will work with no provider at all.
> `hosomaki insight` will work with no provider at all.

## Installation

### Install Ollama

**Native (recommended):**

```bash
curl -fsSL https://ollama.com/install.sh | sh
```

On most distros this registers a systemd service that starts automatically.
If it isn't running yet:

```bash
ollama serve
```

**Docker:**

```bash
docker run -d -p 11434:11434 --name ollama ollama/ollama
```

### Pull a model

```bash
ollama pull llama3
```

Any model works.

If using Docker:

```bash
docker exec -it ollama ollama pull llama3
```

### Install Hosomaki

```bash
git clone https://github.com/rivernova/hosomaki.git
cd hosomaki
make build
sudo make install
```

## Configuration

```yaml
# ~/.config/hosomaki/config.yaml
ai:
  provider: ollama
  endpoint: http://localhost:11434
  model: llama3
  timeout: 120s        # increase for slow hardware or large models
output:
  color: true
  language: en
```

All values can also be set via `.env`:

```bash
HOSOMAKI_AI_MODEL=mistral hosomaki status
HOSOMAKI_AI_ENDPOINT=http://192.168.1.10:11434 hosomaki explain --service nginx
```

## Status

Early development. See [CONTRIBUTING.md](CONTRIBUTING.md) if you want to help.

## License

[Mozilla Public License 2.0](LICENSE)
