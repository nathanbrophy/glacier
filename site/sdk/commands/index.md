---
title: SDK Commands
---

# SDK Commands

Nine commands cover the developer day.

| Command | Description |
|---|---|
| [`glacier init`](./init.md) | Scaffold a new Glacier project |
| [`glacier new`](./new.md) | Add a package, command, or option to an existing project |
| [`glacier generate`](./generate.md) | Run every Glacier code generator (cli, mock, httpmock) |
| [`glacier lint`](./lint.md) | Run gofmt, go vet, staticcheck, and Glacier-specific lints |
| [`glacier test`](./test.md) | Run the Go test suite with a live status panel and summary |
| [`glacier version`](./version.md) | Print the Glacier SDK version; `--check` compares against latest |
| [`glacier explain`](./explain.md) | Print an explanation for a marker, exit code, or config key |
| [`glacier vibe`](./vibe.md) | Run the Glacier vibes animation |
| [`glacier completions`](./completions.md) | Print a shell-completion script for the named shell |

## Global flags

Every command inherits these flags from the root command:

| Flag | Short | Description |
|---|---|---|
| `--help` | `-h` | Print help text |
| `--quiet` | `-q` | Suppress non-error output and animations; keep final summary |
| `--verbose` | `-V` | Raise log level to Debug |
| `--very-verbose` | | Raise log level to Trace |
| `--no-animate` | | Force plain output even on a TTY |
| `--no-banner` | | Suppress the banner on this invocation |
| `--no-color` | | Disable all ANSI color output |
| `--force-color` | | Force color output even when not a TTY |
| `--profile` | | Write pprof CPU, heap, and goroutine profiles |
| `--otel-endpoint` | | Override `OTEL_EXPORTER_OTLP_ENDPOINT` for this invocation |

`--quiet` and `--verbose` / `--very-verbose` are mutually exclusive (exit 2 if combined).

## Exit codes

See [Configuration](../configuration.md#exit-codes) for the full exit-code table. Use `glacier explain <code>` for details on any specific code.
