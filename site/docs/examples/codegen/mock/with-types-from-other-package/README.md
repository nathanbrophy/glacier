# mock/with-types-from-other-package

**Generator:** mock (`mock/gen`)

Demonstrates an interface whose method signatures reference types from imported packages (`net/http`, `context`). The generator resolves fully qualified type names and emits the correct `import` block in the generated file.

## What to look for

- The `HTTPClient` interface has methods that use `*http.Request`, `*http.Response`, and `context.Context`.
- The generated `MockHTTPClient` uses the same fully qualified types.
- The `import` block in `out.go` is complete: no manual edits needed after generation.
- The generated file compiles in isolation (`go build ./...` passes with the `go.mod.test` stub).

## CI check

`tests/codegen-gallery_test.go` runs `mock/gen.Generate` over `in.go` and asserts byte-equality with `out.go`. Drift fails the build.
