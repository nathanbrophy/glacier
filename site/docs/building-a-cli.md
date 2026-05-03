---
title: Building a CLI
---

# Building a CLI

<PackagesUsedBadges :package-names="['cli', 'option', 'errs', 'log', 'term', 'conf']" />

You need a production-ready CLI binary: flag parsing, env-var binding, signal handling, help text, and a branded banner on startup. The [`cli`](/docs/packages/cli) package handles all of it from a plain Go struct. You write the handler; `glaciergen` generates the wiring.

## Walkthrough

### Step 1 - Write a command struct

Define a struct with a `Run(ctx context.Context) error` method. Annotate fields with `+glacier:*` doc-comment markers to declare flags, defaults, and environment bindings.

```go
package main

import "context"

// ServeCmd starts the HTTP server.
//
// +glacier:command name=serve
// +glacier:root
type ServeCmd struct {
    // Port is the TCP port to listen on.
    //
    // +glacier:default 8080
    // +glacier:short p
    // +glacier:env GLACIER_PORT
    Port int

    // Host is the interface address to bind.
    //
    // +glacier:default "0.0.0.0"
    Host string

    // Verbose enables debug logging.
    //
    // +glacier:short v
    Verbose bool

    // Config is the path to the JSON config file.
    //
    // +glacier:required
    // +glacier:env GLACIER_CONFIG
    Config string
}

func (s *ServeCmd) Run(ctx context.Context) error {
    // your server logic here
    return nil
}
```

The struct is your entire surface. Fields without markers are invisible to the codegen and available for dependency injection.

### Step 2 - Run glaciergen

Run the code generator. It discovers every type in your module that satisfies `cli.Command` (i.e., has `Run(ctx context.Context) error`), builds the command tree from the markers, and emits `zz_generated_cli.go`.

```sh
go run github.com/nathanbrophy/glacier/cmd/glaciergen ./...
```

Or add it to a `//go:generate` directive in your package. The generated file registers each command via `cli.Default.Register`. Running `glaciergen --check` in CI detects drift between source and generated file without overwriting.

### Step 3 - Write main

Your `main.go` is three lines. `cli.Default` was populated by `init()` in the generated file.

```go
package main

func main() {
    cli.Default.Main()
}
```

`Main` dispatches `os.Args[1:]`, formats errors in the CLI register (capitalized, actionable), calls `os.Exit` with the right code, and handles `SIGINT`/`SIGTERM` gracefully.

### Step 4 - Embed and render the wordmark

Load configuration at startup and render the branded banner via `//go:embed`. The banner is the canonical bytes from `assets/logo/wordmark.txt`; the CLI module applies the ice gradient at render time when the output is a TTY.

```go
package main

import (
    _ "embed"

    "github.com/nathanbrophy/glacier/cli"
    "github.com/nathanbrophy/glacier/conf"
)

//go:embed assets/logo/wordmark.txt
var wordmark string

func (s *ServeCmd) Run(ctx context.Context) error {
    if err := conf.Load(ctx, conf.WithFile(s.Config), conf.WithEnvPrefix("APP")); err != nil {
        return err
    }
    cli.Default.Banner(wordmark) // renders gradient on TTY; plain on dumb terminal
    return serve(ctx)
}
```

### Step 5 - Layer configuration

Use [`conf`](/docs/packages/conf) inside `Run` to load the full config before any work begins. Registered packages receive their snapshot accessor immediately after `conf.Load` returns.

```go
if err := conf.Load(ctx,
    conf.WithFile(s.Config),
    conf.WithEnvPrefix("APP"),
    conf.WithFlagSource(cli.FlagSource()),
); err != nil {
    return errs.Wrap(err, "serve: load config")
}
```

### Step 6 - Structured logging

Use [`log`](/docs/packages/log) to attach the command name and port to the context before handing off to business logic. Every log call downstream picks up these attributes automatically.

```go
import (
    "log/slog"
    "github.com/nathanbrophy/glacier/log"
)

ctx = log.With(ctx,
    slog.String("cmd", "serve"),
    slog.Int("port", s.Port),
)
slog.InfoContext(ctx, "starting server")
```

## Putting it together

```go
package main

import (
    _ "embed"
    "context"
    "log/slog"

    "github.com/nathanbrophy/glacier/cli"
    "github.com/nathanbrophy/glacier/conf"
    "github.com/nathanbrophy/glacier/errs"
    "github.com/nathanbrophy/glacier/log"
)

//go:embed assets/logo/wordmark.txt
var wordmark string

// ServeCmd starts the HTTP server.
//
// +glacier:command name=serve
// +glacier:root
type ServeCmd struct {
    // Port is the TCP port to listen on.
    //
    // +glacier:default 8080
    // +glacier:short p
    // +glacier:env GLACIER_PORT
    Port int

    // Config is the path to the JSON config file.
    //
    // +glacier:required
    // +glacier:env GLACIER_CONFIG
    Config string
}

func (s *ServeCmd) Run(ctx context.Context) error {
    cli.Default.Banner(wordmark)

    if err := conf.Load(ctx,
        conf.WithFile(s.Config),
        conf.WithEnvPrefix("APP"),
    ); err != nil {
        return errs.Wrap(err, "serve: load config")
    }

    ctx = log.With(ctx,
        slog.String("cmd", "serve"),
        slog.Int("port", s.Port),
    )
    slog.InfoContext(ctx, "server starting")
    return listenAndServe(ctx, s.Port)
}

func main() {
    cli.Default.Main()
}
```

## What's happening underneath

- <TierBadge tier="leaf" /> [`cli`](/docs/packages/cli): dispatches `os.Args`, routes to the matched command struct, handles signal installation and exit codes.
- <TierBadge tier="kernel" /> [`option`](/docs/packages/option): every `cli.With*` constructor is a functional option; the pattern is consistent across the framework.
- <TierBadge tier="kernel" /> [`errs`](/docs/packages/errs): `errs.Wrap` attaches the `"serve: load config: "` prefix while preserving the original error for `errors.Is`/`As`.
- <TierBadge tier="kernel" /> [`log`](/docs/packages/log): context-attribute attachment means downstream helpers log with the command context without any coupling.

## Related

- [Loading config](/docs/loading-config) - layering defaults, env vars, and flags in detail.
- [Structured logging](/docs/structured-logging) - `log.With`, `Trace`/`Notice` levels, and `Redact`.
- [Observability](/docs/observability) - adding `cli.WithMetrics()` to instrument CLI commands.
