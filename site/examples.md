---
title: Examples
aside: false
---

# Examples

Short recipes showing Glacier solving a concrete problem. Each snippet is self-contained. Follow the walkthrough link for the full task page with context and explanation.

---

## Build a CLI in 10 lines

Define a struct with struct tags for flags, implement `Run(ctx)`, call `cli.Main`.

```go
package main

import (
    "context"
    "fmt"
    "github.com/nathanbrophy/glacier/cli"
)

// +glacier:command name=greet
// +glacier:root
type Greet struct {
    // +glacier:default "world"
    // +glacier:usage who to greet
    Name string
}

func (g *Greet) Run(ctx context.Context) error {
    fmt.Printf("Hello, %s!\n", g.Name)
    return nil
}

func main() { cli.Main(&Greet{}) }
```

[Walkthrough: Building a CLI](/docs/building-a-cli)

---

## Mock an HTTP API in tests

Register stubs on an `httpmock.Transport` and inject it into the client under test. Strict mode (the default) fails the test if an unregistered request arrives.

```go
func TestFetchUser(t *testing.T) {
    transport, _ := httpmock.New(httpmock.Strict())
    httpmock.Register(transport,
        httpmock.GET("/users/1").RespondWith(
            httpmock.JSON(200, User{ID: "1", Name: "Alice"}),
        ),
    )

    svc := NewUserService(&http.Client{Transport: transport})
    got, err := svc.Fetch(t.Context(), "1")

    assert.NoError(t, err)
    assert.Equal(t, "Alice", got.Name)
}
```

[Walkthrough: Mocking HTTP](/docs/mocking-http)

---

## Layered config from defaults, env, and flags

Construct a loader with multiple sources. Lower layers provide defaults; higher layers override. `Register[T]` returns a typed accessor - no type assertions at call sites.

```go
loader, err := conf.New(
    conf.WithDefaults(map[string]any{
        "server.port": "8080",
        "log.level":   "info",
    }),
    conf.WithEnvPrefix("APP"),
    conf.WithFlags(flag.CommandLine),
)
if err != nil {
    log.Fatal(err)
}

port    := conf.Register[string](loader, "server.port")
level   := conf.Register[string](loader, "log.level")

fmt.Printf("port=%s level=%s\n", port(), level())
// APP_SERVER_PORT=9090 in env: port=9090 level=info
```

[Walkthrough: Loading config](/docs/loading-config)

---

## Structured logs with context

Attach attributes to the context once. Every downstream log call includes them without passing a logger explicitly.

```go
func HandleRequest(ctx context.Context, r *Request) error {
    ctx = log.With(ctx,
        slog.String("request_id", r.ID),
        slog.String("user_id",    r.UserID),
    )

    log.From(ctx).Info("handling request",
        slog.String("method", r.Method),
        slog.String("path",   r.Path),
    )

    if err := process(ctx, r); err != nil {
        log.From(ctx).Error("request failed", slog.Any("error", err))
        return err
    }

    log.From(ctx).Info("request complete")
    return nil
}
```

[Walkthrough: Structured logging](/docs/structured-logging)

---

## Concurrency with panic recovery

`concur.Group` runs goroutines concurrently, recovers panics, and surfaces them as errors via `Wait`.

```go
func processAll(ctx context.Context, items []Item) error {
    g := concur.NewGroup(ctx)

    for _, item := range items {
        item := item // capture
        g.Go(func(ctx context.Context) error {
            return process(ctx, item)
        })
    }

    return g.Wait() // returns first non-nil error; panics are errors too
}
```

[Walkthrough: Concurrency](/docs/concurrency)

---

## Lazy iteration over a sequence

`fluent` operators compose lazily. No intermediate slices are allocated until a sink (`Collect`, `First`, `Reduce`) pulls from the pipeline.

```go
rows := queryRows(ctx, db) // iter.Seq[Row]

activeNames := fluent.Collect(
    fluent.Take(
        fluent.Map(
            fluent.Filter(rows, func(r Row) bool { return r.Active }),
            func(r Row) string { return r.Name },
        ),
        50,
    ),
)
// activeNames is []string of the first 50 active row names.
```

---

## OpenTelemetry traces and metrics

Construct an `obs.Provider`, defer shutdown, and start spans. When the context carries no tracer, spans are no-ops with no allocation.

```go
provider, err := obs.New(ctx,
    obs.WithOTLPEndpoint("otelcol:4317"),
    obs.WithServiceName("payments"),
)
if err != nil {
    return err
}
defer provider.Shutdown(ctx)

func charge(ctx context.Context, amount int64) error {
    ctx, span := provider.Tracer("payments").Start(ctx, "charge")
    defer span.End()

    span.SetAttributes(attribute.Int64("amount_cents", amount))
    return gateway.Charge(ctx, amount)
}
```

[Walkthrough: Observability](/docs/observability)

---

## Test assertions with smart deep-compare

`assert.Equal[T]` accepts modifiers that adjust comparison semantics without custom comparators.

```go
func TestListUsers(t *testing.T) {
    got, err := svc.List(t.Context(), ListRequest{Active: true})

    assert.NoError(t, err)
    assert.Equal(t, wantUsers, got,
        assert.IgnoreOrder(),            // slice order doesn't matter
        assert.IgnoreFields("UpdatedAt"), // skip volatile timestamps
    )
    assert.Len(t, got, 3)
}
```

[Walkthrough: Writing tests](/docs/writing-tests)
