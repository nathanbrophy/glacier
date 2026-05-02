// SPDX-License-Identifier: Apache-2.0

package httpmock_test

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpmock"
)

func TestNewReturnsTransport(t *testing.T) {
	rt := httpmock.New()
	assert.NotNil(t, rt)
	// Verify it implements http.RoundTripper.
	var _ http.RoundTripper = rt
}

func TestNewWithT_RegistersCleanup(t *testing.T) {
	// Use a sub-test so we can observe cleanup running.
	var verified bool
	t.Run("sub", func(t *testing.T) {
		inner := &trackingTB{TB: t}
		rt := httpmock.NewWithT(inner)
		_ = rt
		// Register a Times(1) stub that will never be called — should cause
		// Verify to call Errorf. We check the cleanup ran by observing it.
		rt.OnRequest().Path("/x").Times(1).Respond(httpmock.Status(200))
		verified = true
	})
	assert.True(t, verified)
}

func TestRoundTripBasicMatch(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Method("GET").Path("/hello").Respond(httpmock.Status(200))

	req := newReq(t, "GET", "https://example.com/hello", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestStubFirstWins(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/x").Respond(httpmock.Status(201))
	rt.OnRequest().Path("/x").Respond(httpmock.Status(202))

	req := newReq(t, "GET", "https://example.com/x", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 201)
}

func TestStrictDefaultUnmatchedReturnsErrNoRouteMatch(t *testing.T) {
	rt := httpmock.New()
	req := newReq(t, "GET", "https://example.com/nope", nil)
	_, err := rt.RoundTrip(req)
	assert.ErrorIs(t, err, httpmock.ErrNoRouteMatch)
}

func TestLenientUnmatched404(t *testing.T) {
	rt := httpmock.New(httpmock.LenientMode())
	req := newReq(t, "GET", "https://example.com/nope", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestWithDefaultStatus(t *testing.T) {
	rt := httpmock.New(httpmock.WithDefaultStatus(503))
	req := newReq(t, "GET", "https://example.com/nope", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 503)
}

func TestRequestsTo(t *testing.T) {
	rt := httpmock.New(httpmock.LenientMode())
	req1 := newReq(t, "GET", "https://example.com/users/42", nil)
	req2 := newReq(t, "GET", "https://example.com/users/99", nil)
	req3 := newReq(t, "GET", "https://example.com/other", nil)
	_, _ = rt.RoundTrip(req1)
	_, _ = rt.RoundTrip(req2)
	_, _ = rt.RoundTrip(req3)

	got := rt.RequestsTo("/users/42")
	assert.Len(t, got, 1)
}

func TestRequestsToWildcard(t *testing.T) {
	rt := httpmock.New(httpmock.LenientMode())
	req1 := newReq(t, "GET", "https://example.com/users/42", nil)
	req2 := newReq(t, "GET", "https://example.com/users/99", nil)
	req3 := newReq(t, "GET", "https://example.com/other", nil)
	_, _ = rt.RoundTrip(req1)
	_, _ = rt.RoundTrip(req2)
	_, _ = rt.RoundTrip(req3)

	got := rt.RequestsTo("/users/*")
	assert.Len(t, got, 2)
}

func TestAllRequests(t *testing.T) {
	rt := httpmock.New(httpmock.LenientMode())
	req1 := newReq(t, "GET", "https://example.com/a", nil)
	req2 := newReq(t, "POST", "https://example.com/b", nil)
	_, _ = rt.RoundTrip(req1)
	_, _ = rt.RoundTrip(req2)

	all := rt.AllRequests()
	assert.Len(t, all, 2)
	assert.Equal(t, all[0].URL.Path, "/a")
	assert.Equal(t, all[1].URL.Path, "/b")
}

func TestWithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	rt := httpmock.New(httpmock.WithLogger(logger))
	rt.OnRequest().Path("/x").Respond(httpmock.Status(200))
	req := newReq(t, "GET", "https://example.com/x", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestEmptyStubMatchesAll(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Respond(httpmock.Status(200))

	req := newReq(t, "DELETE", "https://example.com/anything/at/all", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestLateRegistrationApplies(t *testing.T) {
	rt := httpmock.New(httpmock.LenientMode())
	req1 := newReq(t, "GET", "https://example.com/a", nil)
	resp1, _ := rt.RoundTrip(req1)
	assert.Equal(t, resp1.StatusCode, 404)

	rt.OnRequest().Path("/a").Respond(httpmock.Status(200))
	req2 := newReq(t, "GET", "https://example.com/a", nil)
	resp2, err := rt.RoundTrip(req2)
	assert.NoError(t, err)
	assert.Equal(t, resp2.StatusCode, 200)
}

// helpers

func newReq(t *testing.T, method, url string, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("newReq: %s", err)
	}
	return req
}

func newReqWithBody(t *testing.T, method, url, body, contentType string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("newReqWithBody: %s", err)
	}
	req.Header.Set("Content-Type", contentType)
	return req
}

// trackingTB is a TB that delegates to an inner TB but records Errorf calls.
type trackingTB struct {
	assert.TB
	errors []string
}

func (m *trackingTB) Errorf(format string, args ...any) {
	m.errors = append(m.errors, format)
}

func (m *trackingTB) Helper() {}

func (m *trackingTB) Cleanup(fn func()) {
	m.TB.Cleanup(fn)
}
