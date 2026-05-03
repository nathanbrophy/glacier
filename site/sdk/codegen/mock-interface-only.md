---
title: "mock: interface-only"
---

# mock: interface-only

[ Back to gallery ](./index.md)

**Exhibit path:** `docs/examples/codegen/mock/interface-only/`

A single interface with two methods, annotated with `+glacier:mock`. The minimal working case for the mock generator.

## Input (`in.go`)

<<< @/docs/examples/codegen/mock/interface-only/in.go

## Output (`out.go`)

<<< @/docs/examples/codegen/mock/interface-only/out.go

## What the generator did

- Found the `+glacier:mock` marker on the `Store` interface.
- Generated a `MockStore` struct that implements `Store`.
- Each method has a corresponding `OnXxx` recorder field that test code populates with expected call behavior.
- The mock is typed at compile time: no `interface{}` parameters or return values appear in the generated code.

## Related

[nested-interfaces](./mock-nested-interfaces.md) [with-types-from-other-package](./mock-with-types-from-other-package.md) [glacier generate](../commands/generate.md)
