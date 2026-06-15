# Configuration

Hosomaki is zero-configuration by default. The optional config file lets you override the Ollama endpoint and model.

## Config file

Hosomaki searches for configuration in the following locations:

1. `$XDG_CONFIG_HOME/hosomaki/config.yaml`
2. `~/.hosomaki.yaml`
3. `./hosomaki.yaml` 

### Example

```yaml
# ~/.hosomaki.yaml

# Ollama model
# Any model available via `ollama list` is valid.
model: llama3.2

# Ollama API endpoint
# Override if Ollama is running on a different host or port.
ollama_url: http://localhost:11434
```

## Environment variables

Every config key can be overridden with an environment variable prefixed `HOSOMAKI_`:

```bash
HOSOMAKI_MODEL=qwen2.5 hosomaki status
HOSOMAKI_OLLAMA_URL=http://gpu-box:11434 hosomaki doctor
```

## Audit baseline location

The audit baseline is stored separately from the main config:

- Default: `$XDG_DATA_HOME/hosomaki/audit-baseline.json` (falls back to `~/.local/share/hosomaki/audit-baseline.json`)
- Override per invocation: `hosomaki audit --baseline /path/to/baseline.json`