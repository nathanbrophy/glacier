---
title: glacier new
---

# glacier new    [ SDK ]

[ View source spec → ](../../../specs/0032-sdk.md#commands-new)
**Other commands:** [vibe](./vibe.md) [version](./version.md) [generate](./generate.md) [lint](./lint.md) [test](./test.md) [init](./init.md) [completions](./completions.md) [explain](./explain.md)

<!-- magpie:extract source=specs/0032-sdk.md section=commands subsection=new source-checksum=<TODO> -->
**Synopsis.** Add a package, command, or option to an existing Glacier project.

**Mental model.** `new` is the dual to `init`. It walks the current module via `go/packages`, finds the existing command tree, and writes one or more new files plus surgical AST edits to existing files via `go/ast` + `go/format`. After writing, `new command` invokes `cli/gen.Generate` in-process so the user's tree is in a working state. `--dry-run` prints the plan with per-file unified diffs; no file is written.

**Subcommands.**

| Subcommand | Description |
|---|---|
| `glacier new package <name>` | Create a new Go package skeleton. |
| `glacier new command <name>` | Create a new `+glacier:command` struct and re-run cligen. |
| `glacier new option <TypeName>` | Append a functional-option constructor in the chosen package. |

**Common flags.**

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | Print the plan with diffs; no files written. |
| `--force` | `false` | Overwrite colliding files. |
| `--package` | | Disambiguate when multiple packages match. |

**Exit codes.** `0` success; `2` ambiguous `--package`; `64` codegen re-run failed; `67` no module, collision without `--force`, AST parse failure, invalid name, path violation; `130` SIGINT.
<!-- /magpie:extract -->

## Try it

```
$ glacier new command pause --dry-run
ʕ•_•ʔ glacier new command pause --dry-run

Plan:
  + cmd/myapp/pause.go                  ( 24 lines)
  ~ cmd/myapp/zz_generated_cli.go       (+1, -0)

─── cmd/myapp/pause.go (new) ───────────────────────────────────────
  // PauseCmd pauses the running server.
  //
  // +glacier:command name=pause parent=myapp
  type PauseCmd struct {
      // Force pauses without graceful shutdown.
      //
      // +glacier:short f
      Force bool
  }
  // [...]
─────────────────────────────────────────────────────────────────────

ʕ•ᴥ•ʔ dry run: no files written. Re-run without --dry-run to apply.
```

## Related commands

[init](./init.md) [generate](./generate.md) [explain](./explain.md)
