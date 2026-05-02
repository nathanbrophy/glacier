// SPDX-License-Identifier: Apache-2.0

package httpmock_test

import (
	"fmt"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpmock"
)

func TestPropertySequenceCycleAlgebraic(t *testing.T) {
	// Property: Sequence(a, b, c).Cycle — first 6 calls = [a,b,c,a,b,c].
	statuses := []int{200, 201, 202}
	responders := make([]httpmock.Responder, len(statuses))
	for i, s := range statuses {
		responders[i] = httpmock.Status(s)
	}

	rt := httpmock.New()
	rt.OnRequest().Path("/x").AnyTimes().Respond(httpmock.SequenceCycle(responders...))

	expected := []int{200, 201, 202, 200, 201, 202}
	for i, want := range expected {
		req := newReq(t, "GET", "https://example.com/x", nil)
		resp, err := rt.RoundTrip(req)
		assert.NoError(t, err)
		assert.True(t, resp.StatusCode == want, fmt.Sprintf("call %d: got %d, want %d", i, resp.StatusCode, want))
	}
}

func TestPropertyAllRequestsOrdered(t *testing.T) {
	// Property: AllRequests returns requests in arrival order.
	rt := httpmock.New(httpmock.LenientMode())
	paths := []string{"/a", "/b", "/c", "/d", "/e"}
	for _, p := range paths {
		req := newReq(t, "GET", "https://example.com"+p, nil)
		_, _ = rt.RoundTrip(req)
	}

	all := rt.AllRequests()
	assert.Len(t, all, len(paths))
	for i, r := range all {
		assert.Equal(t, r.URL.Path, paths[i])
	}
}

func TestPropertyRequestsToIsSubset(t *testing.T) {
	// Property: RequestsTo result is always a subset of AllRequests.
	rt := httpmock.New(httpmock.LenientMode())
	for _, p := range []string{"/a", "/b/1", "/b/2", "/c"} {
		req := newReq(t, "GET", "https://example.com"+p, nil)
		_, _ = rt.RoundTrip(req)
	}

	all := rt.AllRequests()
	sub := rt.RequestsTo("/b/*")
	assert.True(t, len(sub) <= len(all))
	assert.Len(t, sub, 2)
}

func TestPropertyCloseIsIdempotent(t *testing.T) {
	rt := httpmock.New()
	for range 5 {
		err := rt.Close()
		assert.NoError(t, err)
	}
}

func TestPropertyBodyExactSymmetric(t *testing.T) {
	// BodyExact(b).Match(b, ...) is always true.
	cases := [][]byte{
		nil,
		{},
		[]byte("hello"),
		[]byte(`{"key":"value"}`),
	}
	for _, b := range cases {
		m := httpmock.BodyExact(b)
		if len(b) == 0 {
			// nil and empty are both represented as empty.
			assert.True(t, m.Match(b, "") || m.Match(nil, ""))
		} else {
			assert.True(t, m.Match(b, ""))
		}
	}
}
