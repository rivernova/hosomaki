# Installation

Hosomaki runs on Linux and requires [Ollama](https://ollama.com).

## Prerequisites

| Requirement | Notes |
|---|---|
| Linux | systemd-based distributions recommended |
| [Ollama](https://ollama.com) | Runs the local language model |
| A compatible Ollama model | `llama3.2` or `qwen2.5` recommended |

### Install Ollama

```bash
curl -fsSL https://ollama.com/install.sh | sh
```

Pull a model before running Hosomaki for the first time:

```bash
ollama pull llama3.2
```

## Install Hosomaki

### From source

```bash
git clone https://github.com/rivernova/hosomaki.git
cd hosomaki
make build
sudo make install
```

### Verify

```bash
hosomaki --version
hosomaki status
```

## Configuration

Hosomaki works out of the box with no configuration required. The default Ollama endpoint (`http://localhost:11434`) and model (`llama3.2`) are used automatically.

To override, create `~/.hosomaki.yaml`:

```yaml
model: qwen2.5
ollama_url: http://localhost:11434
```

See [Configuration](/guide/configuration) for the full reference.

::: tip Root privileges
Some commands produce richer output when run as root, because reading other users' crontabs and some socket metadata requires elevated privileges. Running as a regular user still works, but scope may be limited.
:::