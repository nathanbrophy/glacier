---
title: glacier generate
---

# glacier generate

**Synopsis.** Run every Glacier code generator (cli, mock, httpmock) over the current module.

**Other commands:** [vibe](./vibe.md) [version](./version.md) [lint](./lint.md) [test](./test.md) [init](./init.md) [new](./new.md) [completions](./completions.md) [explain](./explain.md)

## Flags

| Flag | Default | Description |
|---|---|---|
| `[patterns]` | `./...` | go/packages patterns to scan. Pass one or more to restrict to a subtree. |
| `--check` | `false` | Drift-detection mode: no files are written; exits 69 if any generated file is stale. |
| `--only` | (all) | Comma-separated subset of generators to run. Values: `cli`, `mock`, `httpmock`. |
| `--parallel` | `0` | Cap the number of concurrent generators. `0` uses `GOMAXPROCS`. |
| `--no-status` | `false` | Suppress the live status panel animation. |

## Examples

Run all generators over the entire module:

```sh
glacier generate
```

Check for drift without writing (CI gate):

```sh
glacier generate --check
```

Regenerate only mock stubs:

```sh
glacier generate --only=mock
```

Regenerate cli and mock in parallel, capped at 2:

```sh
glacier generate --only=cli,mock --parallel=2
```

Run against a single package:

```sh
glacier generate ./cmd/myapp/...
```

## What it does under the hood

`generate` runs three generators: `cli/gen.Generate`, `mock/gen.Generate`, and `httpmock/gen.Generate`. Each generator is a direct function reference; there is no reflection or plugin system. Generators run concurrently in a `concur.Group`, each owning one row in a `term.StatusBar` rendered by a `term.Animator`. With `--only`, the generator list is filtered before dispatch. With `--check`, the generators run in read-only mode and return an error containing the word "stale" for any drift; the command collects those errors, prints a summary, and exits 69.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | All generators succeeded |
| 2 | Usage error (bad pattern or unknown generator name) |
| 64 | At least one generator failed |
| 69 | Drift detected in `--check` mode |

## See also

- [`glacier lint`](./lint.md) - run after generate to catch style issues
- [`glacier test`](./test.md)
- [`glacier new`](./new.md)
- [`glacier explain +glacier:command`](./explain.md) - marker reference
