---
title: "cli: all-markers"
---

# cli: all-markers

[ Back to gallery ](./index.md)

**Exhibit path:** `docs/examples/codegen/cli/all-markers/`

Every `+glacier:` marker valid on a command struct, in one place. Use this as a reference when checking which markers are supported and how they translate to `cli.With*` calls in the generated file.

## Input (`in.go`)

<<< @/docs/examples/codegen/cli/all-markers/in.go

## Output (`out.go`)

<<< @/docs/examples/codegen/cli/all-markers/out.go

## What the generator did

Markers demonstrated and their generated output:

| Marker | Generated call |
|---|---|
| `+glacier:command name=<verb> parent=<parent>` | `cli.WithName("<verb>")`, parent linkage |
| `+glacier:default <value>` | `cli.WithFlagDefault("<Field>", <value>)` |
| `+glacier:short <char>` | `cli.WithFlagShort("<Field>", '<char>')` |
| `+glacier:env <KEY>` | `cli.WithFlagEnv("<Field>", "<KEY>")` |
| `+glacier:required` | `cli.WithFlagRequired("<Field>")` |
| `+glacier:choices <a>\|<b>` | `cli.WithFlagChoices("<Field>", []string{...})` |
| `+glacier:deprecated <msg>` | `cli.WithFlagDeprecated("<Field>", "<msg>")` |
| `+glacier:validate <funcName>` | `cli.WithFlagValidate("<Field>", <funcName>)` |
| `+glacier:positional` | `cli.WithPositional("<Field>")` |

## Related

[simple](./cli-simple.md) [nested](./cli-nested.md) [glacier explain](../commands/explain.md)
