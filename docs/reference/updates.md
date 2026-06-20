# updates

Explain pending package updates before applying them.

## Usage

```bash
hosomaki updates [flags]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--security-only` | `false` | Show only security-related updates |

## Output

For each pending update, Hosomaki flags:

- Whether it's a **security** fix, a **major** version bump, a **minor** update, or **unknown**
- Whether it requires a reboot to take effect
- Detail explaining what changed, for security and major updates

Read-only. `updates` only lists and explains pending updates. It never applies them.

## Supported package managers

`updates` detects the active package manager and adapts accordingly:

`apt`, `dnf`, `yum`, `pacman`, `zypper`, `apk`, `xbps`, `emerge`, `nix`

If none of these are detected, `updates` returns an error rather than guessing.

## Examples

```bash
hosomaki updates
hosomaki updates --security-only
```