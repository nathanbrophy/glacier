---
title: Observability
---

# Observability

<PackagesUsedBadges :package-names="['obs', 'log', 'httpc']" />

[`obs`](/docs/packages/obs) is Glacier's OpenTelemetry-based observability package. Call `obs.Init` once at program start to configure a `MeterProvider` and `TracerProvider` backed by an OTLP gRPC exporter, then opt individual packages into instrumentation via their options. When instrumentation is disabled, overhead is exactly zero - no allocations, no latency. When a span is active in a context, `trace_id` and `span_id` appear in every log record in that context automatically.

## Walkthrough

### Step 1 - Initialize providers at startup

`obs.Init` configures both providers, sets `obs.Default`, and wires the OTLP gRPC exporter. If `OTEL_EXPORTER_OTLP_ENDPOINT` is not set in the environment, the providers are no-ops with zero overhead.

```go
import (
    "context"
    "github.com/nathanbrophy/glacier/obs"
)

func main() {
    ctx := context.Background()

    prov, err := obs.Init(
        obs.WithResourceAttribute("service.name",    "my-service"),
        obs.WithResourceAttribute("service.version", "v1.2.3"),
        obs.WithSampler(obs.ParentBased(obs.TraceIDRatioBased(0.1))),
    )
    if err != nil {
        panic(err)
    }
    defer prov.Shutdown(ctx)

    // obs.Default is now set; all package-level helpers use it.
    runServer(ctx)
}
```

`prov.Shutdown(ctx)` flushes pending spans and metrics before the process exits. It is idempotent.

### Step 2 - Opt httpc into tracing and metrics

Pass `httpc.WithTracing()` and `httpc.WithMetrics()` when constructing an `httpc.Client`. Every subsequent `Get`, `Post`, or `Do` call emits a span and increments the request counter automatically.

```go
import (
    "github.com/nathanbrophy/glacier/httpc"
    "github.com/nathanbrophy/glacier/obs"
)

client, err := httpc.New(
    httpc.WithTracing(),  // emits a span per request
    httpc.WithMetrics(),  // emits http.requests counter + latency histogram
)
if err != nil {
    return err
}
defer client.Close()

user, _, err := httpc.Get[User](ctx, "https://api.example.com/users/42",
    httpc.WithClient(client),
)
```

`obs.Init` must be called before the client makes its first request; the client captures the tracer and meter at construction time.

### Step 3 - Add your own spans

Three lines to instrument any function:

```go
import "github.com/nathanbrophy/glacier/obs"

func processOrder(ctx context.Context, orderID string) error {
    ctx, span := obs.StartSpan(ctx, "process.order",
        obs.WithSpanKind(obs.SpanKindInternal),
        obs.WithAttributes(obs.String("order.id", orderID)),
    )
    defer span.End()

    if err := validateOrder(ctx, orderID); err != nil {
        span.RecordError(err)
        span.SetStatus(obs.StatusError, err.Error())
        return err
    }
    span.SetStatus(obs.StatusOk, "")
    return nil
}
```

The derived `ctx` carries the active span. Any child span started from it is automatically linked as a child in the trace.

### Step 4 - Declare typed counters and histograms

Generic instrument constructors eliminate cast boilerplate. Declare at package level; `Add` and `Record` are no-ops until `obs.Init` is called.

```go
import "github.com/nathanbrophy/glacier/obs"

var (
    requestCount = obs.Counter[int64]("http.requests",
        obs.WithDescription("Total HTTP requests served"),
        obs.WithUnit("req"),
    )
    requestLatency = obs.Histogram[float64]("http.request.duration",
        obs.WithDescription("HTTP request latency"),
        obs.WithUnit("s"),
    )
)

func handleRequest(ctx context.Context, method string, statusCode int, elapsed float64) {
    requestCount.Add(ctx, 1,
        obs.String(obs.KeyHTTPMethod, method),
        obs.Int(obs.KeyHTTPStatusCode, statusCode),
    )
    requestLatency.Record(ctx, elapsed,
        obs.String(obs.KeyHTTPMethod, method),
    )
}
```

### Step 5 - Automatic trace/log correlation

When a span is active in `ctx`, the `log/` package appends `trace_id` and `span_id` to every log record. You write no correlation code.

```go
ctx, span := obs.StartSpan(ctx, "handle")
defer span.End()

slog.InfoContext(ctx, "processing request")
// record: ... trace_id=4bf92f3577b34da6a3ce929d0e0e4736 span_id=00f067aa0ba902b7
```

## Putting it together

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/nathanbrophy/glacier/httpc"
    "github.com/nathanbrophy/glacier/log"
    "github.com/nathanbrophy/glacier/obs"
)

var requestCount = obs.Counter[int64]("api.requests",
    obs.WithDescription("Outbound API requests"),
    obs.WithUnit("req"),
)

func main() {
    log.SetDefault(slog.New(log.NewJSONHandler(os.Stdout)))

    ctx := context.Background()
    prov, err := obs.Init(
        obs.WithResourceAttribute("service.name",    "gateway"),
        obs.WithResourceAttribute("service.version", "v0.1.0"),
    )
    if err != nil {
        slog.Error("obs init failed", slog.Any("err", err))
        return
    }
    defer prov.Shutdown(ctx)

    client, err := httpc.New(
        httpc.WithTracing(),
        httpc.WithMetrics(),
    )
    if err != nil {
        slog.Error("httpc init failed", slog.Any("err", err))
        return
    }
    defer client.Close()

    if err := fetchUser(ctx, client, "42"); err != nil {
        slog.ErrorContext(ctx, "fetch failed", slog.Any("err", err))
    }
}

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func fetchUser(ctx context.Context, client *httpc.Client, id string) error {
    ctx, span := obs.StartSpan(ctx, "fetch.user",
        obs.WithAttributes(obs.String("user.id", id)),
    )
    defer span.End()

    slog.InfoContext(ctx, "fetching user") // trace_id + span_id auto-appended

    user, _, err := httpc.Get[User](ctx, "https://api.example.com/users/"+id,
        httpc.WithClient(client),
    )
    if err != nil {
        span.RecordError(err)
        span.SetStatus(obs.StatusError, err.Error())
        return err
    }

    requestCount.Add(ctx, 1, obs.Int(obs.KeyHTTPStatusCode, 200))
    span.SetStatus(obs.StatusOk, "")
    slog.InfoContext(ctx, "user fetched", slog.String("name", user.Name))
    return nil
}
```

## What's happening underneath

- <TierBadge tier="mid" /> [`obs`](/docs/packages/obs): manages provider lifecycle, typed instruments (`Counter[T]`, `Histogram[T]`, `Gauge[T]`), and span helpers; backed by the OpenTelemetry Go SDK with an OTLP gRPC exporter.
- <TierBadge tier="kernel" /> [`log`](/docs/packages/log): reads the active span from ctx via the OTel `trace` package and appends `trace_id` / `span_id` to every log record; `obs/` does not import `log/`.
- <TierBadge tier="leaf" /> [`httpc`](/docs/packages/httpc): accepts `httpc.WithTracing()` and `httpc.WithMetrics()` to emit per-request spans and counters from the `obs.Default` provider.

## Related

- [Structured logging](/docs/structured-logging) - context attribute attachment and automatic trace correlation.
- [Building a CLI](/docs/building-a-cli) - `cli.WithMetrics()` to instrument CLI commands.
- [Mocking HTTP](/docs/mocking-http) - test the `httpc` client in isolation without real network calls.
