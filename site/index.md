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

14 packages in three tiers (kernel, mid, leaves). They interlock by design, not by accident. Every package speaks the same functional-options protocol, the same error register, and the same context-propagation rules. Pick one package or all 14 - the seams don't leak.

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

## The 14-package suite

<PackageGrid />

## With Glacier vs without

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
The framework is the 14-package library suite. That's the primary artifact. The Glacier SDK is a separate CLI binary built entirely with the framework's own `cli` package, demonstrating every feature in a real program. It ships after the core packages stabilize. See [/sdk](/sdk) for details.

</div>
