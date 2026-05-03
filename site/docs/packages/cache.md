---
title: cache
---

# cache

<TierBadge tier="leaf" />

<UsedInTasksBadges package-name="cache" />

[View source spec &rarr;](https://github.com/nathanbrophy/glacier/blob/main/specs/0033-cache.md)

## Public summary
<!-- magpie:extract source=specs/0033-cache.md section=public-summary source-checksum=PENDING -->

`cache` is Glacier's generic key-value cache. One interface, three implementations: in-memory, on-disk, and a write-through layered combination. Every implementation is generic over a value type `V`, supports per-key TTL with hybrid defaults, prevents cache stampedes via singleflight on `GetOrLoad`, and exposes hit/miss counters when an OTEL endpoint is configured. The on-disk implementation persists each key as a separate JSON file under a chosen root directory, with advisory file locking so concurrent processes do not corrupt the same key. Zero new direct dependencies: the package leans on `internal/safefile`, `internal/safejson`, `obs`, `errs`, and stdlib. Hot-path `Get` on the in-memory cache is a single map lookup with zero allocations on hit.

<!-- /magpie:extract -->

## Mental model
<!-- magpie:extract source=specs/0033-cache.md section=mental-model source-checksum=PENDING -->

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

<!-- /magpie:extract -->

## API
<!-- magpie:extract source=specs/0033-cache.md section=api source-checksum=PENDING -->

### Cache[V]

```go
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
```

### New[V]

```go
// New constructs an in-memory Cache[V].
func New[V any](opts ...Option) Cache[V]
```

### NewDisk[V]

```go
// NewDisk constructs a disk-backed Cache[V] rooted at path. The directory is
// created with 0o700 if it does not exist. Returns an error only if path is
// not a directory or cannot be created.
func NewDisk[V any](path string, opts ...Option) (Cache[V], error)
```

### NewLayered[V]

```go
// NewLayered composes a primary and a backing cache. Reads consult primary
// first then backing; writes go to both. backing-layer errors degrade the
// composition to primary-only for the failing operation but never propagate
// to the caller.
func NewLayered[V any](primary, backing Cache[V], opts ...Option) Cache[V]
```

### Options

```go
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

### Errors

- `NewDisk` returns the wrapped path-creation or stat error verbatim if the path is unusable. All other paths inside the cache itself never return an error to the caller from `Get`/`Set`/`Delete`. They log via the configured slog.Logger and degrade gracefully.
- `GetOrLoad` returns whatever error the loader returned. Loader errors are **not** cached.
- All sentinel errors live in `errs.Sentinel` form: `cache: corrupt entry: <path>`, `cache: lock timeout: <path>`, etc.

### Observability

When `obs.Provider` is initialized, `cache` calls `obs.Counter("cache.hits", attrs...).Add(1)` etc. without dropping below the zero-allocation budget for `Get` (the obs no-op shim is allocation-free). Counters emitted: `cache.hits`, `cache.misses`, `cache.expirations`, `cache.disk_reads`, `cache.disk_writes`, `cache.singleflight_collapses`.

<!-- /magpie:extract -->

## Examples
<!-- magpie:extract source=specs/0033-cache.md section=examples source-checksum=PENDING -->

In-memory cache with a 5-minute default TTL:

```go
// Example: in-memory cache with 5-minute default TTL.
c := cache.New[string](cache.WithDefaultTTL(5 * time.Minute))
c.Set("greeting", "hello")
v, ok := c.Get("greeting") // "hello", true
```

`GetOrLoad` collapses concurrent misses on the same key onto one loader call:

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

Layered (mem → disk) cache for the SDK version-check:

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

<!-- /magpie:extract -->

## FAQ
<!-- magpie:extract source=specs/0033-cache.md section=faq source-checksum=PENDING -->

<div class="glacier-faq">

**Why not just use `sync.Map`?**

`sync.Map` has no TTL, no disk persistence, and no metrics. The in-memory implementation here is essentially `sync.RWMutex` over a map plus expiry; the value is what comes with that (TTL, layered with disk, mockable, observable).

**Why per-key files instead of a single index file?**

flock granularity. With one index file, every read locks the whole cache. With per-key files, only same-key concurrent operations contend.

**Why JSON on disk and not gob/msgpack?**

Human-readable, debuggable from the shell, no new dependency. The hot path is the in-memory layer; the disk path is for survival across restarts where serialization speed is rarely the bottleneck.

**Can I cache pointer types?**

Yes, but the disk implementation will JSON-marshal the pointed-at value, not the pointer. Round-tripping a pointer type through the disk layer materialises a new pointer.

**Does `GetOrLoad` block other gets on the same key?**

Other concurrent misses on the same key are collapsed onto the in-flight loader. Concurrent **hits** on the same key proceed without blocking.

**How do I expire one entry early?**

Call `Delete(key)`. There is no "expire now" API distinct from delete.

**How do I see hit/miss rates?**

Set `OTEL_EXPORTER_OTLP_ENDPOINT` and the `cache.hits` / `cache.misses` counters appear in your OTEL backend. Without an endpoint, the counter calls are no-ops.

</div>

<!-- /magpie:extract -->
