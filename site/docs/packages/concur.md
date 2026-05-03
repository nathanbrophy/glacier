---
title: concur
---

# concur

<TierBadge tier="mid" />

<UsedInTasksBadges package-name="concur" />

[View source spec &rarr;](https://github.com/nathanbrophy/glacier/blob/main/specs/0007-concur.md)

## Public summary
<!-- magpie:extract source=specs/0007-concur.md section=public-summary source-checksum=PENDING -->

`concur` is Glacier's concurrency-primitives package: a curated set of ctx-aware sync types that complement Go's standard library `sync` and `golang.org/x/sync/errgroup`. Every primitive matches or beats its stdlib equivalent in production builds. `Mutex` and `RWMutex` are byte-for-byte identical to their stdlib counterparts at runtime. `Semaphore` is atomic-counter-backed with a lock-free fast path. When built with the `glacier_debug` build tag, `Mutex` and `RWMutex` gain hold-timeout diagnostics and caller-stack capture that surface stuck locks as structured slog events, with zero overhead in production. `Group` is an errgroup that collects every goroutine error (not first-wins), caps concurrency by default at `runtime.NumCPU() * 64`, recovers panics as typed `PanicError` values, and offers a non-blocking `TryGo` for back-pressure-aware scheduling. `Pool[T]` and `Once[T]` are generic, type-safe wrappers over `sync.Pool` and Go's once semantics. `WaitGroup` embeds `sync.WaitGroup` and adds a ctx-aware `WaitCtx`. Cancellation is uniform: every blocking operation accepts a `context.Context` and returns `ErrCancelled` on cancel.

<!-- /magpie:extract -->

## Mental model
<!-- magpie:extract source=specs/0007-concur.md section=mental-model source-checksum=PENDING -->

There are two modes, and zero cost to choosing wisely:

**Production mode** (default build): `concur.Mutex` is `sync.Mutex` under the hood. There are no extra fields, no goroutines, no timers. `Lock` and `Unlock` are inlined forwarding calls. If you measure performance against a raw `sync.Mutex`, the delta is within noise.

**Debug mode** (`-tags glacier_debug`): the same `Mutex` type gains two extra fields, a caller-stack capture and a hold-timer. When a goroutine holds the lock longer than the configured timeout (default 5 s), a structured slog event fires: who acquired it, who is waiting, how long it has been held. The lock itself is unaffected; diagnostics are observers, not controllers.

```
Default build                      glacier_debug build
──────────────────────             ──────────────────────────────────────
concur.Mutex                       concur.Mutex
  └─ sync.Mutex                      ├─ sync.Mutex
     (zero overhead)                 ├─ callerStack string
                                     └─ holdTimer  *time.Timer
                                              │
                                              └─ slog event on timeout
```

`Group` follows a simple lifecycle: create, schedule, wait.

```
NewGroup(WithLimit(n))
    │
    ├─ Go(fn)    ─── schedules fn in a new goroutine (blocks if at cap)
    ├─ TryGo(fn) ─── schedules if slot available; returns false otherwise
    │
    └─ WaitDone(ctx)
           │
           ├─ waits for all goroutines to finish (or ctx to cancel)
           └─ returns errs.Join over all collected errors
                   (including PanicError for panicking goroutines)
```

`Group.Go` after `WaitDone` has returned panics. This matches `sync.WaitGroup.Add` after `Wait` and is intentional: reusing a finished Group is a programming error, not a runtime condition.

`Semaphore` uses a fast-path atomic CAS when permits are available. The slow path acquires a `sync.Cond` lock and parks. Each slow-path waiter spawns a cancel-watcher goroutine that holds a derived context; on successful acquire, `defer cancel()` cleans up that watcher immediately, with no goroutine leak on the happy path or the cancellation path.

<!-- /magpie:extract -->

## API
<!-- magpie:extract source=specs/0007-concur.md section=api source-checksum=PENDING -->

### Sentinels

```go
// ErrCancelled is returned by any blocking concur operation when the caller's
// context is cancelled. Wrap semantics: errors.Is(err, context.Canceled) and
// errors.Is(err, context.DeadlineExceeded) both hold when appropriate.
var ErrCancelled = errs.Sentinel("concur: cancelled")

// ErrInvalidPermits is returned by Semaphore when n is non-positive or
// exceeds the semaphore's capacity at construction time.
var ErrInvalidPermits = errs.Sentinel("concur: invalid permits")
```

### PanicError

```go
// PanicError is the typed error wrapping a value recovered from a panicking
// goroutine inside Group.Go. It is included in the errs.Join result of
// WaitDone alongside any ordinary errors.
type PanicError struct {
    Value any
}

func (e *PanicError) Error() string
func (e *PanicError) Unwrap() error
```

### Mutex

```go
// Mutex is a mutual-exclusion lock. It implements sync.Locker.
// In default builds, Mutex is byte-equivalent to sync.Mutex.
// In glacier_debug builds, Mutex additionally captures the acquiring
// goroutine's caller stack and starts a hold timer. Must not be copied
// after first use.
type Mutex struct{ /* fields vary by build tag */ }

func (m *Mutex) Lock()
func (m *Mutex) Unlock()

// LockCtx acquires the mutex, blocking until the lock is available or ctx is
// cancelled. Returns ErrCancelled (wrapping ctx.Err()) if ctx fires first.
// Cancellation is best-effort: up to one backoff interval (default 1 ms) may
// elapse after ctx cancels before LockCtx returns.
func (m *Mutex) LockCtx(ctx context.Context) error
```

### RWMutex

```go
// RWMutex is a reader/writer mutual-exclusion lock. In default builds it is
// byte-equivalent to sync.RWMutex. In glacier_debug builds it gains hold-timeout
// diagnostics. Upgrade from RLock to Lock is not supported. Must not be copied
// after first use.
type RWMutex struct{ /* fields vary by build tag */ }

func (m *RWMutex) RLock()
func (m *RWMutex) RUnlock()
func (m *RWMutex) Lock()
func (m *RWMutex) Unlock()

// RLockCtx acquires a shared read lock, returning ErrCancelled if ctx fires first.
func (m *RWMutex) RLockCtx(ctx context.Context) error

// LockCtx acquires the exclusive write lock, returning ErrCancelled if ctx fires first.
func (m *RWMutex) LockCtx(ctx context.Context) error
```

### Group

```go
// NewGroup constructs a Group with the given options.
// Default concurrency cap: runtime.NumCPU() * 64.
// Returns (nil, error) if any option is invalid.
func NewGroup(opts ...option.Option[groupConfig]) (*Group, error)

// WithLimit caps the maximum concurrent goroutines. n must be >= 1.
// n < 0 acts as WithUnlimited().
func WithLimit(n int) option.Option[groupConfig]

// WithUnlimited removes the default concurrency cap.
func WithUnlimited() option.Option[groupConfig]

// Go schedules fn in a new goroutine. Blocks if at the concurrency cap.
// Panics in fn are recovered as *PanicError included in WaitDone's result.
// PANICS if called after WaitDone has returned.
func (g *Group) Go(fn func() error)

// TryGo schedules fn if a slot is available. Returns false if at cap or
// after WaitDone. Never blocks.
func (g *Group) TryGo(fn func() error) bool

// WaitDone blocks until all goroutines finish or ctx cancels.
// Returns errs.Join of all collected errors when all goroutines finish.
// Returns ErrCancelled wrapping ctx.Err() if ctx fires first.
// After WaitDone returns, the Group is terminal: Go panics.
func (g *Group) WaitDone(ctx context.Context) error
```

### Semaphore

```go
// NewSemaphore constructs a counted semaphore with the given capacity.
// Panics if capacity < 1.
func NewSemaphore(capacity int64) *Semaphore

// Acquire blocks until n permits are available or ctx cancels.
// Returns ErrInvalidPermits if n <= 0 or n > capacity.
// Returns ErrCancelled (wrapping ctx.Err()) on cancellation.
// Fast path: single atomic CAS (allocation-free, lock-free).
func (s *Semaphore) Acquire(ctx context.Context, n int64) error

// TryAcquire attempts to acquire n permits without blocking.
// Returns true if acquired, false if not enough permits available.
// Returns ErrInvalidPermits if n <= 0 or n > capacity.
func (s *Semaphore) TryAcquire(n int64) bool

// Release returns n permits to the semaphore. Release(0) is a no-op.
// PANICS if releasing would cause total available permits to exceed capacity.
func (s *Semaphore) Release(n int64)
```

### Pool[T]

```go
// Pool[T] is a typed wrapper over sync.Pool. Get returns T directly,
// eliminating any type-assertion at the call site.
type Pool[T any] struct{ /* ... */ }

// NewPool constructs a Pool. newFn is called when the pool is empty and
// Get has no pooled values. If newFn is nil, Get returns the zero value of T.
func NewPool[T any](newFn func() T) *Pool[T]

func (p *Pool[T]) Get() T
func (p *Pool[T]) Put(v T)
```

### Once[T]

```go
// Once[T] runs a function exactly once and memoizes its (T, error) result.
// ctx is passed to fn on the first call only. If fn panics, the panic propagates
// and Once is not marked done; the next call re-attempts.
type Once[T any] struct{ /* ... */ }

// Do calls fn the first time and memoizes the result.
// All subsequent calls return the memoized (T, error).
func (o *Once[T]) Do(ctx context.Context, fn func(context.Context) (T, error)) (T, error)
```

### WaitGroup

```go
// WaitGroup embeds sync.WaitGroup (inheriting Add, Done, and Wait) and
// adds WaitCtx for context-aware waiting.
type WaitGroup struct {
    sync.WaitGroup
}

// WaitCtx blocks until the counter reaches zero or ctx cancels.
// Returns nil on zero; returns ErrCancelled wrapping ctx.Err() if ctx fires first.
func (wg *WaitGroup) WaitCtx(ctx context.Context) error
```

<!-- /magpie:extract -->

## Examples
<!-- magpie:extract source=specs/0007-concur.md section=examples source-checksum=PENDING -->

Use `Mutex.LockCtx` when acquiring a lock inside a request handler that has a deadline:

```go
func ExampleMutex_LockCtx() {
    var mu concur.Mutex
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    if err := mu.LockCtx(ctx); err != nil {
        // ctx expired before the lock was available
        fmt.Println("lock not acquired:", err)
        return
    }
    defer mu.Unlock()
    // critical section
}
```

Use `Group` to run a bounded fan-out and collect all errors:

```go
func ExampleNewGroup() {
    items := []string{"a", "b", "c", "d", "e"}

    g, err := concur.NewGroup(concur.WithLimit(3)) // at most 3 goroutines running
    if err != nil {
        log.Fatal(err)
    }
    for _, item := range items {
        item := item
        g.Go(func() error {
            return process(item)
        })
    }
    if err := g.WaitDone(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

Use `Semaphore` with `WaitGroup` to rate-limit concurrent HTTP requests:

```go
func ExampleSemaphore() {
    sem := concur.NewSemaphore(10) // max 10 concurrent requests

    var wg concur.WaitGroup
    for _, url := range urls {
        url := url
        if err := sem.Acquire(ctx, 1); err != nil {
            break // ctx cancelled; remaining urls skipped
        }
        wg.Add(1)
        go func() {
            defer wg.Done()
            defer sem.Release(1)
            client.Get(url)
        }()
    }
    wg.WaitCtx(ctx)
}
```

Use `Pool[T]` to reuse buffers with zero allocations on the hot path:

```go
func ExampleNewPool() {
    bufPool := concur.NewPool(func() *bytes.Buffer { return new(bytes.Buffer) })

    buf := bufPool.Get()
    defer func() {
        buf.Reset()
        bufPool.Put(buf)
    }()
    // use buf for temporary work; zero allocations on the hot path
}
```

<!-- /magpie:extract -->

## FAQ
<!-- magpie:extract source=specs/0007-concur.md section=faq source-checksum=PENDING -->

<div class="glacier-faq">

**Why does Group default to `runtime.NumCPU() * 64` goroutines instead of being truly unlimited?**

An unbounded Group is a goroutine-explosion foot-gun: a tight loop calling `Go` can queue millions of goroutines before the work finishes, exhausting memory. `NumCPU * 64` is generous enough for almost all real workloads while preventing accidental abuse. When you genuinely need unlimited goroutines because you manage back-pressure externally, pass `WithUnlimited()` and document the rationale at the call site.

**Why does `Group.Go` panic after `WaitDone` instead of returning an error?**

Same reason `sync.WaitGroup.Add` panics after `Wait`: reusing a finished Group after `WaitDone` is a programming error, not a runtime condition. Errors are for expected failure modes; panics are for invariant violations. A panicking call site is unambiguously wrong; an error-returning `Go` would be silently swallowed by code that forgets to check. If you need a re-usable Group, create a new one.

**`Mutex.LockCtx` says "best-effort cancellation" -- what does that mean?**

`LockCtx` uses a try-lock-with-backoff loop internally. On each iteration it attempts a non-blocking lock acquisition; if that fails and the ctx is still live, it sleeps for a backoff interval (default 1 ms) before retrying. If ctx cancels during a sleep, LockCtx will not detect it until the sleep ends, up to one backoff interval of latency. If you need stricter ctx-response, structure your code so the critical section timeout is much larger than the backoff.

**Does `Once[T]` memoize an error result?**

Yes. `Once[T].Do` memoizes the `(T, error)` pair returned by fn, including when the error is non-nil and T is the zero value. The second caller receives the same `(zero, err)` as the first; the fn is not retried on error. Only a fn that panics causes a re-attempt.

**Why is there no `Close` on `Semaphore`, `Pool`, or `Once`?**

`Close` implies ownership of an external resource (socket, file, goroutine leak) that must be explicitly freed. None of these types own such resources: `Semaphore` is an in-memory counter, `Pool` is GC-managed by `sync.Pool`, and `Once` is a one-shot state machine. Adding `Close` would create a false expectation that not calling it leaks something.

</div>

<!-- /magpie:extract -->
