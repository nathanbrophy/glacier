---
title: Documentation
---

# Documentation

<div class="glacier-sprite-accent">
  <MascotSprite state="idle" :size="72" />
</div>

The docs here are organized by task. Each task page shows a workflow that composes two or more Glacier packages to solve a concrete problem. If you want the full API for a single package, see the **Packages** section in the sidebar.

<div class="glacier-task-grid">

<div class="glacier-task-card">

### [Building a CLI](/docs/building-a-cli)

Define commands as Go structs, run `glaciergen`, and ship a binary with flag parsing, env binding, signal handling, and the Glacier banner - no boilerplate.

`cli` · `option` · `errs` · `log` · `term` · `conf`

</div>

<div class="glacier-task-card">

### [Writing tests](/docs/writing-tests)

Assert values with a smart deep-equal engine, inject fake clocks, and verify interface expectations - all in idiomatic table-driven tests.

`assert` · `fixture` · `mock` · `errs`

</div>

<div class="glacier-task-card">

### [Mocking HTTP](/docs/mocking-http)

Stub HTTP responses with typed stubs, verify every request was made, and keep tests hermetic with zero real network calls.

`httpmock` · `httpc` · `fixture`

</div>

<div class="glacier-task-card">

### [Loading config](/docs/loading-config)

Layer defaults, a JSON file, environment variables, and flags into a single atomic snapshot that any package can read concurrently.

`conf` · `option` · `errs` · `log`

</div>

<div class="glacier-task-card">

### [Structured logging](/docs/structured-logging)

Attach request context once and let it flow into every log record downstream. Use `Trace`/`Notice` levels and `Redact` for secrets.

`log` · `errs` · `obs`

</div>

<div class="glacier-task-card">

### [Observability](/docs/observability)

Initialize OTLP providers once, opt packages into tracing and metrics, and get `trace_id` in every log record automatically.

`obs` · `log` · `httpc`

</div>

<div class="glacier-task-card">

### [Concurrency](/docs/concurrency)

Run bounded goroutine groups with panic recovery, coordinate access with ctx-aware mutexes, and diagnose stuck locks without touching production builds.

`concur` · `errs` · `log`

</div>

</div>

For per-package API reference, use the **Packages** section in the left sidebar.
