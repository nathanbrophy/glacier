---
title: Loading config
---

# Loading config

<PackagesUsedBadges :package-names="['conf', 'option', 'errs', 'log']" />

Production services need configuration that comes from multiple sources at once: defaults baked into the binary, a JSON file on disk, environment variables set by the operator, and command-line flags for one-off overrides. [`conf`](/docs/packages/conf) layers all of those into a single atomic commit, delivers a snapshot accessor to every interested package, and makes concurrent reads safe without any locking at the call site.

## Walkthrough

### Step 1 - Register a configuration section

Each package declares its own configuration struct once, at package init time. `conf.Register[T]` returns a `func() *T` - call it anywhere to get the latest committed snapshot.

```go
package server

import "github.com/nathanbrophy/glacier/conf"

type Config struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

// cfg is the snapshot accessor returned by Register.
var cfg = conf.Register[Config]("server", Config{
    Host: "localhost",
    Port: 8080,
})

// Cfg returns the most recently loaded server configuration snapshot.
// The returned pointer is immutable; do not modify it.
func Cfg() *Config { return cfg() }
```

The defaults supplied to `Register` are the fallback values used when a source does not provide a key. They are never nil.

### Step 2 - Load configuration at startup

Call `conf.Load` once in `main` (or in your CLI command's `Run` method). All registered sections across all imported packages are populated in a single staged-and-replace commit. Torn reads are impossible: a concurrent reader always sees either the full pre-load or the full post-load state.

```go
package main

import (
    "context"
    "log"

    "github.com/nathanbrophy/glacier/conf"

    _ "myapp/db"     // registers "db" section
    _ "myapp/server" // registers "server" section
)

func main() {
    if err := conf.Load(
        context.Background(),
        conf.WithFile("config.json"),
        conf.WithEnvPrefix("APP"),
    ); err != nil {
        log.Fatal(err)
    }
    // server.Cfg() and db.Cfg() now return populated snapshots.
}
```

### Step 3 - Override with environment variables

With prefix `APP`, the field `server.port` is overridden by the environment variable `APP__SERVER__PORT`. The double-underscore separates the prefix, section, and field name.

```sh
APP__SERVER__PORT=9090
APP__DB__MAX_CONNS=50
```

Environment variables override JSON file values; command-line flags override environment variables. The precedence order is fixed:

```
defaults → JSON file → env vars → flags → explicit Set overrides
```

### Step 4 - Add a flag source

To let command-line flags override env vars, pass `conf.WithFlagSource` pointing at the parsed flag set. With the `cli` package, `cli.FlagSource()` returns the right value automatically.

```go
if err := conf.Load(ctx,
    conf.WithFile(s.ConfigPath),
    conf.WithEnvPrefix("APP"),
    conf.WithFlagSource(cli.FlagSource()),
); err != nil {
    return errs.Wrap(err, "serve: load config")
}
```

`errs.Wrap` prepends `"serve: load config: "` to the message while keeping the original error available for `errors.Is` / `errors.As` traversal.

### Step 5 - Log the loaded state

After `conf.Load`, log the active configuration at `Info` or `Notice` level so operators can confirm what the service picked up.

```go
import (
    "log/slog"
    "github.com/nathanbrophy/glacier/log"
)

snap := server.Cfg()
ctx = log.With(ctx, slog.String("host", snap.Host), slog.Int("port", snap.Port))
slog.InfoContext(ctx, "configuration loaded")
```

## Putting it together

```go
package main

import (
    "context"
    "log/slog"

    "github.com/nathanbrophy/glacier/cli"
    "github.com/nathanbrophy/glacier/conf"
    "github.com/nathanbrophy/glacier/errs"
    "github.com/nathanbrophy/glacier/log"

    _ "myapp/db"
    _ "myapp/server"
)

// RunCmd is the root CLI command.
//
// +glacier:command name=run
// +glacier:root
type RunCmd struct {
    // ConfigFile is the path to the JSON config file.
    //
    // +glacier:default "config.json"
    // +glacier:env APP_CONFIG
    ConfigFile string
}

func (r *RunCmd) Run(ctx context.Context) error {
    if err := conf.Load(ctx,
        conf.WithFile(r.ConfigFile),
        conf.WithEnvPrefix("APP"),
        conf.WithFlagSource(cli.FlagSource()),
    ); err != nil {
        return errs.Wrap(err, "run: load config")
    }

    snap := server.Cfg()
    ctx = log.With(ctx,
        slog.String("host", snap.Host),
        slog.Int("port", snap.Port),
    )
    slog.InfoContext(ctx, "configuration loaded")

    return startServer(ctx)
}

func main() { cli.Default.Main() }
```

## What's happening underneath

- <TierBadge tier="leaf" /> [`conf`](/docs/packages/conf): manages the registry of `atomic.Pointer[T]` per section; `Load` builds all new structs and swaps every pointer in one pass.
- <TierBadge tier="kernel" /> [`option`](/docs/packages/option): `conf.WithFile`, `conf.WithEnvPrefix`, and `conf.WithFlagSource` are all functional options built on `option.Option[T]`.
- <TierBadge tier="kernel" /> [`errs`](/docs/packages/errs): wraps load errors with the `"run: load config: "` prefix so the call site is clear in log output.
- <TierBadge tier="kernel" /> [`log`](/docs/packages/log): attaches config values to the context once; every downstream log record carries them automatically.

## Related

- [Building a CLI](/docs/building-a-cli) - how a CLI command's `Run` method drives the full startup sequence.
- [Structured logging](/docs/structured-logging) - attaching context attributes that flow through the entire request lifecycle.
- [Writing tests](/docs/writing-tests) - using `fixture.NewFS` to provide a fake config file in unit tests.
