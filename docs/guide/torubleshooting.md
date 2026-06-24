# Troubleshooting Slow Responses

Hosomaki doesn't run any inference itself. It collects data, sanitises it, and hands it to your local model via Ollama. So when something feels slow, the actual work happening is almost always inside Ollama, not Hosomaki. Here's how to tell the difference and what's usually going on.

```
collect → sanitise → prompt* → validate → render
                         ↑          │
                         └─repair*──┘
                          (up to 3x)

* = calls Ollama
```

## Read the spinner first

Every command shows a spinner, and the label changes as things progress: it starts on something like `thinking…`, switches to `responding…` once tokens start streaming back, and occasionally jumps to `repairing (attempt N)…`. As long as the label keeps moving, nothing is stuck, the model is just working. If it's the timeout you're hitting (`ai.timeout`, 120s by default), you'll get an explicit error rather than a silent hang.

The `repairing` label is worth paying attention to. It means the model's response didn't pass schema validation and Hosomaki had to ask it again. This is an extra full model call, not free. More on that below.

## A quick sanity check

The fastest way to know whether it's Hosomaki or Ollama is to time them separately.

```bash
time ollama run llama3.2:3b "Say hello in one sentence."
time hosomaki status
```

If the raw `ollama run` call is already slow, that's your answer, it's the model or the hardware, and Hosomaki's own overhead (collecting, sanitising, validating) is small in comparison, usually well under a second. If `ollama run` comes back quickly but `hosomaki status` doesn't, that's worth a closer look. Please grab `--debug` output and [open an issue](https://github.com/rivernova/hosomaki/issues).

## Why it might be slow

**No GPU offload.** Ollama silently falls back to CPU if it can't use your GPU, and CPU inference is a different world performance-wise. Run `ollama ps` and check the `PROCESSOR` column. If you get `100% CPU`, that means no GPU is involved at all. If you expected GPU acceleration, check that drivers are installed and that the model actually fits in your VRAM.

**Model too big for available memory.** If a model doesn't fully fit, Ollama either spills to system RAM or splits across CPU/GPU, and both are slow. Try a smaller model as a baseline and see if it changes anything:

```bash
ollama pull llama3.2:3b
HOSOMAKI_AI_MODEL=llama3.2:3b hosomaki status
```

The [model table in the README](https://github.com/rivernova/hosomaki#recommended-ollama-models) has the speed/quality trade-offs across a few common options.

**Cold start.** Ollama unloads a model from memory after a few minutes of inactivity (`OLLAMA_KEEP_ALIVE`), so the first command after a break will always be slower than the next one. If every single run is slow regardless of timing, this isn't your issue. So, again, look at GPU offload or model size instead.

**The model keeps needing repairs.** Hosomaki asks for strict JSON matching a schema, and weaker models occasionally get it wrong on the first try. Each retry is a full extra round-trip to the model — up to three attempts before Hosomaki gives up, so worst case that's four model calls for one command. If you're regularly seeing `repairing (attempt 1)…` or higher, it's usually a sign the model isn't great at structured output rather than anything else. `mistral:7b` and `qwen3:8b` tend to be more reliable here than `llama3.2:3b`.

**Timeout set too low for the hardware.** `ai.timeout` covers the whole request, and on slower setups with bigger models, especially with a repair attempt added in, it can genuinely need most of that window. Bump it before assuming something's broken:

```yaml
ai:
  timeout: 300s
```

## Still not sure?

Run the command with `--debug`:

```bash
hosomaki status --debug
```

This prints the raw model output and validation result for every attempt, repairs included. If the JSON coming back looks fine and arrives promptly but the command still drags, that's actually worth reporting. If instead you see slow token generation or repeated repairs in the debug output, that points back to the model. At this point [opening an issue](https://github.com/rivernova/hosomaki/issues) with that `--debug` output attached is the most useful next step.