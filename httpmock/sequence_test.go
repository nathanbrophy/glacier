// SPDX-License-Identifier: Apache-2.0

package httpmock_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpmock"
)

func TestSequenceBasic(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/x").AnyTimes().Respond(httpmock.Sequence(
		httpmock.Status(200),
		httpmock.Status(201),
		httpmock.Status(202),
	))

	statuses := []int{200, 201, 202}
	for _, want := range statuses {
		req := newReq(t, "GET", "https://example.com/x", nil)
		resp, err := rt.RoundTrip(req)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, want)
	}
}

func TestSequenceCycle(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/x").AnyTimes().Respond(httpmock.SequenceCycle(
		httpmock.Status(200),
		httpmock.Status(201),
	))

	// 4 calls: should cycle [200, 201, 200, 201].
	expected := []int{200, 201, 200, 201}
	for _, want := range expected {
		req := newReq(t, "GET", "https://example.com/x", nil)
		resp, err := rt.RoundTrip(req)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, want)
	}
}

func TestSequenceExhaust(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/x").AnyTimes().Respond(httpmock.SequenceExhaust(
		httpmock.Status(200),
		httpmock.Status(201),
	))

	// First two succeed.
	for _, want := range []int{200, 201} {
		req := newReq(t, "GET", "https://example.com/x", nil)
		resp, err := rt.RoundTrip(req)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, want)
	}

	// Third call: sequence exhausted.
	req := newReq(t, "GET", "https://example.com/x", nil)
	_, err := rt.RoundTrip(req)
	assert.Error(t, err)
}

func TestSequenceSingleElementCycles(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/x").AnyTimes().Respond(httpmock.Sequence(httpmock.Status(200)))

	for range 5 {
		req := newReq(t, "GET", "https://example.com/x", nil)
		resp, err := rt.RoundTrip(req)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)
	}
}

func TestPropertySequenceCycleSixCalls(t *testing.T) {
	a := httpmock.Status(200)
	b := httpmock.Status(201)
	c := httpmock.Status(202)

	rt := httpmock.New()
	rt.OnRequest().Path("/x").AnyTimes().Respond(httpmock.SequenceCycle(a, b, c))

	expected := []int{200, 201, 202, 200, 201, 202}
	for i, want := range expected {
		req := newReq(t, "GET", "https://example.com/x", nil)
		resp, err := rt.RoundTrip(req)
		assert.NoError(t, err)
		if resp.StatusCode != want {
			t.Errorf("call %d: got %d, want %d", i, resp.StatusCode, want)
		}
	}
}

func TestSequenceEmptyPanics(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic for empty Sequence")
	}()
	_ = httpmock.Sequence()
}

func TestSequenceExhaustEmptyPanics(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic for empty SequenceExhaust")
	}()
	_ = httpmock.SequenceExhaust()
}

func TestSequenceExhaustError(t *testing.T) {
	// Use Error responder within a sequence.
	rt := httpmock.New()
	rt.OnRequest().Path("/x").AnyTimes().Respond(httpmock.SequenceExhaust(
		httpmock.Error(errors.New("fail")),
	))

	req := newReq(t, "GET", "https://example.com/x", nil)
	_, err := rt.RoundTrip(req)
	assert.Error(t, err)

	// Second call: exhausted.
	req2 := newReq(t, "GET", "https://example.com/x", nil)
	_, err2 := rt.RoundTrip(req2)
	assert.Error(t, err2)
}

func TestSequenceWithRetry(t *testing.T) {
	type LoginResponse struct {
		Token string `json:"token"`
	}
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("POST").Path("/login").Times(3).Respond(
		httpmock.SequenceExhaust(
			httpmock.Status(http.StatusServiceUnavailable),
			httpmock.Status(http.StatusServiceUnavailable),
			httpmock.JSON(200, LoginResponse{Token: "tok-abc123"}),
		),
	)

	for i := range 3 {
		req := newReq(t, "POST", "https://example.com/login", nil)
		resp, err := rt.RoundTrip(req)
		assert.NoError(t, err)
		if i < 2 {
			assert.Equal(t, resp.StatusCode, http.StatusServiceUnavailable)
		} else {
			assert.Equal(t, resp.StatusCode, 200)
		}
	}
}
