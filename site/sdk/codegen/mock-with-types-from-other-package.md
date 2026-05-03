---
title: "mock: with-types-from-other-package"
---

# mock: with-types-from-other-package

[ Back to gallery ](./index.md)

**Exhibit path:** `docs/examples/codegen/mock/with-types-from-other-package/`

Methods whose signatures reference types from an imported package. Shows that the mock generator preserves the full qualified type name in the generated file and emits the correct import block.

## Input (`in.go`)

<<< @/docs/examples/codegen/mock/with-types-from-other-package/in.go

## Output (`out.go`)

<<< @/docs/examples/codegen/mock/with-types-from-other-package/out.go

## What the generator did

- Resolved the method signatures to their full qualified forms (e.g. `http.Request`, `context.Context`).
- Emitted the correct `import` block in `out.go` with every referenced package.
- The generated mock compiles in isolation: no manual import edits required after generation.

## Related

[interface-only](./mock-interface-only.md) [nested-interfaces](./mock-nested-interfaces.md) [glacier generate](../commands/generate.md)
