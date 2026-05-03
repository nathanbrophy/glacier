# cli/all-markers

**Generator:** cligen (`cli/gen`)

Demonstrates every `+glacier:` marker that is valid on a command struct. Use this as a living reference when checking which markers are supported and how each translates to a `cli.With*` call in the generated file.

## Markers shown

| Marker | Purpose |
|---|---|
| `+glacier:command name=<verb> parent=<parent>` | Register the struct as a command |
| `+glacier:root` | Mark the struct as the root command |
| `+glacier:default <value>` | Set a flag default |
| `+glacier:short <char>` | Add a single-character short flag alias |
| `+glacier:env <KEY>` | Bind a flag to an environment variable |
| `+glacier:required` | Make a flag required (non-empty) |
| `+glacier:choices <a>\|<b>` | Restrict a flag to an enumerated set |
| `+glacier:deprecated <msg>` | Mark a flag as deprecated with a message |
| `+glacier:validate <funcName>` | Run a custom validator at parse time |
| `+glacier:positional` | Mark a field as a positional argument |

## CI check

`tests/codegen-gallery_test.go` runs `cli/gen.Generate` over `in.go` and asserts byte-equality with `out.go`. Drift fails the build.
