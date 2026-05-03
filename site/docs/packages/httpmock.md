---
title: httpmock
---

# httpmock

<TierBadge tier="leaf" />

<UsedInTasksBadges package-name="httpmock" />

[View source spec &rarr;](https://github.com/nathanbrophy/glacier/blob/main/specs/0013-httpmock.md)

## Public summary
<!-- magpie:extract source=specs/0013-httpmock.md section=public-summary source-checksum=PENDING -->

`httpmock` is a programmable `http.RoundTripper` for testing Go code that makes HTTP calls. Plug it into any `*http.Client` in your tests, declare what responses each request should receive via a fluent stub builder, and run your tests with zero real network calls. Stubs are typed (generic `JSON[T]` marshals your structs at compile time), sequenceable (serve different responses for successive calls to the same endpoint), and matchable on method, path, query parameters, headers, and body content. The transport is strict by default: any request that does not match a registered stub is an immediate test failure, surfacing gaps in your mock setup loudly. Fixtures can be loaded from JSON files in `testdata/httpmock/` for scenario-driven test suites. The package ships zero dependencies beyond the Glacier kernel.

<!-- /magpie:extract -->

## Mental model
<!-- magpie:extract source=specs/0013-httpmock.md section=mental-model source-checksum=PENDING -->

`httpmock.Transport` replaces the real network. Once installed as a client's transport, every outgoing HTTP request is intercepted and matched against a list of registered `Stub` entries. A stub is a chain of matchers (method, path, header, body, ...) plus a `Responder` that produces the synthetic response. If the request matches, the response is returned in-process; the TCP stack is never touched.

```
Test code                 http.Client                httpmock.Transport
───────────               ──────────                 ──────────────────
NewWithT(t) ─────────────────────────────────────── *Transport
                                                         │
OnRequest().Method("GET").Path("/users/42")              │ registered stubs
    .Respond(JSON(200, user)) ───────────────────────── stub[0]
                                                         │
client.Get("https://api.example.com/users/42") ─────── RoundTrip(req)
                                                           │ match stub[0]
                                                           │ return scripted *http.Response
                                                         ←─┘
```

Key invariants:

- **Never makes a real network call.** The package imports neither `net/http.DefaultTransport` nor `net.Dial`. This is auditable and verified by `TestNoNetworkImports`.
- **Strict by default.** An unmatched request returns `ErrNoRouteMatch` immediately. Lenient mode (`LenientMode()`) is explicit opt-in.
- **First registered wins.** When multiple stubs match, the earliest-registered stub is used.
- **Single critical section.** Match, increment hit-count, and respond happen atomically under a single mutex, eliminating the `Times(1)` race.
- **`Close()` is idempotent.** A closed transport drops all state and returns `ErrNoRouteMatch` for any subsequent `RoundTrip`.

<!-- /magpie:extract -->

## API
<!-- magpie:extract source=specs/0013-httpmock.md section=api source-checksum=PENDING -->

### Transport constructors

```go
// New constructs a fresh Transport with no registered stubs.
// Strict mode is the default: unmatched requests return ErrNoRouteMatch.
// Concurrency: goroutine-safe after construction.
func New(opts ...option.Option[transportConfig]) *Transport

// NewWithT constructs a Transport and registers Transport.Verify at t.Cleanup.
// Use this in table-driven and per-test setups where automatic expectation
// checking is desired.
// Preconditions: t must be non-nil.
func NewWithT(t assert.TB, opts ...option.Option[transportConfig]) *Transport
```

### Transport methods

```go
// RoundTrip implements http.RoundTripper. Matches req against registered stubs
// in registration order and returns the first matching stub's response.
// The request is always recorded regardless of match outcome.
// Match, hit-count increment, and response production happen in a single
// critical section ensuring Times(1) semantics are race-free.
//
// Error contract:
//   - ErrNoRouteMatch: strict mode and no stub matched.
//   - ScriptError: a matched stub has no Responder configured.
//   - Any Responder error is forwarded as-is.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error)

// OnRequest returns a fresh Stub builder for the next request to match.
// Stubs are matched in registration order.
func (t *Transport) OnRequest() *Stub

// RequestsTo returns all recorded requests whose URL path matches the given
// pattern. Exact match, or simple glob if pattern contains '*'.
func (t *Transport) RequestsTo(pattern string) []*http.Request

// AllRequests returns every recorded request in arrival order.
// The returned slice is a copy; callers may modify it freely.
func (t *Transport) AllRequests() []*http.Request

// Verify checks that all stubs with Times / AtLeast / AtMost / Never
// expectations have been met. Called automatically at t.Cleanup when
// constructed via NewWithT.
func (t *Transport) Verify()

// Close marks the transport as closed and releases held resources.
// Subsequent RoundTrip calls return ErrNoRouteMatch immediately.
// Idempotent: second call returns nil.
func (t *Transport) Close() error
```

### Transport options

```go
func StrictDefault() option.Option[transportConfig]      // strict mode (default)
func LenientMode() option.Option[transportConfig]        // return empty 404 for unmatched
func WithDefaultStatus(status int) option.Option[transportConfig] // custom status for unmatched
func WithLogger(l *slog.Logger) option.Option[transportConfig]    // stub-match trace events
```

### Stub builder (chained)

`OnRequest()` returns a `*Stub`. Chain matcher methods, then finalize with `Respond`. All methods return `*Stub` for chaining.

```go
func (s *Stub) Method(m string) *Stub       // restrict to HTTP method (case-insensitive)
func (s *Stub) Path(p string) *Stub         // exact URL path match
func (s *Stub) PathPrefix(p string) *Stub   // URL path prefix match
func (s *Stub) Regex() *Stub                // reinterpret Path as a regexp (call after Path)
func (s *Stub) Query(name, value string) *Stub  // require query parameter (AND with others)
func (s *Stub) Header(name, value string) *Stub // require header (AND with others; canonical name)
func (s *Stub) Body(matcher BodyMatcher) *Stub  // require body to satisfy matcher

func (s *Stub) Times(n int) *Stub    // expect exactly n calls
func (s *Stub) AtLeast(n int) *Stub  // expect at least n calls
func (s *Stub) AtMost(n int) *Stub   // expect at most n calls
func (s *Stub) AnyTimes() *Stub      // any call count (no Verify failure)
func (s *Stub) Never() *Stub         // assert stub is never matched

func (s *Stub) Respond(r Responder) *Stub  // finalize and register the stub; r must be non-nil
```

### BodyMatcher implementations

```go
func BodyExact(body []byte) BodyMatcher                               // byte-for-byte match
func BodyJSON[T any](want T, opts ...assert.EqualOption) BodyMatcher  // JSON structural match
func BodyContains(s string) BodyMatcher                               // substring match
func BodyMatchFn(f func([]byte) bool) BodyMatcher                     // custom predicate
```

### Responder implementations

```go
func JSON[T any](status int, body T) Responder         // marshal body as JSON; Content-Type set
func JSONFrom[T any](status int, r io.Reader) Responder // pre-encoded JSON from reader, validated as T
func Status(status int) Responder                       // empty body with status code
func Body(status int, body []byte, contentType string) Responder // raw bytes
func Stream(status int, body io.Reader, contentType string) Responder // streaming body; reader closed after
func Error(err error) Responder                         // transport-level error (nil response)
func Sequence(rs ...Responder) Responder                // cycle through rs on each match
func SequenceCycle(rs ...Responder) Responder            // explicit cycle mode (same as Sequence)
func SequenceExhaust(rs ...Responder) Responder          // after exhaustion, return ErrNoRouteMatch
```

### Fixture loading

```go
// LoadFixtures reads testdata/httpmock/<name>.json, parses it as a stubs document,
// and registers each stub on the transport.
//
// Security disciplines applied (Falcon §1.3 / §23.9 / §23.10):
//   - Path canonicalized; traversal attempts rejected.
//   - File size capped at 16 MiB.
//   - JSON decoded via internal/safejson: DisallowUnknownFields, depth cap 32, UTF-8 validated.
//
// On any error, t.Errorf is called and no stubs are registered.
func (t *Transport) LoadFixtures(tb assert.TB, name string) error
```

### Errors

```go
// ErrNoRouteMatch is returned by RoundTrip when strict mode is active and
// no registered stub matches the incoming request.
var ErrNoRouteMatch = errs.Sentinel("httpmock: no route match")

// ScriptError is returned by RoundTrip when a matched stub has no Responder.
type ScriptError struct {
    Step  int    // zero-based registration index of the misconfigured stub
    Cause error
}
func (e *ScriptError) Error() string
func (e *ScriptError) Unwrap() error
```

<!-- /magpie:extract -->

## Examples
<!-- magpie:extract source=specs/0013-httpmock.md section=examples source-checksum=PENDING -->

### Typed JSON response with post-test assertion

`NewWithT` registers `Verify` at `t.Cleanup`. Pair with `Times(n)` to assert exact call counts.

```go
func TestUserFetch(t *testing.T) {
    type User struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }

    rt := httpmock.NewWithT(t)
    rt.OnRequest().
        Method("GET").
        Path("/users/42").
        Times(1).
        Respond(httpmock.JSON(200, User{ID: 42, Name: "Ada"}))

    client := &http.Client{Transport: rt}
    resp, err := client.Get("https://api.example.com/users/42")
    assert.NoError(t, err)
    assert.Equal(t, resp.StatusCode, 200)

    var u User
    assert.NoError(t, json.NewDecoder(resp.Body).Decode(&u))
    assert.Equal(t, u, User{ID: 42, Name: "Ada"})

    seen := rt.RequestsTo("/users/42")
    assert.Len(t, seen, 1)
}
```

### Sequenced responses (retry scenario)

`SequenceExhaust` serves each responder in order and returns `ErrNoRouteMatch` if called again after exhaustion.

```go
func TestRetry(t *testing.T) {
    rt := httpmock.NewWithT(t)
    rt.OnRequest().
        Method("POST").
        Path("/login").
        Times(3).
        Respond(httpmock.SequenceExhaust(
            httpmock.Status(503),                                    // 1st call: server unavailable
            httpmock.Status(503),                                    // 2nd call: still unavailable
            httpmock.JSON(200, LoginResponse{Token: "tok-abc123"}),  // 3rd call: success
        ))

    // ... exercise retry client ...
}
```

### BodyJSON matcher with smart-equal options

`BodyJSON[T]` unmarshals the request body and compares it to `want` using `assert.Equal`, which supports `IgnoreFields` and similar options.

```go
func TestCreateUser(t *testing.T) {
    type CreateRequest struct {
        Name      string    `json:"name"`
        CreatedAt time.Time `json:"created_at"`
    }
    type CreateResponse struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }

    rt := httpmock.NewWithT(t)
    rt.OnRequest().
        Method("POST").
        Path("/users").
        Body(httpmock.BodyJSON(
            CreateRequest{Name: "Ada"},
            assert.IgnoreFields("CreatedAt"), // timestamp fields are non-deterministic
        )).
        Times(1).
        Respond(httpmock.JSON(201, CreateResponse{ID: 42, Name: "Ada"}))

    // ... exercise creation logic ...
}
```

### Transport-level error (simulating network failure)

`httpmock.Error` causes `RoundTrip` to return the given error as a transport-level failure.

```go
func TestClientHandlesNetworkError(t *testing.T) {
    rt := httpmock.NewWithT(t)
    rt.OnRequest().
        Path("/api/data").
        AnyTimes().
        Respond(httpmock.Error(context.DeadlineExceeded))

    client := &http.Client{Transport: rt}
    _, err := client.Get("https://api.example.com/api/data")
    assert.ErrorIs(t, err, context.DeadlineExceeded)
}
```

<!-- /magpie:extract -->

## FAQ
<!-- magpie:extract source=specs/0013-httpmock.md section=faq source-checksum=PENDING -->

<div class="glacier-faq">

**Why no recording mode?**

Recording mode requires a scrubbing pass to redact secrets from headers and bodies before anything touches disk. Getting that scrubbing right is a non-trivial security review, and shipping it half-baked would be worse than shipping nothing. v0 focuses on scripted, deterministic, zero-network responses. Recording lands in `0024-httpmock-record.md` once there is a real consumer use case and a Falcon-reviewed scrubbing design.

**How does httpmock integrate with httpc?**

Both `httpmock` and `httpc` are Tier 2 leaf packages that must not import each other. They are wired together by the test or application at the call site: `httpc.New(httpc.WithTransport(httpmock.New()))`. This keeps both packages independently usable and follows the framework's composition-at-the-leaves rule.

**Why is strict mode the default?**

An unmatched request in lenient mode silently returns a 404 or empty response. A test that relies on lenient behavior may pass even when the code under test is calling the wrong endpoint, has a typo in the path, or is making extra requests. Strict mode makes every unmatched request a loud, immediate failure with `ErrNoRouteMatch`. If you want lenient mode, you must explicitly opt in with `LenientMode()` or `WithDefaultStatus(n)`.

**What happens if I forget to call Respond on a stub?**

If `Respond` is never called, the stub's responder is nil. When `RoundTrip` matches that stub, it returns a `ScriptError`. Use `NewWithT(t)` and `Times(n)`: the `Verify` at `t.Cleanup` will also catch it if the stub was never matched at all.

**Is it safe to share one Transport across concurrent goroutines?**

Yes. All mutable state (stub list, recorded requests, closed flag) is protected by a single `sync.RWMutex`. `RoundTrip` acquires the write lock for the entire match-increment-respond critical section, so `Times(1)` semantics are race-free even under heavy concurrency. Run your tests with `-race`; the package passes cleanly.

**Can I use httpmock outside of tests?**

The package has no `//go:build` constraint limiting it to test contexts, and `New()` does not require a `*testing.T`. In practice, `httpmock` is designed for tests and there is little reason to use it in production code, but nothing prevents it.

</div>

<!-- /magpie:extract -->
