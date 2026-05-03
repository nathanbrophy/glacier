---
title: glacier vibe
---

# glacier vibe

**Synopsis.** Run the Glacier vibes animation: animated polar-bear banner with a rotating tip line. Press any key to exit.

**Other commands:** [version](./version.md) [generate](./generate.md) [lint](./lint.md) [test](./test.md) [init](./init.md) [new](./new.md) [completions](./completions.md) [explain](./explain.md)

## Flags

| Flag | Default | Description |
|---|---|---|
| `--duration` | `0s` | Bound the loop. `0s` means run until a key is pressed or SIGINT arrives. |
| `--no-tips` | `false` | Suppress the rotating tip line. |
| `--seed` | `0` | Seed the tip-rotation order. `0` means time-based. |
| `--ascii` | `false` | Force the kaomoji-only static fallback even on a capable terminal. |

## Examples

Run the animation until you press a key:

```sh
glacier vibe
```

Run for exactly 10 seconds (useful in scripts):

```sh
glacier vibe --duration=10s
```

Show the animation without tips:

```sh
glacier vibe --no-tips
```

Force the static fallback (useful on a terminal that does not support raw mode):

```sh
glacier vibe --ascii
```

Reproducible tip order (useful for screenshots):

```sh
glacier vibe --seed=42 --duration=5s
```

## What it does under the hood

`vibe` detects terminal capabilities via `term.Capability`. On a TTY without `--ascii`, it acquires raw mode with `term.AcquireRaw` so a single keypress exits cleanly, then hands control to a `term.Animator` running at 100ms per frame. Each frame is rendered by `vibeAnimation.Render`: the bear expression cycles through three kaomoji every 30 ticks (3 seconds), and the tip rotates every 50 ticks (5 seconds) from the 12-tip registry in `cmd/glacier/internal/vibetips`. The result is wrapped in a `term.Box` with rounded corners and padding. When stdout is not a TTY, when `--ascii` is set, or when raw mode is unavailable, the command falls back to a single printed frame (kaomoji + wordmark + tagline + one tip) and exits 0.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Success or clean keypress exit |
| 1 | Animation setup failed |
| 130 | SIGINT (Ctrl-C) |
| 143 | SIGTERM |

## See also

- [`glacier explain +glacier:command`](./explain.md) - how command structs are annotated
- [`glacier version`](./version.md)
