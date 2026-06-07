# Security Policy

## Supported Versions

Hosomaki is pre-1.0 and under active development. Only the latest commit on
`main` receives security fixes. There are no backport guarantees for older
commits or tagged releases.

## Threat Model

Hosomaki is a local-only CLI tool. Understanding what it protects against — and
what it does not — helps set the right expectations.

**Within scope:**

- Log data leaving the local machine. Hosomaki never opens a network connection
  except to the Ollama endpoint configured by the user, which defaults to
  `localhost:11434`. No telemetry, analytics, or remote logging exists anywhere
  in the codebase.
- PII reaching the language model. Every pipeline applies a mandatory
  sanitisation pass before log content enters a prompt. This strips IP
  addresses, hostnames, paths, UUIDs, email addresses, MAC addresses, URLs, and
  package version strings, replacing them with opaque placeholders.
- Privilege misuse. Hosomaki runs entirely as the invoking user. It does not
  install setuid binaries, does not write outside `~/.config/hosomaki/`, and
  does not fork persistent background processes.

**Out of scope:**

- The security of the Ollama process itself or the model weights it serves.
  Hosomaki treats Ollama as a trusted local endpoint. Hardening Ollama is the
  user's responsibility.
- Full compromise of the host system. If an attacker already has code execution
  on the machine, Hosomaki provides no additional attack surface worth
  mentioning.
- The accuracy or safety of model output. Hosomaki validates the *shape* of
  model responses against a strict schema, but makes no claims about the
  factual correctness of the analysis produced.

## Known Limitations

Sanitisation is best-effort. The sanitiser operates on text patterns and
heuristics, it cannot guarantee that every sensitive value in every log format
will be detected and masked. Users should not feed logs that contain secrets
they would be seriously harmed by exposing to a local language model, such as
plaintext passwords, private keys, or session tokens that appear literally in
log output.

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Report security issues privately by one of the following methods:

- **GitHub private vulnerability reporting:**
  [Report a vulnerability](https://github.com/rivernova/hosomaki/security/advisories/new)

### What to include

A useful report contains:

- A clear description of the vulnerability and its potential impact
- Steps to reproduce, including OS, kernel version, and Hosomaki version
  (`hosomaki --version`)
- The command invocation or input that triggers the issue
- What you expected to happen and what actually happened
- Any relevant log output (sanitise anything sensitive before sending)

### Response commitment

Fixes for confirmed vulnerabilities will be prioritised based on severity and
merged to `main` as soon as practical. Credit will be given in the release notes
unless you prefer to remain anonymous.