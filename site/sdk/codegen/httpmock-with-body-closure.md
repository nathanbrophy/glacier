---
title: "httpmock: with-body-closure"
---

# httpmock: with-body-closure

[ Back to gallery ](./index.md)

**Exhibit path:** `docs/examples/codegen/httpmock/with-body-closure/`

A handler that uses a body closure to generate response bodies per call. Shows how to return different payloads on repeated requests to the same route without registering multiple handlers.

## Input (`in.go`)

<<< @/docs/examples/codegen/httpmock/with-body-closure/in.go

## Output (`out.go`)

<<< @/docs/examples/codegen/httpmock/with-body-closure/out.go

## What the generator did

- Generated the `MockTransport` scaffold with a `BodyFunc` handler field alongside the standard `Body` field.
- When `BodyFunc` is non-nil, the mock calls it on each request to produce the response body. The closure receives the `*http.Request` so it can vary the response based on query parameters or headers.
- `httpc` body closures are called once per attempt; retries are correct by construction because the closure produces a fresh `io.Reader` on each call.

## Related

[programmable-router](./httpmock-programmable-router.md) [recording-disabled](./httpmock-recording-disabled.md) [glacier generate](../commands/generate.md)
