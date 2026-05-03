---
title: "httpmock: programmable-router"
---

# httpmock: programmable-router

[ Back to gallery ](./index.md)

**Exhibit path:** `docs/examples/codegen/httpmock/programmable-router/`

A router with multiple routes registered at test setup. Shows the standard httpmock pattern: annotate the transport interface, generate the scaffold, register handlers in the test.

## Input (`in.go`)

<<< @/docs/examples/codegen/httpmock/programmable-router/in.go

## Output (`out.go`)

<<< @/docs/examples/codegen/httpmock/programmable-router/out.go

## What the generator did

- Found the `+glacier:mock` marker on the HTTP transport interface.
- Generated a `MockTransport` that implements `http.RoundTripper`.
- The generated `Router` field accepts route registrations (`GET /users`, `POST /users`, etc.) at test-setup time.
- Unregistered routes return a 404 by default, making accidental real-HTTP calls visible immediately.

## Related

[recording-disabled](./httpmock-recording-disabled.md) [with-body-closure](./httpmock-with-body-closure.md) [glacier generate](../commands/generate.md)
