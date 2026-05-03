---
title: "cli: nested"
---

# cli: nested

[ Back to gallery ](./index.md)

**Exhibit path:** `docs/examples/codegen/cli/nested/`

A two-level command tree: a parent command with two subcommands. Shows how `parent=<name>` links subcommands to their parent and how the generated registry wires the full tree.

## Input (`in.go`)

<<< @/docs/examples/codegen/cli/nested/in.go

## Output (`out.go`)

<<< @/docs/examples/codegen/cli/nested/out.go

## What the generator did

- Registered the root command via `+glacier:root`.
- Registered `StartCmd` and `StopCmd` as children of the root using `parent=myapp`.
- The generated registry dispatches `myapp start` and `myapp stop` to the respective `Run` methods.
- `+glacier:short s` on `StartCmd.Port` emitted `cli.WithFlagShort("Port", 's')` in the generated file.

## Related

[simple](./cli-simple.md) [all-markers](./cli-all-markers.md) [glacier new command](../commands/new.md)
