---
title: glacier init
---

# glacier init

**Synopsis.** Scaffold a new Glacier project with signals, banner, version, completions, and optional OTEL wired from day one.

**Other commands:** [vibe](./vibe.md) [version](./version.md) [generate](./generate.md) [lint](./lint.md) [test](./test.md) [new](./new.md) [completions](./completions.md) [explain](./explain.md)

## Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `[dir]` | | `.` (current directory) | Target directory for the new project. Created if it does not exist. |
| `--name` | | (prompted or directory name) | Go module path, e.g. `github.com/acme/myapp`. |
| `--template` | | `cli-app` | Project layout. Values: `library-only`, `cli-app`, `both`. |
| `--license` | | `Apache-2.0` | SPDX license identifier. Values: `Apache-2.0`, `MIT`, `BSD-3-Clause`, `none`. |
| `--mascot` | | `polar_bear` | Project mascot. Values: `polar_bear`, `penguin`, `owl`, `fox`, `otter`, `raccoon`. |
| `--no-git` | | `false` | Skip `git init`. |
| `--yes` | `-y` | `false` | Accept all defaults; non-interactive. Required for CI or scripted setup. |
| `--force` | | `false` | Overwrite existing files. Without `--yes`, prompts for confirmation first. |

## Examples

Interactive scaffold in a new directory:

```sh
glacier init my-app
```

Non-interactive with all defaults (good for CI):

```sh
glacier init my-app --yes
```

Library-only project with a specific module path:

```sh
glacier init my-lib --name=github.com/acme/my-lib --template=library-only --yes
```

Scaffold a project with a penguin mascot and MIT license:

```sh
glacier init my-app --mascot=penguin --license=MIT --yes
```

Force over an existing directory:

```sh
glacier init . --name=github.com/acme/myapp --yes --force
```

## What it does under the hood

`init` collects parameters either interactively (module path via `term.Prompt`, mascot via `term.Select`) or from flags when `--yes` is set. It renders templates from an embedded `embed.FS` (`cmd/glacier/templates/`) using `text/template`, writing each file atomically through `internal/safefile`. The project banner is rendered by `cmd/glacier/internal/figgen` using the app name derived from the module path. The chosen mascot is sourced from `cmd/glacier/internal/mascots`. Without `--force`, any pre-existing output file aborts with exit 67. With `--force` and without `--yes`, a `term.Confirm` prompt guards the overwrite. If no `.git/` exists and `--no-git` is not set, `git init` is run as a subprocess. On success a `term.Box` shows the next steps.

The scaffolded `main.go` is six lines:

```go
package main

import "github.com/nathanbrophy/glacier/cli"

func main() {
    cli.Default.Main()
}
```

All wiring (signals, banner, version subcommand, completions subcommand, OTEL no-op when unset, httpc tracing, cli metrics) lives in the glaciergen-emitted `zz_generated_cli.go`.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Project scaffolded successfully |
| 67 | Scaffolding failed (collision without `--force`, bad name, template error, write failure) |
| 130 | SIGINT during an interactive prompt |

## See also

- [`glacier new`](./new.md) - add packages or commands to an existing project
- [`glacier generate`](./generate.md) - regenerate after editing annotations
- [`glacier explain 67`](./explain.md) - exit code 67 details
