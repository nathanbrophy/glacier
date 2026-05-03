---
title: "httpmock: recording-disabled"
---

# httpmock: recording-disabled

[ Back to gallery ](./index.md)

**Exhibit path:** `docs/examples/codegen/httpmock/recording-disabled/`

Recording mode turned off: all unmatched requests fail immediately with a clear error. Use this exhibit when your test must never let a real HTTP request escape.

## Input (`in.go`)

<<< @/docs/examples/codegen/httpmock/recording-disabled/in.go

## Output (`out.go`)

<<< @/docs/examples/codegen/httpmock/recording-disabled/out.go

## What the generator did

- Generated the same `MockTransport` scaffold as the programmable-router exhibit.
- The `RecordingDisabled` option is set in the generated initialization helper so any unregistered request returns an error immediately rather than being recorded for later inspection.
- This is the safest posture for unit tests: the test fails loudly on any URL that was not explicitly registered, instead of silently recording the call.

## Related

[programmable-router](./httpmock-programmable-router.md) [with-body-closure](./httpmock-with-body-closure.md)
