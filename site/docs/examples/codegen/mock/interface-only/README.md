# mock/interface-only

**Generator:** mock (`mock/gen`)

Demonstrates the minimal case: a single interface with two methods, annotated with `+glacier:mock`. The generated mock is typed at compile time; no `interface{}` appears in the output.

## What to look for

- `+glacier:mock` on the `Store` interface is the only annotation needed.
- The generated `MockStore` struct has one `On<MethodName>` field per method.
- Test code sets up expectations by assigning function literals to the `On*` fields.
- `mock.Of[Store]` returns a `*MockStore` pre-wired with zero-value `On*` fields that fail if called unexpectedly.

## CI check

`tests/codegen-gallery_test.go` runs `mock/gen.Generate` over `in.go` and asserts byte-equality with `out.go`. Drift fails the build.
