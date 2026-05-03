---
title: "cli: simple"
---

# cli: simple

[ Back to gallery ](./index.md)

**Exhibit path:** `docs/examples/codegen/cli/simple/`

A single command with one string flag and one positional argument. This is the minimal working case: one annotated struct, one generated file.

## Input (`in.go`)

<<< @/docs/examples/codegen/cli/simple/in.go

## Output (`out.go`)

<<< @/docs/examples/codegen/cli/simple/out.go

## What the generator did

- Read the `+glacier:command name=serve parent=myapp` marker and registered `ServeCmd` as a subcommand of the root `myapp` command.
- Emitted `cli.WithFlagDefault("Port", 8080)` from the `+glacier:default 8080` marker on the `Port` field.
- Emitted `cli.WithPositional("Config")` from the `+glacier:positional` marker on the `Config` field.
- Wired the struct into the generated registry so `cli.Default` dispatches `serve` to `ServeCmd.Run`.

## Related

[nested](./cli-nested.md) [all-markers](./cli-all-markers.md) [glacier generate](../commands/generate.md)
