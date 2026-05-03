---
title: glacier lint
---

# glacier lint

**Synopsis.** Run gofmt, go vet, staticcheck (when on PATH), and six Glacier-specific lints.

**Other commands:** [vibe](./vibe.md) [version](./version.md) [generate](./generate.md) [test](./test.md) [init](./init.md) [new](./new.md) [completions](./completions.md) [explain](./explain.md)

## Flags

| Flag | Default | Description |
|---|---|---|
| `[patterns]` | `./...` | go/packages patterns to scan. |
| `--fix` | `false` | Apply auto-fixable lint fixes in place (gofmt, `no-em-dash`, marker normalization). |
| `--severity` | `warning` | Minimum severity to report and gate on. Values: `error`, `warning`, `info`. |
| `--format` | `text` | Output format. Values: `text`, `json`, `sarif`. |
| `--no-cache` | `false` | Ignore the per-file content-hash cache (`.glacier/lint-cache.json`). |

## Examples

Run the full lint suite:

```sh
glacier lint
```

Auto-fix gofmt violations and em-dash characters in one pass:

```sh
glacier lint --fix
```

Report only errors (skip warnings):

```sh
glacier lint --severity=error
```

Emit SARIF for upload to GitHub Code Scanning:

```sh
glacier lint --format=sarif > lint.sarif
```

Lint a single package and show JSON output:

```sh
glacier lint ./cli/... --format=json
```

## Glacier-specific lints

| Rule | Severity | What it checks |
|---|---|---|
| `no-em-dash` | error | No U+2014 in `.go`, `.md`, or `.txt` files |
| `panic-in-library` | error | No `panic(` outside `_test.go` in non-`cmd/` packages |
| `library-error-register` | error | Exported `*Error.Error()` strings match `^[a-z][^.]*$` |
| `exported-doc-comment` | warning | Every exported symbol has a doc comment starting with the symbol name |
| `package-example-test` | warning | Every non-internal package has at least one `Example*` function |
| `naked-any` | warning (opt-in) | Function signatures use named interfaces rather than bare `any`/`interface{}` |

The `naked-any` rule is off by default. Enable it via `lint.naked_any.enabled = true` in the config file.

## What it does under the hood

`lint` runs gofmt in-process via `go/format`, shells out to `go vet` for each pattern, and optionally shells out to `staticcheck` when it is found on PATH (skipped silently otherwise). The six Glacier-specific rules run in-process via the `Linter` interface, walking the file tree with `filepath.WalkDir`. Results are cached by content hash in `.glacier/lint-cache.json`; a file that has not changed since the last run reuses its cached findings. Findings are grouped by severity (errors first), then by file. `--fix` rewrites gofmt violations, replaces U+2014 em-dashes with `: `, and normalizes `// +glacier: directive` spacing.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | No findings at or above the severity threshold |
| 2 | Bad pattern |
| 65 | One or more findings at or above the severity threshold |
| 70 | Subprocess failure (go vet or staticcheck failed to start) |

## See also

- [`glacier generate`](./generate.md) - run before lint after editing annotations
- [`glacier test`](./test.md)
- [`glacier explain 65`](./explain.md) - exit code 65 details
