---
title: glacier test
---

# glacier test

**Synopsis.** Run the Go test suite with a live status panel and a color-coded aggregated summary.

**Other commands:** [vibe](./vibe.md) [version](./version.md) [generate](./generate.md) [lint](./lint.md) [init](./init.md) [new](./new.md) [completions](./completions.md) [explain](./explain.md)

## Flags

| Flag | Default | Description |
|---|---|---|
| `[patterns]` | `./...` | go/packages patterns to test. |
| `--race` | `false` | Forward `-race` to `go test`. |
| `--cover` | `false` | Forward `-cover` and write a profile to `.glacier/coverage.out`. |
| `--fuzz` | | Forward `-fuzz=<regexp>` to `go test`. |
| `--bench` | | Forward `-bench=<regexp>` and compare results against the baseline. |
| `--baseline` | `.glacier/bench-baseline.json` | Path to the benchmark baseline file. |
| `--update-baseline` | `false` | Write the current benchmark run as the new baseline. |
| `--format` | `text` | Output format. Values: `text`, `junit`, `sarif`, `json`. |
| `--slowest` | `5` | Number of slowest tests shown in the summary. |
| `--no-status` | `false` | Suppress the live status panel animation. |

## Examples

Run all tests:

```sh
glacier test
```

Run with the race detector and coverage:

```sh
glacier test --race --cover
```

Run only tests in one package:

```sh
glacier test ./cli/...
```

Run benchmarks and record a new baseline:

```sh
glacier test --bench=. --update-baseline
```

Run benchmarks and fail if any regress more than 5%:

```sh
glacier test --bench=.
```

Emit JUnit XML for CI:

```sh
glacier test --format=junit > results.xml
```

Run a fuzz target for 30 seconds:

```sh
glacier test --fuzz=FuzzParse ./conf/...
```

## What it does under the hood

`test` launches `go test -json` as a subprocess and streams its output through a JSON event parser. In-flight packages are shown in a `term.StatusBar` (up to 10 rows) rendered by `term.Animator`. Completed packages stream to the terminal as they finish, colored by pass/fail/skip. When all packages finish, the command renders a summary box via `term.Box` showing package count, pass/fail/skip counts, coverage, wall time, the N slowest tests, and any failure output with an isolation hint. For `--format=json`, every raw `go test -json` event is forwarded verbatim, followed by a Glacier-specific aggregate object. For `--bench`, benchmark output is collected in a separate channel and compared against the stored baseline via `cmd/glacier/internal/benchcmp`; a regression over the threshold (default 5%) prints a box and exits 66.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | All tests passed |
| 2 | Bad pattern or unknown format |
| 66 | One or more tests failed, or a benchmark regressed more than the threshold |
| 70 | `go test` failed to start |
| 130 | SIGINT |

## See also

- [`glacier lint`](./lint.md) - run before test in CI
- [`glacier generate`](./generate.md)
- [`glacier explain 66`](./explain.md) - exit code 66 details
