---
title: Glacier SDK
---

# Glacier SDK

<div class="glacier-sprite-accent">
  <MascotSprite state="wave" :size="96" />
</div>

[ View source spec → ](../../specs/0032-sdk.md)

<!-- magpie:extract source=specs/0032-sdk.md section=public-summary source-checksum=<TODO> -->
The Glacier SDK is a CLI binary, `glacier`, built on every Glacier framework package. It is the framework's longest-running integration test, its public face for new developers, and the source of the "batteries included" experience every project scaffolded by `glacier init` inherits. Nine commands cover the developer day: `vibe`, `version`, `generate`, `lint`, `test`, `init`, `new`, `completions`, `explain`. Each is built with the `cli` package and its codegen tool. Animations route through `term.Animator`. HTTP calls route through `httpc`. Configuration loads through `conf`. Logs flow through `log` with kaomoji-prefixed status messages at command boundaries. Telemetry is opt-in only: when `OTEL_EXPORTER_OTLP_ENDPOINT` is set the SDK emits per-command spans and counters via `obs`. The SDK never phones home and never auto-updates. It demonstrates the Glacier promise that every line you do not write is one Glacier handles for you.
<!-- /magpie:extract -->

## Mental model

<!-- magpie:extract source=specs/0032-sdk.md section=mental-model source-checksum=<TODO> -->
The Glacier SDK is two things at once. To a Glacier developer it is a productivity tool: scaffold a project, generate boilerplate, lint, test, ship. To the framework itself it is a stress test: every package shipped by Glacier is exercised by at least one SDK command, and a Lynx-owned coverage row fails CI if any package falls out of use. The SDK and the framework move in lockstep.
<!-- /magpie:extract -->

## Get started

```sh
go install github.com/nathanbrophy/glacier/cmd/glacier@latest
glacier init my-app --yes
cd my-app && go mod tidy
```

See [Install](./install.md) for per-OS instructions and shell-completion setup.

## Commands

| Command | What it does |
|---|---|
| [`glacier vibe`](./commands/vibe.md) | Glacier vibes animation |
| [`glacier version`](./commands/version.md) | Print version; `--check` compares against latest |
| [`glacier generate`](./commands/generate.md) | Run all code generators |
| [`glacier lint`](./commands/lint.md) | gofmt, vet, staticcheck, Glacier-specific lints |
| [`glacier test`](./commands/test.md) | Live status panel and summary |
| [`glacier init`](./commands/init.md) | Scaffold a new Glacier project |
| [`glacier new`](./commands/new.md) | Add a package, command, or option |
| [`glacier completions`](./commands/completions.md) | Shell completions |
| [`glacier explain`](./commands/explain.md) | Explain a marker, exit code, or config key |

Full command index: [Commands](./commands/index.md)

## Codegen gallery

Nine annotated examples show the three Glacier generators in action. Each exhibit is a self-contained `in.go` + `out.go` pair that CI keeps byte-exact.

[Browse the codegen gallery](./codegen/index.md)

## Configuration

Config file, environment variables, global flags, and exit-code table: [Configuration](./configuration.md)
