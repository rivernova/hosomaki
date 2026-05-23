# Hosomaki
<p align="center">
  <img src="assets/hosomaki_logo.svg" alt="Hosomaki" width="350"/>
</p>

> Local AI intelligence layer for Linux.

Hosomaki reads your system — logs, processes, services — and uses a local AI model to explain what's happening in plain language. No cloud. No telemetry.

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