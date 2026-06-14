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

## Difference from `status`

`status` surfaces observations — what is happening right now. `doctor` goes further and produces concrete suggested investigation steps for each identified issue.

Neither command modifies the system. The AI is explicitly instructed not to suggest specific remediation commands to run, only what to investigate and why.

## Examples

```bash
hosomaki doctor
hosomaki doctor --brief
```