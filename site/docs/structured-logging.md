---
title: Structured logging
---

# Structured logging

<PackagesUsedBadges :package-names="['log', 'errs', 'obs']" />

Glacier's [`log`](/docs/packages/log) package is a thin, opinionated layer over Go's `log/slog`. It adds two extra levels (Trace and Notice), TTY-aware color from the Glacier palette, context-based attribute attachment, and an explicit `Redact` helper for secrets. When an OpenTelemetry span is active in a context, `trace_id` and `span_id` are appended to every log record automatically - no manual correlation.

## Walkthrough

### Step 1 - Configure the handler at startup

Call `log.SetDefault` once, before any logging occurs. The text handler writes human-readable output with color on a TTY; the JSON handler writes newline-delimited JSON for log aggregators.

```go
package main

import (
    "log/slog"
    "os"

    "github.com/nathanbrophy/glacier/log"
)

func main() {
    log.SetDefault(slog.New(log.NewHandler(os.Stderr,
        log.WithLevel(log.LevelInfo),
    )))

    slog.Info("starting", "version", "v0.1")
    // text output: level=INFO msg=starting version=v0.1
}
```

For production, swap to `log.NewJSONHandler(os.Stdout)`. The rest of your code does not change.

### Step 2 - Attach attributes to the context

`log.With` attaches key-value pairs to a context. Every subsequent `slog.*Context` call that uses that context includes those attributes automatically - without the intervening functions knowing or caring they exist.

```go
import (
    "context"
    "log/slog"

    "github.com/nathanbrophy/glacier/log"
)

func handle(ctx context.Context, reqID, userID string) error {
    ctx = log.With(ctx,
        slog.String("request_id", reqID),
        slog.String("user_id",    userID),
    )
    slog.InfoContext(ctx, "handler started")
    // record: level=INFO msg="handler started" request_id=<reqID> user_id=<userID>
    return process(ctx)
}
```

Attach once at the boundary (middleware, CLI command entry, job start); carry the context through.

### Step 3 - Use Trace and Notice levels

Go's standard `slog` has four levels: Debug, Info, Warn, Error. Glacier adds Trace (below Debug) and Notice (between Info and Warn).

```go
import "github.com/nathanbrophy/glacier/log"

// Trace: per-iteration verbose output stripped in production.
slog.Log(ctx, log.LevelTrace, "iterating record", "index", i)

// Notice: important non-warning events worth calling out above Info.
slog.Log(ctx, log.LevelNotice, "config reloaded", "source", "config.json")
```

Glacier's handlers render these as `TRACE` and `NOTICE`. A standard `slog.Handler` that does not know these levels renders them as `DEBUG-4` and `INFO+2`, respectively - they remain valid slog levels everywhere.

### Step 4 - Redact secrets

Any value wrapped in `log.Redact` formats as `[REDACTED]` regardless of which handler is in use. Implement this at the point where the sensitive value enters your code.

```go
import (
    "log/slog"
    "github.com/nathanbrophy/glacier/log"
)

func initDB(ctx context.Context, dsn string) error {
    slog.InfoContext(ctx, "connecting to database",
        slog.String("dsn", log.Redact(dsn).String()),
    )
    // record: level=INFO msg="connecting to database" dsn=[REDACTED]
    return openDB(dsn)
}
```

The redaction is implemented via `slog.LogValuer`, so it works with any `slog.Handler` - text, JSON, or third-party.

### Step 5 - Automatic trace correlation via obs

When you initialize [`obs`](/docs/packages/obs) and start a span, `log.With` automatically appends `trace_id` and `span_id` to every log record in that context. You write no correlation code.

```go
import (
    "context"
    "log/slog"

    "github.com/nathanbrophy/glacier/obs"
)

ctx, span := obs.StartSpan(ctx, "handle.request")
defer span.End()

slog.InfoContext(ctx, "processing request")
// record includes: trace_id=4bf92f3577b34da6a3ce929d0e0e4736 span_id=00f067aa0ba902b7
```

`log/` reads the trace context from the active span in ctx; `obs/` never imports `log/`. The dependency flows in one direction only.

## Putting it together

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/nathanbrophy/glacier/log"
    "github.com/nathanbrophy/glacier/obs"
)

func main() {
    log.SetDefault(slog.New(log.NewHandler(os.Stderr,
        log.WithLevel(log.LevelInfo),
    )))

    prov, err := obs.Init(
        obs.WithResourceAttribute("service.name", "my-service"),
    )
    if err != nil {
        slog.Error("obs init failed", slog.Any("error", err))
        return
    }
    defer prov.Shutdown(context.Background())

    handleRequest(context.Background(), "req-123", "u-42")
}

func handleRequest(ctx context.Context, reqID, userID string) {
    ctx = log.With(ctx,
        slog.String("request_id", reqID),
        slog.String("user_id",    userID),
    )

    ctx, span := obs.StartSpan(ctx, "handle.request")
    defer span.End()

    slog.InfoContext(ctx, "handler started")
    // record: request_id=req-123 user_id=u-42 trace_id=... span_id=...

    apiKey := getAPIKey()
    slog.DebugContext(ctx, "outbound call",
        slog.String("api_key", log.Redact(apiKey).String()),
    )
    // record: api_key=[REDACTED]
}
```

## What's happening underneath

- <TierBadge tier="kernel" /> [`log`](/docs/packages/log): thin layer over `log/slog`; adds Trace/Notice levels, TTY color, `Redact`, and context-attribute attachment via `log.With`.
- <TierBadge tier="kernel" /> [`errs`](/docs/packages/errs): errors wrapped with `errs.Wrap` carry a `"package: action: "` prefix that appears cleanly in `slog.Any("error", err)` output.
- <TierBadge tier="mid" /> [`obs`](/docs/packages/obs): when a span is active in ctx, `log.With` appends `trace_id` and `span_id` automatically; no manual correlation needed.

## Related

- [Observability](/docs/observability) - initialize providers, opt packages into tracing and metrics.
- [Building a CLI](/docs/building-a-cli) - attaching command context to the logger at CLI entry.
- [Loading config](/docs/loading-config) - logging the active configuration snapshot after `conf.Load`.
