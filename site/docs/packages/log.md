---
title: log
---

# log

<TierBadge tier="kernel" />

<UsedInTasksBadges package-name="log" />

[View source spec &rarr;](https://github.com/nathanbrophy/glacier/blob/main/specs/0005-log.md)

## Public summary
<!-- magpie:extract source=specs/0005-log.md section=public-summary source-checksum=PENDING -->

`glacier/log` is a thin, opinionated layer over Go's `log/slog`. It adds two extra levels (Trace and Notice), glacier-attribute ordering in text and JSON output, TTY-aware color from the Glacier palette, context-based attribute attachment, explicit logger injection and retrieval, and a `Redact` helper for marking secrets. Every Glacier package emits structured logs through this surface so the framework's diagnostic output is consistent wherever it runs, from a developer's terminal to a JSON log aggregator in production.

<!-- /magpie:extract -->

## Mental model
<!-- magpie:extract source=specs/0005-log.md section=mental-model source-checksum=PENDING -->

**"ctx carries attrs, never handlers."**

The central design decision: context values carry log *attributes* (key-value pairs), not log *handlers*. When you call `log.With(ctx, slog.String("request_id", id))`, the request ID follows the context into every function that logs with it, without those functions knowing or caring that the attribute exists. The handler is configured once, at program start, not threaded through every call.

The separation maps cleanly onto two use cases:

- **Handler injection** (`Inject` / `From`): middleware or test setup that wants to swap the handler entirely calls `log.Inject(ctx, l)`. Code that wants the current handler calls `log.From(ctx)`.
- **Attribute attachment** (`With`): request handlers, jobs, and background tasks annotate their context with identifying attributes once; every subsequent log call in that context carries them automatically.

```
                  +----------------------------------------------+
                  |                  Program start               |
                  |  log.SetDefault(slog.New(log.NewHandler(...)))|
                  +--------------------+--------------------------+
                                       |
                    +-----------------\/-----------------+
                    |          HTTP middleware            |
                    |  ctx = log.Inject(ctx, reqLogger)  |
                    |  ctx = log.With(ctx, requestID)    |
                    +------------------+-----------------+
                                       |
                    +-----------------\/-----------------+
                    |         Business logic              |
                    |  l := log.From(ctx)                |
                    |  l.Info("handled", ...)            |
                    |  // record: requestID auto-appended|
                    +------------------------------------+
```

**Six levels.** Trace and Notice fill gaps in the standard four-level set. Use Trace for very-verbose iteration tracing you want stripped in production; use Notice for important non-warning events (config reloads, connection established) that stand out above Info but don't warrant a warning. Standard handlers render them as `DEBUG-4` and `INFO+2` respectively; Glacier's handlers render `TRACE` and `NOTICE` by name.

**Redact.** Marking a sensitive value is one explicit call: `log.Redact(apiKey)`. The wrapped value always formats as `[REDACTED]` regardless of which handler is in use, because the redaction is implemented via the stdlib `slog.LogValuer` contract.

<!-- /magpie:extract -->

## API
<!-- magpie:extract source=specs/0005-log.md section=api source-checksum=PENDING -->

### Level constants

```go
// LevelTrace is below Debug -- for very-verbose tracing stripped in production.
// Glacier handlers render this as "TRACE". Stdlib handlers render it as "DEBUG-4".
const LevelTrace slog.Level = -8

// LevelDebug mirrors slog.LevelDebug (-4). Provided for symmetry.
const LevelDebug slog.Level = slog.LevelDebug

// LevelInfo mirrors slog.LevelInfo (0). Provided for symmetry.
const LevelInfo slog.Level = slog.LevelInfo

// LevelNotice is between Info and Warn -- for important non-warning events
// (config reloaded, connection established). Rendered as "NOTICE"; stdlib
// handlers render it as "INFO+2".
const LevelNotice slog.Level = 2

// LevelWarn mirrors slog.LevelWarn (4). Provided for symmetry.
const LevelWarn slog.Level = slog.LevelWarn

// LevelError mirrors slog.LevelError (8). Provided for symmetry.
const LevelError slog.Level = slog.LevelError
```

### Logger access

```go
// Default returns slog.Default(). Provided for symmetry with From.
func Default() *slog.Logger

// SetDefault sets slog's global default logger.
//
//   log.SetDefault(slog.New(log.NewHandler(os.Stderr, log.WithLevel(log.LevelInfo))))
//
// Preconditions: l must not be nil.
// Concurrency: goroutine-safe (delegates to stdlib).
func SetDefault(l *slog.Logger)

// From returns the logger associated with ctx via Inject, or slog.Default()
// if none has been injected. Never returns nil.
//
//   l := log.From(ctx)
//   l.Info("handler started")
func From(ctx context.Context) *slog.Logger

// Inject returns a new context carrying l, retrievable via From. Used by
// middleware that wants to scope a logger to a request lifetime.
//
//   ctx = log.Inject(ctx, log.From(ctx).With("request_id", id))
func Inject(ctx context.Context, l *slog.Logger) context.Context
```

### Context attribute attachment

```go
// With returns a new context that carries the supplied attrs in addition to
// any already attached. When code logs through a Glacier handler using this
// ctx, the carried attrs are appended to the record automatically.
//
//   ctx = log.With(ctx, slog.String("request_id", id))
//   ctx = log.With(ctx, slog.String("user_id", uid))
//   slog.InfoContext(ctx, "handled")
//   // record contains both request_id and user_id
//
// Note: ctx-attached attrs are appended by Glacier handlers only.
// Stdlib handlers do not inspect the context for attrs.
func With(ctx context.Context, attrs ...slog.Attr) context.Context
```

### Handler construction

```go
// ColorMode controls whether the text handler emits ANSI color escapes.
type ColorMode int

const (
    // ColorAuto enables color when w is a TTY and neither GLACIER_NO_COLOR
    // nor NO_COLOR is set. This is the default for NewHandler.
    ColorAuto ColorMode = iota
    // ColorAlways forces color on regardless of TTY status.
    // Still suppressed when GLACIER_NO_COLOR=1 is set.
    ColorAlways
    // ColorNever forces color off regardless of TTY status or env vars.
    ColorNever
)

// NewHandler returns Glacier's text slog.Handler.
//
// Canonical attribute order: level, msg, package, op, error, then user attrs.
// Color palette (24-bit ANSI, pre-computed at construction):
//   TRACE/DEBUG: text-muted #8B949E
//   INFO:        cyan       #22D3EE
//   NOTICE:      teal       #2DD4BF
//   WARN:        warning    #FBBF24
//   ERROR:       error      #F87171
//
//   h := log.NewHandler(os.Stderr,
//       log.WithLevel(log.LevelDebug),
//       log.WithSource(),
//   )
func NewHandler(w io.Writer, opts ...option.Option[handlerConfig]) slog.Handler

// NewJSONHandler returns Glacier's JSON slog.Handler. Attribute ordering
// matches NewHandler; no color is emitted. WithLevel and WithSource are honored.
//
//   log.SetDefault(slog.New(log.NewJSONHandler(os.Stderr,
//       log.WithLevel(log.LevelInfo),
//       log.WithSource(),
//   )))
func NewJSONHandler(w io.Writer, opts ...option.Option[handlerConfig]) slog.Handler

// WithLevel sets the handler's minimum log level. Default: slog.LevelInfo.
func WithLevel(l slog.Leveler) option.Option[handlerConfig]

// WithSource enables source-location attribution on every record.
// Off by default; adds approximately 30% latency per log call.
func WithSource() option.Option[handlerConfig]

// WithColor overrides the text handler's color mode. Ignored by NewJSONHandler.
func WithColor(m ColorMode) option.Option[handlerConfig]
```

### Redaction

```go
// Redact wraps v in a slog.LogValuer that always renders as "[REDACTED]"
// regardless of formatter (text, JSON, or any third-party slog handler).
//
//   slog.Info("login",
//       slog.String("user", username),
//       slog.Any("password", log.Redact(password)),
//   )
//   // -> level=INFO msg=login user=ada password=[REDACTED]
//
// Redact(nil) renders as [REDACTED].
func Redact(v any) slog.LogValuer
```

<!-- /magpie:extract -->

## Examples
<!-- magpie:extract source=specs/0005-log.md section=examples source-checksum=PENDING -->

### Default setup at program start

```go
package main

import (
	"log/slog"
	"os"

	"github.com/nathanbrophy/glacier/log"
)

func main() {
	log.SetDefault(slog.New(log.NewHandler(os.Stderr,
		log.WithLevel(log.LevelInfo),
	)))

	slog.Info("starting", "version", "v0.1")
	// Output (to stderr): level=INFO msg=starting version=v0.1
}
```

### Context attribute attachment (request scoping)

```go
package main

import (
	"context"
	"log/slog"

	"github.com/nathanbrophy/glacier/log"
)

func handle(ctx context.Context, reqID, userID string) error {
	ctx = log.With(ctx,
		slog.String("request_id", reqID),
		slog.String("user_id", userID),
	)
	slog.InfoContext(ctx, "handler started")
	// -> level=INFO msg="handler started" request_id=<reqID> user_id=<userID>

	if err := process(ctx); err != nil {
		slog.ErrorContext(ctx, "handler failed", slog.Any("error", err))
		return err
	}
	return nil
}
```

### Redaction (explicit secret marking)

```go
package main

import (
	"log/slog"

	"github.com/nathanbrophy/glacier/log"
)

func ExampleRedact() {
	slog.Info("auth",
		slog.String("user", "ada"),
		slog.Any("api_key", log.Redact("sk-secret-1234")),
	)
	// Output:
	// level=INFO msg=auth user=ada api_key=[REDACTED]
}
```

### JSON output for production

```go
package main

import (
	"log/slog"
	"os"

	"github.com/nathanbrophy/glacier/log"
)

func main() {
	log.SetDefault(slog.New(log.NewJSONHandler(os.Stderr,
		log.WithLevel(log.LevelInfo),
		log.WithSource(),
	)))

	slog.Info("server started", "port", 8080)
	// {"level":"INFO","source":{"function":"main.main","file":"main.go","line":15},
	//  "msg":"server started","port":8080}
}
```

<!-- /magpie:extract -->

## FAQ
<!-- magpie:extract source=specs/0005-log.md section=faq source-checksum=PENDING -->

<div class="glacier-faq">

**Why does `log.With` only work with Glacier handlers? My existing `slog.NewTextHandler` doesn't pick up the attrs.**

Context-attribute attachment (`log.With`) is a Glacier-handler feature. The stdlib `slog.NewTextHandler` does not inspect the context for extra attrs. To get ctx-attr injection, use `log.NewHandler` or `log.NewJSONHandler`.

**What is the difference between `Inject`/`From` and `With`?**

`Inject` and `From` carry a whole `*slog.Logger` (which bundles a handler and pre-set attrs). `With` carries only `slog.Attr` key-value pairs. Use `Inject` when middleware wants to swap the handler entirely. Use `With` when you want to annotate all logs in a context with identifying attributes without changing the handler.

**Does `log.Redact` prevent the value from appearing anywhere?**

`Redact` replaces the value in the `slog` record before it reaches any handler, via the `slog.LogValuer` contract. Any handler that correctly resolves `LogValuer` will see `[REDACTED]`. `Redact` is a call-site discipline tool, not a global filter.

**Why are color escape sequences pre-computed at construction time?**

Logging is a hot path. Calling `fmt.Sprintf` for every log record would add an allocation and a format operation on the critical path. Each level's ANSI escape sequence is computed once when the handler is constructed and written verbatim by `Handle`. The `BenchmarkColorEscape` benchmark verifies that escape emission costs zero allocations.

**Can I use the six level constants with a non-Glacier handler?**

Yes. The constants are `slog.Level` values and work with any `slog.Handler`. The only difference is rendering: non-Glacier handlers will format `LevelTrace` and `LevelNotice` as `DEBUG-4` and `INFO+2` respectively. For canonical rendering ("TRACE", "NOTICE"), use `log.NewHandler` or `log.NewJSONHandler`.

</div>

<!-- /magpie:extract -->
