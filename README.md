# Hosomaki
<p align="center">
  <img src="assets/hosomaki_logo.svg" alt="Hosomaki" width="350"/>
</p>

> Local intelligence layer for Linux — with and without AI.

Hosomaki reads your system — logs, processes, services — and helps you understand what's happening in plain language.  
It works in two modes:

- **AI‑powered mode**: uses a local model (via Ollama) to explain issues, correlate events and summarise system behaviour.  
- **Insight mode**: a deterministic, AI‑free analysis path based on rules, heuristics, diffs and structured inspection.

No cloud. No telemetry. Your system, your data, your choice.


## Commands

### `explain` — understand what's going on

The most flexible command. Several ways to use it, no copy-pasting required:

```bash
# Pipe any log output directly
journalctl -p err -n 20 | hosomaki explain
dmesg | tail -50         | hosomaki explain

# By systemd service — hosomaki fetches the logs for you
hosomaki explain --service nginx
hosomaki explain --service postgresql --lines 100

# Errors from the last boot (useful after a crash)
hosomaki explain --boot
hosomaki explain --boot -1        # the boot before that

# Kernel messages (OOM, hardware errors, driver issues)
hosomaki explain --dmesg

# Any log file
hosomaki explain --file /var/log/nginx/error.log
hosomaki explain --file /var/log/syslog

# Quick one-liner
hosomaki explain "kernel: OOM killer activated on process nginx"
```

### `status` — system health at a glance

Collects a snapshot of uptime, memory, disk, failed services and recent errors, then asks the AI to summarise what's going on.

```bash
hosomaki status           # paragraph summary
hosomaki status --brief   # single sentence
```

### `shell-integration` — explain failures automatically

Installs a shell function that wraps any command and explains it if it fails:

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

If the command fails, hosomaki explains the error automatically. If it succeeds, output passes through unchanged.

---

### `completion` — shell autocompletion

Generate and install a shell completion script so your shell can tab-complete `hosomaki` subcommands and flags.

Example installation commands:

```bash
# Bash
hosomaki completion bash >> ~/.bash_completion.d/hosomaki
source ~/.bash_completion.d/hosomaki

# Zsh
hosomaki completion zsh > ~/.zsh/completions/_hosomaki
fpath=(~/.zsh/completions $fpath)
autoload -U compinit && compinit

# Fish
hosomaki completion fish > ~/.config/fish/completions/hosomaki.fish
```

### Coming soon

These commands are planned and actively being designed.

**`hosomaki doctor`**  
Full system diagnosis with concrete suggested actions. Goes beyond `status` — instead of describing what it sees, it tells you what to actually do about it.

**`hosomaki predict`**  
Spots potential failures before they happen by analysing patterns across logs, services, and system behaviour over time. Useful for catching things like slow disk degradation, memory leaks, or services on the edge of crashing.

**`hosomaki audit`**  
Surfaces invisible system changes: files modified, services added or removed, permission changes, new processes, package updates. Answers the question "what changed since yesterday?".

**`hosomaki trace <process>`**  
Intelligent tracing of a running process — syscalls, resource usage, relevant events — with a plain-language explanation of what it's doing and whether anything looks wrong.

---

**`hosomaki explain --since / --until`**  
Time‑bounded log analysis (journalctl‑compatible), allowing explanations restricted to a specific time window.

**`hosomaki explain --context`**  
Pull logs from multiple related services at once (e.g. nginx + php‑fpm + postgres) to understand cascading failures.

**`hosomaki explain --diff`**  
Compare logs from two boots side by side and explain what changed.

---

**`hosomaki watch <service>`**  
Real‑time log tailing with noise suppression — only surfaces lines the AI considers noteworthy.

**`hosomaki compare --service <name> --boot -1`**  
Compare a service’s behaviour between the current boot and the previous one.

**`hosomaki why <exit-code> --service <name>`**  
Given a nonzero exit code, pull surrounding context and explain the failure chain.

**`hosomaki summarise --since <time>`**  
Digest of everything that went wrong in a time window, grouped by severity.

---

**`hosomaki history`**  
Local log of past explanations (stored in `~/.local/share/hosomaki/`), so you can revisit insights without re-running the model.

**`hosomaki alias`**  
Save long invocations under short names (e.g. `hosomaki alias nginx-errors "explain --service nginx --lines 100"`).

**`--output json`**  
Structured output across all commands for piping into other tools or scripts.

---

**`hosomaki ports`**  
List listening ports with process names; AI flags anything unusual or unexpected for the system’s profile.

**`hosomaki crons`**  
Parse system + user crontabs, explain what each job does, and when it last ran or failed.

**`hosomaki mounts`**  
Check mount health, detect stale NFS mounts, full disks approaching thresholds, and slow mountpoints.

**`hosomaki timers`**  
Systemd timer inspection — the modern equivalent of cron analysis.

---

**`hosomaki env-check`**  
Scan `/proc/<pid>/environ` for common misconfigurations (empty secrets, default passwords, exposed tokens). Detection is rule‑based; AI explains the risk and impact.

---

**`hosomaki insight`**  
Toggle a deterministic, AI‑free analysis mode. When enabled, all commands run without invoking a model: log filtering, correlation, anomaly detection, diffing and heuristics are performed using rule‑based logic only.  
Useful for servers without local models, restricted environments, or users who prefer predictable, explainable behaviour.  
Can be activated or deactivated at any time:

hosomaki insight on     # all analysis runs without AI
hosomaki insight off    # restore AI-powered explanations

---

**Multi‑provider AI support**  
Hosomaki’s AI layer will become fully pluggable.  
In addition to the current local‑first Ollama integration, support is planned for:

- **OpenAI** (API key, optional, never sending logs unless explicitly allowed)
- **Anthropic** (Claude models)
- **Other local or remote providers** via a unified interface

The goal is to let users choose their preferred backend, switch providers at runtime, or fall back to `hosomaki insight` for deterministic analysis.  
AI becomes an optional module — not a requirement.

---

## Requirements

- Linux (systemd-based distro recommended)
- Go 1.22+
- [Ollama](https://ollama.com) running locally with a model pulled (e.g. `ollama pull llama3`)

## Installation

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

All values can also be set via environment variables:

```bash
HOSOMAKI_AI_MODEL=mistral hosomaki status
HOSOMAKI_AI_ENDPOINT=http://192.168.1.10:11434 hosomaki explain --service nginx
```

## Status

Early development. See [CONTRIBUTING.md](CONTRIBUTING.md) if you want to help.

## License

[Mozilla Public License 2.0](LICENSE)