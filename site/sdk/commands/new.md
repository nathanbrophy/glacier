---
title: glacier new
---

# glacier new

**Synopsis.** Add a package, command, or functional-option constructor to an existing Glacier project.

**Other commands:** [vibe](./vibe.md) [version](./version.md) [generate](./generate.md) [lint](./lint.md) [test](./test.md) [init](./init.md) [completions](./completions.md) [explain](./explain.md)

## Subcommands

| Subcommand | Description |
|---|---|
| `glacier new package <name>` | Create a new Go package skeleton (`doc.go`, `<name>.go`, `<name>_test.go`) |
| `glacier new command <name>` | Create a new `+glacier:command` struct and re-run `cli/gen.Generate` |
| `glacier new option <TypeName>` | Append a functional-option constructor to `options.go` in the target package |

## Flags for `new package`

| Flag | Default | Description |
|---|---|---|
| `<name>` | (required) | Package directory name |
| `--pkg` | (same as name) | Go package name if different from the directory name |
| `--dry-run` | `false` | Print what would be created without writing any files |
| `--force` | `false` | Overwrite existing files |

## Flags for `new command`

| Flag | Default | Description |
|---|---|---|
| `<name>` | (required) | Command name (becomes the verb) |
| `--parent` | `root` | Parent command name in the tree |
| `--dry-run` | `false` | Print the plan without writing files |
| `--force` | `false` | Overwrite existing files |

## Flags for `new option`

| Flag | Default | Description |
|---|---|---|
| `<TypeName>` | (required) | The type the option configures |
| `--pkg` | `.` | Package path where `options.go` lives |
| `--dry-run` | `false` | Print the generated code without writing it |
| `--force` | `false` | Overwrite existing files |

## Examples

Create a new package skeleton:

```sh
glacier new package auth
```

Preview a new command without writing it:

```sh
glacier new command pause --dry-run
```

Add a command under a specific parent:

```sh
glacier new command pause --parent=serve
```

Generate a functional-option constructor for `ServerConfig`:

```sh
glacier new option ServerConfig
```

Preview the option code without writing:

```sh
glacier new option ServerConfig --dry-run
```

## What it does under the hood

`new package` validates the name as a Go identifier, walks to the module root by finding `go.mod`, and writes three files atomically via `internal/safefile`. `new command` locates the directory containing `zz_generated_cli.go` by walking the module tree, writes the command stub, then calls `cli/gen.Generate` in-process to regenerate the wiring. `new option` reads the existing package name from a sibling `.go` file, generates the `<TypeName>Option` type and `With<TypeName>` constructor, formats with `go/format`, and appends to `options.go` (creating it if absent).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 2 | Usage error |
| 64 | Codegen re-run failed after `new command` |
| 67 | No module found, file collision without `--force`, invalid name, or write failure |
| 130 | SIGINT |

## See also

- [`glacier init`](./init.md) - scaffold a brand-new project
- [`glacier generate`](./generate.md) - re-run all generators after adding annotations
- [`glacier explain +glacier:command`](./explain.md) - annotation reference
