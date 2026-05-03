---
title: Codegen Gallery
---

# Codegen Gallery

Nine annotated examples cover the three Glacier code generators. Each exhibit is a self-contained pair: `in.go` (annotated source) and `out.go` (expected generated file). CI runs each generator over its `in.go` and asserts byte-equality with `out.go`; drift fails the build.

Source lives under `docs/examples/codegen/`.

## cli (cligen)

Generates the command-wiring file `zz_generated_cli.go` from `+glacier:command` annotated structs.

| Exhibit | What it shows |
|---|---|
| [simple](./cli-simple.md) | A single command with one flag and one positional argument |
| [nested](./cli-nested.md) | A two-level command tree (parent + two subcommands) |
| [all-markers](./cli-all-markers.md) | Every `+glacier:` marker valid on a command struct in one place |

## mock

Generates typed mock implementations from `+glacier:mock` annotated interfaces.

| Exhibit | What it shows |
|---|---|
| [interface-only](./mock-interface-only.md) | A single interface with two methods |
| [nested-interfaces](./mock-nested-interfaces.md) | An interface that embeds another interface |
| [with-types-from-other-package](./mock-with-types-from-other-package.md) | Methods whose signatures reference types from an imported package |

## httpmock

Generates `httpmock` handler scaffolding from `+glacier:mock` annotated HTTP interfaces.

| Exhibit | What it shows |
|---|---|
| [programmable-router](./httpmock-programmable-router.md) | A router with multiple routes registered at test setup |
| [recording-disabled](./httpmock-recording-disabled.md) | Recording mode turned off; all unmatched requests fail immediately |
| [with-body-closure](./httpmock-with-body-closure.md) | A handler that uses a body closure to generate response bodies per call |

## How to read an exhibit

Each exhibit page shows:
1. The input source (`in.go`) with its `+glacier:` markers.
2. The expected generated output (`out.go`).
3. A short explanation of what the generator did and why.

To reproduce any exhibit locally:

```sh
cd docs/examples/codegen/<exhibit>
glacier generate ./...
```

The generated file should be byte-for-byte identical to `out.go`.
