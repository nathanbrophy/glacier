---
title: httpc
---

# httpc

<TierBadge tier="leaf" />

<UsedInTasksBadges package-name="httpc" />

[View source spec &rarr;](https://github.com/nathanbrophy/glacier/blob/main/specs/0015-httpc.md)

## Public summary
<!-- magpie:extract source=specs/0015-httpc.md section=public-summary source-checksum=PENDING -->

`httpc` is Glacier's typed, retry-aware, dry-run-capable HTTP client. Write `user, resp, err := httpc.Get[User](ctx, url)` and the framework reads the response body, JSON-unmarshals it into your type, and hands it back with no boilerplate read-all-unmarshal loop. Mutating methods take closure-generated bodies so retry is safe even for large or streamed payloads: each attempt gets a fresh body from your closure, no seeking required. Retry policies compose declaratively: `MaxAttempts`, `ExponentialBackoff`, `Jittered`, `RetryOn`, `RetryIf`. A CLI's `--dry-run` flag propagates through `context.Context` so every `httpc` call inside becomes plan-only without a single conditional at the call site. The package wraps stdlib `net/http`, carries no third-party dependencies, and composes directly with `httpmock` for hermetic, network-free tests.

<!-- /magpie:extract -->

## Mental model
<!-- magpie:extract source=specs/0015-httpc.md section=mental-model source-checksum=PENDING -->

Three ideas hold `httpc` together.

**Typed methods auto-unmarshal.** `Get[T]`, `Post[T]`, `Put[T]`, `Patch[T]`, and `Delete[T]` are generic functions. The type parameter `T` names the Go type you want back. The framework reads and decodes the response body for you. When `T` is `[]byte`, the raw body is returned unchanged. When `T` is anything else, the body is decoded via `internal/safejson` (depth-capped, size-limited, UTF-8-validating). You get back the decoded value alongside a `*Response` wrapper that exposes the original `*http.Response`, the body bytes, and the elapsed duration.

**Closure-generated bodies make retry correct.** Retry requires re-sending the same body on every attempt. HTTP bodies are `io.Reader` (one-shot, not rewindable). `httpc` solves this by requiring callers to provide a closure that produces the body, not the body itself. `JSONBody[T](func() T)` is called once per attempt. `MultipartBody(func(*multipart.Writer) error)` gets a fresh `multipart.Writer` per attempt. Your closure is called serially, never concurrently for the same request.

**Dry-run propagates through `context.Context`.** Attaching dry-run to a context with `httpc.WithDryRun(ctx, httpc.WithPlanSink(fn))` makes every `httpc` call inside that context skip the network and emit a structured `*RequestPlan` to your sink function instead. There is no conditional code at call sites. A CLI command's `--dry-run` flag sets the context attribute once; every downstream `httpc` call is automatically audit-only. The plan's rendered headers are scrubbed of sensitive values by default.

```
Client.Get[T](ctx, url, opts...)
        │
        ├─ dry-run? ──yes──► emit *RequestPlan to sink ──► return zero T, nil, nil (or ErrDryRun)
        │
        ├─ build *http.Request (apply base URL, headers, body closure)
        │
        ├─ retry loop ──────────────────────────────────────────────┐
        │       invoke body closure (fresh per attempt)             │
        │       RoundTrip(req) via configured transport             │
        │       check RetryOn / RetryIf                             │
        │       ctx cancelled? ──yes──► short-circuit, stop loop    │
        │       backoff sleep ◄──────────────────────────────────────┘
        │
        ├─ status non-2xx? ──► return zero T, *Response, *StatusError
        │
        ├─ decode body via internal/safejson (size cap, depth cap, UTF-8)
        │
        └─ return T, *Response, nil
```

**Relationship to `httpmock`.** `httpc` is production HTTP client code. `httpmock` is the testing transport. They are both Tier 2 leaf packages and must not import each other. Consumers wire them together at the test level: `httpc.New(httpc.WithTransport(httpmock.NewWithT(t)))`.

<!-- /magpie:extract -->

## API
<!-- magpie:extract source=specs/0015-httpc.md section=api source-checksum=PENDING -->

### Sentinel errors

```go
// ErrDryRun is returned by typed methods when WithDryRunErrors() is set and
// the context carries a dry-run attribute.
var ErrDryRun = errs.Sentinel("httpc: dry run")

// ErrMaxAttempts is returned when the retry loop exhausts its attempt budget.
var ErrMaxAttempts = errs.Sentinel("httpc: max attempts")

// ErrMaxElapsed is returned when the retry loop exceeds its overall time budget.
var ErrMaxElapsed = errs.Sentinel("httpc: max elapsed")
```

### Client

```go
// Client is a configured HTTP client. The zero value is not usable; construct via New.
// A single Client is goroutine-safe: concurrent calls share the underlying transport.
type Client struct { /* unexported */ }

// Default is the package-level shared Client, equivalent to New() with all defaults.
// Package-level functions (Get, Post, etc.) delegate to Default.
var Default = New()

// New constructs a Client from the given options. If no WithTransport option is
// provided, New uses http.DefaultTransport and owns the transport for Close purposes.
func New(opts ...option.Option[clientConfig]) *Client

// Close releases resources held by the client. If the client owns its transport,
// Close closes it. Idempotent: calling it more than once is safe and returns nil.
func (c *Client) Close() error
```

### Client options

```go
func WithTransport(rt http.RoundTripper) option.Option[clientConfig]  // custom transport (e.g. httpmock)
func WithTimeout(d time.Duration) option.Option[clientConfig]         // per-request deadline
func WithBaseURL(rawURL string) option.Option[clientConfig]           // prepended to relative URLs
func WithHeaders(h http.Header) option.Option[clientConfig]           // headers sent on every request
func WithRetry(opts ...RetryOption) option.Option[clientConfig]       // client-level default retry policy
func WithLogger(l *slog.Logger) option.Option[clientConfig]           // lifecycle event logger
```

### Typed methods

Package-level functions delegate to `Default`. `(c *Client)` receiver versions are identical in shape.

```go
// Get sends GET, decodes response body into T. When T is []byte, raw bytes returned.
// Error contract: *StatusError (non-2xx), *BodyParseError (decode failure),
// ErrMaxAttempts, ErrMaxElapsed, ErrDryRun, context.Canceled.
func Get[T any](ctx context.Context, url string, opts ...RequestOption) (T, *Response, error)

// Head sends HEAD. No body is read or returned.
func Head(ctx context.Context, url string, opts ...RequestOption) (*Response, error)

// Post sends POST. Body via opts (e.g., JSONBody). Response decoded into T.
func Post[T any](ctx context.Context, url string, opts ...RequestOption) (T, *Response, error)

// Put sends PUT. Same body and decode semantics as Post.
func Put[T any](ctx context.Context, url string, opts ...RequestOption) (T, *Response, error)

// Patch sends PATCH. Same body and decode semantics as Post.
func Patch[T any](ctx context.Context, url string, opts ...RequestOption) (T, *Response, error)

// Delete sends DELETE. Same body and decode semantics as Post.
func Delete[T any](ctx context.Context, url string, opts ...RequestOption) (T, *Response, error)

// Do sends a raw *http.Request with no auto-decode, retry, or base URL joining.
// Escape hatch for callers who construct the full request themselves.
func Do(ctx context.Context, req *http.Request) (*Response, error)
```

### Response

```go
// Response wraps *http.Response with httpc-specific metadata.
type Response struct {
    *http.Response
    Body    []byte        // bytes read by typed methods; nil for Head and Do
    Elapsed time.Duration // wall-clock time from first byte sent to last byte of body read
}

// Drain discards and closes any unread response body to release the TCP connection.
func (r *Response) Drain() error
```

### Body builders (RequestOption)

All body builders return a `RequestOption`. The enclosed closure is called once per retry attempt, serially, never concurrently for the same request.

```go
// JSONBody sets Content-Type: application/json and calls gen() per attempt.
// gen must be idempotent; it may be called multiple times during retry.
func JSONBody[T any](gen func() T) RequestOption

// MultipartBody sets multipart/form-data with a fresh boundary per attempt.
func MultipartBody(gen func(*multipart.Writer) error) RequestOption

// RawBody passes raw bytes and content type from a closure per attempt.
func RawBody(gen func() ([]byte, string, error)) RequestOption

// StreamBody provides a fresh io.ReadCloser per attempt. Previous attempt's
// ReadCloser is closed before the closure is invoked again.
func StreamBody(gen func() (io.ReadCloser, string, error)) RequestOption

// FormBody sets application/x-www-form-urlencoded. gen is called per attempt.
func FormBody(gen func() url.Values) RequestOption

// WithRequestHeaders merges h on top of client-level headers for this call only.
func WithRequestHeaders(h http.Header) RequestOption

// WithRetry attaches a per-call retry policy that merges with the client-level default.
func WithRetry(opts ...RetryOption) RequestOption

// WithMaxResponseBytes overrides the default 32 MiB response-body cap for this call.
func WithMaxResponseBytes(n int64) RequestOption

// WithUnboundedResponse removes the default response-body size cap.
// Use only when the caller has an independent bound. Falcon sign-off required at framework call sites.
func WithUnboundedResponse() RequestOption
```

### Retry options

```go
func MaxAttempts(n int) RetryOption                   // total attempts including first; default 1 (no retry)
func ExponentialBackoff(base time.Duration) RetryOption // base * 2^attempt
func LinearBackoff(d time.Duration) RetryOption        // fixed delay between attempts
func Jittered() RetryOption                            // +-25% uniform jitter; apply after backoff
func RetryOn(statuses ...int) RetryOption              // status codes that trigger retry; replaces default [500,502,503,504,429]
func RetryIf(fn func(*Response, error) bool) RetryOption // custom predicate; either RetryOn or RetryIf triggers retry
func MaxElapsed(d time.Duration) RetryOption           // overall wall-clock budget for retry loop
```

### Dry-run

```go
// WithDryRun derives a context that makes every httpc call skip the network.
// Instead, each call emits a *RequestPlan to the configured sink and returns immediately.
// Default: zero T, nil response, nil error. WithDryRunErrors changes error to ErrDryRun.
// Headers in *RequestPlan are scrubbed of sensitive values by default.
func WithDryRun(ctx context.Context, opts ...DryRunOption) context.Context

func WithPlanSink(fn func(*RequestPlan)) DryRunOption   // receive each emitted plan; replaces default slog.Debug sink
func WithDryRunErrors() DryRunOption                    // return ErrDryRun instead of nil
func WithPlanIncludeSecrets() DryRunOption              // disable header redaction (debugging only)
func IsDryRun(ctx context.Context) bool                 // report whether ctx carries dry-run

// RequestPlan is the audit record produced during dry-run mode.
// Header values are scrubbed unless WithPlanIncludeSecrets is set.
// Redacted headers include: Authorization, Cookie, Set-Cookie, X-Api-Key,
// X-Auth-Token, Proxy-Authorization, and any name matching (?i)auth|key|token|cookie|secret.
type RequestPlan struct {
    Request *http.Request  // fully prepared request; body not attached (see Body field)
    Body    []byte         // bytes the body closure would have produced; nil for HEAD/GET
    Retry   retryConfig    // copy of the effective retry policy
    Timeout time.Duration  // effective per-request timeout (0 means none)
}
```

### Error types

```go
// StatusError is returned when the server responds with a non-2xx status code
// after the retry policy (if any) is exhausted.
type StatusError struct {
    Status int     // HTTP status code
    Body   []byte  // raw response body; NOT included in Error() string
    Cause  error
}
func (e *StatusError) Error() string
func (e *StatusError) Unwrap() error

// BodyParseError is returned when the response body cannot be decoded into T.
// Named BodyParseError (not ParseError) to avoid collision with cli.FlagParseError.
type BodyParseError struct {
    Cause       error   // underlying decode error; never nil
    Body        []byte  // first 1 KiB of response body; NOT included in Error() string
    ContentType string
}
func (e *BodyParseError) Error() string
func (e *BodyParseError) Unwrap() error
```

<!-- /magpie:extract -->

## Examples
<!-- magpie:extract source=specs/0015-httpc.md section=examples source-checksum=PENDING -->

### Typed GET

`Get[User]` decodes the response body into `User` automatically. On non-2xx status, err is `*httpc.StatusError`. On decode failure, err is `*httpc.BodyParseError`.

```go
// ExampleGet demonstrates a typed GET that decodes the response body into a
// User struct. On non-2xx status, err is *httpc.StatusError. On decode
// failure, err is *httpc.BodyParseError.
func ExampleGet() {
    type User struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }

    ctx := context.Background()
    user, resp, err := httpc.Get[User](ctx, "https://api.example.com/users/42")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Drain()

    fmt.Printf("%s (status %s)\n", user.Name, resp.Status)
    // Output: Ada Lovelace (status 200 OK)
}
```

### POST with closure body and retry

The closure is called once per retry attempt, so the body is always fresh. No seeking required.

```go
// ExamplePost_JSONBody demonstrates a POST with a JSONBody closure. The
// closure is called once per retry attempt, so the body is always fresh.
func ExamplePost_JSONBody() {
    type NewUser struct {
        Name string `json:"name"`
        Age  int    `json:"age"`
    }
    type User struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }

    ctx := context.Background()
    created, _, err := httpc.Post[User](ctx, "https://api.example.com/users",
        httpc.JSONBody(func() NewUser {
            return NewUser{Name: "Ada", Age: 36}
        }),
        httpc.WithRetry(
            httpc.MaxAttempts(3),
            httpc.ExponentialBackoff(100*time.Millisecond),
            httpc.Jittered(),
            httpc.RetryOn(500, 502, 503),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("created user %d\n", created.ID)
}
```

### CLI dry-run flag propagation

Set dry-run on the context once; every `httpc` call inside respects it with no conditional code at call sites.

```go
// ExampleWithDryRun demonstrates how a CLI command's --dry-run flag propagates
// through context to all httpc calls. No conditional code is needed at call sites.
func ExampleWithDryRun() {
    type DeployCmd struct {
        DryRun bool `json:"dry_run"` // +glacier:default false
    }

    func (d *DeployCmd) Run(ctx context.Context) error {
        if d.DryRun {
            var plans []*httpc.RequestPlan
            ctx = httpc.WithDryRun(ctx, httpc.WithPlanSink(func(p *httpc.RequestPlan) {
                plans = append(plans, p)
            }))
            defer func() {
                for _, p := range plans {
                    fmt.Printf("[dry-run] would %s %s\n", p.Request.Method, p.Request.URL)
                }
            }()
        }
        return runDeployPipeline(ctx) // every httpc call inside respects dry-run
    }
}
```

### Composition with httpmock for tests

`httpc` and `httpmock` are both leaves; wire them together at the test level, never at the package level.

```go
// ExampleNew_withHttpmock demonstrates wiring httpmock as the transport for
// hermetic, network-free testing.
func ExampleNew_withHttpmock() {
    // In a real test: t *testing.T
    rt := httpmock.NewWithT(t)
    rt.OnRequest().Method("GET").Path("/users/42").
        Respond(httpmock.JSON(200, User{ID: 42, Name: "Ada"}))

    client := httpc.New(httpc.WithTransport(rt))

    user, _, err := client.Get[User](ctx, "https://api.example.com/users/42")
    assert.NoError(t, err)
    assert.Equal(t, "Ada", user.Name)
}
```

<!-- /magpie:extract -->

## FAQ
<!-- magpie:extract source=specs/0015-httpc.md section=faq source-checksum=PENDING -->

<div class="glacier-faq">

**Why is there a default 32 MiB response-body cap?**

Unbounded response body reads are a class of denial-of-service: a malicious or misconfigured server can stream gigabytes of data until the process runs out of memory. The default cap of 32 MiB covers the vast majority of REST API responses. Callers that need larger responses opt out with `WithUnboundedResponse()`. The cap is implemented as an `io.LimitReader` before any JSON parsing. gzip responses are capped both before and after decompression to prevent zip-bomb attacks.

**How does dry-run propagate? Do I need to change my handler code?**

No. `httpc.WithDryRun` attaches a flag to the `context.Context` value. Every `httpc` call that receives a context carrying that flag checks it at the start of the request dispatch path and emits a `*RequestPlan` to the configured sink instead of making a network call. Your handler code never needs an `if dryRun { ... }` conditional.

**Why closure-generated bodies instead of accepting `io.Reader` directly?**

`io.Reader` is one-shot. If an HTTP request fails and the client retries, the original reader is exhausted and cannot be replayed. By requiring a closure that returns a fresh body on each call, `httpc` guarantees that every retry attempt gets the same body. The closure also lets callers use stateful body construction (generating a new multipart boundary, computing a fresh HMAC) that would be impossible to replay from a fixed reader.

**How do I test code that uses httpc?**

Inject `httpmock.NewWithT(t)` as the transport: `client := httpc.New(httpc.WithTransport(httpmock.NewWithT(t)))`. Register stubs on the transport for each expected request. Use `httpmock.JSON[T](statusCode, value)` for typed JSON responses. The mock transport fails the test immediately on any unexpected request, so gaps in your stub setup surface loudly.

**What happens when my body closure returns an error?**

The request is never sent. `httpc` calls the body closure as the first step of request construction. If the closure returns an error, that error is returned directly to the caller and no `RoundTrip` is attempted. The retry loop does not start: a closure failure is treated as a construction error, not a transient failure.

**Can I use a package-level function like `httpc.Get[T]` and also have a test-configured client?**

Yes, but they are separate. `httpc.Get[T](ctx, url)` delegates to `httpc.Default`. In tests you can replace `Default`: `httpc.Default = httpc.New(httpc.WithTransport(rt))` and restore it after the test. For cleaner isolation, prefer constructing an explicit `*Client` and threading it through your code; the explicit client approach avoids mutating shared package state.

**Why is the error type named `BodyParseError` instead of `ParseError`?**

`cli.FlagParseError` and `httpc.BodyParseError` were both originally named `ParseError`, causing an ambiguity when consumers import both packages and call `errors.As`. The renaming gives each type a distinct, self-documenting name.

</div>

<!-- /magpie:extract -->
