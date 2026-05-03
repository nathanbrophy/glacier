---
title: glacier test
---

# glacier test    [ SDK ]

[ View source spec → ](../../../specs/0032-sdk.md#commands-test)
**Other commands:** [vibe](./vibe.md) [version](./version.md) [generate](./generate.md) [lint](./lint.md) [init](./init.md) [new](./new.md) [completions](./completions.md) [explain](./explain.md)

<!-- magpie:extract source=specs/0032-sdk.md section=commands subsection=test source-checksum=<TODO> -->
**Synopsis.** Run the Go test suite with a live status panel and an aggregated summary.

**Mental model.** `test` wraps `go test -json` as a subprocess and stream-parses the JSON event stream. Completed packages stream upward in the result block (most recent at the top). Active packages occupy the status panel below (up to 10 rows). The summary block at the end gives pass/fail counts, coverage vs threshold, slowest tests, and failure details. `--bench=<re>` runs benchmarks and compares against the per-project baseline at `<repo>/.glacier/bench-baseline.json`; regression > 5% exits 66. `--update-baseline` rewrites the baseline atomically. Output formats: text (default on TTY), JUnit XML, SARIF, JSON.

**Flags.**

| Flag | Default | Description |
|---|---|---|
| `[patterns]` | `./...` | go/packages patterns to test. |
| `--race` | `false` | Forward `-race` to `go test`. |
| `--cover` | `false` | Forward `-cover` and merge coverage across packages. |
| `--fuzz` | | Forward `-fuzz=<re>` to `go test`. |
| `--bench` | | Forward `-bench=<re>` and run benchstat against the baseline. |
| `--baseline` | `.glacier/bench-baseline.json` | Override the default baseline path. |
| `--update-baseline` | `false` | Write the current run as the new baseline. |
| `--format` | `text` | Output format. Values: `text`, `junit`, `sarif`, `json`. |
| `--slowest` | `5` | Count of slowest-tests entries in the summary. |
| `--no-status` | `false` | Suppress the live status panel. |

**`--format=json` schema.** Each line is a `go test -json` event verbatim. The final line is a Glacier-specific aggregate object with keys `action`, `packages`, `pass`, `fail`, `skip`, `coverage`, `wall_seconds`, `slowest`, and `failures`. Tools that already consume `go test -json` work unchanged.

**Exit codes.** `0` all pass; `2` bad pattern or unknown format; `66` tests failed or bench regression > 5%; `70` `go test` failed to start; `130` SIGINT.
<!-- /magpie:extract -->

## Try it

```asciinema
site/public/casts/test.cast
```

The cast shows a focused run with one failing test and the isolation hint.

## Related commands

[lint](./lint.md) [generate](./generate.md) [explain](./explain.md)
