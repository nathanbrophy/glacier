---
layout: page
title: Glacier
titleTemplate: ':title'
---

<HeroSection />

<div class="glacier-pillars-section">

## Why Glacier

<div class="glacier-pillars">

<PillarCard title="Less plumbing. More Go.">

Idiomatic handler code stays on your side. The rest stays on Glacier's. You write a struct with a <code>Run(ctx)</code> method; <code>glacier.Run</code> handles argument parsing, signal handling, configuration layering, and lifecycle teardown. Nothing generic belongs on your side.

</PillarCard>

<PillarCard title="Curated suite, designed together.">

15 packages in three tiers (kernel, mid, leaves). They interlock by design, not by accident. Every package speaks the same functional-options protocol, the same error register, and the same context-propagation rules. Pick one package or all 15 - the seams don't leak.

</PillarCard>

<PillarCard title="Generics-first ergonomics.">

Type-parameterized helpers replace marshal/decode/cast boilerplate. `any` is a fallback, not a default. `conf.Register[T]` gives you a typed accessor. `fluent.Map[A, B]` maps without casts. `assert.Equal[T]` compares without reflection surprises.

</PillarCard>

<PillarCard title="Test-first, dogfooded helpers.">

`go test` gives release confidence. Glacier's own packages are tested with Glacier's own helpers: `assert`, `fixture`, `mock`, `httpmock`. If a helper isn't good enough for Glacier itself, it doesn't ship.

</PillarCard>

</div>

</div>

<PromiseSection />

<SdkSpotlight />

## The 15-package suite

<PackageGrid />

## With Glacier vs without

The framework's value shows up in concrete diffs. Below: the plumbing you stop writing, the test ergonomics you gain, and the HTTP boundaries you stop hand-rolling.

### Building a CLI

<CodeCompare withTitle="With Glacier" withoutTitle="Without Glacier">

<template #without>

```go
package main

import (
    "context"
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    port := flag.String("port", "8080", "listen port")
    debug := flag.Bool("debug", false, "verbose logging")
    flag.Parse()

    host := os.Getenv("APP_HOST")
    if host == "" {
        host = "localhost"
    }

    ctx, stop := signal.NotifyContext(
        context.Background(),
        syscall.SIGINT, syscall.SIGTERM,
    )
    defer stop()

    if *debug {
        log.SetFlags(log.LstdFlags | log.Lshortfile)
    }
    log.Printf("starting on %s:%s", host, *port)

    if err := serve(ctx, host, *port); err != nil {
        log.Fatal(err)
    }
}
```

</template>

<template #with>

```go
package main

import (
    "context"
    "github.com/nathanbrophy/glacier/cli"
)

// +glacier:command name=app
// +glacier:root
type App struct {
    // +glacier:default 8080
    // +glacier:usage listen port
    Port string

    // +glacier:short v
    // +glacier:usage verbose logging
    Debug bool
}

func (a *App) Run(ctx context.Context) error {
    return serve(ctx, a.Port)
}

func main() { cli.Main(&App{}) }
```

</template>

</CodeCompare>

### Mocking an interface

<CodeCompare withTitle="With Glacier" withoutTitle="Without Glacier">

<template #without>

```go
// Hand-rolled fake. Every new method = another field + another assertion.
type fakeStore struct {
    getCalls  []string
    getReturn map[string]Order
    getErr    error

    putCalls  []Order
    putErr    error
}

func (f *fakeStore) Get(_ context.Context, id string) (Order, error) {
    f.getCalls = append(f.getCalls, id)
    return f.getReturn[id], f.getErr
}
func (f *fakeStore) Put(_ context.Context, o Order) error {
    f.putCalls = append(f.putCalls, o)
    return f.putErr
}

func TestProcess(t *testing.T) {
    s := &fakeStore{getReturn: map[string]Order{"o-1": {ID: "o-1"}}}
    if err := Process(context.Background(), s, "o-1"); err != nil {
        t.Fatal(err)
    }
    if len(s.getCalls) != 1 || s.getCalls[0] != "o-1" {
        t.Fatalf("Get not called as expected: %v", s.getCalls)
    }
}
```

</template>

<template #with>

```go
// +glacier:mock
type Store interface {
    Get(ctx context.Context, id string) (Order, error)
    Put(ctx context.Context, o Order) error
}

func TestProcess(t *testing.T) {
    s := mock.New[Store](t)
    s.On("Get", mock.Any, "o-1").
        Return(Order{ID: "o-1"}, nil).Once()

    err := Process(context.Background(), s, "o-1")
    assert.NoError(t, err)
    s.AssertExpectations(t)
}
```

</template>

</CodeCompare>

### Mocking HTTP

<CodeCompare withTitle="With Glacier" withoutTitle="Without Glacier">

<template #without>

```go
func TestFetchUser(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            if r.URL.Path != "/users/42" {
                http.NotFound(w, r)
                return
            }
            if r.Header.Get("Authorization") == "" {
                http.Error(w, "no auth", http.StatusUnauthorized)
                return
            }
            w.Header().Set("Content-Type", "application/json")
            _, _ = w.Write([]byte(`{"id":42,"name":"ada"}`))
        }))
    defer srv.Close()

    c := &Client{BaseURL: srv.URL, Token: "t"}
    u, err := c.FetchUser(context.Background(), 42)
    if err != nil {
        t.Fatal(err)
    }
    if u.Name != "ada" {
        t.Fatalf("want ada, got %s", u.Name)
    }
}
```

</template>

<template #with>

```go
func TestFetchUser(t *testing.T) {
    rt := httpmock.NewRouter(t).
        GET("/users/42").
        WithHeader("Authorization", httpmock.Any).
        RespondJSON(200, User{ID: 42, Name: "ada"}).
        Once()

    c := &Client{BaseURL: "http://api", Token: "t",
        HTTP: &http.Client{Transport: rt}}

    u, err := c.FetchUser(context.Background(), 42)
    assert.NoError(t, err)
    assert.Equal(t, "ada", u.Name)
}
```

</template>

</CodeCompare>

### Layered configuration

<CodeCompare withTitle="With Glacier" withoutTitle="Without Glacier">

<template #without>

```go
// Flag, then env, then YAML, then default. Repeated everywhere.
func loadPort() int {
    if v := flag.Lookup("port").Value.String(); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            return n
        }
    }
    if v := os.Getenv("APP_PORT"); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            return n
        }
    }
    var cfg struct{ Port int `yaml:"port"` }
    if b, err := os.ReadFile("app.yaml"); err == nil {
        _ = yaml.Unmarshal(b, &cfg)
        if cfg.Port != 0 {
            return cfg.Port
        }
    }
    return 8080
}
```

</template>

<template #with>

```go
type Config struct {
    Port int    `glacier:"port"  default:"8080"`
    Host string `glacier:"host"  default:"localhost"`
    DB   string `glacier:"db.url" env:"DATABASE_URL"`
}

cfg, err := conf.Load[Config](ctx,
    conf.FromFlags(),
    conf.FromEnv("APP"),
    conf.FromYAML("app.yaml"),
    conf.WithDefaults(),
)
```

</template>

</CodeCompare>

## Frequently asked

<div class="glacier-faq">

**Why Go?**
Go's static types, fast compilation, and straightforward concurrency model make it a practical choice for building reliable services and tools. Glacier assumes you already chose Go and gives you the plumbing you'd otherwise write yourself.

**Why a framework at all?**
Every Go project above a certain size re-invents the same five things: flag parsing, config layering, structured logging, mock infrastructure, and signal handling. Glacier solves each once, in a way that composes, so you don't solve it again in the next project.

**What about performance?**
Hot paths target zero allocations per operation. Each package ships benchmark tests with `benchstat` regression gates in CI. The `option` and `errs` kernel packages carry no runtime overhead beyond a function call.

**Is Glacier stable?**
Not yet. URL stability and API stability are not guaranteed before v1.0.0. Use it, file issues, and expect the seams to shift before the first stable release. The spec-first process means nothing ships to `main` without a reviewed design.

**When can I depend on it?**
The v0 libraries are usable today. Commit to `go.sum` pins and expect to absorb breaking changes until v1.0.0. The changelog will be honest about what breaks and why.

**How is the Glacier SDK CLI different from the framework?**
The framework is the 15-package library suite. The Glacier SDK is a separate CLI binary, `glacier`, built entirely with the framework's own `cli` package. It scaffolds projects (`glacier init`), runs the three code generators (`glacier generate`), lints, tests with bench gating (`glacier test`), and explains every marker, exit code, and config key. The SDK is also the framework's longest-running integration test: every shipping package is exercised by at least one SDK command. See [/sdk/](/sdk/) for the full reference.

</div>
