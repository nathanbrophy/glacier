---
title: glacier explain
---

# glacier explain    [ SDK ]

[ View source spec → ](../../../specs/0032-sdk.md#commands-explain)
**Other commands:** [vibe](./vibe.md) [version](./version.md) [generate](./generate.md) [lint](./lint.md) [test](./test.md) [init](./init.md) [new](./new.md) [completions](./completions.md)

<!-- magpie:extract source=specs/0032-sdk.md section=commands subsection=explain source-checksum=<TODO> -->
**Synopsis.** Print an explanation for a marker, exit code, or config key.

**Mental model.** `explain <topic>` reads from an `embed.FS` of pre-rendered topic files at `cmd/glacier/internal/explain/topics/<slug>.md`. The files are generated at build time by `cmd/glacier/internal/explaingen/` from spec sections; CI byte-equality check ensures spec-to-impl freshness. The command renders the topic as a `term.Box` with title, body, and a "see also" rows block. The kaomoji on the title row reflects category (marker, exit code, config key). `--list` prints all topics.

**Argument.**

```
glacier explain <topic>
glacier explain --list
```

`<topic>` is one of: a marker name (e.g. `+glacier:command`), an exit code (e.g. `66`), or a config key (e.g. `versioncheck.cache_ttl`).

**Flags.**

| Flag | Default | Description |
|---|---|---|
| `--list` | `false` | Enumerate all topics and exit. |

**Exit codes.** `0` success; `2` unknown topic (with did-you-mean hint if Levenshtein distance <= 2); `1` stdout write failure.
<!-- /magpie:extract -->

## Try it

```
$ glacier explain 66
╭─ ʕ× ×ʔ exit code 66 ─────────────────────────────────────────────────╮
│                                                                       │
│ One or more tests failed, OR a benchmark regressed by more than       │
│ the configured threshold (default 5%).                                │
│                                                                       │
│ Common next steps:                                                    │
│   - Run the failing test in isolation: `glacier test ./<pkg> -run X`  │
│   - Inspect the bench delta: `glacier test --bench=. --baseline=...`  │
│   - Update the baseline if the regression is intended:                │
│     `glacier test --bench=. --update-baseline`                        │
│                                                                       │
│ See also:                                                             │
│   exit code 65 (lint findings)                                        │
│   exit code 70 (subprocess failure)                                   │
│   key test.regression_pct                                             │
╰───────────────────────────────────────────────────────────────────────╯
```

## Related commands

[version](./version.md) [lint](./lint.md) [test](./test.md) [completions](./completions.md)
