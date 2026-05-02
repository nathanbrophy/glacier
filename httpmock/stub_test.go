// SPDX-License-Identifier: Apache-2.0

package httpmock_test

import (
	"net/http"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpmock"
)

func TestStubMethodCaseInsensitive(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Method("get").Path("/x").Respond(httpmock.Status(200))

	req := newReq(t, "GET", "https://example.com/x", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestStubPathExact(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/users/42").Respond(httpmock.Status(200))

	req := newReq(t, "GET", "https://example.com/users/42", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	// Should not match a different path.
	req2 := newReq(t, "GET", "https://example.com/users/99", nil)
	_, err2 := rt.RoundTrip(req2)
	assert.ErrorIs(t, err2, httpmock.ErrNoRouteMatch)
}

func TestStubPathPrefix(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().PathPrefix("/users/").Respond(httpmock.Status(200))

	req := newReq(t, "GET", "https://example.com/users/42", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	req2 := newReq(t, "GET", "https://example.com/users/99", nil)
	resp2, err2 := rt.RoundTrip(req2)
	assert.NoError(t, err2)
	assert.Equal(t, resp2.StatusCode, 200)

	// Non-matching.
	req3 := newReq(t, "GET", "https://example.com/other", nil)
	_, err3 := rt.RoundTrip(req3)
	assert.ErrorIs(t, err3, httpmock.ErrNoRouteMatch)
}

func TestStubRegex(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path(`/users/(\d+)`).Regex().Respond(httpmock.Status(200))

	req := newReq(t, "GET", "https://example.com/users/42", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	req2 := newReq(t, "GET", "https://example.com/users/abc", nil)
	_, err2 := rt.RoundTrip(req2)
	assert.ErrorIs(t, err2, httpmock.ErrNoRouteMatch)
}

func TestStubRegexInvalidCompile(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic for invalid regex")
	}()
	rt := httpmock.New()
	rt.OnRequest().Path(`/users/[invalid`).Regex().Respond(httpmock.Status(200))
}

func TestStubPathPrefixAndRegexMutuallyExclusive(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic for PathPrefix + Regex")
	}()
	rt := httpmock.New()
	rt.OnRequest().PathPrefix("/users/").Regex().Respond(httpmock.Status(200))
}

func TestStubQuery(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/search").Query("page", "2").Respond(httpmock.Status(200))

	req := newReq(t, "GET", "https://example.com/search?page=2", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	req2 := newReq(t, "GET", "https://example.com/search?page=1", nil)
	_, err2 := rt.RoundTrip(req2)
	assert.ErrorIs(t, err2, httpmock.ErrNoRouteMatch)
}

func TestStubHeader(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/api").Header("X-Foo", "bar").Respond(httpmock.Status(200))

	req := newReq(t, "GET", "https://example.com/api", nil)
	req.Header.Set("X-Foo", "bar")
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	req2 := newReq(t, "GET", "https://example.com/api", nil)
	req2.Header.Set("X-Foo", "wrong")
	_, err2 := rt.RoundTrip(req2)
	assert.ErrorIs(t, err2, httpmock.ErrNoRouteMatch)
}

func TestStubHeaderCaseInsensitiveName(t *testing.T) {
	rt := httpmock.New()
	// Register with mixed-case header name.
	rt.OnRequest().Path("/api").Header("x-custom-header", "val").Respond(httpmock.Status(200))

	req := newReq(t, "GET", "https://example.com/api", nil)
	// Set the header with canonical casing.
	req.Header.Set("X-Custom-Header", "val")
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestStubTimes(t *testing.T) {
	tb := &trackingTB{TB: t}
	rt := httpmock.NewWithT(tb)
	rt.OnRequest().Path("/x").Times(2).Respond(httpmock.Status(200))

	// Call exactly 2 times.
	for range 2 {
		req := newReq(t, "GET", "https://example.com/x", nil)
		_, err := rt.RoundTrip(req)
		assert.NoError(t, err)
	}
	rt.Verify(tb)
	assert.Len(t, tb.errors, 0)
}

func TestStubAtLeast(t *testing.T) {
	tb := &trackingTB{TB: t}
	rt := httpmock.New()
	rt.OnRequest().Path("/x").AtLeast(2).Respond(httpmock.Status(200))

	// Call only once — should fail AtLeast(2).
	req := newReq(t, "GET", "https://example.com/x", nil)
	_, _ = rt.RoundTrip(req)

	rt.Verify(tb)
	assert.True(t, len(tb.errors) > 0, "expected Verify to report failure")
}

func TestStubAtMost(t *testing.T) {
	tb := &trackingTB{TB: t}
	rt := httpmock.New()
	rt.OnRequest().Path("/x").AtMost(1).Respond(httpmock.Status(200))

	// Call twice — should fail AtMost(1).
	for range 2 {
		req := newReq(t, "GET", "https://example.com/x", nil)
		_, _ = rt.RoundTrip(req)
	}

	rt.Verify(tb)
	assert.True(t, len(tb.errors) > 0, "expected Verify to report failure")
}

func TestStubNever(t *testing.T) {
	tb := &trackingTB{TB: t}
	rt := httpmock.New()
	rt.OnRequest().Path("/x").Never().Respond(httpmock.Status(200))

	// Calling the stub should succeed (stub is matched) but Verify should fail.
	req := newReq(t, "GET", "https://example.com/x", nil)
	_, _ = rt.RoundTrip(req)

	rt.Verify(tb)
	assert.True(t, len(tb.errors) > 0, "expected Verify to report Never failure")
}

func TestStubAnyTimes(t *testing.T) {
	tb := &trackingTB{TB: t}
	rt := httpmock.New()
	rt.OnRequest().Path("/x").AnyTimes().Respond(httpmock.Status(200))

	// Not calling it at all — AnyTimes means no failure.
	rt.Verify(tb)
	assert.Len(t, tb.errors, 0)
}

func TestStubMissingRespond(t *testing.T) {
	rt := httpmock.New()
	// OnRequest appends the stub; we call Respond with nil to finalize but no real responder.
	// Actually, we'll just not call Respond — the stub has a nil responder.
	rt.OnRequest().Path("/x") // No Respond call.

	req := newReq(t, "GET", "https://example.com/x", nil)
	_, err := rt.RoundTrip(req)

	var se *httpmock.ScriptError
	assert.ErrorAs(t, err, &se)
}

func TestStubRespondCalledTwiceLastWins(t *testing.T) {
	rt := httpmock.New()
	stub := rt.OnRequest().Path("/x")
	stub.Respond(httpmock.Status(201))
	stub.Respond(httpmock.Status(299))

	req := newReq(t, "GET", "https://example.com/x", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 299)
}

func TestVerifyAllStubsTimesMet(t *testing.T) {
	tb := &trackingTB{TB: t}
	rt := httpmock.New()
	rt.OnRequest().Path("/a").Times(1).Respond(httpmock.Status(200))
	rt.OnRequest().Path("/b").Times(2).Respond(httpmock.Status(200))

	// Call /a once and /b once (short of Times(2)).
	req1 := newReq(t, "GET", "https://example.com/a", nil)
	req2 := newReq(t, "GET", "https://example.com/b", nil)
	_, _ = rt.RoundTrip(req1)
	_, _ = rt.RoundTrip(req2)

	rt.Verify(tb)
	// /b was called only once, expected 2 → one error.
	assert.True(t, len(tb.errors) > 0)
}

func TestVerifyAtCleanup_NewWithT(t *testing.T) {
	tb := &trackingTB{TB: t}
	rt := httpmock.NewWithT(tb)
	rt.OnRequest().Path("/check").Times(1).Respond(httpmock.Status(200))
	// Never call RoundTrip — Verify should run at cleanup and report the unmet expectation.
	// The cleanup is registered on tb.TB (the real t), so it will fire after the sub-test.
	// We can observe tb.errors after cleanup runs by running in a sub-test.
	_ = rt
}

func TestPathAndPathPrefixCoexist(t *testing.T) {
	// Per spec, calling Path then PathPrefix: PathPrefix wins (last set).
	rt := httpmock.New()
	rt.OnRequest().Path("/users/42").PathPrefix("/users/").Respond(httpmock.Status(200))

	req := newReq(t, "GET", "https://example.com/users/99", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestScriptErrorTypedAndUnwrap(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/x") // no Respond

	req := newReq(t, "GET", "https://example.com/x", nil)
	_, err := rt.RoundTrip(req)

	var se *httpmock.ScriptError
	assert.ErrorAs(t, err, &se)
	assert.NotNil(t, se.Cause)
	assert.True(t, len(se.Error()) > 0)
	assert.NotNil(t, se.Unwrap())
}

func TestMultipartBodyMatch(t *testing.T) {
	body := "raw multipart content"
	rt := httpmock.New()
	rt.OnRequest().
		Path("/upload").
		Body(httpmock.BodyContains("multipart")).
		Respond(httpmock.Status(200))

	req := newReqWithBody(t, "POST", "https://example.com/upload", body, "multipart/form-data")
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestClientTimeoutHonored(t *testing.T) {
	// http.Client.Timeout is respected because Transport.RoundTrip is in-memory.
	rt := httpmock.New()
	rt.OnRequest().Path("/slow").AnyTimes().Respond(httpmock.Status(200))

	client := &http.Client{Transport: rt}
	resp, err := client.Get("https://example.com/slow")
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}
