# Configuration

Hosomaki is zero-configuration by default. The optional config file lets you override the LLM provider, model, endpoint, and output behavior.

## Config file

Hosomaki searches for a config file named `config.yaml` in:

1. `~/.config/hosomaki/`
2. `$HOME`

### Example

```yaml
# ~/.config/hosomaki/config.yaml
ai:
  provider: ollama
  endpoint: http://localhost:11434
  model: llama3.1:8b
  timeout: 120s        # increase for slow hardware or large models
output:
  color: true
  language: en
```

| Key | Default | Description |
|---|---|---|
| `ai.provider` | `ollama` | AI backend (currently only `ollama` is supported) |
| `ai.endpoint` | `http://localhost:11434` | Ollama API endpoint |
| `ai.model` | `llama3.1:8b` | Any model available via `ollama list` |
| `ai.timeout` | `120s` | Request timeout |
| `output.color` | `true` | Colorized terminal output |
| `output.language` | `en` | Language for AI responses |

## Environment variables

Every config key can be overridden with an environment variable prefixed `HOSOMAKI_`, using the nested key path:

```bash
HOSOMAKI_AI_MODEL=mistral hosomaki status
HOSOMAKI_AI_ENDPOINT=http://gpu-box:11434 hosomaki doctor
```