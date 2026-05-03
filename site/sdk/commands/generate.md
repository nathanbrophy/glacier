---
title: glacier generate
---

# glacier generate    [ SDK ]

[ View source spec → ](../../../specs/0032-sdk.md#commands-generate)
**Other commands:** [vibe](./vibe.md) [version](./version.md) [lint](./lint.md) [test](./test.md) [init](./init.md) [new](./new.md) [completions](./completions.md) [explain](./explain.md)

<!-- magpie:extract source=specs/0032-sdk.md section=commands subsection=generate source-checksum=<TODO> -->
**Synopsis.** Run every Glacier code generator (cli, mock, httpmock) over the current module.

**Mental model.** `generate` is the umbrella over Glacier's three v0 generators. The generator registry is an explicit list of (`name`, `fn`) pairs; functions are direct references to `cli/gen.Generate`, `mock/gen.Generate`, `httpmock/gen.Generate`. No reflection. Each generator runs in a `concur.Group` slot, owns one row in a `term.StatusBar`, and reports progress through a thin channel. `--check` mode swaps the panel for a check-list and emits unified diffs for every stale file. Exit 64 on generator failure; exit 69 on drift detected by `--check`.

**Flags.**

| Flag | Default | Description |
|---|---|---|
| `[patterns]` | `./...` | go/packages patterns to scan. |
| `--check` | `false` | Drift-detection mode: no files written; exit 69 on drift. |
| `--only` | (all) | Restrict generators by name. Values: `cli`, `mock`, `httpmock`. |
| `--parallel` | `0` | Cap concurrent generators. `0` resolves to `min(GOMAXPROCS, 8)`. |
| `--no-status` | `false` | Suppress the live status panel. |

**Use in CI.** Run `glacier generate --check` as a gate to catch drift between source annotations and generated files.

**Exit codes.** `0` success; `2` usage error (bad pattern or unknown generator); `64` generator failure; `69` drift detected.
<!-- /magpie:extract -->

## Try it

```asciinema
site/public/casts/generate.cast
```

The cast shows a full run followed by `--check` with one stale file.

## Codegen gallery

See the [codegen gallery](../codegen/index.md) for annotated input/output examples for each generator.

## Related commands

[lint](./lint.md) [test](./test.md) [new](./new.md) [explain](./explain.md)
