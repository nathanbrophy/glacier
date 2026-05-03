---
title: Mocking HTTP
---

# Mocking HTTP

<PackagesUsedBadges :package-names="['httpmock', 'httpc', 'fixture']" />

When your code calls an external HTTP API, you want tests that run without a real network: no flakiness, no credentials, no rate limits. [`httpmock`](/docs/packages/httpmock) is a programmable `http.RoundTripper` that intercepts every outgoing request, matches it against registered stubs, and returns a scripted response - all in-process. Pair it with [`httpc`](/docs/packages/httpc) to test typed, retry-aware HTTP calls end-to-end.

## Walkthrough

### Step 1 - Create a transport and stub a route

`httpmock.NewWithT(t)` creates a transport bound to your test. When the test ends, `Verify` runs automatically via `t.Cleanup` and reports any unmet `Times` expectations.

```go
import (
    "testing"
    "github.com/nathanbrophy/glacier/httpmock"
)

func TestFetchUser(t *testing.T) {
    rt := httpmock.NewWithT(t)

    type User struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }

    rt.OnRequest().
        Method("GET").
        Path("/users/42").
        Times(1).
        Respond(httpmock.JSON(200, User{ID: 42, Name: "Ada"}))
}
```

The transport is strict by default: any request that does not match a registered stub is an immediate test failure. There is no "pass-through to real network" mode.

### Step 2 - Wire the transport into httpc

`httpc` and `httpmock` are both Tier 2 leaf packages and do not import each other. You wire them at the test level: pass the mock transport to `httpc.New`.

```go
import (
    "context"
    "net/http"
    "testing"

    "github.com/nathanbrophy/glacier/assert"
    "github.com/nathanbrophy/glacier/httpc"
    "github.com/nathanbrophy/glacier/httpmock"
)

func TestFetchUser(t *testing.T) {
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

    client, err := httpc.New(httpc.WithTransport(rt))
    assert.NoError(t, err)

    user, resp, err := httpc.Get[User](
        context.Background(),
        "https://api.example.com/users/42",
        httpc.WithClient(client),
    )
    assert.NoError(t, err)
    assert.Equal(t, resp.StatusCode, 200)
    assert.Equal(t, user.Name, "Ada")
}
```

`httpc.Get[User]` reads and JSON-decodes the response body for you. No boilerplate read-all-unmarshal loop.

### Step 3 - Test retry behavior with sequenced responses

`httpmock.SequenceExhaust` returns responses in order, then fails the test if the sequence is exhausted unexpectedly. Use it to verify your retry logic hits the right status codes in the right order.

```go
rt.OnRequest().
    Method("POST").
    Path("/login").
    Times(3).
    Respond(httpmock.SequenceExhaust(
        httpmock.Status(503),
        httpmock.Status(503),
        httpmock.JSON(200, LoginResponse{Token: "tok-abc123"}),
    ))
```

### Step 4 - Use fixture for golden-file HTTP responses

For complex response payloads, store the response body in `testdata/` and load it with [`fixture`](/docs/packages/fixture).

```go
import (
    "testing"
    "github.com/nathanbrophy/glacier/fixture"
    "github.com/nathanbrophy/glacier/httpmock"
)

func TestFetchProfile(t *testing.T) {
    body := fixture.Load(t, "testdata/httpmock/profile_response.json")

    rt := httpmock.NewWithT(t)
    rt.OnRequest().
        Method("GET").
        Path("/profile").
        Respond(httpmock.Raw(200, "application/json", body))
}
```

Run with `GLACIER_GOLDEN_UPDATE=1` to create the fixture file on first run.

### Step 5 - Assert which requests were made

After the code under test runs, inspect the recorded requests.

```go
seen := rt.RequestsTo("/users/42")
assert.Len(t, seen, 1)
assert.Equal(t, seen[0].Method, "GET")
```

## Putting it together

```go
package client_test

import (
    "context"
    "testing"

    "github.com/nathanbrophy/glacier/assert"
    "github.com/nathanbrophy/glacier/assert/require"
    "github.com/nathanbrophy/glacier/httpc"
    "github.com/nathanbrophy/glacier/httpmock"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func TestUserClient_Get(t *testing.T) {
    rt := httpmock.NewWithT(t)
    rt.OnRequest().
        Method("GET").
        Path("/users/42").
        Times(1).
        Respond(httpmock.JSON(200, User{ID: 42, Name: "Ada"}))

    httpClient, err := httpc.New(httpc.WithTransport(rt))
    require.NoError(t, err)

    user, resp, err := httpc.Get[User](
        context.Background(),
        "https://api.example.com/users/42",
        httpc.WithClient(httpClient),
    )
    require.NoError(t, err)
    assert.Equal(t, resp.StatusCode, 200)
    assert.Equal(t, user.Name, "Ada")

    seen := rt.RequestsTo("/users/42")
    assert.Len(t, seen, 1)
    // Verify() fires at t.Cleanup: confirms Times(1) was satisfied.
}
```

## What's happening underneath

- <TierBadge tier="leaf" /> [`httpmock`](/docs/packages/httpmock): an in-memory `http.RoundTripper` that never dials the network; strict by default, auto-verifies at `t.Cleanup`.
- <TierBadge tier="leaf" /> [`httpc`](/docs/packages/httpc): typed generic HTTP methods that auto-unmarshal the response; accepts any `http.RoundTripper` via `WithTransport`.
- <TierBadge tier="mid" /> [`fixture`](/docs/packages/fixture): golden-file helpers for storing and comparing response payloads from `testdata/`.

## Related

- [Writing tests](/docs/writing-tests) - assert, mock interfaces, and inject fake clocks.
- [Observability](/docs/observability) - add `httpc.WithTracing()` and `httpc.WithMetrics()` in production code.
- [Building a CLI](/docs/building-a-cli) - dry-run propagation through `context.Context` for audit-only HTTP calls.
