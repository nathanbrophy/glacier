---
id: 0013
title: HTTPMock
slug: httpmock
status: verified
owner-agent: otter
created: 2026-05-01
last-updated: 2026-05-02
supersedes: []
superseded-by: null
reviewers:
  - { agent: otter,  required: true,  signed-off-at: "2026-05-01T00:00:00Z" }
  - { agent: lynx,   required: true,  signed-off-at: "2026-05-01T00:00:00Z" }
  - { agent: falcon, required: true,  signed-off-at: "2026-05-01T00:00:00Z" }
  - { agent: magpie, required: false, signed-off-at: null }
implementing-commits: [6657420]
verified-at: "2026-05-02T00:00:00Z"
docs-extract:
  - public-summary
  - mental-model
  - api
  - examples
  - faq
---

# HTTPMock

<!--
  Section headers below are STABLE ANCHORS. Magpie extracts content by header,
  so do not rename or reorder them. Doing so is a process change requiring its
  own spec.

  Sections marked **Public** are extracted by Magpie for the public site.
  Sections marked **Internal** are engineering-only and never appear in published docs.
-->

## Public Summary

<!-- **Public.** One paragraph in end-user voice. The canonical description for the site and README. -->

`httpmock` is a programmable `http.RoundTripper` for testing Go code that makes HTTP calls. Plug it into any `*http.Client` in your tests, declare what responses each request should receive via a fluent stub builder, and run your tests with zero real network calls :  ever. Stubs are typed (generic `JSON[T]` marshals your structs at compile time), sequenceable (serve different responses for successive calls to the same endpoint), and matchable on method, path, query parameters, headers, and body content. The transport is strict by default: any request that does not match a registered stub is an immediate test failure, surfacing gaps in your mock setup loudly. Fixtures can be loaded from JSON files in `testdata/httpmock/` for scenario-driven test suites. The package ships zero dependencies beyond the Glacier kernel.

## Mental Model

<!-- **Public.** The conceptual frame a developer should hold while using this. Mermaid diagrams welcome. Source for the "Concepts" page on the site. -->

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

- **NEVER makes a real network call.** The package imports neither `net/http.DefaultTransport` nor `net.Dial`. This is auditable and verified by `TestNoNetworkImports`.
- **Strict by default.** An unmatched request returns `ErrNoRouteMatch` immediately. Lenient mode (`LenientMode()`) is explicit opt-in.
- **First registered wins.** When multiple stubs match, the earliest-registered stub is used.
- **Single critical section.** Match, increment hit-count, and respond happen atomically under a single mutex, eliminating the `Times(1)` race.
- **`Close()` releases the stub-list lock.** A closed transport drops all state and returns `ErrNoRouteMatch` for any subsequent `RoundTrip`.

## Goals

<!-- **Internal.** Bulleted list. -->

- Provide a fully in-memory `http.RoundTripper` that never makes real network calls.
- Let tests declare per-request scripted responses via a fluent, chained `Stub` builder.
- Supply typed generic responders (`JSON[T]`, `JSONFrom[T]`) and typed generic body matchers (`BodyJSON[T]`) that eliminate marshal/unmarshal boilerplate.
- Support ordered sequences of responses (`Sequence`, `SequenceCycle`, `SequenceExhaust`) for retry-testing patterns.
- Load stubs from JSON fixture files in `testdata/httpmock/` for scenario-driven and golden-fixture workflows.
- Enforce the full Falcon §1.3 / §23.9 / §23.10 disciplines on fixture loading (path canonicalization, 16 MiB cap, `DisallowUnknownFields`, schema validation, UTF-8 validation, depth cap ≤ 32).
- Record every request for post-test assertion (`RequestsTo`, `AllRequests`).
- Auto-verify stub expectations at `t.Cleanup` when using `NewWithT`.
- Be goroutine-safe for concurrent `RoundTrip` calls under `-race`.
- Stay within the LOC budget: ≤ 800 lines of production code, ≤ 1100 lines of tests.

## Non-Goals

<!-- **Internal.** Bulleted list. What this spec deliberately excludes. -->

- **Recording mode.** httpmock is replay-only at v0. No capturing of real network traffic. Recording is deferred to `0024-httpmock-record.md` with a full Falcon scrubbing review. The scrubbing rules from Falcon §1.3 (allowlist-based header redaction, body-redaction hooks) are documented as forward-pointers in the `## Security & Supply-Chain Notes` section.
- **Real-network proxying or man-in-the-middle.** Out of scope, not a v0 goal, and architecturally incompatible with the "never dials" invariant.
- **TLS certificate management.** The transport intercepts before TLS :  the test code constructs `http.Client` with this transport, so TLS handshake never occurs.
- **gRPC / WebSocket / SSE transport mocking.** Those are streaming protocols with distinct framing; deferred to `0025-httpc-streams.md`.
- **Integration with `httpc` at the package level.** `httpc` and `httpmock` are both Tier 2 leaves; they must not import each other. Consumers wire them together at the test level: `httpc.New(httpc.WithTransport(httpmock.New()))`. No architectural coupling.
- **YAML or TOML fixture formats.** JSON only at v0, per Falcon's supply-chain ruling (D25).

## Architecture

<!-- **Internal.** Mermaid diagram + prose. Package layout, data flow, lifecycle. -->

`httpmock` is a Tier 2 leaf package. It imports from Tier 0 kernel packages only (`option`, `errs`, `log`, `assert`) plus `fluent` (Tier 1 mid, used internally for stub-list iteration). It imports no other Tier 2 package.

### File layout

```
httpmock/
├── doc.go              package-level doc comment; package declaration
├── transport.go        Transport type; New, NewWithT, RoundTrip, OnRequest,
│                       RequestsTo, AllRequests, Verify, Close
├── stub.go             Stub type; Method, Path, PathPrefix, Regex, Query,
│                       Header, Body, Times, AtLeast, AtMost, AnyTimes, Never, Respond
├── responders.go       Responder interface; JSON[T], JSONFrom[T], Status,
│                       Body, Stream, Error, Sequence, SequenceCycle, SequenceExhaust
├── body_matchers.go    BodyMatcher interface; BodyExact, BodyJSON[T],
│                       BodyContains, BodyMatchFn
├── fixtures.go         LoadFixtures; fixture JSON schema; internal/safefile + safejson
└── errors.go           ErrNoRouteMatch sentinel; ScriptError typed error
```

### Test file layout

```
httpmock/
├── transport_test.go        Transport, RoundTrip, OnRequest, recording
├── stub_test.go             Stub fluent chain
├── responders_test.go       every responder constructor
├── sequence_test.go         Sequence / SequenceCycle / SequenceExhaust
├── body_matchers_test.go    BodyExact, BodyJSON[T], BodyContains, BodyMatchFn
├── fixtures_test.go         LoadFixtures, schema validation, size cap, path traversal
├── concurrency_test.go      concurrent RoundTrip (-race), Times(1) race fix
├── lifecycle_test.go        Close idempotency, NewWithT cleanup
├── import_audit_test.go     audit: no net.Dial / no http.DefaultTransport
├── property_test.go         property-based / algebraic-identity tests
├── fuzz_test.go             FuzzFixture
├── bench_test.go            benchmarks
├── example_test.go          godoc examples
└── testdata/httpmock/       golden fixture JSONs + malformed / oversize variants
```

### Lifecycle

```
New(opts)  ─────────────────────────────────────────────────────► *Transport
                │
                ├─ OnRequest() ──► *Stub (builder chain) ──► Respond(r) ──► registered
                │
                ├─ RoundTrip(req) ──► match stubs ──► return *http.Response or error
                │                     (single mutex: match + increment + respond)
                │
                ├─ RequestsTo(path) ──► []*http.Request
                ├─ AllRequests()    ──► []*http.Request
                ├─ Verify()         ──► check Times expectations (auto at Cleanup if NewWithT)
                │
                └─ Close() error   ──► release stub-list mutex; idempotent
```

### Concurrency model

A single `Transport` is goroutine-safe. The stub list is protected by a `sync.RWMutex`:

- `OnRequest` (stub registration) acquires the write lock.
- `RoundTrip` acquires the write lock for the entire match-increment-respond critical section, ensuring `Times(1)` semantics are race-free (§23.14).
- `RequestsTo` and `AllRequests` acquire the read lock.
- `Verify` acquires the read lock.
- `Close` acquires the write lock, marks the transport closed, and returns nil on subsequent calls (idempotent, §23.16).

### Dependency posture

| Direction | Package | Why |
|---|---|---|
| Imports | `option` (Tier 0) | `option.Option[transportConfig]` constructor options |
| Imports | `errs` (Tier 0) | `errs.Sentinel` for `ErrNoRouteMatch`; `errs.Join` in Close |
| Imports | `log` (Tier 0) | injected `*slog.Logger` for stub-match trace events |
| Imports | `assert` (Tier 0) | `assert.TB` constraint for `NewWithT`; `assert.EqualOption` for `BodyJSON[T]` |
| Imports | `fluent` (Tier 1) | internal stub-list iteration in test-only helpers |
| Imports | `internal/safefile` | path-canonicalization for `LoadFixtures` (§23.10) |
| Imports | `internal/safejson` | JSON decode with size cap + depth cap + DisallowUnknownFields + UTF-8 (§23.10) |

## Schema

<!-- **Internal.** Go types with invariants stated as `// invariant: ...` comments on each field. -->

```go
// transportConfig is the internal configuration struct for Transport.
// All fields are set via functional options at construction time.
type transportConfig struct {
    // invariant: if defaultStatus == 0, strict mode is active (unmatched → ErrNoRouteMatch).
    // invariant: if defaultStatus != 0, must be a valid HTTP status code (100–599).
    defaultStatus int

    // invariant: logger is never nil; defaults to log.Default() if not set.
    logger *slog.Logger
}

// Transport implements http.RoundTripper. It is the central type of the
// httpmock package.
type Transport struct {
    // invariant: mu protects all mutable fields below.
    mu sync.RWMutex

    // invariant: stubs is ordered by registration time; first match wins.
    stubs []*Stub

    // invariant: recorded is appended-to only; never re-ordered or removed.
    recorded []*http.Request

    // invariant: closed, once true, never becomes false.
    closed bool

    cfg transportConfig
}

// Stub is the chained builder for a single registered request expectation.
type Stub struct {
    // invariant: method, if non-empty, is the upper-cased HTTP verb.
    method string

    // invariant: exactly one of pathExact, pathPrefix, pathRegex is set at a time.
    // pathPrefix and pathRegex are mutually exclusive; both set → panic at registration.
    pathExact   string
    pathPrefix  string
    pathRegex   *regexp.Regexp

    // invariant: queryParams entries are ANDed (all must match).
    queryParams map[string]string

    // invariant: headers entries are ANDed (all must match); names are canonical (http.CanonicalHeaderKey).
    headers map[string]string

    // invariant: bodyMatcher may be nil (no body check).
    bodyMatcher BodyMatcher

    // invariant: exactly one of timesMin / timesMax / anyTimes is active.
    // timesExact > 0 implies timesMin == timesExact && timesMax == timesExact.
    timesMin  int   // -1 means no lower bound (AnyTimes)
    timesMax  int   // -1 means no upper bound (AnyTimes or AtLeast)
    hitCount  int   // invariant: protected by the Transport's mu during RoundTrip

    // invariant: responder must be non-nil when RoundTrip is called.
    // A nil responder causes RoundTrip to return ScriptError{Cause: "stub missing response"}.
    responder Responder
}

// BodyMatcher matches a request body. Implementations must be safe for
// concurrent calls.
type BodyMatcher interface {
    // Match is called with the raw request body bytes and the Content-Type header value.
    // invariant: body is the fully-read body; the original io.ReadCloser has been restored.
    Match(body []byte, contentType string) bool

    // String returns a human-readable description of the matcher, used in failure messages.
    String() string
}

// Responder produces an *http.Response (or transport-level error) for a
// matched request. Implementations must be safe for concurrent calls.
type Responder interface {
    // Respond is called once per matched RoundTrip. The returned response's
    // Body must be a non-nil io.ReadCloser (even for empty bodies).
    // invariant: exactly one of (*http.Response, nil) or (nil, non-nil error) is returned.
    Respond(req *http.Request) (*http.Response, error)
}

// ScriptError is the typed error for stub-script configuration failures
// (e.g., a stub was registered without a Responder).
// It follows the library error register: "httpmock: script step <n>: <cause>".
type ScriptError struct {
    // Step is the zero-based index of the stub in the registration order.
    // invariant: Step >= 0.
    Step  int

    // Cause is the underlying configuration error.
    // invariant: Cause is non-nil.
    Cause error
}
```

## API

<!--
  **Public.** Every exported symbol introduced by this spec.
  For each: signature, doc comment (which becomes godoc), preconditions, postconditions,
  error contract, concurrency notes (goroutine-safe? blocking?), lifecycle hooks.
  Magpie extracts signatures + doc comments verbatim to the API reference page.
-->

### Package declaration

```go
// Package httpmock provides a programmable http.RoundTripper for testing
// code that makes HTTP calls. The transport never makes real network calls;
// every response is scripted via stubs declared during test setup.
//
// Usage pattern:
//
//	rt := httpmock.NewWithT(t)
//	rt.OnRequest().Method("GET").Path("/users/42").
//	    Respond(httpmock.JSON(200, User{ID: 42, Name: "Ada"}))
//	client := &http.Client{Transport: rt}
//	// ... exercise code under test ...
//	// Verify is called automatically at t.Cleanup.
package httpmock
```

---

### Transport constructors

```go
// New constructs a fresh Transport with no registered stubs.
// The returned *Transport implements http.RoundTripper.
// Strict mode is the default: unmatched requests return ErrNoRouteMatch.
//
// Preconditions: none.
// Postconditions: the returned transport is ready for stub registration and RoundTrip calls.
// Concurrency: goroutine-safe after construction.
// Lifecycle: call Close when done to release held resources.
func New(opts ...option.Option[transportConfig]) *Transport
```

```go
// NewWithT constructs a Transport and registers Transport.Verify at t.Cleanup.
// Use this in table-driven and per-test setups where automatic expectation
// checking is desired.
//
// Preconditions: t must be non-nil.
// Postconditions: Verify will be called automatically when the test ends.
// Concurrency: goroutine-safe after construction.
// Lifecycle: Close is NOT called automatically; call it explicitly if the
// transport holds resources beyond the stub list.
func NewWithT(t assert.TB, opts ...option.Option[transportConfig]) *Transport
```

---

### Transport methods

```go
// RoundTrip implements http.RoundTripper. It matches req against registered
// stubs in registration order and returns the first matching stub's response.
// The request is always recorded (regardless of match outcome).
//
// Match, hit-count increment, and response production happen in a single
// critical section under the transport's mutex, ensuring Times(1) semantics
// are race-free (§23.14).
//
// Error contract:
//   - ErrNoRouteMatch: strict mode and no stub matched.
//   - ScriptError: a matched stub has no Responder configured.
//   - Any error returned by a Responder is forwarded as-is.
//
// Concurrency: goroutine-safe; multiple goroutines may call RoundTrip concurrently.
// Blocking: non-blocking (in-memory; no I/O).
// Lifecycle: returns ErrNoRouteMatch immediately if Close has been called.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error)
```

```go
// OnRequest returns a fresh Stub builder for the next request to match.
// Stubs are matched in registration order; the last call to Respond on the
// returned Stub finalizes the registration.
//
// Concurrency: goroutine-safe; acquires the write lock.
func (t *Transport) OnRequest() *Stub
```

```go
// RequestsTo returns all recorded requests whose URL path matches the given
// pattern. An exact string match is performed unless pattern contains a
// '*' wildcard, in which case simple glob matching is used
// (e.g., "/users/*" matches "/users/42").
//
// Concurrency: goroutine-safe; acquires the read lock.
func (t *Transport) RequestsTo(pattern string) []*http.Request
```

```go
// AllRequests returns every recorded request in arrival order.
// The returned slice is a copy; callers may modify it freely.
//
// Concurrency: goroutine-safe; acquires the read lock.
func (t *Transport) AllRequests() []*http.Request
```

```go
// Verify checks that all registered stubs with Times / AtLeast / AtMost /
// Never expectations have had their expectations met. For each unmet
// expectation, t.Errorf is called with the stub description and the actual
// vs. expected call counts.
//
// Verify is called automatically at t.Cleanup when the transport was
// constructed via NewWithT.
//
// Concurrency: goroutine-safe; acquires the read lock.
func (t *Transport) Verify()
```

```go
// Close releases the stub-list lock and marks the transport as closed.
// Subsequent RoundTrip calls return ErrNoRouteMatch immediately.
// Close is idempotent: a second call returns nil. Multiple concurrent calls
// are safe.
//
// Error contract: returns nil on success; errors from resource cleanup are
// joined via errs.Join.
//
// Concurrency: goroutine-safe.
func (t *Transport) Close() error
```

---

### Transport options

```go
// StrictDefault configures the transport to return ErrNoRouteMatch for any
// unmatched request. This is the default mode; the option exists to make
// explicit strict intent readable in test setup.
//
// §23.15: renamed from Strict() to StrictDefault to disambiguate from
// fixture.StrictLeaks and option.ApplyStrict.
func StrictDefault() option.Option[transportConfig]
```

```go
// LenientMode configures the transport to return an empty response with
// HTTP 404 for unmatched requests, rather than ErrNoRouteMatch.
// Equivalent to WithDefaultStatus(http.StatusNotFound).
//
// §23.15: renamed from Lenient() to LenientMode.
func LenientMode() option.Option[transportConfig]
```

```go
// WithDefaultStatus configures the transport to return an empty response
// with the given HTTP status code for unmatched requests (lenient mode).
//
// Preconditions: status must be a valid HTTP status code (100–599).
// Panics at construction time if status is outside [100, 599].
func WithDefaultStatus(status int) option.Option[transportConfig]
```

```go
// WithLogger injects the slog.Logger used for stub-match trace events.
// Defaults to log.Default() if not provided.
//
// Preconditions: l must be non-nil. Panics at construction if l is nil.
func WithLogger(l *slog.Logger) option.Option[transportConfig]
```

---

### Stub builder (chained)

All Stub methods return the same `*Stub` to enable chaining. Stub registration
is finalized when `Respond` is called. Calling `OnRequest()` again starts a new
independent stub.

```go
// Method restricts the stub to requests with the given HTTP method.
// Case-insensitive: "get", "GET", and "Get" all match GET requests.
//
// Preconditions: m must be non-empty.
// Chaining: returns the same *Stub for further configuration.
func (s *Stub) Method(m string) *Stub
```

```go
// Path restricts the stub to requests whose URL path equals p exactly.
// Mutually composable with Method, Query, Header, and Body.
// Mutually exclusive with PathPrefix + Regex on the same stub
// only in the sense that at most one path-matching mode is active;
// calling both Path and PathPrefix sets the last one wins (see PathPrefix doc).
//
// Preconditions: p must begin with '/'.
func (s *Stub) Path(p string) *Stub
```

```go
// PathPrefix restricts the stub to requests whose URL path starts with p.
// Composable with Method, Query, Header, and Body.
// Mutually exclusive with Regex: calling both PathPrefix and Regex on the
// same stub panics at registration time with a library-register message.
//
// Preconditions: p must be non-empty.
func (s *Stub) PathPrefix(p string) *Stub
```

```go
// Regex reinterprets the most recently set Path value as a regular expression
// anchored at the start (re2 / Go regexp syntax). Must be called after Path.
// Mutually exclusive with PathPrefix: calling both panics at registration.
//
// Preconditions: the Path set on this stub must be a valid Go regexp.
// Panics at registration time with a library-register message if the regexp
// does not compile.
func (s *Stub) Regex() *Stub
```

```go
// Query restricts the stub to requests that include the given query parameter
// with the given value. Multiple Query calls are ANDed together.
//
// Preconditions: name must be non-empty.
func (s *Stub) Query(name, value string) *Stub
```

```go
// Header restricts the stub to requests that include the given header with
// the given value. Header name matching is canonical (http.CanonicalHeaderKey).
// Multiple Header calls are ANDed together.
//
// Preconditions: name must be non-empty.
func (s *Stub) Header(name, value string) *Stub
```

```go
// Body restricts the stub to requests whose body satisfies the given
// BodyMatcher. The request body is fully read, compared, and then restored
// so that downstream code sees an intact io.ReadCloser.
//
// Built-in matchers: BodyExact, BodyJSON[T], BodyContains, BodyMatchFn.
//
// Preconditions: matcher must be non-nil.
func (s *Stub) Body(matcher BodyMatcher) *Stub
```

```go
// Times sets the exact expected call count for this stub.
// Verify (and auto-Verify at t.Cleanup) fails if the stub was not called
// exactly n times.
//
// Preconditions: n > 0. Panics at registration if n <= 0.
// Exclusive with AtLeast, AtMost, AnyTimes, Never on the same stub.
func (s *Stub) Times(n int) *Stub
```

```go
// AtLeast sets a minimum expected call count. Verify fails if the stub was
// called fewer than n times.
//
// Preconditions: n >= 0.
func (s *Stub) AtLeast(n int) *Stub
```

```go
// AtMost sets a maximum expected call count. Verify fails if the stub was
// called more than n times.
//
// Preconditions: n >= 0.
func (s *Stub) AtMost(n int) *Stub
```

```go
// AnyTimes removes all call-count expectations. The stub may be called any
// number of times, including zero, without triggering a Verify failure.
func (s *Stub) AnyTimes() *Stub
```

```go
// Never asserts that this stub must not be matched. Verify fails if the stub
// is called one or more times. Useful for asserting that a code path does not
// make a specific request.
func (s *Stub) Never() *Stub
```

```go
// Respond sets the Responder that produces the HTTP response for this stub.
// The stub is finalized and registered on the Transport when Respond is called.
// Calling Respond twice on the same stub overrides the first (last wins; no error).
//
// Preconditions: r must be non-nil.
func (s *Stub) Respond(r Responder) *Stub
```

---

### BodyMatcher implementations

```go
// BodyMatcher matches a request body. Implementations must be goroutine-safe.
type BodyMatcher interface {
    Match(body []byte, contentType string) bool
    String() string
}
```

```go
// BodyExact returns a BodyMatcher that requires the request body to be
// byte-for-byte identical to body.
//
// String() returns: `body exact: <hex-encoded prefix>`.
func BodyExact(body []byte) BodyMatcher
```

```go
// BodyJSON returns a BodyMatcher that unmarshals the request body as JSON
// into a zero-value T and compares it to want using assert.Equal with the
// supplied options. T is verified at compile time.
//
// opts may include assert.EqualOption values such as IgnoreFields(...) to
// perform partial structural comparison.
//
// String() returns: `body JSON: <T type name>`.
//
// §23.17: typed via generics; T is load-bearing for compile-time safety.
func BodyJSON[T any](want T, opts ...assert.EqualOption) BodyMatcher
```

```go
// BodyContains returns a BodyMatcher that requires the request body (interpreted
// as UTF-8) to contain the given substring.
//
// String() returns: `body contains: "<s>"`.
func BodyContains(s string) BodyMatcher
```

```go
// BodyMatchFn returns a BodyMatcher that delegates to f.
// f receives the raw body bytes; return true if the body matches.
//
// Preconditions: f must be non-nil.
// String() returns: `body fn: <func pointer>`.
func BodyMatchFn(f func([]byte) bool) BodyMatcher
```

---

### Responder implementations

```go
// Responder produces an *http.Response (or transport-level error) for a
// matched request. Implementations must be goroutine-safe.
type Responder interface {
    Respond(req *http.Request) (*http.Response, error)
}
```

```go
// JSON returns a Responder that marshals body as JSON, sets
// Content-Type: application/json, and returns a response with the given
// status code.
//
// Preconditions: status must be a valid HTTP status code (100–599).
// body must be JSON-marshallable.
// Panics at call time if json.Marshal(body) fails :  this is a programming
// error (mismatched type); library-register message:
//   "httpmock: JSON: marshal <T>: <cause>".
//
// §23.17: typed via generics; T is load-bearing for compile-time safety.
func JSON[T any](status int, body T) Responder
```

```go
// JSONFrom returns a Responder that reads all bytes from r, treats them as
// pre-encoded JSON, and decodes them into a zero-value T to validate the
// encoding. On each RoundTrip, the bytes are re-served (the Reader is fully
// consumed at construction time).
//
// Preconditions: r must be non-nil and contain valid JSON decodable into T.
// Panics at call time if reading r or decoding fails.
//
// §23.17: typed via generics.
func JSONFrom[T any](status int, r io.Reader) Responder
```

```go
// Status returns a Responder that produces an empty-body response with the
// given HTTP status code.
//
// Preconditions: status must be a valid HTTP status code (100–599).
func Status(status int) Responder
```

```go
// Body returns a Responder that produces a response with the given status
// code, raw body bytes, and Content-Type header.
//
// Preconditions: status must be a valid HTTP status code (100–599).
// body may be nil (treated as empty). contentType may be empty.
func Body(status int, body []byte, contentType string) Responder
```

```go
// Stream returns a Responder that produces a response with the given status
// code and a streaming body read from r. r is closed after RoundTrip
// completes. Use this for large or lazily-generated response bodies.
//
// Preconditions: status must be a valid HTTP status code (100–599). r must be non-nil.
// Error contract: if r.Read returns an error during RoundTrip, the error is
// wrapped as: "httpmock: stream: read: <cause>".
func Stream(status int, body io.Reader, contentType string) Responder
```

```go
// Error returns a Responder that causes RoundTrip to return err as a
// transport-level error (response is nil). Use this to simulate network
// errors, timeouts, or connection refused scenarios.
//
// Preconditions: err must be non-nil.
func Error(err error) Responder
```

```go
// Sequence returns a Responder that serves rs[0] on the first match,
// rs[1] on the second, etc. After all responders are exhausted, the
// sequence cycles back to rs[0] (cycle mode). See SequenceExhaust for
// non-cycling behavior.
//
// Equivalent to SequenceCycle(rs...).
//
// Preconditions: len(rs) >= 1. Panics if rs is empty.
// Concurrency: goroutine-safe; the index is advanced under the transport's mutex.
func Sequence(rs ...Responder) Responder
```

```go
// SequenceCycle is identical to Sequence: exhausted sequence cycles to start.
// Provided for explicit readability.
func SequenceCycle(rs ...Responder) Responder
```

```go
// SequenceExhaust returns a Responder that serves each responder in order.
// After all are exhausted, subsequent matches return ErrNoRouteMatch.
//
// Preconditions: len(rs) >= 1. Panics if rs is empty.
// Concurrency: goroutine-safe; the index is advanced under the transport's mutex.
func SequenceExhaust(rs ...Responder) Responder
```

---

### Static fixture loading

```go
// LoadFixtures reads testdata/httpmock/<name>.json (where <name> must not
// contain '..' components or absolute path separators), parses it as a stubs
// document, and registers each stub on the transport.
//
// Falcon §1.3 / §23.9 / §23.10 disciplines applied:
//   - Path is canonicalized via internal/safefile; traversal attempts return
//     a typed error and call t.Errorf (no stubs are registered).
//   - File size is capped at 16 MiB via io.LimitReader.
//   - JSON is decoded via internal/safejson: DisallowUnknownFields, depth ≤ 32,
//     UTF-8 validated.
//   - Schema is validated against the httpmock fixture schema.
//
// On any error (file not found, too large, malformed JSON, unknown field,
// schema violation), t.Errorf is called and no stubs are registered from
// the failing file.
//
// Fixture file format (JSON schema):
//
//	{
//	  "stubs": [
//	    {
//	      "method":   "GET",        // optional; any method if absent
//	      "path":     "/users/42",  // optional; any path if absent
//	      "respond":  {
//	        "status": 200,
//	        "headers": { "Content-Type": "application/json" },
//	        "body":   { "id": 42, "name": "Ada" }
//	      }
//	    }
//	  ]
//	}
//
// Preconditions: t must be non-nil. name must be a non-empty relative path
// component (no '/', no '..', no null bytes).
// Returns: the first error encountered, or nil on success.
func (t *Transport) LoadFixtures(tb assert.TB, name string) error
```

---

### Errors

```go
// ErrNoRouteMatch is returned by RoundTrip when strict mode is active and
// no registered stub matches the incoming request.
// Library register format: "httpmock: no route match".
var ErrNoRouteMatch = errs.Sentinel("httpmock: no route match")
```

```go
// ScriptError is the typed error returned by RoundTrip when a matched stub
// has an invalid configuration (e.g., missing Responder).
// Library register format: "httpmock: script step <n>: <cause>".
type ScriptError struct {
    // Step is the zero-based registration index of the misconfigured stub.
    Step  int
    // Cause is the underlying configuration error.
    Cause error
}

func (e *ScriptError) Error() string { /* "httpmock: script step <n>: <cause>" */ }
func (e *ScriptError) Unwrap() error { return e.Cause }
```

## Examples

<!--
  **Public.** Runnable Go examples in fenced ```go blocks.
  Each example is self-contained and `go test ./...`-compatible (valid Example functions).
  Magpie transcludes verbatim into tutorials.
-->

### Typed JSON response with post-test assertion

```go
func ExampleTransport_OnRequest() {
    // This example requires *testing.T; shown as a narrative example.
    // In a real test file this is a func TestUserFetch(t *testing.T) body.

    // rt := httpmock.NewWithT(t)  // Verify runs automatically at t.Cleanup.
    //
    // rt.OnRequest().
    //     Method("GET").
    //     Path("/users/42").
    //     Times(1).
    //     Respond(httpmock.JSON(200, User{ID: 42, Name: "Ada"}))
    //
    // client := &http.Client{Transport: rt}
    // resp, err := client.Get("https://api.example.com/users/42")
    // assert.NoError(t, err)
    //
    // var u User
    // assert.NoError(t, json.NewDecoder(resp.Body).Decode(&u))
    // assert.Equal(t, u, User{ID: 42, Name: "Ada"})
    //
    // seen := rt.RequestsTo("/users/42")
    // assert.Len(t, seen, 1)
    //
    // // Verify() fires at t.Cleanup :  confirms Times(1) was satisfied.
}
```

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

### Static fixture loading

```go
// testdata/httpmock/login_flow.json contains a stubs document with
// pre-configured responses for a multi-step login sequence.

func TestLoginFlow(t *testing.T) {
    rt := httpmock.NewWithT(t)
    err := rt.LoadFixtures(t, "login_flow")
    assert.NoError(t, err)

    // All stubs declared in login_flow.json are now registered.
    // Exercise the login client ...
}
```

### Transport-level error (simulating network failure)

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

### Composing with httpc (consumer-level wiring, no package coupling)

```go
func TestTypedGet(t *testing.T) {
    type Item struct{ ID int `json:"id"` }

    rt := httpmock.NewWithT(t)
    rt.OnRequest().Method("GET").Path("/items/1").
        Respond(httpmock.JSON(200, Item{ID: 1}))

    // httpc and httpmock are both leaves; wired together by the test, never by packages.
    c := httpc.New(httpc.WithTransport(rt), httpc.WithBaseURL("https://api.example.com"))
    item, err := httpc.Get[Item](context.Background(), c, "/items/1")
    assert.NoError(t, err)
    assert.Equal(t, item.ID, 1)
}
```

## Test Matrix

<!--
  **Internal.** Owned by Lynx.
  Table: scenario × input × expected outcome × covered-by-test-name.
-->

Source: `specs/test-matrices/leaves.md` § `## Package: httpmock/`. All 70 rows from that matrix are binding. The table below is the full matrix reproduced here as the spec's authoritative test contract.

| # | Name | Spec ref | Type | Description | Helpers |
|---|---|---|---|---|---|
| 1 | TestNewReturnsTransport | §21.11 F1 | unit | `New()` returns *Transport implementing http.RoundTripper | assert |
| 2 | TestNewWithT_RegistersCleanup | §21.11 F7 | unit | `NewWithT(t)` registers Verify at t.Cleanup | assert |
| 3 | TestRoundTripBasicMatch | §21.11 F2 | unit | OnRequest+Method+Path → scripted response returned | assert |
| 4 | TestRoundTripNeverDials | §21.11 NF3 / E14 | unit | Package-import audit + runtime: no real-network call | go-list-deps audit |
| 5 | TestStubFirstWins | §21.11 E2 | unit | First registered stub wins when multiple match | assert |
| 6 | TestStubMethodCaseInsensitive | §21.11 F9 | unit | "get" matches "GET" | assert |
| 7 | TestStubPathExact | §21.11 F10 | unit | Path("/users/42") exact match | assert |
| 8 | TestStubPathPrefix | §21.11 F11 | unit | PathPrefix("/users/") matches /users/42 | assert |
| 9 | TestStubRegex | §21.11 F12 | unit | `Path(`/users/(\d+)`).Regex()` matches /users/42 | assert |
| 10 | TestStubRegexInvalidCompile | §21.11 F12 | unit | Invalid regexp → panic at registration, library-register message | assert.Panics |
| 11 | TestStubPathPrefixAndRegexMutuallyExclusive | §21.11 F12 | unit | Both PathPrefix and Regex on same stub → panic | assert.Panics |
| 12 | TestStubQuery | §21.11 F13 | unit | Query("page","2") matches | assert |
| 13 | TestStubHeader | §21.11 F14 | unit | Header("X-Foo","bar") matches | assert |
| 14 | TestStubHeaderCaseInsensitiveName | §21.11 F14 | unit | Header name matched canonically | assert |
| 15 | TestBodyExact | §21.11 F15 | unit | BodyExact(bytes) matches byte-for-byte | assert |
| 16 | TestBodyJSONSmartEqual | §21.11 F15 | unit | BodyJSON[T](want, IgnoreFields("CreatedAt")) works | assert |
| 17 | TestBodyJSONTypeParam | §23.17 | unit | BodyJSON[User] carries T at compile time | compile-only check |
| 18 | TestBodyContains | §21.11 F15 | unit | Substring match | assert |
| 19 | TestBodyMatchFn | §21.11 F15 | unit | Predicate function match | assert |
| 20 | TestRespondJSON | §21.11 F18 | unit | JSON[T](200, body) marshals + sets Content-Type | assert |
| 21 | TestRespondJSONFrom | §21.11 F19 | unit | JSONFrom[T](200, reader) reads + validates | assert |
| 22 | TestRespondJSONFromMalformed | §21.11 F19 | unit | Malformed JSON in reader → panic at registration | assert.Panics |
| 23 | TestRespondStatus | §21.11 F20 | unit | Empty body, status code set correctly | assert |
| 24 | TestRespondBody | §21.11 F21 | unit | Raw bytes + content type served | assert |
| 25 | TestRespondStream | §21.11 F22 | unit | Streaming reader honored; reader closed after RoundTrip | assert, fixture |
| 26 | TestRespondStreamReaderError | §21.11 E6 | unit | Reader error mid-read → wrapped transport error | assert |
| 27 | TestRespondError | §21.11 F23 | unit | Error(err) surfaces err to http.Client caller | assert |
| 28 | TestSequenceBasic | §21.11 F24 | unit | rs[0], rs[1], ... served in order | assert |
| 29 | TestSequenceCycle | §21.11 E9 | unit | Exhausted cycle wraps back to rs[0] | assert |
| 30 | TestSequenceExhaust | §21.11 E10 | unit | SequenceExhaust → ErrNoRouteMatch after exhaustion | assert |
| 31 | PropertySequenceCycleSixCalls | property | property | Sequence(a,b,c).Cycle: first 6 calls = [a,b,c,a,b,c] | rapid |
| 32 | TestStubTimes | §21.11 F16 | unit | Times(2): stub must be called exactly 2 times | assert |
| 33 | TestStubAtLeast | §21.11 F16 | unit | AtLeast(2): Verify fails if called fewer than 2 times | assert |
| 34 | TestStubAtMost | §21.11 F16 | unit | AtMost(2): Verify fails if called more than 2 times | assert |
| 35 | TestStubNever | §21.11 F16 | unit | Never: Verify fails if stub is matched at all | mock(TB) |
| 36 | TestStubAnyTimes | §21.11 F16 | unit | AnyTimes: any call count passes Verify | assert |
| 37 | TestStubMissingRespond | §21.11 E4 | unit | Stub without Respond → ScriptError on RoundTrip | assert |
| 38 | TestStubRespondCalledTwiceLastWins | §21.11 E3 | unit | Second Respond call overrides first; no error | assert |
| 39 | TestStrictDefaultUnmatchedReturnsErrNoRouteMatch | §21.11 NF5 | unit | Unmatched request in strict mode → ErrNoRouteMatch | assert |
| 40 | TestLenientUnmatched404 | §21.11 F8 / §23.15 | unit | LenientMode() returns empty 404 response | assert |
| 41 | TestWithDefaultStatus | §21.11 F8 | unit | WithDefaultStatus(503) returns empty 503 for unmatched | assert |
| 42 | TestWithLogger | §21.11 F8 | unit | Injected logger receives stub-match trace events | mock(slog handler) |
| 43 | TestRequestsTo | §21.11 F4 | unit | RequestsTo("/users/42") returns matching requests | assert |
| 44 | TestRequestsToWildcard | §21.11 F4 | unit | "/users/*" wildcard matches /users/42 | assert |
| 45 | TestAllRequests | §21.11 F5 | unit | Returns every request in arrival order | assert |
| 46 | TestVerifyAllStubsTimesMet | §21.11 F6 / E7 | unit | Verify reports each unmet stub expectation via t.Errorf | mock(TB) |
| 47 | TestVerifyAtCleanup_NewWithT | §21.11 F7 | unit | NewWithT auto-Verify fires at t.Cleanup | mock(TB) |
| 48 | TestLoadFixturesBasic | §21.11 F25 | integration | Reads testdata/httpmock/<name>.json; registers stubs | fixture |
| 49 | TestLoadFixturesMalformedJSON | §21.11 E11 | unit | Malformed JSON → t.Errorf; no stubs registered | fixture |
| 50 | TestLoadFixturesTooLarge | §21.11 E12 | unit | File > 16 MiB → typed error; no stubs registered | fixture (16 MiB+ file) |
| 51 | TestLoadFixturesPathTraversal | §21.11 E13 / §23.10 | unit | `../etc/passwd` → typed error; no stubs registered | fixture |
| 52 | TestLoadFixturesAbsolutePath | §23.10 | unit | Absolute path rejected | fixture |
| 53 | TestLoadFixturesUNCRejected | §23.10 / X-platform | unit | Windows UNC path rejected (build-tagged) | fixture |
| 54 | TestLoadFixturesUnknownFields | §21.11 NF6 / §23.10 | unit | DisallowUnknownFields: extra keys → rejection | fixture |
| 55 | TestLoadFixturesDepthCap | §23.10 | unit | JSON nesting depth > 32 → rejection | fixture |
| 56 | TestLoadFixturesUTF8Validated | §23.10 | unit | Malformed UTF-8 in fixture → rejection | fixture |
| 57 | TestNoNetworkImports | §21.11 NF3 | audit | Package imports neither `net.Dial` nor `http.DefaultTransport` | go-list-deps |
| 58 | TestConcurrentRoundTrip | §21.11 NF4 / E8 | concurrency | Many goroutines calling RoundTrip simultaneously; -race clean | -race |
| 59 | TestTimes1RaceFix | §23.14 | concurrency | Match + increment + respond in single critical section; -race | -race |
| 60 | TestTransportClose | §23.16 | lifecycle | Close() flushes recorded requests; subsequent RoundTrip → ErrNoRouteMatch | assert |
| 61 | TestTransportCloseIdempotent | §23.16 | lifecycle | Close() called twice returns nil both times | assert |
| 62 | TestMultipartBodyMatch | §21.11 E15 | unit | Body matchers operate on raw multipart bytes | assert |
| 63 | TestScriptErrorTypedAndUnwrap | §21.11 F28 | unit | ScriptError.Error() matches register; Unwrap returns Cause | assert.ErrorAs |
| 64 | FuzzFixture | §23.10 / §21.11 F25 | fuzz | Malformed JSON, oversized content, control chars in fixture loader | testing.F |
| 65 | TestImportSurfaceFluentInternal | §21.11 NF9 | audit | `fluent` present in transitive import set for internal use | go-list |
| 66 | BenchmarkRoundTripFirstStubMatches | §23.13 / NF1 | bench | Simple match-and-respond: target ≤ 5 µs/op, ≤ 5 allocs/op | testing.B |
| 67 | BenchmarkRoundTripScanThirty | §21.11 NF1 | bench | 30-stub linear scan: bounded latency | testing.B |
| 68 | BenchmarkBodyJSONMatch | §21.11 NF1 | bench | BodyJSON[T] smart-equal cost bounded | testing.B |
| 69 | BenchmarkResponseJSON | §21.11 NF1 | bench | JSON[T] response build cost bounded | testing.B |
| 70 | ExampleTransport_OnRequest | §21.11 examples | doc-test | Headline godoc example | godoc |

### Additional edge-case tests (Lynx additions)

| # | Name | Spec ref | Type | Description | Helpers |
|---|---|---|---|---|---|
| 71 | TestEmptyStubMatchesAll | leaves.md edge case | unit | Stub with no matchers matches every request | assert |
| 72 | TestPathAndPathPrefixCoexist | leaves.md edge case | unit | Path + PathPrefix on same stub: AND-match semantics | assert |
| 73 | TestEmptyBodyMatcherSemantics | leaves.md edge case | unit | BodyContains("") on request with no body: matches | assert |
| 74 | TestSequenceSingleElementCycles | leaves.md edge case | unit | Sequence(a).Cycle: every call returns a | assert |
| 75 | TestLateRegistrationApplies | leaves.md edge case | unit | Stub registered after first RoundTrip applies to subsequent | assert |
| 76 | TestClientTimeoutHonored | leaves.md edge case | unit | http.Client.Timeout respected; Transport does not override | assert, fixture.NewClock |

### Coverage targets

- **Line coverage:** 95% minimum.
- **Public API coverage:** 100%.

## Dependency Justification

<!--
  **Internal.** Owned by Falcon.
  One row per new direct dependency. The empty table is the goal.
  Required answers: license, last-release-date, maintainer count, alternatives considered, why we don't roll our own.
-->

| Module | Version | License | Last release | Maintainers | Alternatives considered | Why we can't roll our own |
|---|---|---|---|---|---|---|

No new direct dependencies introduced by this spec. `httpmock` imports only Glacier kernel packages (`option`, `errs`, `log`, `assert`), one Glacier mid-tier package (`fluent`, internal use), and stdlib. Falcon §1.3 supply-chain posture: zero third-party deps.

## Security & Supply-Chain Notes

<!-- **Internal.** Untrusted-input handling, sandboxing implications, secrets handling, vuln-scan considerations. -->

### Untrusted-input register rows (§23.9)

This package touches two untrusted-input register rows:

| Row | Input source | Parser / decoder | Size cap | Validation rule |
|---|---|---|---|---|
| 5 | HTTP fixture file bytes | `internal/safejson`: `encoding/json` + `DisallowUnknownFields` + depth ≤ 32 + UTF-8 validation | 16 MiB via `io.LimitReader` | Schema-validated against httpmock fixture schema |
| 6 | HTTP fixture file path | `internal/safefile` path canonicalization | as row 5 | Resolved path must be under `testdata/httpmock/`; `..` components rejected; absolute paths rejected; Windows UNC paths rejected |

### Path safety (§23.10 / §7.7)

`LoadFixtures` routes all path operations through `internal/safefile`:

1. `filepath.Clean` canonicalization.
2. Reject any post-clean component equal to `..`.
3. Reject `filepath.IsAbs` paths (unless caller passes `safefile.AllowAbsolute()`, which `LoadFixtures` never does).
4. Reject Windows UNC paths (`strings.HasPrefix(p, "\\\\")` or `\\?\`).
5. Open-then-fstat (never stat-then-open) to avoid TOCTOU.
6. `Lstat`-and-reject non-regular files.

The resolved absolute path must have the `testdata/httpmock/` directory as a prefix; any attempt to escape this root returns a typed error and calls `t.Errorf`.

### JSON fixture hardening (§23.10)

All fixture JSON decoding passes through `internal/safejson.Decode[T]`, which applies in one call:

- `io.LimitReader(r, 16*1024*1024)` :  16 MiB cap; excess triggers `ErrFixtureTooLarge`.
- `json.NewDecoder(...).DisallowUnknownFields()` :  unknown fields are rejected.
- Depth cap ≤ 32 :  deeply-nested JSON (via adversarial fixture) is rejected.
- UTF-8 validation :  malformed byte sequences are rejected.

### "Never makes a real network call" (NF3 / E14)

The package's import graph must never include `net/http.DefaultTransport` or `net.Dial`. This is an architectural invariant, not merely a runtime check. `TestNoNetworkImports` (row 57 in the test matrix) verifies it by running `go list -deps` on the package and asserting the absence of those symbols. This test is run in CI on every PR.

As a defense-in-depth measure, the `RoundTrip` implementation never calls `http.DefaultTransport.RoundTrip` or any function that resolves to a real dialer. The entire response surface is the in-memory stub list.

### Recording mode (forward-pointer)

v0 ships replay only. When `0024-httpmock-record.md` is accepted, the scrubbing rules from Falcon §1.3 (allowlist-based header redaction for `Authorization`, `Cookie`, `Set-Cookie`, and matching `(?i)auth|key|token|cookie|secret`; body-redaction hook API) will be applied to any persisted traffic. No recording infrastructure is added in this spec; no stubs for it are laid.

### Secrets in fixture files

Fixture files are committed to the repository under `testdata/httpmock/`. Developers are responsible for ensuring no real secrets appear in fixtures. Glacier's CI will add a secrets-scanning gate (e.g., `gitleaks`) separate from this package spec. This spec does not itself scan fixture content for secrets.

## FAQ

<!-- **Public.** Anticipated user questions with answers. Magpie extracts to the public docs FAQ. -->

**Why no recording mode?**

Recording mode :  capturing real HTTP traffic and persisting it as fixtures :  requires a scrubbing pass to redact secrets from headers and bodies before anything touches disk. Getting that scrubbing right is a non-trivial security review, and shipping it half-baked would be worse than shipping nothing. v0 focuses on the case that matters most for unit testing: scripted, deterministic, zero-network responses. Recording lands in `0024-httpmock-record.md` once there is a real consumer use case and a Falcon-reviewed scrubbing design.

**How does httpmock integrate with httpc?**

Both `httpmock` and `httpc` are Tier 2 leaf packages. Leaves must not import each other (D12). They are wired together by the test or application at the call site: `httpc.New(httpc.WithTransport(httpmock.New()))`. This keeps both packages independently usable :  a project that uses `httpc` for production HTTP does not need to pull in `httpmock`, and a project using `httpmock` to test a hand-rolled `*http.Client` does not need `httpc`. The wiring is deliberate and follows the framework's composition-at-the-leaves rule.

**Why is strict mode the default?**

An unmatched request in lenient mode silently returns a 404 or empty response. A test that relies on lenient behavior may pass even when the code under test is calling the wrong endpoint, has a typo in the path, or is making extra requests it shouldn't. Strict mode makes every unmatched request a loud, immediate failure with `ErrNoRouteMatch`. If you want lenient mode, you must explicitly opt in with `LenientMode()` or `WithDefaultStatus(n)` :  the API makes the intent visible in the test code.

**What happens if I forget to call Respond on a stub?**

If `Respond` is never called, the stub's `responder` field is nil. When `RoundTrip` matches that stub, it returns a `ScriptError` (row 37 in the test matrix). Use `NewWithT(t)` and `Times(n)` :  the `Verify` at `t.Cleanup` will also catch it if the stub was never matched at all. Both failure modes surface the problem before the test can pass accidentally.

**Is it safe to share one Transport across concurrent goroutines?**

Yes. All mutable state (stub list, recorded requests, closed flag) is protected by a single `sync.RWMutex`. `RoundTrip` acquires the write lock for the entire match-increment-respond critical section, so `Times(1)` semantics are race-free even under heavy concurrency. `RequestsTo`, `AllRequests`, and `Verify` acquire the read lock. Run your tests with `-race`; the package passes cleanly.

**Can I use httpmock outside of tests?**

The package only uses `net/http` from the standard library for its `RoundTripper` interface implementation; there is no `//go:build` constraint limiting it to test contexts. The `NewWithT` constructor takes an `assert.TB` (which `*testing.T` satisfies), but `New()` does not. In practice, `httpmock` is designed for tests and there is little reason to use it in production code :  but nothing prevents it.

## Decisions & Rationale

<!-- **Internal.** Why-this-and-not-that for non-obvious choices. Folded-in ADR. -->

**D-HM-1: Fluent chained builder over declarative struct literals.**
A struct literal approach (`Stub{Method: "GET", Path: "/users/42"}`) is simpler to parse but harder to extend without breaking callers. The chained builder lets future specs add new matcher predicates without changing existing call sites. It also reads naturally at the test site: left-to-right, one predicate per line.

**D-HM-2: First registered wins (not best match).**
"Best match" semantics (longest path, most-specific method, etc.) require a ranking function that is opaque to the test author. First-registered is predictable: if you add a catch-all stub at the end, it catches everything that earlier stubs don't. Test authors can read their stub list top to bottom and reason about which stub fires for a given request.

**D-HM-3: Times(1) race fix via single critical section (§23.14).**
The naive implementation acquires the read lock to find a matching stub, then acquires the write lock to increment the hit count. Between those two acquisitions, another goroutine can match the same stub, causing both to fire when Times(1) was expected to fire exactly once. The fix: `RoundTrip` acquires the write lock for the entire match-increment-respond sequence. This is slightly more contended but correct, and the performance target (≤ 5 µs/op) is achievable under a write lock for in-memory operations.

**D-HM-4: StrictDefault / LenientMode names (§23.15).**
The original `Strict()` / `Lenient()` names collided with `fixture.StrictLeaks()` and `option.ApplyStrict()`. Renaming to `StrictDefault` / `LenientMode` makes the domain context explicit at the call site and eliminates cross-package naming confusion. `StrictDefault` conveys "strict is the default; this option makes that explicit." `LenientMode` conveys "activate lenient behavior."

**D-HM-5: Close() as a lifecycle method (§23.16).**
`Transport` holds a `sync.RWMutex` and a slice of recorded requests. Providing `Close()` satisfies the Glacier lifecycle rule (D17) and signals to consumers that the transport has a defined teardown point. `Close()` is idempotent (second call returns nil) and marks the transport closed so subsequent `RoundTrip` calls return `ErrNoRouteMatch` immediately. `NewWithT` does not auto-close; callers who care about explicit teardown call `Close` themselves via `defer`.

**D-HM-6: JSON[T] panics at registration, not at RoundTrip.**
If `json.Marshal(body)` fails (e.g., the caller passed a channel or a function), the error is surfaced at the `JSON[T](...)` call site in test setup, not at the first `RoundTrip` call. Panicking at setup time gives a clear stack trace pointing to the test setup line, not to a confusing runtime error deep inside the transport. This is consistent with how mock.OnCall panics on invalid method names.

**D-HM-7: Fixture path rooted under testdata/httpmock/ (Falcon §1.3 / §23.10).**
The fixture root is hardcoded to `testdata/httpmock/` relative to the caller's working directory (i.e., the package under test). This is the same convention as `fixture.Load` and the Go test tooling itself. Rooting under `testdata/` ensures `go build` never includes fixture files in the binary. The `httpmock/` subdirectory disambiguates httpmock fixtures from fixture-package golden files.

**D-HM-8: Replay-only at v0 (NF6 / §23.19).**
The original Falcon §1.3 analysis focused on recording mode :  scrubbing headers and bodies before persisting traffic. Since v0 ships replay only, those scrubbing rules become forward-pointers. Deferring recording simplifies the v0 surface, eliminates a non-trivial security-review scope, and lets the recording design be informed by real usage patterns once the replay story is mature.

### §23 amendments incorporated

| Amendment | Effect on this spec |
|---|---|
| §23.9 row 5 | HTTP fixture file bytes: 16 MiB cap, `DisallowUnknownFields`, depth ≤ 32, UTF-8 validation via `internal/safejson`. |
| §23.9 row 6 | HTTP fixture file path: canonicalization + root-prefix enforcement via `internal/safefile`. |
| §23.10 | Path-safety convention (§7.7) applied to `LoadFixtures` in full. |
| §23.13 | Performance targets recalibrated: simple match-and-respond ≤ 5 µs/op; ≤ 5 allocs/op. |
| §23.14 | Times(1) race fix: match-AND-increment-AND-respond in single write-lock critical section. `TestTimes1RaceFix` added to matrix (row 59). |
| §23.15 | `Strict()` renamed `StrictDefault()`; `Lenient()` renamed `LenientMode()`. |
| §23.16 | `Transport.Close() error` added; idempotent; `errs.Join` over resource errors. `TestTransportClose` and `TestTransportCloseIdempotent` added to matrix (rows 60–61). |
| §23.17 | `JSON[T any]`, `JSONFrom[T any]`, `BodyJSON[T any]` confirmed as generic, compile-time typed. `TestBodyJSONTypeParam` (row 17) is a compile-only correctness check. |

## Open Questions

<!--
  **Internal.** Unresolved items.
  MUST be empty before this spec moves to `accepted`.
-->

None.

## Verification

<!-- **Internal.** Concrete steps to prove the change works end-to-end. Run when the spec moves to `verified`. -->

The following verification steps are performed when the spec moves from `implemented` to `verified`.

1. **Full test suite passes under race detector.**
   ```
   go test -race ./httpmock/...
   ```
   All 76 test rows (including the 6 edge-case additions) pass. No data races reported.

2. **No-network import audit.**
   ```
   go list -deps github.com/nathanbrophy/glacier/httpmock
   ```
   Assert that neither `net` (containing `Dial`) nor `net/http` `DefaultTransport` symbol appears as a direct package-level reference. `TestNoNetworkImports` (row 57) automates this check via `go-list-deps` in CI.

3. **Fuzz target runs clean for 30 seconds (PR gate) and 10 minutes (nightly).**
   ```
   go test -fuzz=FuzzFixture -fuzztime=30s ./httpmock/
   ```
   No crash or finding reported. Seed corpus covers: valid fixture, malformed JSON, truncated JSON, depth-bomb (>32 levels), oversized fixture (>16 MiB), path traversal in name field, control characters in fixture strings.

4. **Benchmark targets met.**
   ```
   go test -bench=. -benchmem -count=10 ./httpmock/
   ```
   Confirm:
   - `BenchmarkRoundTripFirstStubMatches`: ≤ 5 µs/op, ≤ 5 allocs/op.
   - `BenchmarkRoundTripScanThirty`: linear growth with stub count; no quadratic behavior.
   - `BenchmarkBodyJSONMatch`: cost proportional to struct size; no unbounded allocation.
   - `BenchmarkResponseJSON`: marshal cost dominated by JSON encoder, not httpmock overhead.

5. **LOC budget respected.**
   ```
   find httpmock -name '*.go' ! -name '*_test.go' | xargs wc -l
   ```
   Production code total ≤ 800 lines. Test code total ≤ 1100 lines.

6. **Coverage thresholds met.**
   ```
   go test -coverprofile=cover.out ./httpmock/...
   go tool cover -func=cover.out
   ```
   Line coverage ≥ 95%. Public API coverage = 100%.

7. **Path traversal rejection verified.**
   `TestLoadFixturesPathTraversal` (row 51), `TestLoadFixturesAbsolutePath` (row 52), and `TestLoadFixturesUNCRejected` (row 53, build-tagged for Windows) all pass.

8. **Lifecycle Close idempotency verified.**
   `TestTransportClose` (row 60) and `TestTransportCloseIdempotent` (row 61) pass without error. A transport closed twice returns nil on the second call.

9. **godoc examples compile and run.**
   ```
   go test -run=^Example ./httpmock/
   ```
   All `Example*` functions in `example_test.go` pass.

10. **staticcheck and go vet pass.**
    ```
    go vet ./httpmock/...
    staticcheck ./httpmock/...
    ```
    Zero findings.
