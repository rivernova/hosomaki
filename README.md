# Hosomaki
<p align="center">
  <img src="assets/hosomaki_logo.svg" alt="Hosomaki" width="350"/>
</p>

> Local AI intelligence layer for Linux.

Hosomaki reads your system — logs, processes, services — and uses a local AI model to explain what's happening in plain language. No cloud. No telemetry.

```bash
$ journalctl -p err -n 20 | hosomaki explain
$ hosomaki status
$ hosomaki explain "kernel: OOM killer activated on process nginx"
```

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
output:
  color: true
  language: en
```

## Status

Early development. See [CONTRIBUTING.md](CONTRIBUTING.md) if you want to help.

## License

[Mozilla Public License 2.0](LICENSE)