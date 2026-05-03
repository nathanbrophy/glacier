# cli/simple

**Generator:** cligen (`cli/gen`)

Demonstrates the minimal case: a single command annotated with `+glacier:command`, one integer flag with a default, and one positional string argument. The generated file wires the command into `cli.Default` with no manual boilerplate.

## What to look for

- `+glacier:command name=serve parent=myapp` links `ServeCmd` to the root `myapp` command.
- `+glacier:default 8080` on `Port` produces `cli.WithFlagDefault("Port", 8080)` in the output.
- `+glacier:positional` on `Config` produces `cli.WithPositional("Config")` in the output.
- `main.go` is six lines; all wiring is in `zz_generated_cli.go`.

## CI check

`tests/codegen-gallery_test.go` runs `cli/gen.Generate` over `in.go` and asserts byte-equality with `out.go`. Drift fails the build.
