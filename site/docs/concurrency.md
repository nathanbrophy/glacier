---
title: Concurrency
---

# Concurrency

<PackagesUsedBadges :package-names="['concur', 'errs', 'log']" />

[`concur`](/docs/packages/concur) is Glacier's concurrency-primitives package: ctx-aware sync types that complement Go's standard library. Every blocking operation accepts a `context.Context` and returns `ErrCancelled` on cancel. In production builds, `concur.Mutex` is byte-for-byte identical to `sync.Mutex` - zero overhead. Build with `-tags glacier_debug` to add hold-timeout diagnostics that surface stuck locks as structured log events, without touching production performance.

## Walkthrough

### Step 1 - Run a bounded goroutine group

`concur.Group` runs a set of goroutines, caps concurrency, collects every error (not first-wins), and recovers panics as typed `PanicError` values. Use it anywhere you fan out work over a slice.

```go
import (
    "context"
    "github.com/nathanbrophy/glacier/concur"
    "github.com/nathanbrophy/glacier/errs"
)

func processAll(ctx context.Context, items []string) error {
    g, err := concur.NewGroup(concur.WithLimit(8)) // at most 8 goroutines
    if err != nil {
        return errs.Wrap(err, "processAll: new group")
    }

    for _, item := range items {
        item := item
        g.Go(func() error {
            return processItem(ctx, item)
        })
    }

    if err := g.WaitDone(ctx); err != nil {
        // err is errs.Join over all goroutine errors, including PanicError.
        return errs.Wrap(err, "processAll: wait")
    }
    return nil
}
```

`g.Go` blocks when the concurrency cap is reached; `g.TryGo` returns false immediately when no slot is available, for back-pressure-aware scheduling.

### Step 2 - Control access with a ctx-aware mutex

`concur.Mutex` has a standard `Lock`/`Unlock` pair plus `LockCtx`, which returns `context.Err()` if the context expires before the lock is available.

```go
import (
    "context"
    "github.com/nathanbrophy/glacier/concur"
)

type Store struct {
    mu   concur.Mutex
    data map[string]string
}

func (s *Store) Set(ctx context.Context, key, val string) error {
    if err := s.mu.LockCtx(ctx); err != nil {
        return err // context cancelled or deadline exceeded
    }
    defer s.mu.Unlock()
    s.data[key] = val
    return nil
}
```

In production builds `concur.Mutex` is `sync.Mutex`; no extra fields, no goroutines, no timers.

### Step 3 - Bound parallelism with a semaphore

`concur.NewSemaphore(n)` limits how many goroutines may proceed past `Acquire` at once. The fast path is a lock-free atomic CAS; the slow path parks on a `sync.Cond`.

```go
import (
    "context"
    "github.com/nathanbrophy/glacier/concur"
)

func fetchAll(ctx context.Context, urls []string) error {
    sem := concur.NewSemaphore(10) // max 10 concurrent requests

    var wg concur.WaitGroup
    for _, url := range urls {
        url := url
        if err := sem.Acquire(ctx, 1); err != nil {
            break // ctx cancelled; skip remaining
        }
        wg.Add(1)
        go func() {
            defer wg.Done()
            defer sem.Release(1)
            fetch(ctx, url)
        }()
    }
    wg.WaitCtx(ctx)
    return ctx.Err()
}
```

### Step 4 - Enable mutex diagnostics in development

Build with `-tags glacier_debug` to gain hold-timeout diagnostics. When a goroutine holds a `concur.Mutex` longer than the configured threshold (default 5 s), a structured `slog` event fires showing who acquired the lock, who is waiting, and how long the lock has been held.

```sh
go test -tags glacier_debug ./...
go run  -tags glacier_debug ./cmd/myapp
```

No code change is required. The diagnostic fields are compiled in only under the build tag; production binaries are unaffected. Log output arrives via the `log/` package, so it flows through whatever handler your application has configured.

```go
// In glacier_debug builds, this log event fires when the lock is held > 5s:
// level=WARN msg="mutex held too long" holder="myapp/store.go:42" duration=5.001s waiters=2
```

### Step 5 - Log group errors with context

Wrap group errors with `errs.Wrap` and log them with the request context so `trace_id` and any request-scoped attributes appear in the output.

```go
import (
    "context"
    "log/slog"

    "github.com/nathanbrophy/glacier/concur"
    "github.com/nathanbrophy/glacier/errs"
    "github.com/nathanbrophy/glacier/log"
)

func processJob(ctx context.Context, jobID string, items []string) error {
    ctx = log.With(ctx, slog.String("job_id", jobID))

    g, _ := concur.NewGroup(concur.WithLimit(4))
    for _, item := range items {
        item := item
        g.Go(func() error { return processItem(ctx, item) })
    }

    if err := g.WaitDone(ctx); err != nil {
        slog.ErrorContext(ctx, "job failed", slog.Any("error", err))
        return errs.Wrap(err, "processJob: wait")
    }
    slog.InfoContext(ctx, "job complete", slog.Int("items", len(items)))
    return nil
}
```

## Putting it together

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"

    "github.com/nathanbrophy/glacier/concur"
    "github.com/nathanbrophy/glacier/errs"
    "github.com/nathanbrophy/glacier/log"
)

func main() {
    log.SetDefault(slog.New(log.NewHandler(os.Stderr, log.WithLevel(log.LevelInfo))))

    urls := []string{
        "https://api.example.com/a",
        "https://api.example.com/b",
        "https://api.example.com/c",
    }

    if err := fetchAllBounded(context.Background(), urls); err != nil {
        slog.Error("fetch failed", slog.Any("error", err))
        os.Exit(1)
    }
}

func fetchAllBounded(ctx context.Context, urls []string) error {
    sem := concur.NewSemaphore(2)

    g, err := concur.NewGroup(concur.WithLimit(4))
    if err != nil {
        return errs.Wrap(err, "fetchAllBounded: new group")
    }

    for _, url := range urls {
        url := url
        if err := sem.Acquire(ctx, 1); err != nil {
            break
        }
        g.Go(func() error {
            defer sem.Release(1)
            slog.InfoContext(ctx, "fetching", slog.String("url", url))
            result, err := fetch(ctx, url)
            if err != nil {
                return errs.Wrap(err, "fetch "+url)
            }
            fmt.Println(result)
            return nil
        })
    }

    if err := g.WaitDone(ctx); err != nil {
        return errs.Wrap(err, "fetchAllBounded: wait")
    }
    return nil
}
```

## What's happening underneath

- <TierBadge tier="kernel" /> [`concur`](/docs/packages/concur): ctx-aware `Mutex`, `RWMutex`, `Semaphore`, `Group`, `WaitGroup`, `Pool[T]`, and `Once[T]`; zero overhead in production; `glacier_debug` tag adds hold-timeout diagnostics.
- <TierBadge tier="kernel" /> [`errs`](/docs/packages/errs): `errs.Wrap` chains error context across goroutine boundaries; `errs.Join` (used internally by `Group`) collects all goroutine errors without losing any.
- <TierBadge tier="kernel" /> [`log`](/docs/packages/log): context-attribute attachment means job and request IDs appear in every log record emitted from inside a goroutine, as long as the goroutine receives the enriched context.

## Related

- [Structured logging](/docs/structured-logging) - attaching context attributes that flow into goroutine log output.
- [Mocking HTTP](/docs/mocking-http) - combining `concur.Semaphore` with `httpc` for bounded parallel API calls.
- [Writing tests](/docs/writing-tests) - using `fixture.GuardLeaks` to detect goroutine leaks left by a `Group` that was never drained.
