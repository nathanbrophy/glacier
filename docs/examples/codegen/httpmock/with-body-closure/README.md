# httpmock/with-body-closure

**Generator:** httpmock (`httpmock/gen`)

Demonstrates a handler that uses a body closure to generate response bodies per call. The closure receives the `*http.Request` and returns an `io.Reader`, so different callers or repeated calls can get different response bodies from the same route.

## What to look for

- The `+glacier:httpmock body=closure` marker triggers generation of the `BodyFunc` field alongside the standard `Body` field on the route handler.
- `httpc` body closures are called once per attempt: retries always get a fresh `io.Reader`, preventing double-read errors on a closed body.
- The closure is set in test code; the generated scaffold only provides the field and the dispatch logic.

## CI check

`tests/codegen-gallery_test.go` runs `httpmock/gen.Generate` over `in.go` and asserts byte-equality with `out.go`. Drift fails the build.
