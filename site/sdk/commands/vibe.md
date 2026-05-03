---
title: glacier vibe
---

# glacier vibe    [ SDK ]

[ View source spec → ](../../../specs/0032-sdk.md#commands-vibe)
**Other commands:** [version](./version.md) [generate](./generate.md) [lint](./lint.md) [test](./test.md) [init](./init.md) [new](./new.md) [completions](./completions.md) [explain](./explain.md)

<!-- magpie:extract source=specs/0032-sdk.md section=commands subsection=vibe source-checksum=<TODO> -->
**Synopsis.** Run the Glacier vibes animation: dancing polar bear plus wordmark banner. Press any key to exit.

**Mental model.** `vibe` is the SDK's flagship dogfood of `term.Animator`. The animator owns a slog handler, intercepts log records, runs a 10 Hz frame loop. Each frame composes the polar-bear block-character mascot (5 lines, left side) with the ANSI Shadow GLACIER wordmark (right side, gradient applied). The bear's expression cycles every 3 seconds (calm -> confident -> thinking -> calm). The wordmark gradient offsets one line per 100 ms tick, producing a slow vertical shimmer. Below the banner a tip line rotates every 5 seconds from a curated 12-tip registry. Pressing any key (or Ctrl-C, or sending SIGTERM) exits cleanly.

**Flags.**

| Flag | Default | Description |
|---|---|---|
| `--duration` | `0s` | Bound the loop. `0s` means run until key press. |
| `--no-tips` | `false` | Suppress the rotating tip line. |
| `--seed` | `0` | Seed the tip-rotation order. `0` means time-based. |
| `--ascii` | `false` | Force the kaomoji-only fallback even on capable terminals. |

**Non-TTY behavior.** When stdout is not a TTY, when `NO_COLOR` or `GLACIER_NO_COLOR` is set, or when `--ascii` is passed, vibe degrades to a one-shot static frame: kaomoji bear + plain wordmark + tagline + a single tip + a `(static)` annotation; prints once and exits 0.

**Exit codes.** `0` success; `70` raw-mode acquire failed; `1` animator setup failed; `130` SIGINT; `143` SIGTERM.
<!-- /magpie:extract -->

## Try it

```asciinema
site/public/casts/vibe.cast
```

The cast above runs `glacier vibe --seed=0 --duration=10s`. Use `--ascii` to see the static fallback on any terminal.

## Related commands

[version](./version.md) [generate](./generate.md) [explain](./explain.md)
