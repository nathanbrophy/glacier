---
title: Writing tests
---

# Writing tests

<PackagesUsedBadges :package-names="['assert', 'fixture', 'mock', 'errs']" />

Glacier's test packages let you write table-driven tests that verify values precisely, inject deterministic fakes for time and the filesystem, and set fluent expectations on interface mocks - all without importing a third-party test framework. The packages compose cleanly: `assert` checks values, `fixture` manages test resources, and `mock` verifies interface behavior.

## Walkthrough

### Step 1 - Assert values with smart equality

Import `assert` for checks that continue on failure, or `assert/require` when a nil return would crash the next assertion. The `Equal[T]` function accepts option values that configure the comparison engine.

```go
import (
    "testing"
    "github.com/nathanbrophy/glacier/assert"
    "github.com/nathanbrophy/glacier/assert/require"
)

func TestUserNames(t *testing.T) {
    got  := []string{"carol", "alice", "bob"}
    want := []string{"alice", "bob", "carol"}

    // IgnoreOrder treats slices as multisets.
    assert.Equal(t, got, want, assert.IgnoreOrder())

    // require stops the test immediately on failure.
    require.NoError(t, loadUser(t, "u-1"))
}
```

`assert.Equal` calls `t.Helper()` and reports every failure in one run. `require.NoError` calls `t.FailNow` - use it when continuing past a nil value would panic.

### Step 2 - Inject a fake clock

Code that calls `time.Now()` is hard to test deterministically. Accept a `fixture.Clock` interface instead and inject `fixture.NewClock(t, start)` in tests. Calling `clk.Advance` drives timers without wall-clock sleeping.

```go
import (
    "context"
    "testing"
    "time"

    "github.com/nathanbrophy/glacier/assert"
    "github.com/nathanbrophy/glacier/fixture"
)

func TestRetryTimeout(t *testing.T) {
    start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
    clk   := fixture.NewClock(t, start)
    deadline := start.Add(5 * time.Second)

    calls := 0
    result := retryUntil(clk, deadline, func() bool {
        calls++
        clk.Advance(100 * time.Millisecond)
        return calls >= 3
    })

    assert.True(t, result)
    assert.Equal(t, calls, 3)
}
```

`fixture.NewClock` registers cleanup with `t.Cleanup` - your test stays linear with no deferred teardown to write.

### Step 3 - Mock an interface

Pass any interface type parameter to `mock.Of[T]` to get a mock whose expectations you set with a fluent builder. Matchers are type-safe at compile time.

```go
import (
    "context"
    "testing"

    "github.com/nathanbrophy/glacier/assert"
    "github.com/nathanbrophy/glacier/mock"
)

type Repo interface {
    FindUser(ctx context.Context, id string) (User, error)
    SaveUser(ctx context.Context, u User) error
}

func TestService_FindUser(t *testing.T) {
    m := mock.Of[Repo](t)

    m.OnCall("FindUser").
        With(mock.Any[context.Context](), mock.Eq[string]("u-42")).
        Return(User{ID: "u-42", Name: "Alice"}, nil).
        Times(1)

    m.OnCall("SaveUser").Never()

    svc := NewService(m.Interface())
    got, err := svc.FindUser(context.Background(), "u-42")
    assert.Equal(t, got.Name, "Alice")
    assert.NoError(t, err)
    // Verify() runs automatically at t.Cleanup.
}
```

`mock.Eq[string]` will not compile if the corresponding parameter is not a `string` - the check is at compile time, not runtime.

### Step 4 - Verify error types

Use `errs.Wrap` in production code and `assert.ErrorIs` / `assert.ErrorAs` in tests to check the error chain without losing information.

```go
import (
    "errors"
    "testing"

    "github.com/nathanbrophy/glacier/assert"
    "github.com/nathanbrophy/glacier/errs"
)

func TestLoadConfig_MissingFile(t *testing.T) {
    err := loadConfig("/nonexistent.json")

    // Check the sentinel at any depth in the chain.
    assert.True(t, errors.Is(err, errs.ErrNotFound))

    // Extract typed detail.
    var pe *errs.PathError
    if assert.True(t, errors.As(err, &pe)) {
        assert.Equal(t, pe.Path, "/nonexistent.json")
    }
}
```

## Putting it together

```go
package service_test

import (
    "context"
    "testing"
    "time"

    "github.com/nathanbrophy/glacier/assert"
    "github.com/nathanbrophy/glacier/assert/require"
    "github.com/nathanbrophy/glacier/fixture"
    "github.com/nathanbrophy/glacier/mock"
)

type Cache interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

func TestService_CacheHit(t *testing.T) {
    t.Parallel()

    start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
    clk   := fixture.NewClock(t, start)

    mc := mock.Of[Cache](t)
    mc.OnCall("Get").
        With(mock.Any[context.Context](), mock.Eq[string]("user:42")).
        Return([]byte(`{"id":42,"name":"Ada"}`), nil).
        Times(1)

    svc := NewService(mc.Interface(), clk)

    user, err := svc.User(context.Background(), "42")
    require.NoError(t, err)
    assert.Equal(t, user.Name, "Ada")
    assert.Equal(t, clk.Now(), start, "clock must not advance on cache hit")
}
```

## What's happening underneath

- <TierBadge tier="kernel" /> [`assert`](/docs/packages/assert): smart deep-equal engine with pointer deref, map-order insensitivity, float tolerance, and field exclusion; `require` sub-package adds `t.FailNow` semantics.
- <TierBadge tier="mid" /> [`fixture`](/docs/packages/fixture): manages test resources (fake clock, in-memory FS, goroutine leak guard) and registers all cleanup with `t.Cleanup`.
- <TierBadge tier="mid" /> [`mock`](/docs/packages/mock): reflect-based runtime mocks for any exported interface; optional `+glacier:mock` codegen path for full IDE autocomplete.
- <TierBadge tier="kernel" /> [`errs`](/docs/packages/errs): wraps and composes errors in a way that keeps the full chain traversable via `errors.Is` and `errors.As`.

## Related

- [Mocking HTTP](/docs/mocking-http) - stub HTTP servers for tests that call external APIs.
- [Loading config](/docs/loading-config) - test config loading with an in-memory FS fixture.
- [Concurrency](/docs/concurrency) - testing goroutine groups and checking for goroutine leaks.
