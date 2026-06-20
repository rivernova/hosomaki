# Hosomaki

<img src="assets/hosomaki_logo.png" alt="Hosomaki"/>

Hosomaki reads your system and helps you understand what's happening in plain language. No cloud. No telemetry. Your system, your data, your choice.

It uses a local model via [Ollama](https://ollama.com) and never sends anything off your machine.

📖 **[See full documentation here→](https://rivernova.github.io/hosomaki/guide/introduction)**

---

## Commands

| Command | What it does |
|---|---|
| `explain` | Explain errors from a service, boot, log file, pipe, inline text, or a running process |
| `status` | Quick summary of current system health |
| `doctor` | Full diagnosis with concrete suggested actions |
| `audit` | Surface changes since a baseline snapshot |
| `watch` | Tail a service journal and explain errors in real time |
| `why` | Given an exit code and service, reconstruct the full failure chain |
| `ports` | List listening ports and flag anything unexpected |
| `timers` | Inspect all systemd timers and flag failures or overdue schedules |
| `crons` | Read all crontabs and explain what each job does |
| `mounts` | Inspect active mounts, detect stale NFS, and flag disks approaching capacity |
| `updates` | Explain pending package updates before applying them |
| `history` | Review past diagnostic results |
| `shell-integration` | Install a shell wrapper that explains failed commands automatically |

Run `hosomaki <command> --help` for flags and usage details.

---

### Quick examples

```bash
# General health check
hosomaki status
hosomaki status --brief                          # one-sentence summary
hosomaki doctor
hosomaki doctor --brief                          # one-sentence summary

# Explain logs. Adapts to whatever you throw at it
hosomaki explain --service nginx
hosomaki explain --service nginx --lines 100     # control how many lines to read
hosomaki explain --service nginx --since "1 hour ago"
hosomaki explain --service nginx --since "2024-01-15 14:00" --until "2024-01-15 15:00"
hosomaki explain --boot                          # current boot
hosomaki explain --boot -1                       # previous boot
hosomaki explain --dmesg                         # kernel ring buffer
hosomaki explain --file /var/log/syslog
hosomaki explain --context nginx,postgresql      # correlate multiple services at once
hosomaki explain --diff -1                       # compare current boot with the previous one
hosomaki explain --diff -2:-1                    # compare any two boots
hosomaki explain --pid 1234                      # what is this process doing right now
journalctl -p err -n 50 | hosomaki explain       # pipe any log output
hosomaki explain "kernel: OOM killer activated"  # inline message

# Explain why a service exited
hosomaki why 1 --service nginx
hosomaki why 137 --service myapp --lines 100
hosomaki why 1 --service nginx --since "10 min ago"

# Surface what changed on the system
hosomaki audit --init                            # take a baseline
hosomaki audit                                   # diff against it
hosomaki audit --dirs /etc,/usr/local/bin        # track additional directories
hosomaki audit --baseline /tmp/my-baseline.json  # use a custom baseline path

# Watch a service and explain errors as they arrive
hosomaki watch nginx
hosomaki watch nginx --lines 20                  # seed with number of lines. 0 skips seed
hosomaki watch nginx --window 10s --max-lines 30 # tune batching behaviour

# Review scheduled work
hosomaki ports
hosomaki timers
hosomaki crons

# Mount health, stale NFS, disk thresholds
 hosomaki mounts        
 
# Pending package updates
hosomaki updates
hosomaki updates --security-only                  # only security-related updates

# Review past diagnostic results
hosomaki history
hosomaki history --command explain                # filter by source command
hosomaki history --since 7d                       # only entries from the last week
hosomaki history --clear                          # wipe the log                          

# Auto-explain failed commands
hosomaki shell-integration --shell bash >> ~/.bashrc && source ~/.bashrc
explain make build
```

---

## Data Privacy & Security

Hosomaki is designed around a simple principle: your data should remain yours. Before anything reaches the local model, a mandatory sanitisation layer strips IPs, hostnames, paths, UUIDs, emails, and credentials from collected data.

See [Data Privacy](https://rivernova.github.io/hosomaki/guide/privacy), [Sanitisation](https://rivernova.github.io/hosomaki/guide/sanitisation), and [SECURITY.md](SECURITY.md) for the full data handling policy and threat model.

---

## Data Flow & Processing Pipeline

<img src="assets/hosomaki_flowchart.svg" alt="Dataflow"/>

Each command sanitises its input locally, sends it to the local model, then validates and repairs the response against a strict schema before rendering. See [Architecture](https://rivernova.github.io/hosomaki/guide/architecture) for the full pipeline breakdown.

---

## Coming soon

See the [Roadmap](https://github.com/rivernova/hosomaki/wiki) for the full plan.

---

## Requirements

- Linux (systemd-based distro recommended)
- Go 1.23+
- [Ollama](https://ollama.com) running locally with a model pulled

---

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
ollama pull gemma3:4b
```

Any model works. Larger models produce better results, smaller models are faster.

If using Docker:

```bash
docker exec -it ollama ollama pull gemma3:4b
```

### Install Hosomaki

```bash
git clone https://github.com/rivernova/hosomaki.git
cd hosomaki
make build
sudo make install
```

### Recommended Ollama Models

Hosomaki works best with instruction-tuned local models for text generation, summarisation, and log parsing. Model choice depends on your hardware and desired trade-off between speed and quality.

| Model | Best for | Notes |
|---|---|---|
| `llama3.2:3b` | Fast responses, low resource | Lightweight summarisation and log tasks |
| `gemma3:4b` | Balanced | Large context window, multilingual support |
| `mistral:7b` | General-purpose | Strong instruction-following 7B model |
| `qwen3:8b` | Higher-quality reasoning & summaries | Requires more RAM/VRAM |

---

### Configuration

```bash
mkdir -p ~/.config/hosomaki && cp config.example.yml ~/.config/hosomaki/config.yaml
```

See the [configuration guide](https://rivernova.github.io/hosomaki/guide/configuration) for the full schema and environment variable overrides.

## Development

```bash
make build    # build binary to ./bin/hosomaki
make test     # run tests
make lint     # run linter (requires golangci-lint)
make dev      # run without building (go run)
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

---

## Status

Early development. The core commands (`explain`, `status`, `doctor`, `shell-integration`) are stable. Everything else is in progress.

## License

[Mozilla Public License 2.0](LICENSE)
