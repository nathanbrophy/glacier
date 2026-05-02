---
id: 0007
title: Concur
slug: concur
status: verified
owner-agent: otter
created: 2026-05-01
last-updated: 2026-05-02
supersedes: []
superseded-by: null
reviewers:
  - { agent: otter,  required: true,  signed-off-at: "2026-05-01T00:00:00Z" }
  - { agent: lynx,   required: true,  signed-off-at: "2026-05-01T00:00:00Z" }
  - { agent: falcon, required: true,  signed-off-at: "2026-05-01T00:00:00Z" }
  - { agent: magpie, required: false, signed-off-at: null }
implementing-commits: [0c5ac6f]
verified-at: "2026-05-02T00:00:00Z"
docs-extract:
  - public-summary
  - mental-model
  - api
  - examples
  - faq
---

# Concur

<!--
  Section headers below are STABLE ANCHORS. Magpie extracts content by header,
  so do not rename or reorder them. Doing so is a process change requiring its
  own spec.

  Sections marked **Public** are extracted by Magpie for the public site.
  Sections marked **Internal** are engineering-only and never appear in published docs.
-->

## Public Summary

<!-- **Public.** One paragraph in end-user voice. The canonical description for the site and README. -->

`concur` is Glacier's concurrency-primitives package: a curated set of ctx-aware sync types that complement Go's standard library `sync` and `golang.org/x/sync/errgroup`. Every primitive matches or beats its stdlib equivalent in production builds — `Mutex` and `RWMutex` are byte-for-byte identical to their stdlib counterparts at runtime; `Semaphore` is atomic-counter-backed with a lock-free fast path. When built with the `glacier_debug` build tag, `Mutex` and `RWMutex` gain hold-timeout diagnostics and caller-stack capture that surface stuck locks as structured slog events — with zero overhead in production. `Group` is an errgroup that collects every goroutine error (not first-wins), caps concurrency by default at `runtime.NumCPU() * 64`, recovers panics as typed `PanicError` values, and offers a non-blocking `TryGo` for back-pressure-aware scheduling. `Pool[T]` and `Once[T]` are generic, type-safe wrappers over `sync.Pool` and Go's once semantics. `WaitGroup` embeds `sync.WaitGroup` and adds a ctx-aware `WaitCtx`. Cancellation is uniform: every blocking operation accepts a `context.Context` and returns `ErrCancelled` on cancel.

## Mental Model

<!-- **Public.** The conceptual frame a developer should hold while using this. Mermaid diagrams welcome. Source for the "Concepts" page on the site. -->

There are two modes, and zero cost to choosing wisely:

**Production mode** (default build): `concur.Mutex` is `sync.Mutex` under the hood. There are no extra fields, no goroutines, no timers. `Lock` and `Unlock` are inlined forwarding calls. If you measure performance against a raw `sync.Mutex`, the delta is within noise.

**Debug mode** (`-tags glacier_debug`): the same `Mutex` type gains two extra fields — a caller-stack capture and a hold-timer. When a goroutine holds the lock longer than the configured timeout (default 5 s), a structured slog event fires: who acquired it, who is waiting, how long it has been held. The lock itself is unaffected; diagnostics are observers, not controllers.

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

`Group.Go` after `WaitDone` has returned **panics** — this matches `sync.WaitGroup.Add` after `Wait` and is intentional: reusing a finished Group is a programming error, not a runtime condition.

`Semaphore` uses a fast-path atomic CAS when permits are available. The slow path acquires a `sync.Cond` lock and parks. Each slow-path waiter spawns a cancel-watcher goroutine that holds a derived context; on successful acquire, `defer cancel()` cleans up that watcher immediately — no goroutine leak on the happy path or the cancellation path.

## Goals

<!-- **Internal.** Bulleted list. -->

- Provide ctx-aware wrappers for every common sync primitive so cancellation is uniform across the framework.
- Match stdlib performance in production builds (zero overhead for Mutex/RWMutex; allocation-free Semaphore fast path).
- Make hold-timeout diagnostics available without any production cost via the `glacier_debug` build tag.
- Collect all goroutine errors in Group (no first-wins); recover panics as typed errors.
- Cap Group concurrency by default (`runtime.NumCPU() * 64`) to prevent goroutine explosions; expose `WithUnlimited()` for explicit opt-out.
- Provide type-safe generic wrappers (`Pool[T]`, `Once[T]`) that eliminate `any` conversions at the call site.
- Keep the package import-minimal: `option`, `errs`, `log` (debug builds only); no tier-1 peers, no leaves.

## Non-Goals

<!-- **Internal.** Bulleted list. What this spec deliberately excludes. -->

- Typed channels, actor frameworks, futures/promises — not in v0; require a separate spec.
- `OnceErr` — stdlib `sync.OnceValues` covers the use case; adding a duplicate here violates less-code-is-more.
- Pool with TTL or eviction policy — deferred.
- Rate-limiter primitives — Semaphore is the building block; a higher-level rate-limiter is a future spec.
- `sync.Map` wrapper — callers who need a concurrent map should evaluate `sync.Map` directly; an ergonomic wrapper has not been validated by a real consumer use case.
- Upgrading a read lock to a write lock (`RWMutex` upgrade) — stdlib semantics prohibit this; `concur` follows.

## Architecture

<!-- **Internal.** Mermaid diagram + prose. Package layout, data flow, lifecycle. -->

### File layout

```
concur/
├── doc.go                  package-level doc comment
├── mutex.go                Mutex + RWMutex (non-debug builds; build tag: !glacier_debug)
├── mutex_debug.go          Mutex + RWMutex with hold-timeout diagnostics (build tag: glacier_debug)
├── panic.go                PanicError type + Error() + Unwrap()
├── group.go                Group + NewGroup + WithLimit + WithUnlimited + groupConfig
├── semaphore.go            Semaphore + NewSemaphore
├── pool.go                 Pool[T] + NewPool[T]
├── once.go                 Once[T]
├── waitgroup.go            WaitGroup + WaitCtx
├── errors.go               ErrCancelled + ErrInvalidPermits sentinels
│
├── mutex_test.go           Mutex/RWMutex correctness + LockCtx + race
├── mutex_debug_test.go     //go:build glacier_debug — hold-timeout slog event capture
├── mutex_bench_test.go     Lock/Unlock vs stdlib parity (NF1, §23.13)
├── group_test.go           Group correctness (Go, TryGo, WaitDone, panic recovery, limit)
├── group_panic_test.go     Go-after-WaitDone PANICS (§23.14)
├── group_bench_test.go     Per-Go alloc budget (NF3)
├── semaphore_test.go       Acquire/Release/TryAcquire + invalid permits
├── semaphore_watcher_test.go  cancel-watcher leak guard (§23.14)
├── semaphore_bench_test.go Fast-path ≤ 50 ns/op zero allocs (§23.13)
├── pool_test.go            Pool[T] round-trip + zero-on-empty
├── pool_bench_test.go      Allocation parity vs sync.Pool
├── once_test.go            Memoization, panic-doesn't-memoize, ctx first-call-only
├── waitgroup_test.go       WaitCtx correctness
├── race_test.go            //go:build race — combined race-detector matrix
├── lifecycle_doc_test.go   Documents Group has no Close (§23.16)
└── example_test.go         Godoc examples for every primitive
```

### Import graph

`concur` sits at Tier 1 (mid). It may import any Tier 0 kernel package; it may not import another Tier 1 package or any Tier 2 leaf.

```
concur
  ├─ github.com/nathanbrophy/glacier/errs   (error helpers, sentinels)
  ├─ github.com/nathanbrophy/glacier/option (functional options for Group config)
  └─ github.com/nathanbrophy/glacier/log    (debug builds only — slog events)
```

In non-debug builds, `log` is not imported. The `glacier_debug` build tag gates the `mutex_debug.go` file entirely.

### Lifecycle audit (§23.16)

| Type | Has Close? | Idempotent? | errs.Join on errors? | Notes |
|---|---|---|---|---|
| `Group` | No | n/a | n/a | `WaitDone` is the lifecycle boundary. After `WaitDone` returns, the Group is terminal — `Go` panics. Rationale: `Group` is a run-to-completion primitive, not a reusable resource. No close needed because it owns no external resources (sockets, files, etc.). |
| `Semaphore` | No | n/a | n/a | Stateless counter. Callers hold permits and release them; no destructor required. |
| `Pool[T]` | No | n/a | n/a | Wraps `sync.Pool`; GC-managed. |
| `Once[T]` | No | n/a | n/a | Run-to-completion once; no resources to close. |
| `WaitGroup` | No | n/a | n/a | Embeds stdlib `sync.WaitGroup`; stdlib has no Close. |
| `Mutex` / `RWMutex` | No | n/a | n/a | Debug-mode timer fires and auto-cancels; no explicit Close. |

All `concur` types omit `Close` because they own no external resources; in-memory state is managed by the Go runtime. Full rationale in DR6 (Decisions & Rationale).

## Schema

<!-- **Internal.** Go types with invariants stated as `// invariant: ...` comments on each field. -->

```go
// errors.go
var (
    // ErrCancelled is returned by blocking operations when the context is
    // cancelled. Wraps ctx.Err() so errors.Is(err, context.Canceled) holds.
    ErrCancelled = errs.Sentinel("concur: cancelled")

    // ErrInvalidPermits is returned by Semaphore operations when n is
    // non-positive or exceeds the semaphore's capacity.
    ErrInvalidPermits = errs.Sentinel("concur: invalid permits")
)

// panic.go
// PanicError carries a value recovered from a panicking goroutine inside Group.
// It is added to the Group's collected errors via errs.Join.
type PanicError struct {
    // Value is the argument passed to the panic() call.
    // invariant: Value is never nil (a nil panic would not recover as PanicError).
    Value any
}

// mutex.go (non-debug, build tag !glacier_debug)
// Mutex is a mutual-exclusion lock. In default builds it is byte-equivalent to
// sync.Mutex — the struct layout is a single embedded sync.Mutex with no
// additional fields. LockCtx adds ctx-aware acquisition via a try-lock-with-
// backoff loop; cancellation is best-effort (up to one backoff window, default
// 1 ms, may elapse after ctx cancels before the error is returned).
type Mutex struct {
    // invariant: mu is the only field in non-debug builds; sizeof(Mutex) == sizeof(sync.Mutex).
    mu sync.Mutex
}

// mutex_debug.go (build tag glacier_debug)
// Mutex augments sync.Mutex with hold-timeout diagnostics. Fields are only
// present in glacier_debug builds; production callers pay zero overhead.
type Mutex struct {
    mu          sync.Mutex
    // invariant: acquiredAt is zero when the lock is not held.
    acquiredAt  time.Time
    // invariant: callerStack is empty ("") when the lock is not held.
    callerStack string
    // invariant: holdTimeout > 0; zero value is replaced by defaultHoldTimeout (5s).
    holdTimeout time.Duration
    timer       *time.Timer
}

// group.go
// groupConfig holds the resolved configuration for a Group.
type groupConfig struct {
    // invariant: limit >= 1 OR limit == unlimited (-1 sentinel).
    // Default: runtime.NumCPU() * 64.
    limit int
}

// Group is an errgroup-equivalent. After WaitDone returns, the Group is
// terminal: any call to Go panics.
type Group struct {
    // invariant: sem is non-nil if and only if cfg.limit != unlimited.
    sem    *Semaphore
    // invariant: wg.counter == 0 after WaitDone returns.
    wg     sync.WaitGroup
    mu     sync.Mutex // guards errs
    // invariant: collected errors are appended under mu; never read until WaitDone.
    errs   []error
    // invariant: done is closed exactly once, by the first WaitDone call.
    done   chan struct{}
    // invariant: closed == true after WaitDone returns; never reset.
    closed bool
}

// semaphore.go
// Semaphore is a counted semaphore with an atomic-counter fast path.
type Semaphore struct {
    // invariant: capacity > 0.
    capacity int64
    // invariant: 0 <= count <= capacity at all observable points outside a critical section.
    count    atomic.Int64
    mu       sync.Mutex  // guards cond; only taken on slow path
    cond     *sync.Cond
}

// pool.go
// Pool[T] is a typed wrapper over sync.Pool. Get returns T; Put accepts T.
// No any-conversion at the call site.
type Pool[T any] struct {
    p sync.Pool
}

// once.go
// Once[T] memoizes the (T, error) result of the first successful call to Do.
// A panicking Do does not count as "done"; the next call re-attempts.
type Once[T any] struct {
    mu   sync.Mutex
    done bool
    val  T
    err  error
}

// waitgroup.go
// WaitGroup embeds sync.WaitGroup to inherit Add, Done, and Wait, and
// adds WaitCtx for ctx-aware waiting.
type WaitGroup struct {
    sync.WaitGroup
}
```

## API

<!--
  **Public.** Every exported symbol introduced by this spec.
  For each: signature, doc comment (which becomes godoc), preconditions, postconditions,
  error contract, concurrency notes (goroutine-safe? blocking?), lifecycle hooks.
  Magpie extracts signatures + doc comments verbatim to the API reference page.
-->

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
//
// Callers who want to inspect the recovered value use errors.As:
//
//   var pe *concur.PanicError
//   if errors.As(err, &pe) {
//       log.Printf("goroutine panicked: %v", pe.Value)
//   }
//
// Security note: pe.Value may contain sensitive data if the panicking code
// had access to secrets. Callers should redact before logging in production
// (see log.RedactValue).
type PanicError struct {
    // Value is the argument supplied to the panic() call.
    Value any
}

// Error implements the error interface.
// Format: "concur: panic in group goroutine: <value>".
func (e *PanicError) Error() string

// Unwrap returns a synthesized error whose message is e.Error(), enabling
// errors.As traversal to reach PanicError from a wrapped errs.Join chain.
func (e *PanicError) Unwrap() error
```

**Concurrency**: PanicError values are immutable after creation; safe to share across goroutines.

### Mutex

```go
// Mutex is a mutual-exclusion lock. It implements sync.Locker.
//
// In default builds, Mutex is byte-equivalent to sync.Mutex: a single
// embedded field with no additional overhead. Lock and Unlock are inlining-
// eligible forwarding calls.
//
// In glacier_debug builds (-tags glacier_debug), Mutex additionally captures
// the acquiring goroutine's caller stack and starts a hold timer. If the lock
// is held longer than the configured timeout (default 5 s), a structured slog
// event is emitted at Warn level containing: holder caller, any waiting
// caller, and elapsed hold duration. The lock is NOT released by this
// diagnostic; the holder continues normally.
//
// The zero value of Mutex is an unlocked mutex. Mutex must not be copied
// after first use (same constraint as sync.Mutex).
type Mutex struct{ /* fields vary by build tag; see Schema */ }

// Lock acquires the mutex, blocking until it is available.
// Semantics are identical to sync.Mutex.Lock.
func (m *Mutex) Lock()

// Unlock releases the mutex.
// Panics if the mutex is not locked (stdlib parity).
func (m *Mutex) Unlock()

// LockCtx acquires the mutex, blocking until the lock is available or ctx is
// cancelled. Returns ErrCancelled (wrapping ctx.Err()) if ctx fires before
// the lock is acquired.
//
// Precondition: ctx must be non-nil.
// Cancellation is best-effort: a try-lock-with-backoff loop is used internally
// (default backoff: 1 ms). If ctx cancels during the backoff window, up to one
// backoff interval may elapse before LockCtx returns. Callers must not rely on
// immediate return on cancellation.
//
// Concurrency: goroutine-safe.
// Blocking: yes, until lock acquired or ctx cancelled.
func (m *Mutex) LockCtx(ctx context.Context) error
```

### RWMutex

```go
// RWMutex is a reader/writer mutual-exclusion lock. It implements sync.Locker
// (via Lock/Unlock). In default builds it is byte-equivalent to sync.RWMutex.
// In glacier_debug builds it gains the same hold-timeout diagnostics as Mutex.
//
// Upgrade from RLock to Lock is not supported (same restriction as sync.RWMutex).
// The zero value is an unlocked mutex. Must not be copied after first use.
type RWMutex struct{ /* fields vary by build tag */ }

// RLock acquires a shared read lock.
func (m *RWMutex) RLock()

// RUnlock releases a shared read lock.
func (m *RWMutex) RUnlock()

// Lock acquires the exclusive write lock.
func (m *RWMutex) Lock()

// Unlock releases the exclusive write lock.
func (m *RWMutex) Unlock()

// RLockCtx acquires a shared read lock, returning ErrCancelled if ctx
// fires first. Best-effort cancellation; same backoff caveat as Mutex.LockCtx.
func (m *RWMutex) RLockCtx(ctx context.Context) error

// LockCtx acquires the exclusive write lock, returning ErrCancelled if ctx
// fires first. Best-effort cancellation; same backoff caveat as Mutex.LockCtx.
func (m *RWMutex) LockCtx(ctx context.Context) error
```

### Group

```go
// NewGroup constructs a Group with the given options.
//
// Default concurrency cap: runtime.NumCPU() * 64. Pass WithUnlimited() to
// remove the cap entirely. Pass WithLimit(n) to set an explicit cap.
//
// Precondition: opts must be valid (WithLimit(0) returns an option error).
// Returns (nil, error) if any option is invalid.
//
// Concurrency: the returned *Group is goroutine-safe.
func NewGroup(opts ...option.Option[groupConfig]) (*Group, error)

// WithLimit caps the maximum number of goroutines that Group.Go will run
// concurrently. Go blocks when at the cap; TryGo returns false.
//
// n must be >= 1. n == 0 returns an option error.
// n < 0 is treated as WithUnlimited() (alias for the explicit-unlimited case).
func WithLimit(n int) option.Option[groupConfig]

// WithUnlimited removes the default concurrency cap, allowing an unbounded
// number of goroutines. Use when the caller manages back-pressure externally.
func WithUnlimited() option.Option[groupConfig]

// Go schedules fn to run in a new goroutine. If a concurrency cap is set,
// Go blocks until a slot is available (a goroutine finishes and releases
// its permit).
//
// Panics in fn are recovered and converted to *PanicError, which is included
// in the errs.Join result of WaitDone. Callers who want raw panic propagation
// must wrap fn themselves with a recover call.
//
// PANICS if called after WaitDone has returned. This matches
// sync.WaitGroup.Add-after-Wait and is intentional: a finished Group is
// terminal. Panic message: "concur: Go after WaitDone".
//
// Concurrency: goroutine-safe.
// Blocking: may block if at concurrency cap.
func (g *Group) Go(fn func() error)

// TryGo schedules fn in a new goroutine if a slot is available.
// Returns true if fn was scheduled, false if the group is at its concurrency
// cap or WaitDone has already been called.
//
// Unlike Go, TryGo never blocks. It is the caller's responsibility to handle
// the false case (e.g., backoff, buffer, or drop).
//
// Concurrency: goroutine-safe.
// Blocking: never.
func (g *Group) TryGo(fn func() error) bool

// WaitDone blocks until all scheduled goroutines have completed or ctx is
// cancelled, whichever comes first.
//
// Return value semantics:
//   - All goroutines finish before ctx fires: returns errs.Join over every
//     collected error (nil if none). Does NOT return the first error only —
//     all errors are preserved.
//   - ctx fires before all goroutines finish: returns ErrCancelled wrapping
//     ctx.Err(). Goroutines that were already running continue to completion
//     in the background (WaitDone does not cancel in-flight goroutines).
//
// After WaitDone returns, the Group is terminal. Any subsequent Go call
// panics.
//
// Concurrency: goroutine-safe; may be called from one goroutine at a time.
// Blocking: yes, until all goroutines finish or ctx fires.
func (g *Group) WaitDone(ctx context.Context) error
```

### Semaphore

```go
// NewSemaphore constructs a counted semaphore with the given capacity.
//
// Precondition: capacity >= 1. Panics if capacity < 1 (programming error).
// Concurrency: the returned *Semaphore is goroutine-safe.
func NewSemaphore(capacity int64) *Semaphore

// Acquire blocks until n permits are available or ctx is cancelled.
//
// Error contract:
//   - Returns nil on success (n permits have been acquired).
//   - Returns ErrCancelled (wrapping ctx.Err()) if ctx fires before permits
//     are available.
//   - Returns ErrInvalidPermits if n <= 0 or n > capacity.
//
// Implementation: fast path is a single atomic CAS (allocation-free, lock-free).
// Slow path parks on sync.Cond. Each slow-path waiter spawns one cancel-watcher
// goroutine that holds a derived context; on successful acquire, defer cancel()
// cleans up the watcher — no goroutine leak on the happy path or cancel path.
//
// Concurrency: goroutine-safe.
// Blocking: may block until permits are available.
func (s *Semaphore) Acquire(ctx context.Context, n int64) error

// TryAcquire attempts to acquire n permits without blocking.
// Returns true and decrements the counter if n permits are available.
// Returns false without modifying state if not enough permits are available.
//
// Error contract: returns ErrInvalidPermits if n <= 0 or n > capacity.
// (The bool+error form is used so callers can distinguish "not enough permits"
// from "invalid call".)
//
// Concurrency: goroutine-safe.
// Blocking: never.
func (s *Semaphore) TryAcquire(n int64) bool

// Release returns n permits to the semaphore and wakes any blocked Acquire
// calls.
//
// Release(0) is a permitted no-op.
// PANICS if releasing would cause the total available permits to exceed
// capacity (over-release is a programming error). Panic message:
// "concur: release: over-release".
//
// Concurrency: goroutine-safe.
// Blocking: never (brief lock acquisition to broadcast to waiters).
func (s *Semaphore) Release(n int64)
```

### Pool[T]

```go
// Pool[T] is a typed wrapper over sync.Pool. Get returns T directly,
// eliminating any any-conversion at the call site.
//
// The zero value is not useful; use NewPool.
type Pool[T any] struct{ /* ... */ }

// NewPool constructs a Pool. newFn is called when the pool is empty and
// Get is invoked with no pooled values available (matching sync.Pool.New
// semantics). If newFn is nil, Get returns the zero value of T when the
// pool is empty.
func NewPool[T any](newFn func() T) *Pool[T]

// Get retrieves a value from the pool. If the pool is empty and newFn is
// non-nil, newFn is called and its result returned. If the pool is empty
// and newFn is nil, the zero value of T is returned.
//
// Concurrency: goroutine-safe (sync.Pool semantics).
func (p *Pool[T]) Get() T

// Put returns a value to the pool for future reuse.
//
// Concurrency: goroutine-safe (sync.Pool semantics).
func (p *Pool[T]) Put(v T)
```

### Once[T]

```go
// Once[T] runs a function exactly once and memoizes its (T, error) result.
// Subsequent calls to Do return the memoized values without invoking fn.
//
// ctx is passed to fn on the first call only. Later callers' contexts are
// ignored — the memoized result is always returned regardless of whether
// those callers' ctxs are cancelled.
//
// If fn panics, the panic propagates and Once is not marked "done". The
// next call to Do re-attempts. This matches sync.Once semantics.
//
// If fn returns normally (even if it returns an error), the result is
// memoized. A (zero, error) result is memoized the same as a (value, nil)
// result.
//
// The zero value is a valid, unused Once.
type Once[T any] struct{ /* ... */ }

// Do calls fn the first time it is invoked and memoizes the result.
// All subsequent calls return the memoized (T, error) without calling fn.
//
// Concurrency: goroutine-safe. Concurrent first calls block until fn returns.
// Blocking: the first call blocks until fn returns; subsequent calls are a
//   single atomic load.
func (o *Once[T]) Do(ctx context.Context, fn func(context.Context) (T, error)) (T, error)
```

### WaitGroup

```go
// WaitGroup embeds sync.WaitGroup (inheriting Add, Done, and Wait) and adds
// WaitCtx for context-aware waiting.
//
// The zero value is a valid, unused WaitGroup.
type WaitGroup struct {
    sync.WaitGroup
}

// WaitCtx blocks until the WaitGroup counter reaches zero or ctx is cancelled,
// whichever comes first.
//
// Returns nil if the counter reaches zero before ctx fires.
// Returns ErrCancelled wrapping ctx.Err() if ctx fires first (the counter
// may still be > 0 in this case — the goroutines continue running).
//
// WaitCtx does NOT cancel any in-flight goroutines.
//
// Add-during-WaitCtx has the same undefined-behavior / data-race semantics as
// sync.WaitGroup.Add-during-Wait. Documented as caller responsibility.
//
// Concurrency: goroutine-safe.
// Blocking: yes, until counter reaches zero or ctx fires.
func (wg *WaitGroup) WaitCtx(ctx context.Context) error
```

## Examples

<!--
  **Public.** Runnable Go examples in fenced ```go blocks.
  Each example is self-contained and `go test ./...`-compatible (valid Example functions).
  Magpie transcludes verbatim into tutorials.
-->

### Mutex — production lock

```go
func ExampleMutex() {
    var mu concur.Mutex

    mu.Lock()
    defer mu.Unlock()
    // critical section: safe to read/write shared state here
}
```

### Mutex — ctx-aware acquisition

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

### Group — bounded concurrency pipeline

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

### Semaphore — rate-limited HTTP

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

### Pool[T] — buffer reuse

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

### Once[T] — one-time async initialization

```go
var dbOnce concur.Once[*sql.DB]

func DB(ctx context.Context) (*sql.DB, error) {
    return dbOnce.Do(ctx, func(ctx context.Context) (*sql.DB, error) {
        return sql.Open("postgres", os.Getenv("DATABASE_URL"))
    })
}
```

### WaitGroup — ctx-aware shutdown

```go
func ExampleWaitGroup_WaitCtx() {
    var wg concur.WaitGroup

    for _, w := range workers {
        wg.Add(1)
        go func(w *Worker) {
            defer wg.Done()
            w.Run(ctx)
        }(w)
    }

    if err := wg.WaitCtx(ctx); err != nil {
        return errs.Wrap(err, "shutdown: workers did not finish before deadline")
    }
}
```

## Test Matrix

<!--
  **Internal.** Owned by Lynx.
  Table: scenario × input × expected outcome × covered-by-test-name.
-->

| # | Test name | Spec ref | Kind | Scenario | Assertions |
|---|---|---|---|---|---|
| 1 | `TestMutexLockUnlockBasic` | §21.5 F2 | Unit (positive) | Single goroutine Lock→Unlock returns no error. | `assert.NoError`, `assert.True` |
| 2 | `TestMutexLockCtxAlreadyCancelled` | §21.5 E1 | Unit (negative) | `LockCtx` with already-cancelled ctx returns `ErrCancelled` immediately, never acquires. | `assert.ErrorIs`, `fixture.NewClock` |
| 3 | `TestMutexLockCtxCancelledMidWait` | §21.5 E2 | Unit (negative) | Mutex held by g1; g2 calls `LockCtx`; ctx cancelled; g2 returns `ErrCancelled` wrapping `ctx.Err()`. | `assert.ErrorIs`, `concur.WaitGroup` |
| 4 | `TestMutexLockCtxTryLockBackoffWindow` | §23.14 | Unit (concurrency lock-in) | `LockCtx` with quickly-cancelled ctx during backoff window verifies best-effort-cancellation — must observe `ctx.Err()` within 2× configured backoff. | `fixture.NewClock`, `assert.Eventually` |
| 5 | `TestMutexLockCtxNoLeakAfterCancel` | §23.14 | Race + leak | Repeated cancelled `LockCtx` leaves no orphan goroutines. | `fixture.GuardLeaks`, `fixture.WatchGoroutines` |
| 6 | `TestMutexUnlockUnheldPanics` | stdlib parity | Unit (negative) | `Unlock` without prior `Lock` panics with stdlib message. | `assert.Panics` |
| 7 | `TestMutexDebugHoldTimeoutEmitsSlog` | §21.5 F1, E3 | Unit (`glacier_debug`) | Hold > timeout emits structured slog event with holder caller, waiter caller, elapsed. Lock not released by diagnostic. | `fixture.Capture`, custom slog handler sink, `assert.Contains` |
| 8 | `TestMutexDebugProductionByteEquivalent` | §21.5 NF1, §23.13 | Bench/Test | `unsafe.Sizeof(concur.Mutex{}) == unsafe.Sizeof(sync.Mutex{})` in non-debug builds. | `assert.Equal` |
| 9 | `TestRWMutexParallelReadersExclusiveWriter` | §21.5 F3 | Unit (positive) | N readers proceed concurrently; `Lock` waits until all `RUnlock`. | `concur.WaitGroup`, `assert.Equal` |
| 10 | `TestRWMutexRLockCtxCancelled` | §21.5 F3 | Unit (negative) | `RLockCtx` returns `ErrCancelled` when lock is held exclusively. | `assert.ErrorIs` |
| 11 | `TestRWMutexLockCtxCancelled` | §21.5 F3 | Unit (negative) | `LockCtx` returns `ErrCancelled` when readers hold. | `assert.ErrorIs` |
| 12 | `TestRWMutexUpgradeDeadlockDocumented` | §21.5 (no upgrade) | Unit (edge) | Documenting test: `Lock` while `RLock`-held by same goroutine deadlocks (stdlib semantics). Uses ctx-bounded wait. | `assert.ErrorIs` |
| 13 | `TestGroupAllPass` | §21.5 F4 | Unit (positive) | All goroutines succeed; `WaitDone` returns nil. | `assert.NoError` |
| 14 | `TestGroupOneError` | §21.5 F8 | Unit (negative) | One goroutine errors; `WaitDone` returns `errs.Join` with that single error. | `assert.ErrorIs`, `errs.Chain` |
| 15 | `TestGroupAllErrorsCollected` | §21.5 F8 | Unit (negative) | N goroutines all error; `errs.Join` collects every error (not first-wins). | `fluent.Count`, `assert.Equal` |
| 16 | `TestGroupCtxCancelDuringWait` | §21.5 F8 | Unit (negative) | `WaitDone` returns `ErrCancelled` wrapping `ctx.Err()` if ctx fires before goroutines finish. | `assert.ErrorIs`, `fixture.NewClock` |
| 17 | `TestGroupPanicRecoveredAsPanicError` | §21.5 F6, E7, F18 | Unit (negative) | Goroutine panics with non-error value; recovered as `*PanicError{Value:...}`; appended via `errs.Join`. | `assert.ErrorAs`, `errs.Chain` |
| 18 | `TestGroupPanicErrorMessage` | §21.5 F18 | Unit (positive) | `(&PanicError{Value:"boom"}).Error()` formats exactly `"concur: panic in group goroutine: <boom>"`. | `assert.Equal` |
| 19 | `TestGroupWithLimitBlocks` | §21.5 F5, E5 | Unit (positive) | `WithLimit(2)`; 3rd `Go` blocks until first finishes. Verified via timing. | `fixture.NewClock`, `assert.Eventually` |
| 20 | `TestGroupWithLimitZeroReturnsOptionError` | §21.5 E5 | Unit (negative) | `NewGroup(WithLimit(0))` returns option error. | `assert.ErrorIs` |
| 21 | `TestGroupWithLimitNegativeIsUnlimited` | §21.5 E6 | Unit (positive) | `WithLimit(-1)` acts as `WithUnlimited()`. | `assert.NoError` |
| 22 | `TestGroupWithUnlimitedExplicit` | §23.14 | Unit (positive) | Explicit `WithUnlimited()` overrides default cap. | `assert.NoError` |
| 23 | `TestGroupDefaultLimitIsNumCPU64` | §23.14 | Unit (positive) | Internal default of `runtime.NumCPU() * 64` verified by introspection or back-pressure observation. | `assert.Equal` |
| 24 | `TestGroupTryGoAtCapReturnsFalse` | §21.5 F7 | Unit (negative) | `TryGo` with limit at capacity returns false; fn not scheduled. | `assert.False` |
| 25 | `TestGroupTryGoAfterWaitDoneReturnsFalse` | §21.5 F7 | Unit (negative) | `TryGo` after `WaitDone` returns false. | `assert.False` |
| 26 | `TestGroupGoAfterWaitDonePanics` | §23.14 | Unit (concurrency lock-in) | `Group.Go` after `WaitDone` PANICS with message `"concur: Go after WaitDone"`. | `assert.PanicsWithMessage` |
| 27 | `TestGroupGoAfterWaitDonePanicMessage` | §23.14 | Unit (positive) | Asserts exact panic string for the Go-after-Wait case. | `assert.Equal` |
| 28 | `TestGroupConcurrentGoFromManyGoroutines` | §21.5 NF5 | Race | 1000 goroutines call `Go`; all errors collected; race-clean. | `concur.WaitGroup`, `fixture.GuardLeaks`, `-race` |
| 29 | `PropertyGroupErrorCountEqualsGoroutineErrors` | §21.5 F8 | Property | For N in [0, 200] with K errors and N-K nils, `len(errs.Chain(WaitDone)) == K`. | property generator over `fluent.Range`, `assert.Equal` |
| 30 | `TestSemaphoreAcquireFastPath` | §21.5 F9 | Unit (positive) | `Acquire` when permits available returns nil without blocking. | `assert.NoError` |
| 31 | `TestSemaphoreAcquireSlowPathBlocks` | §21.5 F9 | Unit (positive) | `Acquire` when no permits blocks until `Release`. | `concur.WaitGroup`, `assert.Eventually` |
| 32 | `TestSemaphoreAcquireCancelled` | §21.5 F10 | Unit (negative) | `Acquire` blocks; ctx cancels; returns `ErrCancelled` wrapping `ctx.Err()`. | `assert.ErrorIs` |
| 33 | `TestSemaphoreInvalidPermitsZero` | §21.5 E8 | Unit (negative) | `Acquire(ctx, 0)` returns `ErrInvalidPermits`. | `assert.ErrorIs` |
| 34 | `TestSemaphoreInvalidPermitsNegative` | §21.5 F10 | Unit (negative) | `Acquire(ctx, -1)` returns `ErrInvalidPermits`. | `assert.ErrorIs` |
| 35 | `TestSemaphoreInvalidPermitsExceedsCapacity` | §21.5 E9 | Unit (negative) | `Acquire(ctx, capacity+1)` returns `ErrInvalidPermits`. | `assert.ErrorIs` |
| 36 | `TestSemaphoreTryAcquireSuccess` | §21.5 F11 | Unit (positive) | `TryAcquire(n)` with available permits returns true. | `assert.True` |
| 37 | `TestSemaphoreTryAcquireFailure` | §21.5 F11 | Unit (negative) | `TryAcquire` when not enough permits returns false; permits unchanged. | `assert.False` |
| 38 | `TestSemaphoreReleaseZeroNoOp` | §21.5 E10 | Unit (positive) | `Release(0)` is a permitted no-op; counter unchanged. | `assert.Equal` |
| 39 | `TestSemaphoreOverReleasePanics` | §21.5 E11 | Unit (negative) | `Release(n)` where total released > acquired panics with `"concur: release: over-release"`. | `assert.PanicsWithMessage` |
| 40 | `TestSemaphoreCtxWatcherNoLeak` | §23.14 | Race + leak | 10k Acquire→success cycles; cancel-watcher cleanup via `defer cancel()` on per-acquire derived ctx; zero leaked goroutines. | `fixture.GuardLeaks(WatchGoroutines, WithDrainTimeout(500ms))` |
| 41 | `TestSemaphoreCtxWatcherNoLeakOnCancel` | §23.14 | Race + leak | 10k cancelled-Acquire cycles; cancel-watcher exits; no leaked goroutines. | `fixture.GuardLeaks` |
| 42 | `TestSemaphoreManyGoroutinesAcquireRelease` | §21.5 NF5 | Race | 1000 goroutines acquire/release; final counter == 0. | `-race`, `concur.WaitGroup` |
| 43 | `FuzzSemaphoreAcquireRelease` | §21.5 F9–F12 | Fuzz | Fuzz-driven sequence of Acquire/Release/TryAcquire with capacity in [1,32]; invariant: counter never exceeds capacity, never negative; every panic is documented over-release. | `assert.True`, `testing.F` |
| 44 | `TestPoolGetPutRoundTripPreservesType` | §21.5 F13, E12 | Unit (positive) | `NewPool[*Buf](newFn).Get()` → mutate → `Put` → `Get` returns same kind. | `assert.IsType` |
| 45 | `TestPoolGetEmptyNoNewReturnsZero` | §21.5 E12 | Unit (negative) | `NewPool[int](nil).Get()` returns 0. | `assert.Equal[int]` |
| 46 | `TestPoolConcurrentGetPut` | §21.5 NF5 | Race | 1000 goroutines `Get`/`Put`; race-clean. | `-race` |
| 47 | `TestOnceMemoizesValueAndError` | §21.5 F14 | Unit (positive) | First call's `(value, error)` is memoized; subsequent calls return same pair. | `assert.Equal`, `assert.ErrorIs` |
| 48 | `TestOnceFirstCallCtxOnly` | §21.5 F14 | Unit (positive) | ctx of first call is threaded to fn; later callers' ctxs are ignored. | `assert.Equal` |
| 49 | `TestOncePanicDoesNotMemoize` | §21.5 E13 | Unit (negative) | First call panics; panic propagates; Once not completed; second `Do` re-attempts and succeeds. | `assert.Panics`, `assert.NoError` |
| 50 | `TestOnceConcurrentFirstCallWins` | §21.5 NF5 | Race | 100 goroutines call `Do` simultaneously; fn invoked exactly once; all observe same memoized result. | `concur.WaitGroup`, `-race`, atomic counter |
| 51 | `TestWaitGroupWaitCtxAlreadyZero` | §21.5 E14 | Unit (positive) | `WaitCtx` returns nil immediately if counter == 0. | `assert.NoError` |
| 52 | `TestWaitGroupWaitCtxCancelled` | §21.5 F15 | Unit (negative) | Counter > 0; ctx cancels; returns `ErrCancelled` wrapping `ctx.Err()`. | `assert.ErrorIs` |
| 53 | `TestWaitGroupWaitCtxNormalCompletion` | §21.5 F15 | Unit (positive) | Counter reaches zero before ctx cancel; returns nil. | `assert.NoError` |
| 54 | `TestWaitGroupRaceAddDuringWait` | §21.5 E15 | Race (documented) | Documents stdlib semantics — racy `Add` during `WaitCtx`. | `-race` (expected clean since stdlib `sync.WaitGroup` governs) |
| 55 | `TestErrCancelledIsSentinelStable` | §21.5 F16 | Unit (positive) | `errors.Is(wrapped, concur.ErrCancelled)` is true; pointer equality is false (wrapping). | `assert.ErrorIs` |
| 56 | `TestErrInvalidPermitsSentinelStable` | §21.5 F17 | Unit (positive) | `errors.Is(err, concur.ErrInvalidPermits)` after `Acquire(ctx, 0)`. | `assert.ErrorIs` |
| 57 | `TestPanicErrorUnwrapToSynthesized` | §21.5 F18 | Unit (positive) | `PanicError.Unwrap()` returns a synthesized error reflecting `Value`. | `assert.NotNil`, `assert.ErrorAs` |
| 58 | `TestErrFormatRegisterCompliance` | §21.5 NF6 | Unit (cross-cutting) | Every error string matches `^concur: [a-z]+(?:: [a-z ]+)*$` — lowercase, no period, package prefix. | `assert.Regexp` |
| 59 | `TestNoOnceErr` | D23 charter | Unit (architecture) | Verify `concur.OnceErr` does NOT exist (`sync.OnceValues` is the answer). | reflection / compile-time guard |
| 60 | `TestGroupHasNoCloseDocumented` | §23.16 | Unit (lifecycle) | `*Group` does NOT have a `Close()` method. Compile-time and reflection check. | reflection check |
| 61 | `TestConfPointerSnapshotOrdering` | §23.14 (cross-link conf) | Property | Atomic snapshot accessor in `conf` never returns torn state under contention. Primitive-level invariant lives here. | `concur.Group`, `assert.Equal` |
| EX1 | `TestGroupReentrantGoSingleRecover` | edge | Unit (edge) | `Group.Go` called from inside `Go`'s own fn (re-entrant) — panic-recovery deferred frame does not double-fire. | `assert.Equal` (panic count) |
| EX2 | `TestSemaphoreReleaseOverflowPanics` | edge | Unit (negative) | `Semaphore.Release(MaxInt64)` when only 1 acquired — verify panic, not silent overflow. | `assert.PanicsWithMessage` |
| EX3 | `TestOnceZeroValueMemoized` | edge | Unit (positive) | `Once[T].Do` where fn returns zero value AND non-nil error — second call returns same `(zero, err)`. | `assert.Equal`, `assert.ErrorIs` |
| EX4 | `TestPoolInterfaceNilNotTypedNil` | edge | Unit (edge) | `Pool[T]` with T = interface — Get on empty pool returns untyped nil, not typed nil. | `assert.Nil` |
| EX5 | `TestRWMutexUnlockMismatchPanics` | edge | Unit (negative) | `RWMutex.LockCtx` then `Unlock` (not `RUnlock`) — mismatch panics per stdlib semantics. | `assert.Panics` |
| EX6 | `TestGroupLimitOneSerializes` | edge | Unit (positive) | `Group` with `WithLimit(1)` serializes goroutines — liveness holds. | `assert.Eventually` |
| EX7 | `TestGroupsIndependentLimits` | edge | Unit (positive) | Two `Group`s sharing nothing — their Semaphore-backed limits are independent. | `assert.NoError` |
| B1 | `BenchmarkMutexLockUnlock` | §21.5 NF1, §23.13 | Benchmark | Bytes/op == 0; ns/op within 5% of `sync.Mutex` baseline. | `testing.B`, `benchstat` |
| B2 | `BenchmarkRWMutexRLockRUnlock` | §21.5 NF1 | Benchmark | Within 5% of `sync.RWMutex`. | `testing.B` |
| B3 | `BenchmarkSemaphoreAcquireReleaseUncontended` | §21.5 NF2, §23.13 | Benchmark | ≤ 50 ns/op, zero allocs (via `testing.AllocsPerRun`). | `testing.AllocsPerRun` |
| B4 | `BenchmarkSemaphoreAcquireReleaseContended` | §21.5 NF2 | Benchmark | Documents slow-path cost; baseline for regressions. | `testing.B` |
| B5 | `BenchmarkGroupGoWithLimit` | §21.5 NF3 | Benchmark | Per-Go alloc count within stated bound (1 closure + 1 recover frame). | `testing.AllocsPerRun` |
| B6 | `BenchmarkGroupTryGo` | §21.5 F7 | Benchmark | Tracks scheduling latency. | `testing.B` |
| B7 | `BenchmarkPoolGetPut` | §21.5 NF4 | Benchmark | Allocation-equivalent to `sync.Pool`. | `testing.AllocsPerRun` |
| B8 | `BenchmarkOnceDoFastPath` | §21.5 F14 | Benchmark | Post-memoization `Do` is single atomic load. | `testing.B` |
| B9 | `BenchmarkWaitGroupWaitCtxFastPath` | §21.5 F15 | Benchmark | Counter == 0 path is constant-time. | `testing.B` |

### Testing notes

- **Debug-tag tests** (`mutex_debug_test.go`) run in a separate CI job with `-tags glacier_debug`. The hold-timeout test must not flake — use a `fixture.NewClock`-driven internal clock injection seam (exposed test-only via `concur_test` package) OR use generous wall-clock timeouts (50 ms hold vs 10 ms threshold).
- **NF1 byte-equivalence** is tested via both `unsafe.Sizeof` AND a `benchstat`-driven alert: `BenchmarkMutexLockUnlock` is compared head-to-head with a sibling `BenchmarkStdlibMutex`; CI fails if delta > 5%.
- **§23.16 Group lifecycle**: test #60 is a compile-time and reflection-time gate that `*Group` does not expose `Close()`. This is documentation-as-test.
- **Goroutine-leak tests** (#40, #41, #5) use `fixture.GuardLeaks` with `fixture.WatchGoroutines`. Drain timeout of 500 ms is sufficient to let cancel-watchers clean up on acquire.
- **Race detector**: all tests in `race_test.go` run with `-race`; CI gate requires zero races on every PR.

## Dependency Justification

<!--
  **Internal.** Owned by Falcon.
  One row per new direct dependency. The empty table is the goal.
-->

| Module | Version | License | Last release | Maintainers | Alternatives considered | Why we can't roll our own |
|---|---|---|---|---|---|---|

No new direct dependencies. `concur` imports only stdlib and two Glacier kernel packages (`option`, `errs`) in production builds, plus `log` in debug builds. All three are in-tree.

## Security & Supply-Chain Notes

<!-- **Internal.** Untrusted-input handling, sandboxing implications, secrets handling, vuln-scan considerations. -->

- **PanicError.Value**: the `Value any` field carries whatever value the panicking goroutine passed to `panic()`. In multi-tenant or security-sensitive services this value may contain secrets, PII, or internal implementation details. The `Error()` method formats `Value` via `%v` — callers must redact before logging. Suggested pattern: `log.RedactValue(pe.Value)` before writing to any slog handler in production.
- **Race detector mandatory for Group/Semaphore tests**: all concurrent-path tests run under `-race` in CI. Zero race reports is a hard gate.
- **No unsafe pointer arithmetic**: the package uses `sync/atomic` via `atomic.Int64` (safe wrapper); no `unsafe.Pointer` manipulation.
- **No external inputs**: `concur` primitives accept only Go values from the caller's own code. There is no deserialization surface, no file access, and no network I/O. Untrusted-input row is not needed in the register.
- **Semaphore over-release panic**: a programming error surfaces as a panic with a fixed message (`"concur: release: over-release"`). The message contains no user-supplied data.
- **`glacier_debug` build tag**: debug builds emit slog events containing caller stack frames. Caller stacks may include function names that reveal internal module structure. Teams that enable `glacier_debug` in production (unusual; this is a developer-debugging feature) should consider whether the slog output goes to a secure handler.

## FAQ

<!-- **Public.** Anticipated user questions with answers. Magpie extracts to the public docs FAQ. -->

**Q: Why does Group default to `runtime.NumCPU() * 64` goroutines instead of being truly unlimited?**

An unbounded Group is a goroutine-explosion foot-gun: a tight loop calling `Go` can queue millions of goroutines before the work finishes, exhausting memory. `NumCPU * 64` is generous enough for almost all real workloads (3200+ goroutines on a 50-core machine) while preventing accidental abuse. When you genuinely need unlimited goroutines — because you manage back-pressure externally — pass `WithUnlimited()` and document the rationale at the call site.

**Q: Why does `Group.Go` panic after `WaitDone` instead of returning an error?**

Same reason `sync.WaitGroup.Add` panics after `Wait`: reusing a finished Group after `WaitDone` is a programming error, not a runtime condition. Errors are for expected failure modes; panics are for invariant violations. A panicking call site is unambiguously wrong; an error-returning `Go` would be silently swallowed by code that forgets to check. If you need a re-usable Group, create a new one.

**Q: `Mutex.LockCtx` says "best-effort cancellation" — what does that mean?**

`LockCtx` uses a try-lock-with-backoff loop internally. On each iteration it attempts a non-blocking lock acquisition; if that fails and the ctx is still live, it sleeps for a backoff interval (default 1 ms) before retrying. If ctx cancels during a sleep, LockCtx will not detect it until the sleep ends — up to one backoff interval of latency. This is documented behavior. If you need stricter ctx-response, structure your code so the critical section timeout is much larger than the backoff.

**Q: Does `Once[T]` memoize an error result?**

Yes. `Once[T].Do` memoizes the `(T, error)` pair returned by fn, including when the error is non-nil and T is the zero value. The second caller receives the same `(zero, err)` as the first — the fn is not retried on error. Only a fn that panics causes a re-attempt. If you need retry-on-error semantics, do not use `Once[T]`; manage the state yourself with a mutex.

**Q: Why is there no `Close` on `Semaphore`, `Pool`, or `Once`?**

`Close` implies ownership of an external resource (socket, file, goroutine leak) that must be explicitly freed. None of these types own such resources: `Semaphore` is an in-memory counter, `Pool` is GC-managed by `sync.Pool`, and `Once` is a one-shot state machine. Adding `Close` would create a false expectation that not calling it leaks something. The only type with a natural lifecycle terminal is `Group`, whose terminal state is "all goroutines finished" — signalled by `WaitDone` returning, not a `Close` call.

## Decisions & Rationale

<!-- **Internal.** Why-this-and-not-that for non-obvious choices. Folded-in ADR. -->

**DR1: Go-after-WaitDone panics (§23.14 amendment to E4)**

The original plan (§21.5 E4) described Go-after-WaitDone as scheduling the goroutine into an orphaned state ("result error appended to next group's join"). This was ambiguous and surprised reviewers: there is no "next group" concept. The §23.14 amendment locks in the panic semantics matching `sync.WaitGroup.Add`-after-`Wait`. The reasoning: `Group` is a run-to-completion primitive. Once `WaitDone` returns, all goroutines have finished and the group's internal channel is closed. Accepting a new `Go` call after this point would require re-opening closed state or silently dropping work — both wrong. Panicking is the correct response to an unambiguous programming error.

**DR2: Default concurrency cap = `runtime.NumCPU() * 64` (§23.14)**

An uncapped Group is a footgun for workloads that loop over large inputs. The cap is not a performance optimization — it is a safety rail. `NumCPU * 64` is empirically generous: Go's scheduler handles thousands of goroutines without degradation on modern hardware. Callers who exceed this and have a genuine reason (I/O-heavy fan-out with external back-pressure) use `WithUnlimited()` and accept the responsibility. The default is documented so it does not surprise callers who observe back-pressure they did not expect.

**DR3: Semaphore cancel-watcher cleanup via `defer cancel()` (§23.14)**

The original design spawned a cancel-watcher goroutine for each slow-path `Acquire`. Without cleanup, a successful acquire that never cancelled would leave the watcher goroutine alive until the per-acquire derived context timed out or was GC'd. At scale (many semaphore acquires per second), this is a goroutine leak. The fix — `defer cancel()` on the derived context at the top of the slow-path block — ensures the watcher goroutine is signalled to exit immediately on successful acquire. `TestSemaphoreCtxWatcherNoLeak` (10k cycles, `fixture.WatchGoroutines`) is the regression guard.

**DR4: No `OnceErr` type; use `sync.OnceValues`**

`sync.OnceValues` (available since Go 1.21) provides exactly the `(T, error)` once-with-result pattern. Adding a parallel `concur.OnceErr` would duplicate stdlib with no differentiation. `concur.Once[T]` adds ctx-pass-through to the first call, which `sync.OnceValues` does not provide — that differentiation justifies the type. `OnceErr` without the ctx affordance would not justify the export.

**DR5: Mutex/RWMutex byte-equivalence in production (NF1)**

`concur.Mutex` embeds `sync.Mutex` as its only field in non-debug builds. The compiler inlines the forwarding `Lock`/`Unlock` calls, producing machine code identical to direct `sync.Mutex` usage. Verified by `TestMutexDebugProductionByteEquivalent` (`unsafe.Sizeof` comparison) and `BenchmarkMutexLockUnlock` (`benchstat` ≤ 5% delta).

**DR6: Group has no Close — WaitDone is the lifecycle (§23.16)**

`Group` owns no external resources (goroutines terminate naturally; the internal Semaphore is in-memory). `Close` would imply a resource-release that does not exist. The lifecycle is `NewGroup → (Go/TryGo)* → WaitDone`; after `WaitDone` returns, the Group is unreachable and GC-eligible. A `Close` would be either a confusing no-op or redundant with the Go-after-WaitDone panic.

**DR7: `LockCtx` uses try-lock-with-backoff, not a goroutine-plus-channel design (§23.14)**

The goroutine-plus-channel alternative allocates one goroutine and one channel per `LockCtx` call, and leaks the goroutine if the mutex is never released. Try-lock-with-backoff allocates nothing and eliminates the leak. The "best-effort cancellation" caveat (up to one backoff interval) is acceptable for the shutdown-path and test-teardown use cases where `LockCtx` is typically called.

**DR8: `Pool[T]` wraps `sync.Pool` generically (D36)**

Direct `sync.Pool` requires a type assertion (`pool.Get().(*MyType)`) at every call site — not compile-time safe. `Pool[T]` moves the single unavoidable assertion inside `Get()`, which is guaranteed to succeed because `Put` accepts only `T`. The generics-first mandate (D36) applies: when the type is statically known, consumers should not write type assertions by hand.

## Open Questions

<!--
  **Internal.** Unresolved items.
  MUST be empty before this spec moves to `accepted` (per CLAUDE.md core directive 1 / D11).
-->

None.

## Verification

<!-- **Internal.** Concrete steps to prove the change works end-to-end. Run when the spec moves to `verified`. -->

1. **`go test ./concur/... -race -count=1`** exits with code 0 on Linux, macOS, and Windows. No race reports, no test failures.
2. **`go test ./concur/... -tags glacier_debug -race -count=1`** exits with code 0. `TestMutexDebugHoldTimeoutEmitsSlog` captures the expected slog event.
3. **`go test ./concur/... -race -count=100`** (no-flake gate) exits with code 0 on all 100 runs. Zero flakes across: `TestGroupGoAfterWaitDonePanics`, `TestSemaphoreCtxWatcherNoLeak`, `TestSemaphoreCtxWatcherNoLeakOnCancel`, `TestMutexLockCtxNoLeakAfterCancel`, `TestGroupConcurrentGoFromManyGoroutines`, `TestOnceConcurrentFirstCallWins`.
4. **Named §23.14 tests pass**: `TestSemaphoreCtxWatcherNoLeak`, `TestSemaphoreCtxWatcherNoLeakOnCancel`, `TestGroupGoAfterWaitDonePanics`, `TestGroupGoAfterWaitDonePanicMessage`, `TestGroupDefaultLimitIsNumCPU64`, `TestMutexLockCtxTryLockBackoffWindow`. Each is run individually to confirm targeted behavior.
5. **Benchmark gates** (`go test ./concur/... -bench=. -benchmem -count=10 -benchstat`):
   - `BenchmarkMutexLockUnlock` vs `BenchmarkStdlibMutex`: delta ≤ 5% ns/op, 0 allocs/op.
   - `BenchmarkSemaphoreAcquireReleaseUncontended`: ≤ 50 ns/op, 0 allocs/op.
   - `BenchmarkGroupGoWithLimit`: allocs/op ≤ 2 (1 closure + 1 recover frame).
   - `BenchmarkPoolGetPut`: allocs/op == allocs/op for `sync.Pool`.
6. **Fuzz gate** (`go test ./concur/... -run=FuzzSemaphoreAcquireRelease -fuzz=. -fuzztime=30s`): no crashes, no unrecovered panics, all panics are the documented over-release panic.
7. **Byte-equivalence** (`go test ./concur/ -run=TestMutexDebugProductionByteEquivalent`): `unsafe.Sizeof(concur.Mutex{}) == unsafe.Sizeof(sync.Mutex{})` in non-debug builds.
8. **Lifecycle gate** (`go test ./concur/ -run=TestGroupHasNoCloseDocumented`): reflection confirms `*Group` has no `Close` method.
9. **Layering check** (`go test ./internal/laytest/ -run=TestLayeringInvariants`): F4 holds — `concur` imports no other Tier 1 package.
10. **Error-format gate** (`go test ./concur/ -run=TestErrFormatRegisterCompliance`): every error string matches `^concur: [a-z]+(?:: [a-z ]+)*$`.
11. **No testify** (`grep -r "github.com/stretchr/testify" ./concur/`): returns zero matches.
