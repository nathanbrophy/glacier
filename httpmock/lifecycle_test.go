// SPDX-License-Identifier: Apache-2.0

package httpmock_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpmock"
)

func TestTransportClose(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/x").Respond(httpmock.Status(200))

	err := rt.Close()
	assert.NoError(t, err)

	// Subsequent RoundTrip must return ErrNoRouteMatch.
	req := newReq(t, "GET", "https://example.com/x", nil)
	_, err2 := rt.RoundTrip(req)
	assert.ErrorIs(t, err2, httpmock.ErrNoRouteMatch)
}

func TestTransportCloseIdempotent(t *testing.T) {
	rt := httpmock.New()
	err1 := rt.Close()
	assert.NoError(t, err1)

	err2 := rt.Close()
	assert.NoError(t, err2)
}

func TestTransportCloseRecordsRequest(t *testing.T) {
	// Requests to a closed transport are still recorded.
	rt := httpmock.New()
	rt.Close()

	req := newReq(t, "GET", "https://example.com/x", nil)
	_, _ = rt.RoundTrip(req)

	all := rt.AllRequests()
	assert.Len(t, all, 1)
}
