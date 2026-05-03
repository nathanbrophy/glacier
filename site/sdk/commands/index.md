---
title: SDK Commands
---

# SDK Commands

Nine commands cover the developer day.

## Create

| Command | Description |
|---|---|
| [`glacier init`](./init.md) | Scaffold a new Glacier project in the current directory |
| [`glacier new`](./new.md) | Add a package, command, or option to an existing project |

## Develop

| Command | Description |
|---|---|
| [`glacier generate`](./generate.md) | Run every Glacier code generator (cli, mock, httpmock) over the current module |
| [`glacier lint`](./lint.md) | Run gofmt, go vet, staticcheck, and Glacier-specific lints |
| [`glacier test`](./test.md) | Run the Go test suite with a live status panel and aggregated summary |

## Inspect

| Command | Description |
|---|---|
| [`glacier version`](./version.md) | Print the Glacier SDK version; `--check` compares against latest |
| [`glacier explain`](./explain.md) | Print an explanation for a marker, exit code, or config key |

## Utility

| Command | Description |
|---|---|
| [`glacier vibe`](./vibe.md) | Run the Glacier vibes animation |
| [`glacier completions`](./completions.md) | Print a shell-completion script for the named shell |

## Global flags

Every command inherits these flags from glaciergen:

| Flag | Short | Description |
|---|---|---|
| `--help` | `-h` | Print help text |
| `--version` | `-v` | Print version info (root only) |
| `--quiet` | `-q` | Suppress non-error output and animations |
| `--verbose` | `-V` | Raise log level to debug |
| `--very-verbose` | | Raise log level to trace |
| `--no-animate` | | Force plain output even on TTY |
| `--no-banner` | | Suppress banner on this invocation |
| `--profile` | | Write CPU+heap+goroutine profiles |
| `--otel-endpoint` | | Override `OTEL_EXPORTER_OTLP_ENDPOINT` |

## Exit codes

See [Configuration](../configuration.md#exit-codes) for the full exit-code table. Use `glacier explain <code>` for details on any code.
