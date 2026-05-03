# mock/nested-interfaces

**Generator:** mock (`mock/gen`)

Demonstrates an interface that embeds another interface. The mock generator flattens the method set: the generated struct directly implements all methods, including those from the embedded interface. No intermediate mock structs are generated.

## What to look for

- `ReadWriter` embeds `Reader`; only `ReadWriter` has the `+glacier:mock` annotation.
- The generated `MockReadWriter` has `OnRead`, `OnWrite`, and `OnClose` fields.
- `Reader` is not separately mocked because no standalone `+glacier:mock` annotation appears on it.
- This keeps test code flat: one mock variable, one field per method.

## CI check

`tests/codegen-gallery_test.go` runs `mock/gen.Generate` over `in.go` and asserts byte-equality with `out.go`. Drift fails the build.
