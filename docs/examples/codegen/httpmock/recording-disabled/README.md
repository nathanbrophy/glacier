# httpmock/recording-disabled

**Generator:** httpmock (`httpmock/gen`)

Demonstrates recording mode turned off. In this posture, any request that does not match a registered route returns an error immediately rather than being recorded. Use this when a test must never let a real HTTP call escape.

## What to look for

- The generated `NewMockTransport` call passes `httpmock.WithRecordingDisabled()`.
- Unregistered routes return an error: `httpmock: no route matched GET /unregistered`.
- This is the safest default for unit tests: a missed route registration shows up as an immediate test failure, not a silent real HTTP call.

## CI check

`tests/codegen-gallery_test.go` runs `httpmock/gen.Generate` over `in.go` and asserts byte-equality with `out.go`. Drift fails the build.
