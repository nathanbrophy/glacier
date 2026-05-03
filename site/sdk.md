---
title: Glacier SDK
aside: false
---

# Glacier SDK

<div class="glacier-sprite-accent">
  <MascotSprite state="thinking" :size="96" />
</div>

The Glacier SDK CLI binary is deferred to spec 0032. This page is a placeholder.

## What the SDK will be

The Glacier SDK is a separate deliverable from the Glacier framework libraries. Where the framework is a suite of importable Go packages, the SDK is a CLI binary - `glacier` - that you install once and use across projects.

The SDK has two main components:

**The `glacier` CLI.** A developer-facing command-line tool that covers project scaffolding, code generation (`glaciergen`), and local development workflows. Every subcommand is built with the `cli` package from the Glacier framework. Running `glacier --version` renders the full wordmark via the `cli` package's banner feature. Running `glacier --help` demonstrates the framework's help-text conventions. The `glacier` binary is itself the proof that the `cli` package is production-grade.

**The `glaciergen` codegen tool.** A code generation command that reads `+glacier:cmd`, `+glacier:mock`, and other marker annotations from Go source and emits the wiring code you would otherwise write by hand. It ships as both a standalone `go run` target and as `glacier gen` in the SDK CLI.

## How it dogfoods the framework

The SDK is the framework's own longest-running integration test.

- The `cli` package's banner feature (`//go:embed assets/logo/wordmark.txt`) is exercised on every `glacier --version` call.
- Signal handling routes through `internal/sigh`, the same package all other `cli`-based programs use.
- Configuration for the `glacier` tool itself is loaded via `conf`, using the same layered-sources pattern the framework recommends.
- Structured logs from the SDK's subcommands flow through `log`, demonstrating context-attribute propagation in a real program.
- Any `httpc` call the SDK makes (for checking latest releases, for example) uses the same `httpc` client with retry logic the framework provides.

If a framework feature doesn't work well enough for the SDK to use it cleanly, that's a bug in the framework - surfaced immediately by dogfooding before it reaches framework consumers.

## When

The SDK ships after the core packages stabilize at v0. The `cli` package (spec 0011), `mock` codegen (spec 0012), and `glaciergen` shape (deferred from spec 0002) need to be accepted and implemented first. Spec 0032 will define the full SDK scope and CLI surface.

In the meantime, the framework libraries are available at `github.com/nathanbrophy/glacier`. The 14-package suite is the primary artifact. The SDK is icing on the cake.
