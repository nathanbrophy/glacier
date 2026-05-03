---
title: Features
aside: false
---

# Features

14 packages across three tiers. Each package stands alone. Together they are Glacier. The tiers enforce a no-cycles constraint: kernel packages have no internal dependencies; mid-tier packages depend only on the kernel; leaf packages depend only on kernel and mid-tier, never on each other. See [Concepts](/concepts) for the full mental model.

---

## Tier 0: Kernel

The five kernel packages are universal. Every consumer of any Glacier package transitively depends on these. They are small, fast, and stable by design.

### option <TierBadge tier="kernel" />

Functional options. `Apply`, `Validate`, `Required`, `Option[T]`: every Glacier package configurable at construction speaks this protocol. `option.OptionFunc[T]` adapts a closure into an `Option[T]` without boilerplate. `option.Apply` applies a slice of options to a zero-valued config struct, stopping at first error by default or collecting all errors with `Strict()` mode.

```go
func WithTimeout(d time.Duration) option.Option[clientConfig] {
    return option.OptionFunc[clientConfig](func(c *clientConfig) error {
        if d <= 0 {
            return errors.New("httpc: WithTimeout: duration must be positive")
        }
        c.timeout = d
        return nil
    })
}
```

[Read the docs](/docs/packages/option)

---

### errs <TierBadge tier="kernel" />

Error stories without ceremony. `Wrap`, `Join`, `Chain`, `Sentinel`, `IsAny`, `Retryable`, `Coded`: tree-walking, classification, library and CLI registers. `errs.Wrap` attaches a prefix and optional stack trace. `errs.Chain` produces a lazy `iter.Seq[error]` for depth-first error-tree traversal. `errs.Sentinel` panics at construction time if the error message violates the library register, keeping error format discipline automatic.

```go
var ErrNotFound = errs.Sentinel("store: record not found")

func (s *Store) Get(ctx context.Context, id string) (*Record, error) {
    r, err := s.db.Query(ctx, id)
    if err != nil {
        return nil, errs.Wrap(err, "store: get")
    }
    return r, nil
}
```

[Read the docs](/docs/packages/errs)

---

### log <TierBadge tier="kernel" />

Structured logging on `log/slog` with `Trace` and `Notice` levels, brand-palette TTY color, context attribute attachment, and a `Redact` helper. `log.NewHandler` and `log.NewJSONHandler` construct `slog.Handler` implementations with Glacier's level set and color support. `log.With(ctx, attrs...)` attaches attributes to the context so every downstream log call includes them without threading a logger.

```go
ctx = log.With(ctx,
    slog.String("request_id", reqID),
    slog.String("user_id", userID),
)
// All log calls below automatically include request_id and user_id.
log.From(ctx).Info("handling request", slog.String("path", r.URL.Path))
```

[Read the docs](/docs/packages/log)

---

### assert <TierBadge tier="kernel" />

Test assertions and runtime invariants. `Equal[T]` with smart deep-compare, `Must` for init-time panics, `require/` for halt-on-failure. `assert.Equal` accepts option modifiers: `IgnoreOrder()`, `IgnoreCase()`, `WithDelta(d)`, `IgnoreFields(names...)`. The `require` sub-package mirrors every assertion with `t.FailNow()` semantics for cases where continuing after a failure is misleading.

```go
assert.Equal(t, wantUsers, gotUsers,
    assert.IgnoreOrder(),
    assert.IgnoreFields("UpdatedAt"),
)
assert.NoError(t, err)
assert.Match(t, `^user:\d+$`, got.ID)
```

[Read the docs](/docs/packages/assert)

---

### term <TierBadge tier="kernel" />

Terminal as first-class output. Capability detection, 24-bit ANSI styling, glyph registry, beauty-writer layout, prompts, animation. `term.Detect()` reads `COLORTERM`, `TERM`, `NO_COLOR`, `GLACIER_NO_COLOR` and resolves the right color mode. `term.Style` wraps a string with the correct ANSI SGR escape for a given palette token. The beauty-writer handles truncation, padding, and column alignment for structured terminal output.

```go
w := term.NewWriter(os.Stdout, term.WithColor(term.ColorAuto))
w.Println(term.Style("Ready", term.TokenSuccess))
w.Printf("Listening on :%s\n", port)
```

[Read the docs](/docs/packages/term)

---

## Tier 1: Mid

The five mid-tier packages are independent of each other. Each depends only on the kernel. Pick one, pick all five - you won't drag in unrelated packages.

### concur <TierBadge tier="mid" />

Concurrency primitives that play with context: `Mutex`, `RWMutex`, `Group` with panic recovery, `Semaphore`, `Pool[T]`, `Once[T]`, `WaitGroup`. `concur.Group` runs goroutines and recovers panics, surfacing them as errors rather than crashing the process. `concur.Mutex.LockCtx(ctx)` acquires with cancellation support - useful when a lock contention could outlast a request deadline.

```go
g := concur.NewGroup(ctx)
g.Go(func(ctx context.Context) error {
    return processChunk(ctx, chunk)
})
if err := g.Wait(); err != nil {
    return err
}
```

[Read the docs](/docs/packages/concur)

---

### fluent <TierBadge tier="mid" />

Lazy `iter.Seq` pipeline operators: `Map`, `Filter`, `Take`, `Window`, `GroupBy`, joins, set ops, aggregations. Generics-first; zero deps. Sequences are lazy by default - `fluent.Of` wraps any `iter.Seq[T]`, and operators chain without allocating intermediate slices. Evaluation happens only when a sink (`Collect`, `First`, `Reduce`) pulls from the sequence.

```go
results := fluent.Collect(
    fluent.Take(
        fluent.Filter(
            fluent.Map(rows, parseRow),
            func(r Row) bool { return r.Active },
        ),
        100,
    ),
)
```

[Read the docs](/docs/packages/fluent)

---

### conf <TierBadge tier="mid" />

Layered configuration with atomic snapshots. Defaults, JSON file, env, flags, overrides: `Register[T]` returns a typed accessor. Layers are applied in priority order (defaults lowest, explicit overrides highest). The registry stores an atomic snapshot; reads never block writers. `conf.Register[T]` binds a typed key - the returned accessor func always returns the current snapshot value.

```go
loader, _ := conf.New(
    conf.WithFile("config.json"),
    conf.WithEnvPrefix("APP"),
    conf.WithFlags(flag.CommandLine),
)

logLevel := conf.Register[string](loader, "log.level",
    conf.WithDefault("info"),
)
fmt.Println(logLevel()) // "debug" if APP_LOG_LEVEL=debug in env
```

[Read the docs](/docs/packages/conf)

---

### fixture <TierBadge tier="mid" />

Test resources: golden files, typed snapshots, deterministic fake clocks, in-memory filesystems, leak guards for goroutines, FDs, env vars. `fixture.Golden` reads a `.golden` file from `testdata/` and diffs against it, auto-updating with `-update`. `fixture.Clock` returns a fake `time.Time` source that advances on demand. `fixture.GuardLeaks` asserts no goroutine, file descriptor, or environment variable outlives the test.

```go
func TestRenderReport(t *testing.T) {
    got := renderReport(fakeData())
    fixture.Golden(t, "report", got) // testdata/TestRenderReport/report.golden
}
```

[Read the docs](/docs/packages/fixture)

---

### obs <TierBadge tier="mid" />

Opt-in OpenTelemetry: `MeterProvider` and `TracerProvider` via OTLP gRPC, instrumentation hooks for `httpc`, `cli`, `conf`, zero overhead when off. `obs.New` constructs a provider pair. When the context carries no tracer, span creation is a no-op with no allocation. Instrument hooks attach to `httpc`, `cli`, and `conf` via the standard `option.Option[T]` mechanism - no global state involved.

```go
provider, _ := obs.New(ctx,
    obs.WithOTLPEndpoint("otelcol:4317"),
    obs.WithServiceName("my-service"),
)
defer provider.Shutdown(ctx)

ctx, span := provider.Tracer("my-service").Start(ctx, "handle-request")
defer span.End()
```

[Read the docs](/docs/packages/obs)

---

## Tier 2: Leaves

The four leaf packages are large enough to justify isolation. They depend on kernel and mid-tier packages only and never import each other. Consumers who need only `httpmock` for tests don't pull in `cli`.

### cli <TierBadge tier="leaf" />

Build CLIs from a struct and a `Run` method. Comment markers and `glaciergen` codegen emit flag parsing, help text, and routing. The `+glacier:command` marker on a struct generates the wiring code at `go generate` time. Per-field markers like `+glacier:default`, `+glacier:env`, `+glacier:short`, and `+glacier:required` configure each flag. The banner feature embeds `assets/logo/wordmark.txt` via `//go:embed` and renders it with the 6-stop ice gradient on TTY output.

```go
// +glacier:command name=serve
// +glacier:root
type Serve struct {
    // +glacier:default 8080
    // +glacier:short p
    Port string

    // +glacier:env ENABLE_METRICS
    Metrics bool
}

func (s *Serve) Run(ctx context.Context) error {
    return listenAndServe(ctx, ":"+s.Port, s.Metrics)
}
```

[Read the docs](/docs/packages/cli)

---

### mock <TierBadge tier="leaf" />

Interface mocks: `mock.Of[T]` reflect-based, or `+glacier:mock` codegen for typed wrappers. Fluent expectation builder, automatic `Verify` on cleanup. `mock.Of[T]` returns a value satisfying interface `T` backed by a recorder. `mock.Expect(m, "Method").With(...).Return(...)` registers an expectation. `t.Cleanup` triggers `Verify` automatically - no explicit call needed in tests.

```go
db := mock.Of[Database](t)
mock.Expect(db, "GetUser").
    With(mock.AnyContext(), "user-42").
    Return(&User{ID: "user-42"}, nil)

svc := NewService(db)
u, err := svc.FindUser(t.Context(), "user-42")
assert.NoError(t, err)
assert.Equal(t, "user-42", u.ID)
```

[Read the docs](/docs/packages/mock)

---

### httpmock <TierBadge tier="leaf" />

A programmable `http.RoundTripper` for tests. Stub builder, generic `JSON[T]` responses, strict-by-default; testdata fixtures land in JSON. `httpmock.New()` returns a `Transport` that serves registered stubs in order. Strict mode (the default) fails the test if a request arrives with no matching stub. `httpmock.JSON[T](200, v)` serializes `v` and sets the correct `Content-Type` header.

```go
transport, _ := httpmock.New(httpmock.Strict())
httpmock.Register(transport,
    httpmock.GET("/api/users/1").
        RespondWith(httpmock.JSON(200, User{ID: "1", Name: "Alice"})),
)

client := &http.Client{Transport: transport}
resp, _ := client.Get("https://api.example.com/api/users/1")
```

[Read the docs](/docs/packages/httpmock)

---

### httpc <TierBadge tier="leaf" />

Typed HTTP client: `Get[T]`, `Post[T]`, `Put[T]` auto-unmarshal. Retry with backoff, closure-body retry-safe payloads, dry-run via context. `httpc.Get[T](ctx, url)` fetches and unmarshals into `T` in one call. The retry loop re-invokes the body closure on each attempt so the request body is always readable. Dry-run mode (injected via context) records calls without making real network requests.

```go
type User struct { ID string; Name string }

user, err := httpc.Get[User](ctx, "https://api.example.com/users/1",
    httpc.WithRetry(3, httpc.ExponentialBackoff(100*time.Millisecond)),
)
if err != nil {
    return err
}
fmt.Println(user.Name) // Alice
```

[Read the docs](/docs/packages/httpc)
