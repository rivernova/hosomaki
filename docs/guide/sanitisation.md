# Sanitisation

Hosomaki's sanitisation layer is the first and mandatory step of every AI pipeline invocation. It strips sensitive data from system output before it ever reaches the language model.

## What gets scrubbed

| Category | Examples |
|---|---|
| IPv4 addresses | `192.168.1.1`, `10.0.0.254` |
| IPv6 addresses | `::1`, `fe80::1` |
| Hostnames and FQDNs | `myserver.internal`, `db.prod.example.com` |
| Filesystem paths | `/home/alice/`, `/etc/secrets/` |
| UUIDs | `550e8400-e29b-41d4-a716-446655440000` |
| Usernames | As appearing in log lines |
| Credentials and tokens | API keys, passwords in log output |

## Architectural position

Sanitisation happens in the `cmd/` layer — inside each command's `RunE` function — before the prompt package is called:

```go
san := sanitiser.Default()
sanitisedLogs := san.Sanitise(rawLogs)

generationPrompt := prompt.Explain(prompt.ExplainInput{
    Logs: sanitisedLogs,
    // ...
})
```

The `sanitiser` package is deliberately never imported by `internal/prompt`. The prompt package always receives pre-sanitised data and has no knowledge of what was stripped.

## Per-line sanitisation

For streaming use cases like `watch`, `sanitiser.DefaultPerLine()` applies the same scrubbing rules to each journal line individually as it arrives from the tail, before it enters the batch buffer.

## What sanitisation does not do

- It does not anonymise log structure or timing information
- It does not prevent the model from inferring general system characteristics
- It is not a substitute for ensuring your Ollama instance is secured at the network level

## Privacy guarantee

Because Ollama runs entirely on your machine, and the sanitisation layer removes identifiable data before prompting, Hosomaki provides a strong practical privacy guarantee: **no sensitive system data leaves your machine**. The model never sees raw IPs, credentials, hostnames, or user paths.

See [Data Privacy](/guide/privacy) for the full data handling policy.