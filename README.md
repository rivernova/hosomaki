# Hosomaki

**Your Linux system has a story to tell. Hosomaki is the moment it finally finds its voice.**

<p align="center">
  <img src="assets/hosomaki_logo.png" alt="Hosomaki"/>
</p>

> Local intelligence layer for Linux.

Hosomaki reads your system and helps you understand what's happening in plain language. No cloud. No telemetry. Your system, your data, your choice.

It uses a local model via [Ollama](https://ollama.com) and never sends anything off your machine.

---

## Commands

### `explain`

Understands what's going on. Adapts to whatever you throw at it.

```bash
# Pipe any log output directly
journalctl -p err -n 20 | hosomaki explain
dmesg | tail -50         | hosomaki explain

# By systemd service. hosomaki fetches the logs for you
hosomaki explain --service nginx
hosomaki explain --service postgresql --lines 100

# Errors from a specific boot. Useful after a crash
hosomaki explain --boot
hosomaki explain --boot -1        # the boot before that

# Kernel messages
hosomaki explain --dmesg

# Any log file
hosomaki explain --file /var/log/nginx/error.log
hosomaki explain --file /var/log/syslog

# Quick one-liner
hosomaki explain "kernel: OOM killer activated on process nginx"
```

### `status`

Quick health snapshot. Collects uptime, memory, disk, failed services, and recent errors, then summarises everything.

```bash
hosomaki status           # paragraph summary
hosomaki status --brief   # single sentence
```

### `doctor`

Full system diagnosis with concrete suggested actions. Unlike `status`, which only describes what it sees, for each detected issue it explains the likely cause and proposes specific actions like commands to run, files to inspect, configuration values to change.

If a suggested action is potentially disruptive or irreversible, the output says so explicitly. Doctor never modifies the system itself.

```bash
hosomaki doctor           # full diagnosis with suggested actions
hosomaki doctor --brief   # one sentence per issue
```

### `shell-integration`

Installs a small shell wrapper. Any command prefixed with `explain` will be analysed automatically if it fails.

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

---

## Coming soon

See the [Roadmap](https://github.com/rivernova/hosomaki/wiki) for the full plan.

---

## Requirements

- Linux (systemd-based distro recommended)
- Go 1.23+
- [Ollama](https://ollama.com) running locally with a model pulled

## Installation

### Install Ollama

**Native (recommended):**

```bash
curl -fsSL https://ollama.com/install.sh | sh
```

On most distros this registers a systemd service that starts automatically. If it isn't running yet:

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

Any model works. If using Docker:

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

Copy the example config and edit as needed:

```bash
cp config.example.yml ~/.config/hosomaki/config.yaml
```

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

Environment variables are also supported, prefixed with `HOSOMAKI_`:

```bash
HOSOMAKI_AI_MODEL=mistral hosomaki explain --service nginx
```

## Development

```bash
make build    # build binary to ./bin/hosomaki
make test     # run tests
make lint     # run linter (requires golangci-lint)
make dev      # run without building (go run)
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## Status

Early development. The core commands (`explain`, `status`, `doctor`, `shell-integration`) are stable. Everything else is in progress.

## License

[Mozilla Public License 2.0](LICENSE)