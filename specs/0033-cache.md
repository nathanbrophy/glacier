---
id: 0033
title: Cache
slug: cache
status: accepted
owner-agent: otter
created: 2026-05-03
last-updated: 2026-05-03
supersedes: []
superseded-by: null
reviewers:
  - { agent: otter,  required: true,  signed-off-at: "2026-05-03" }
  - { agent: lynx,   required: true,  signed-off-at: "2026-05-03" }
  - { agent: falcon, required: true,  signed-off-at: "2026-05-03" }
  - { agent: gopher, required: false, signed-off-at: "2026-05-03" }
  - { agent: magpie, required: false, signed-off-at: null }
implementing-commits: []
verified-at: null
docs-extract:
  - public-summary
  - mental-model
  - api
  - examples
  - faq
---

# Cache

## Public Summary

`cache` is Glacier's generic key-value cache. One interface, three implementations: in-memory, on-disk, and a write-through layered combination. Every implementation is generic over a value type `V`, supports per-key TTL with hybrid defaults, prevents cache stampedes via singleflight on `GetOrLoad`, and exposes hit/miss counters when an OTEL endpoint is configured. The on-disk implementation persists each key as a separate JSON file under a chosen root directory, with advisory file locking so concurrent processes do not corrupt the same key. Zero new direct dependencies: the package leans on `internal/safefile`, `internal/safejson`, `obs`, `errs`, and stdlib. Hot-path `Get` on the in-memory cache is a single map lookup with zero allocations on hit.

## Mental Model

A `Cache[V]` is a fast lookup table for values that are expensive to compute or fetch. Every key has a TTL: when the entry expires, it disappears on the next read. The in-memory implementation is the default and is enough for most uses; the disk-backed and layered implementations exist so a process can survive a restart without losing its caches.

```mermaid
flowchart LR
    Caller[caller] -->|Get/Set/GetOrLoad| Cache[(Cache[V])]
    Cache -.->|in-memory| Mem[map+ttl]
    Cache -.->|on-disk| Disk[per-key JSON files]
    Cache -.->|layered| Layered[mem -> disk]
    Disk -. flock .-> Lock[advisory file locks]
    Cache -.->|optional| Obs[obs counters]
```

A typical SDK use looks like:

```go
versionCache := cache.NewLayered(
    cache.New[ghreleases.Release](cache.WithDefaultTTL(24*time.Hour)),
    cache.NewDisk[ghreleases.Release](cacheDir),
)
release, err := versionCache.GetOrLoad(ctx, "github:nathanbrophy/glacier", func(ctx context.Context) (ghreleases.Release, error) {
    return fetcher.Latest(ctx, "nathanbrophy/glacier")
})
```

`GetOrLoad` collapses concurrent misses on the same key onto a single loader call; everyone waits for that one fetch. Behind the scenes, the in-memory layer satisfies repeat reads in microseconds, the disk layer survives restarts, and the OTEL counters expose hit ratio for ops dashboards when telemetry is enabled.

The cache is **not** a database, a job queue, or a message broker. It is a transient store with bounded staleness.

## Goals

- **One interface, three implementations.** `Cache[V]` is the contract; `MemCache[V]`, `DiskCache[V]`, and `LayeredCache[V]` are the implementations. All carry `+glacier:mock` so tests in dependent packages can drop in a mock without touching disk or memory.
- **Generics-first.** Every public function and type is parameterized over the value type. `any`-typed payloads are not part of the public surface.
- **Zero new direct dependencies.** Stdlib + Glacier internal helpers (`safefile`, `safejson`, `obs`, `errs`, `option`) only.
- **Zero allocations on warm hot path.** `MemCache[V].Get` allocates zero bytes when the key is present and unexpired (verified by a `testing.AllocsPerRun` row).
- **Cross-process safety on disk.** Per-key advisory file lock so two SDK processes can both `glacier version --check` without racing each other to corrupt `versioncheck.json`.
- **Stampede-proof.** `GetOrLoad` collapses concurrent misses for the same key onto one loader call via singleflight.
- **Observability when wanted, free when not.** Counter emission via `obs` is automatic when an OTEL endpoint is configured; otherwise the metric paths are no-ops with no allocation.
- **D-S22 unblock.** Replace the SDK's `TODO(spec 0033 cache)` stub at `cmd/glacier/commands/version.go` with a real `cache.NewLayered` call so version-check has the 24h TTL the SDK spec promised.

## Non-Goals

- **Distributed caching.** No Redis-style network protocol, no replication, no consistent hashing. A future spec may layer that on top.
- **LRU eviction at v0.** TTL-only. The `WithMaxEntries(n)` LRU option is reserved for a follow-up amendment if a real use case appears.
- **Pluggable serializers.** v0 is JSON via `internal/safejson` for disk; in-memory is simple value copy. gob and msgpack are explicitly out of scope.
- **Background sweeper goroutine.** Expired entries are pruned lazily on the next `Get(key)` call. No background scanner.
- **Refresh-ahead / probabilistic re-fetch.** Defer to v1 if real use cases appear.
- **Negative caching as a first-class concept.** A miss is a miss. Callers that want negative caching can `Set` a sentinel value with a short TTL.
- **Encryption at rest.** Files written through `internal/safefile` are written unencrypted. Callers must not put secrets in the cache.

## Architecture

```
cache/
├── cache.go            # Cache[V] interface + +glacier:mock marker; common types
├── doc.go              # package doc; mental model excerpt
├── options.go          # Option type + With* constructors (option pattern via /option)
├── mem.go              # MemCache[V] implementation; ~80 LOC
├── disk.go             # DiskCache[V] implementation; ~120 LOC including flock dance
├── layered.go          # LayeredCache[V] write-through composition; ~50 LOC
├── singleflight.go     # tiny generic singleflight; ~40 LOC
├── observability.go    # obs counter emission when endpoint configured; ~30 LOC
├── example_test.go     # canonical examples
├── mem_test.go
├── disk_test.go
├── layered_test.go
├── singleflight_test.go
├── bench_test.go       # zero-alloc Get; 1us SetWithTTL ceiling
└── zz_generated_mock.go  # produced by glacier generate (mock pkg)
```

### Layering and lifecycle

- **`MemCache[V]`** is goroutine-safe via a single `sync.RWMutex`. Entries are `(value, expiry)` pairs. `Get` takes RLock, checks expiry, returns `(value, true)` if live or `(zero, false)` if expired or absent. Expired entries are pruned on read (lazy cleanup; no goroutine).
- **`DiskCache[V]`** writes one JSON file per key at `<root>/<sha256(key)>.json`. The hash filename keeps untrusted keys from escaping the root via path traversal. Each write goes through `internal/safefile.WriteFileAtomic` (write-to-temp + rename). Each read takes a shared advisory lock; each write takes an exclusive advisory lock. The lock file is `<root>/<sha256(key)>.lock` (sibling to the data file). Lock acquisition uses the existing `internal/lockfile` if present; otherwise this spec adds it as a new internal helper at `internal/lockfile/lockfile.go` (~60 LOC stdlib `golang.org/x/sys/unix.Flock` and the windows `LockFileEx` equivalent).
- **`LayeredCache[V]`** holds a primary (mem) and a backing (disk). `Get` consults primary first; on miss, consults backing; on hit, populates primary. `Set` writes to both, primary first. `Delete` deletes from both, primary first. Errors from the backing layer are logged via `obs` and degrade the cache to mem-only for that operation; they never propagate to the caller.
- **`GetOrLoad`** is a default method on `Cache[V]` provided by an embedded helper struct so each implementation gets it for free. It uses a per-instance `singleflight[V]` to collapse concurrent misses.

### Concurrency contract

| Operation | Concurrency notes |
|---|---|
| `MemCache[V].Get` | Goroutine-safe; RLock; allocation-free on hit |
| `MemCache[V].Set` | Goroutine-safe; Lock; one allocation per entry |
| `DiskCache[V].Get` | Goroutine-safe and **process-safe** via flock(SH); one disk read |
| `DiskCache[V].Set` | Goroutine-safe and process-safe via flock(EX); one safefile rename |
| `LayeredCache[V].Get` | Inherits both layers' contracts |
| `Cache[V].GetOrLoad` | Goroutine-safe; concurrent misses on same key share one loader |

### Observability

When `OTEL_EXPORTER_OTLP_ENDPOINT` is set at process startup, `obs.Provider` is initialized and `cache` emits these counters via `obs.Counter`:

- `cache.hits{layer="mem"|"disk", impl="..."}`: successful Get
- `cache.misses{layer, impl}`
- `cache.expirations{layer, impl}`
- `cache.disk_reads`, `cache.disk_writes`
- `cache.singleflight_collapses`: incremented when a concurrent miss is collapsed

When the OTEL endpoint is unset, the obs API is a no-op shim and these counter calls are inlined-out at run-time. A test row asserts `testing.AllocsPerRun(100, func() { c.Get("k") })` is zero.

## Schema

```go
package cache

import (
    "context"
    "log/slog"
    "time"
)

// Cache is the generic key-value cache contract. All implementations in this
// package satisfy Cache[V]; tests in dependent packages can mock it.
//
// invariant: zero-value V is returned alongside ok=false on any miss
// invariant: Get is goroutine-safe in every implementation
//
// +glacier:mock
type Cache[V any] interface {
    // Get returns the value for key and whether it was found and unexpired.
    Get(key string) (value V, ok bool)

    // Set stores value under key with the cache's default TTL. If no default
    // TTL is configured, the entry is stored without expiry (TTL == 0).
    Set(key string, value V)

    // SetWithTTL stores value under key with an explicit TTL. ttl <= 0 stores
    // the entry without expiry.
    SetWithTTL(key string, value V, ttl time.Duration)

    // Delete removes the entry for key. No-op if absent.
    Delete(key string)

    // GetOrLoad returns the value for key. On miss, loader is called and the
    // result is stored before being returned. Concurrent misses on the same
    // key share a single loader call (singleflight). The loader's context is
    // ctx; loader errors are not cached.
    GetOrLoad(ctx context.Context, key string, loader func(context.Context) (V, error)) (V, error)
}

// Option configures a cache implementation at construction time.
type Option = option.Option[config]

// config is the internal aggregation target for Option values.
// invariant: defaultTTL >= 0
// invariant: clock != nil after Apply
type config struct {
    defaultTTL time.Duration
    clock      func() time.Time
    logger     *slog.Logger
}
```

The disk format on disk per key is:

```json
{
  "value":     <V serialized via internal/safejson>,
  "stored_at": "2026-05-03T12:34:56Z",
  "ttl_ns":    86400000000000
}
```

Files with malformed JSON or unexpected schema versions are treated as a miss, deleted, and logged via the configured logger.

## API

```go
// New constructs an in-memory Cache[V].
func New[V any](opts ...Option) Cache[V]

// NewDisk constructs a disk-backed Cache[V] rooted at path. The directory is
// created with 0o700 if it does not exist. Returns an error only if path is
// not a directory or cannot be created.
func NewDisk[V any](path string, opts ...Option) (Cache[V], error)

// NewLayered composes a primary and a backing cache. Reads consult primary
// first then backing; writes go to both. backing-layer errors degrade the
// composition to primary-only for the failing operation but never propagate
// to the caller.
func NewLayered[V any](primary, backing Cache[V], opts ...Option) Cache[V]

// WithDefaultTTL sets the default TTL for entries stored via Set.
// Default: 0 (no expiry).
func WithDefaultTTL(d time.Duration) Option

// WithLogger sets the slog.Logger used for non-fatal messages
// (e.g. corrupt cache file). Default: slog.Default().
func WithLogger(l *slog.Logger) Option

// WithClock injects a clock function. Tests pass a deterministic clock to
// exercise expiry logic without sleeping. Default: time.Now.
func WithClock(c func() time.Time) Option
```

### Error contract

- `NewDisk` returns the wrapped path-creation or stat error verbatim if the path is unusable. All other paths inside the cache itself never return an error to the caller from `Get`/`Set`/`Delete`. They log via the configured slog.Logger and degrade gracefully.
- `GetOrLoad` returns whatever error the loader returned. Loader errors are **not** cached.
- All sentinel errors live in `errs.Sentinel` form: `cache: corrupt entry: <path>`, `cache: lock timeout: <path>`, etc.

### Observability hooks

When `obs.Provider` is initialized, `cache` calls `obs.Counter("cache.hits", attrs...).Add(1)` etc. without dropping below the zero-allocation budget for `Get` (the obs no-op shim is allocation-free). Counters are documented under § Verification.

## Examples

```go
// Example: in-memory cache with 5-minute default TTL.
c := cache.New[string](cache.WithDefaultTTL(5 * time.Minute))
c.Set("greeting", "hello")
v, ok := c.Get("greeting") // "hello", true
```

```go
// Example: GetOrLoad collapses concurrent misses.
type Release struct{ Tag string }

c := cache.New[Release](cache.WithDefaultTTL(24 * time.Hour))
loader := func(ctx context.Context) (Release, error) {
    return Release{Tag: "v1.2.3"}, nil // expensive in real life
}

// Two goroutines miss at once. loader is called exactly once.
var wg sync.WaitGroup
for i := 0; i < 2; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        _, _ = c.GetOrLoad(ctx, "latest", loader)
    }()
}
wg.Wait()
```

```go
// Example: layered (mem -> disk) cache for the SDK version-check.
mem := cache.New[ghreleases.Release](cache.WithDefaultTTL(24 * time.Hour))
disk, err := cache.NewDisk[ghreleases.Release](filepath.Join(userCacheDir, "glacier"))
if err != nil { return err }

c := cache.NewLayered(mem, disk)
release, err := c.GetOrLoad(ctx, "github:nathanbrophy/glacier", func(ctx context.Context) (ghreleases.Release, error) {
    return fetcher.Latest(ctx, "nathanbrophy/glacier")
})
```

## Test Matrix

(Lynx owns the full matrix during sign-off. Plan locks the row categories below; Lynx may add or merge rows.)

| Scenario | Input | Expected | Covered by |
|---|---|---|---|
| **Mem hit** | Set then Get same key | (value, true), 0 allocs | `TestMemHit`, `BenchmarkMemHit` |
| **Mem miss** | Get of unknown key | (zero, false) | `TestMemMiss` |
| **Mem expiry** | Set with TTL=1ms; advance clock; Get | (zero, false) | `TestMemExpiry` |
| **Mem default TTL** | Cache with default 1h, Set, advance 30m, Get | hit | `TestMemDefaultTTL` |
| **Mem zero TTL** | Cache with default 0, Set, advance 100y, Get | hit | `TestMemZeroTTLNoExpiry` |
| **Mem Set overwrites** | Set k=A; Set k=B; Get k | B | `TestMemSetOverwrites` |
| **Mem Delete** | Set, Delete, Get | miss | `TestMemDelete` |
| **Mem concurrency** | 100 goroutines Get/Set | no race, no panic | `TestMemRace` (with -race) |
| **Mem zero alloc on hit** | testing.AllocsPerRun(100, Get) | == 0 | `TestMemZeroAllocOnHit` |
| **Disk hit** | Set, Get | hit; file exists at <hash>.json | `TestDiskHit` |
| **Disk miss** | Get of unknown | miss; no file written | `TestDiskMiss` |
| **Disk expiry** | Set with TTL=1ms; advance clock; Get | miss; file deleted | `TestDiskExpiry` |
| **Disk persistence** | Set; new DiskCache on same path; Get | hit | `TestDiskPersistsAcrossInstances` |
| **Disk corrupt file** | Manually write bad JSON to <hash>.json; Get | miss; file deleted; warning logged | `TestDiskCorruptFile` |
| **Disk path traversal** | Key="../etc/passwd" | file lands inside root, not at /etc | `TestDiskPathTraversalBlocked` |
| **Disk cross-process flock** | Two processes Set same key concurrently | both succeed; final value is one of them; no torn write | `TestDiskFlockTwoProcesses` |
| **Disk read while write** | Goroutine writing; goroutine reading | reader sees old or new value, never partial | `TestDiskReadDuringWrite` |
| **Disk safefile rename** | Inject safefile.WriteFileAtomic failure | Set returns silently; warning logged | `TestDiskSafefileFailureDegrades` |
| **Layered hit primary** | Set, Get | hit from mem; disk not touched | `TestLayeredHitPrimary` |
| **Layered hit backing** | Disk has key; mem cold; Get | hit; mem populated | `TestLayeredHitBacking` |
| **Layered backing failure** | Disk Get returns error | falls through to mem-only | `TestLayeredBackingErrorDegrades` |
| **Layered Delete both** | Set, Delete, Get | miss in both | `TestLayeredDeleteBoth` |
| **GetOrLoad miss** | Empty cache, GetOrLoad with loader | loader called once, value cached | `TestGetOrLoadMissCallsLoader` |
| **GetOrLoad hit** | Pre-populate, GetOrLoad | loader not called | `TestGetOrLoadHitSkipsLoader` |
| **GetOrLoad singleflight** | 100 goroutines GetOrLoad same key | loader called exactly once | `TestGetOrLoadCollapsesConcurrent` |
| **GetOrLoad loader error** | Loader returns error | error returned, value not cached | `TestGetOrLoadErrorNotCached` |
| **GetOrLoad context cancel** | Cancel ctx before loader returns | error == ctx.Err() | `TestGetOrLoadContextCancel` |
| **OTEL counters emitted** | OTEL endpoint configured; Get/Set | counters increment | `TestObsCounterEmission` |
| **OTEL counters no-op when unset** | No endpoint; Get/Set | testing.AllocsPerRun unchanged | `TestObsCounterNoOpZeroAlloc` |
| **Mock satisfies Cache[V]** | mock.Of[Cache[string]]() | compiles; all methods callable | `TestMockSatisfiesContract` |
| **WithClock determinism** | Two caches, same fake clock | identical observable behaviour | `TestWithClockDeterministic` |
| **Logger overrides slog.Default** | WithLogger; trigger corrupt-file warning | warning lands on injected logger | `TestLoggerOverride` |
| **Hash filename collision-free** | 10k random keys | 10k distinct files | `TestHashFilenameUnique` |

Bench rows (≥ 5):

| Bench | Budget |
|---|---|
| `BenchmarkMemHit` | ≤ 50 ns/op, 0 B/op, 0 allocs/op |
| `BenchmarkMemSet` | ≤ 1 µs/op, 1 alloc/op |
| `BenchmarkDiskGet` (warm OS cache) | ≤ 50 µs/op |
| `BenchmarkDiskSet` | ≤ 200 µs/op |
| `BenchmarkGetOrLoadCollapse` | 100 concurrent misses ≤ 2× single loader latency |

Plan-locked total: ≥ 30 rows + 5 bench rows. Lynx may add or merge during sign-off.

## Dependency Justification

| Module | Version | License | Last release | Maintainers | Alternatives considered | Why we can't roll our own |
|---|---|---|---|---|---|---|
|  |  |  |  |  |  |  |

(Empty: no new direct dependencies. flock uses `golang.org/x/sys` which is already an indirect of stdlib's network packages and is on the existing accepted-dep list. Confirm during Falcon review.)

## Security & Supply-Chain Notes

- **Path traversal:** every disk file lands at `<root>/<sha256(key)>.json`. The hash filename neutralises any `..`/`/`/null bytes in the key. A test row asserts `Set("../etc/passwd", v)` writes inside root.
- **Untrusted JSON:** disk reads use `internal/safejson` which has size and depth limits. Corrupt files are deleted and logged.
- **flock advisory only:** advisory locking on Unix and Windows is best-effort. Two processes that ignore the lock could still corrupt a file. The cache assumes cooperative callers and does not provide stronger guarantees.
- **No secrets:** files are written unencrypted at 0o600. Callers must not store credentials in cached values. A test row inspects the disk-cache `embed` directive isn't used and rest mode flags 0o600 perms after every Set.
- **Disk pressure:** the cache does not bound total disk usage at v0. A misbehaving caller that Sets many distinct keys with no TTL fills the disk. Documented in FAQ; mitigated by callers via `WithDefaultTTL`.
- **No network:** `cache` issues no network calls. The `obs` counter emission is local; the OTEL exporter is owned by `obs`.

## Migration & Compatibility

- **D-S22 in spec 0032 is unblocked.** Spec 0032 left a `TODO(spec 0033 cache)` stub in `cmd/glacier/commands/version.go`. After this spec moves to `accepted` and the package is implemented, the stub becomes:

```go
mem := cache.New[ghreleases.Release](cache.WithDefaultTTL(24 * time.Hour))
disk, _ := cache.NewDisk[ghreleases.Release](filepath.Join(userCacheDir, "glacier"))
versionCache := cache.NewLayered(mem, disk)
release, err := versionCache.GetOrLoad(ctx, "github:" + repo, func(ctx context.Context) (ghreleases.Release, error) {
    return c.releaseFetcher().Latest(ctx, repo)
})
```

The lint cache and bench baseline (D-S23 in spec 0032) similarly switch to this package. Each call site is updated in the implementing commit; no spec amendments are needed in 0032.

- **Spec 0010 (mock):** `Cache[V]` carries `+glacier:mock`, so generated mocks ship under `cache/zz_generated_mock.go`. No amendment to spec 0010 is needed; this is the documented use case.

- **Backwards compat:** none required (new package).

## FAQ

**Q: Why not just use sync.Map?**
A: sync.Map has no TTL, no disk persistence, and no metrics. The in-memory implementation here is essentially `sync.RWMutex` over a map plus expiry; the value is what comes with that (TTL, layered with disk, mockable, observable).

**Q: Why per-key files instead of a single index file?**
A: flock granularity. With one index file, every read locks the whole cache. With per-key files, only same-key concurrent operations contend.

**Q: Why JSON on disk and not gob/msgpack?**
A: human-readable, debuggable from the shell, no new dependency. The hot path is the in-memory layer; the disk path is for survival across restarts where serialization speed is rarely the bottleneck.

**Q: Can I cache pointer types?**
A: Yes, but the disk implementation will JSON-marshal the pointed-at value, not the pointer. Round-tripping a pointer type through the disk layer materialises a new pointer.

**Q: Does GetOrLoad block other gets on the same key?**
A: Other concurrent misses on the same key are collapsed onto the in-flight loader. Concurrent **hits** on the same key proceed without blocking.

**Q: How do I expire one entry early?**
A: Call `Delete(key)`. There is no "expire now" API distinct from delete.

**Q: How do I see hit/miss rates?**
A: Set `OTEL_EXPORTER_OTLP_ENDPOINT` and the `cache.hits` / `cache.misses` counters appear in your OTEL backend. Without an endpoint, the counter calls are no-ops.

## Decisions & Rationale

| ID | Decision | Rationale |
|---|---|---|
| **D-C1** | Spec ID 0033, owner Otter, initial status `proposed`. | Next free ID; Otter owns architecture per spec 0000 sign-off matrix. |
| **D-C2** | Required reviewers: Otter, Lynx, Falcon. Optional: Gopher, Magpie. | Component spec; Lynx for test matrix; Falcon for the supply-chain claim ("zero new direct deps"). |
| **D-C3** | One interface, three impls. | Strongest DI story; tests in dependent packages mock `Cache[V]` instead of stubbing files. |
| **D-C4** | Generic `Cache[V any]`. | Project memory: generics-first; eliminates marshal/unmarshal boilerplate. |
| **D-C5** | TTL-only eviction at v0. | Less code is more. LRU adds a heap + size accounting + 50+ LOC; no current use case needs it. |
| **D-C6** | `GetOrLoad` with built-in singleflight. | Common pattern; ~30 LOC; eliminates a footgun callers would otherwise hit. |
| **D-C7** | JSON on disk via `internal/safejson`. | Already dogfooded; human-readable; no new dep. |
| **D-C8** | Per-key files at `<root>/<sha256(key)>.json`. | flock granularity; path-traversal safety. |
| **D-C9** | flock per file via `internal/lockfile`. | D-S22 in spec 0032 explicitly calls for `versioncheck.lock`; flock is the cooperative cross-process primitive. |
| **D-C10** | Hybrid TTL: per-instance default + per-Set override. | Most ergonomic for caller. Default is the common case; override is the safety valve. |
| **D-C11** | `Get` returns `(V, bool)`; no Status enum. | v0 simplicity. Negative caching is achievable by Set with a sentinel. |
| **D-C12** | OTEL counters auto-emit when endpoint configured. | Zero overhead when unset (verified by alloc-count test row). |
| **D-C13** | Lazy expiry (no background sweeper). | Less code, fewer fail modes, no goroutine lifecycle. Memory hold during the gap between expiry and next Get is bounded by caller behaviour. |
| **D-C14** | No refresh-ahead. | v0; revisit if a use case appears. |
| **D-C15** | Errors stay inside (logged via slog), Get/Set/Delete never return them. | Cache errors should not break callers' hot paths. NewDisk returns errors at construction time only. |
| **D-C16** | Layered backing-layer errors degrade to primary-only. | Mem cache survives even if disk is full or read-only. |
| **D-C17** | Disk file mode 0o600. | Default-private; cache may contain user-specific data. |
| **D-C18** | Hash filename uses sha256(key). | Eliminates path traversal and 2^256 keys before collision is acceptable. |
| **D-C19** | No encryption at rest. | Out of scope; documented in Security notes. Callers must not put secrets in the cache. |
| **D-C20** | `Option = option.Option[config]` from /option package. | Dogfood the framework's own option pattern (project memory). |

## Open Questions

None.

## Verification

The spec moves to `verified` when all of these hold:

1. `go test ./cache/...` passes with `-race` on linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64.
2. `BenchmarkMemHit` reports 0 allocs/op and ≤ 50 ns/op on commodity hardware.
3. `TestMemZeroAllocOnHit` confirms `testing.AllocsPerRun(100, func() { c.Get("k") }) == 0`.
4. `TestObsCounterNoOpZeroAlloc` confirms unchanged alloc count when no OTEL endpoint is configured.
5. `TestDiskFlockTwoProcesses` runs two `os.exec` subprocesses contending on the same key and observes no torn write.
6. `TestDiskPathTraversalBlocked` confirms `Set("../etc/passwd", v)` writes inside the cache root.
7. `TestGetOrLoadCollapsesConcurrent` confirms loader is called exactly once across 100 concurrent missing readers.
8. `TestMockSatisfiesContract` confirms `mock.Of[Cache[string]]()` compiles and all interface methods are callable on the mock.
9. `cmd/glacier/commands/version.go` no longer contains the `TODO(spec 0033 cache)` marker; the version-check uses the layered cache; `glacier version --check` does not re-fetch from GitHub on a second run within 24h (validated by an httpmock-backed integration test).
10. `go.mod` direct-dependency count is unchanged from before this spec landed (Falcon-verified during sign-off).
11. The `cache` package is registered in `internal/reflectx`'s package-allow-list if applicable (otherwise no-op).
12. `glacier lint --severity=error ./cache/...` returns clean.
13. The package ships at least one `Example*()` function (per spec 0008 example-test policy).
