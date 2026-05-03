# cli/nested

**Generator:** cligen (`cli/gen`)

Demonstrates a two-level command tree: a root command and two subcommands. Shows how `parent=<name>` links subcommands to their parent and how short flags are declared with `+glacier:short`.

## What to look for

- Two commands share `parent=myapp`: `StartCmd` and `StopCmd`.
- `+glacier:short s` on `StartCmd.Port` produces `cli.WithFlagShort("Port", 's')`.
- The generated file contains one `init()` block that registers all three commands in order.
- Dispatch (`myapp start`, `myapp stop`) is handled by the generated registry with no hand-written switch statement.

## CI check

`tests/codegen-gallery_test.go` runs `cli/gen.Generate` over `in.go` and asserts byte-equality with `out.go`. Drift fails the build.
