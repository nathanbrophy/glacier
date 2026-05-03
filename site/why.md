---
title: Why Glacier
aside: false
---

# Why Glacier

<div class="glacier-sprite-accent">
  <MascotSprite state="wave" :size="80" />
</div>

Glacier is a Go framework for developers who have already decided to write Go and would rather not re-implement the same plumbing in every project. This page walks the four pillars you saw on the landing page and shows the concrete tradeoff behind each one.

## Less plumbing. More Go. {#less-plumbing}

Every non-trivial Go program needs at least five things before it can do useful work: argument parsing, configuration loading, structured logging, lifecycle management, and signal handling. Most teams write these in the first week, and most teams write them slightly differently each time.

Glacier draws a sharp line. Code unique to your problem stays on your side. Everything generic stays on Glacier's side.

```go
// Your side: a struct and a Run method.
// +glacier:command name=server
// +glacier:root
type Server struct {
    // +glacier:default 8080
    // +glacier:usage listen port
    Port string

    // +glacier:env DATABASE_URL
    // +glacier:required
    DSN string
}

func (s *Server) Run(ctx context.Context) error {
    db, err := sql.Open("pgx", s.DSN)
    if err != nil {
        return err
    }
    defer db.Close()
    return listenAndServe(ctx, s.Port, db)
}

// Glacier's side: everything else.
func main() { cli.Main(&Server{}) }
```

The `cli.Main` call wires flag parsing, environment variable binding, config layering, `SIGINT`/`SIGTERM` cancellation into the context, structured log initialization, and exit-code mapping. None of that is your problem.

When the boundary feels wrong - when something generic is on your side or something domain-specific is on Glacier's side - that's the spec to file.

## Curated suite, designed together. {#curated-suite}

15 packages. Three tiers. Zero cycles.

<TierBadge tier="kernel" /> Five kernel packages that every consumer transitively depends on: `option`, `errs`, `log`, `assert`, `term`. They have no dependencies on each other except where the DAG explicitly allows it.

<TierBadge tier="mid" /> Five mid-tier packages - `concur`, `fluent`, `conf`, `fixture`, `obs` - each independent of the others, depending only on the kernel.

<TierBadge tier="leaf" /> Five leaf packages - `cli`, `mock`, `httpmock`, `httpc`, `cache` - large enough to justify isolation, never importing each other.

Every package configurable at construction uses the same `option.Option[T]` protocol. Every error from every package follows the same library register: `lowercase, no trailing period, package: action: cause`. Every package that logs injects via `WithLogger(*slog.Logger)`. These are testable invariants, not aspirational guidelines. A Lynx-owned layering test rejects forbidden import edges on every PR.

```go
// option.Option[T] is the shared construction protocol.
// Every package accepts ...option.Option[fooConfig] in its constructor.

transport, err := httpmock.New(
    httpmock.WithLogger(logger),
    httpmock.Strict(), // zero-arg boolean option
)

loader, err := conf.New(
    conf.WithFile("config.json"),
    conf.WithEnvPrefix("APP"),
)
```

Because every package speaks the same options protocol, you never have to learn a different construction idiom per package.

## Generics-first ergonomics. {#generics-first}

Go 1.18 introduced generics. Glacier uses them where they remove boilerplate, not where they add complexity.

`conf.Register[T]` gives you a typed accessor. You call `Register[string]("log.level", ...)` once and get back a `func() string` - no type assertions at call sites.

`fluent.Map[A, B]` maps a lazy `iter.Seq[A]` to an `iter.Seq[B]` without any `any` in the call. `fluent.Filter[T]`, `fluent.Take[T]`, `fluent.GroupBy[K, V]` follow the same pattern.

`assert.Equal[T]` compares two values of the same type with smart deep-comparison (ignoring order when you ask, applying a delta for floats, ignoring fields you mark). The type parameter is inferred; you don't write it.

`mock.Of[T]` returns a mock of any interface `T` without code generation, powered by reflection. For production use, `+glacier:mock` codegen emits a typed wrapper that removes the last reflection call at test time.

```go
// Typed accessor - no cast at call sites.
logLevel := conf.Register[string](loader, "log.level",
    conf.WithDefault("info"),
)
fmt.Println(logLevel()) // "info" or whatever is in the config

// Lazy sequence pipeline - all generics, no any.
nums := fluent.Of(iter.Seq[int](func(yield func(int) bool) {
    for i := 0; i < 100; i++ { yield(i) }
}))
evens := fluent.Filter(nums, func(n int) bool { return n%2 == 0 })
first10 := fluent.Take(evens, 10)

// Type-safe deep equality in tests.
assert.Equal(t, want, got, assert.IgnoreOrder())
```

`any` shows up in Glacier where the type genuinely is unknown at compile time. Everywhere a type is known, generics carry it.

## Test-first, dogfooded helpers. {#test-first}

Glacier's goal is that `go test ./...` gives you full release confidence. No manual testing needed to release.

Every Glacier package is tested using Glacier's own test helpers. `assert` handles equality and invariant checks. `fixture` provides golden files, typed snapshots, deterministic fake clocks, and in-memory filesystems. `mock` handles interface fakes. `httpmock` handles HTTP transport faking.

If a helper isn't good enough for Glacier itself, it doesn't ship. Dogfooding is the strongest quality signal in the project.

```go
func TestUserService_Create(t *testing.T) {
    db := mock.Of[Database](t)
    mock.Expect(db, "Insert").
        With(mock.AnyContext(), mock.MatchType[User]()).
        Return(User{ID: "u1"}, nil)

    svc := NewUserService(db)
    got, err := svc.Create(t.Context(), CreateRequest{Name: "Alice"})

    assert.NoError(t, err)
    assert.Equal(t, User{ID: "u1", Name: "Alice"}, got)
    // mock.Verify is called automatically via t.Cleanup - no explicit call needed.
}
```

Goroutine leak detection, environment-variable isolation, and fake filesystem resets happen at the fixture layer, not scattered across individual test files. The test itself stays readable.

See [/concepts](/concepts) for how the three-tier DAG shapes which packages a test legitimately imports.
