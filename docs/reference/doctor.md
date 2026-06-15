# doctor

Full system diagnosis with concrete suggested actions.

## Usage

```bash
hosomaki doctor [flags]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--brief` | `false` | One-sentence summary instead of full output |
| `--debug` | `false` | Print raw model response to stderr |

## Scope

`doctor` goes further than `status` and produces concrete suggested actions for each identified issue.

## Examples

```bash
hosomaki doctor
hosomaki doctor --brief
```