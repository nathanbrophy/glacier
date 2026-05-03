---
title: "mock: nested-interfaces"
---

# mock: nested-interfaces

[ Back to gallery ](./index.md)

**Exhibit path:** `docs/examples/codegen/mock/nested-interfaces/`

An interface that embeds another interface. Shows how the mock generator handles embedded interfaces: it flattens the method set and generates a single mock struct that satisfies the full embedded interface.

## Input (`in.go`)

<<< @/docs/examples/codegen/mock/nested-interfaces/in.go

## Output (`out.go`)

<<< @/docs/examples/codegen/mock/nested-interfaces/out.go

## What the generator did

- Found `ReadWriter` which embeds `Reader` and `Writer`.
- Flattened the method set: `Read`, `Write`, and `Close` all appear on `MockReadWriter`.
- The embedded interface is not represented as a separate mock field; the generated struct directly implements all methods. This keeps test code flat: one mock, one `On*` field per method.

## Related

[interface-only](./mock-interface-only.md) [with-types-from-other-package](./mock-with-types-from-other-package.md)
