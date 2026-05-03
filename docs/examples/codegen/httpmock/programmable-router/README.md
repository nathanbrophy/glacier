# httpmock/programmable-router

**Generator:** httpmock (`httpmock/gen`)

Demonstrates the standard httpmock pattern: a router with multiple routes registered at test setup. Each route is registered with a method, a path pattern, and a fixed response. Unregistered routes return 404.

## What to look for

- The `Transport` interface annotated with `+glacier:mock` tells the generator to produce an `httpmock`-backed transport.
- The generated `MockTransport` implements `http.RoundTripper`.
- Test code calls `mock.Route("GET", "/users", ...)` to register handlers before making requests through `httpc`.
- `httpmock` and `httpc` share no imports; they are wired together only at the test boundary.

## CI check

`tests/codegen-gallery_test.go` runs `httpmock/gen.Generate` over `in.go` and asserts byte-equality with `out.go`. Drift fails the build.
