# Installation

Hosomaki runs on Linux and requires [Ollama](https://ollama.com) with a model pulled.

## Prerequisites

| Requirement | Notes                                                                                      |
|---|--------------------------------------------------------------------------------------------|
| Linux | systemd-based distributions recommended                                                    |
| [Ollama](https://ollama.com) | Runs the local language model                                                              |
| A compatible Ollama model | `llama3.2:3b` (default) or `qwen2.5` recommended                                           |
| Go 1.23+ | **Only** needed to build from source. Not necessary for the `.deb`/`.rpm` packages |

## Install Ollama

### Native (recommended)

```bash
curl -fsSL https://ollama.com/install.sh | sh
```

On most distributions this registers a systemd service that starts automatically. If it isn't running yet, start it with `ollama serve`.

### Docker with GPU (optional)

If you prefer to run Ollama in a container, the repository ships an optional Compose file at [`deploy/docker-compose.yml`](https://github.com/rivernova/hosomaki/blob/main/deploy/docker-compose.yml) with an NVIDIA GPU profile. Hosomaki never starts it for you, so bring it up explicitly:

```bash
make ollama-up      # docker compose up -d, using the GPU if available
make ollama-down    # stop it
```

The GPU profile requires the [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html). Remove the `deploy.resources` block in the Compose file to run on CPU only.

### Pull a model

```bash
ollama pull llama3.2:3b
```

Any instruction-tuned model works, larger models produce better results, smaller ones are faster.

## Recommended models

Hosomaki works best with instruction-tuned local models. Model choice trades off speed against quality for your hardware:

| Model | Best for | Notes |
|---|---|---|
| `llama3.2:3b` | Fast responses, low resource | Default; lightweight summarisation and log tasks |
| `gemma3:4b` | Balanced | Large context window, multilingual support |
| `mistral:7b` | General-purpose | Strong instruction-following 7B model |
| `qwen3:8b` | Higher-quality reasoning & summaries | Requires more RAM/VRAM |

If a command feels slow, it's almost always Ollama or the model, not Hosomaki — see [Troubleshooting](/guide/troubleshooting).

## Install Hosomaki

### From a package (recommended)

Download the `.deb` or `.rpm` for your architecture from the [latest release](https://github.com/rivernova/hosomaki/releases), then install it:

```bash
# Debian / Ubuntu
sudo apt install ./hosomaki_<version>_amd64.deb

# Fedora / RHEL
sudo dnf install ./hosomaki_<version>_amd64.rpm
```

Packages install the binary to `/usr/bin/hosomaki` and need no Go toolchain. Release checksums are signed with [cosign](https://github.com/sigstore/cosign). See the release assets to verify the download.

### From source

Building from source requires **Go 1.23 or newer**.

```bash
git clone https://github.com/rivernova/hosomaki.git
cd hosomaki
make build      # compiles to ./bin/hosomaki (verifies your Go version first)
make install    # copies to /usr/local/bin (prompts for sudo on the copy step only)
```

Do **not** run `make build` with `sudo`. Compiling as root pollutes the Go cache under `/root` and leaves root-owned artifacts. Only the copy in `make install` needs elevated privileges, and the target handles that itself.

### Verify

```bash
hosomaki --version
hosomaki status
```

## Configuration

Hosomaki is zero-configuration by default and works out of the box. To customise the model, endpoint, output, or environment-variable overrides, see the **[Configuration guide](/guide/configuration)**.

::: tip Root privileges
Some commands produce richer output when run as root, because reading other users' crontabs and some socket metadata requires elevated privileges. Running as a regular user still works, but scope may be limited.
:::