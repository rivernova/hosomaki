# Configuration

Hosomaki is zero-configuration by default. The optional config file lets you override the Ollama endpoint and model.

## Config file

Hosomaki uses [Viper](https://github.com/spf13/viper) and searches for configuration in the following locations (highest priority first):

1. `$XDG_CONFIG_HOME/hosomaki/config.yaml`
2. `~/.hosomaki.yaml`
3. `./hosomaki.yaml` (current directory)

### Example

```yaml
# ~/.hosomaki.yaml

# Ollama model to use for all commands.
# Any model available via `ollama list` is valid.
model: llama3.2

# Ollama API endpoint.
# Override if Ollama is running on a different host or port.
ollama_url: http://localhost:11434
```

## Available settings

| Key | Default | Description |
|---|---|---|
| `model` | `llama3.2` | Ollama model tag |
| `ollama_url` | `http://localhost:11434` | Ollama API base URL |

## Environment variables

Every config key can be overridden with an environment variable prefixed `HOSOMAKI_`:

```bash
HOSOMAKI_MODEL=qwen2.5 hosomaki status
HOSOMAKI_OLLAMA_URL=http://gpu-box:11434 hosomaki doctor
```

## Per-command flags

Some commands expose additional flags that take precedence over config:

```bash
hosomaki explain --service nginx --lines 100 --since "2 hours ago"
hosomaki watch nginx --window 15s --max-lines 50
hosomaki audit --baseline /tmp/custom-baseline.json
```

Run `hosomaki <command> --help` for the full flag reference.

## Audit baseline location

The audit baseline is stored separately from the main config:

- Default: `$XDG_DATA_HOME/hosomaki/audit-baseline.json` (falls back to `~/.local/share/hosomaki/audit-baseline.json`)
- Override per invocation: `hosomaki audit --baseline /path/to/baseline.json`