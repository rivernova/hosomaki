# Daemon Configuration

::: tip Roadmap
Daemon mode is planned for **Phase 3** of the Hosomaki roadmap.
:::

## Planned configuration schema

The daemon configuration will extend the existing `~/.hosomaki.yaml` with a `daemon` stanza:

```yaml
# Existing CLI config
model: gemma3:4b
ollama_url: http://localhost:11434

# Planned daemon config
daemon:
  socket: /run/user/1000/hosomaki.sock
  watch:
    - nginx
    - postgresql
    - myapp
  schedule:
    daily_digest: "07:00"
  thresholds:
    disk_warn_pct: 80
    memory_warn_pct: 90
```

This schema is subject to change before implementation.