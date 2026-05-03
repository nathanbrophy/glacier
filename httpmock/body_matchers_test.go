// SPDX-License-Identifier: Apache-2.0

package httpmock_test

import (
	"strings"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpmock"
)

func TestBodyExact(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/x").Body(httpmock.BodyExact([]byte("hello"))).Respond(httpmock.Status(200))

	req := newReqWithBody(t, "POST", "https://example.com/x", "hello", "text/plain")
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	req2 := newReqWithBody(t, "POST", "https://example.com/x", "world", "text/plain")
	_, err2 := rt.RoundTrip(req2)
	assert.ErrorIs(t, err2, httpmock.ErrNoRouteMatch)
}

func TestBodyContains(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/x").Body(httpmock.BodyContains("ello")).Respond(httpmock.Status(200))

	req := newReqWithBody(t, "POST", "https://example.com/x", "hello world", "text/plain")
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	req2 := newReqWithBody(t, "POST", "https://example.com/x", "goodbye", "text/plain")
	_, err2 := rt.RoundTrip(req2)
	assert.ErrorIs(t, err2, httpmock.ErrNoRouteMatch)
}

func TestBodyMatchFn(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/x").Body(httpmock.BodyMatchFn(func(b []byte) bool {
		return strings.Contains(string(b), "magic")
	})).Respond(httpmock.Status(200))

	req := newReqWithBody(t, "POST", "https://example.com/x", "magic word", "text/plain")
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	req2 := newReqWithBody(t, "POST", "https://example.com/x", "no keyword here", "text/plain")
	_, err2 := rt.RoundTrip(req2)
	assert.ErrorIs(t, err2, httpmock.ErrNoRouteMatch)
}

func TestBodyJSONSmartEqual(t *testing.T) {
	type CreateRequest struct {
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
	}

	rt := httpmock.New()
	rt.OnRequest().
		Path("/users").
		Body(httpmock.BodyJSON(
			CreateRequest{Name: "Ada"},
			assert.IgnoreFields("CreatedAt"),
		)).
		Respond(httpmock.Status(201))

	// Body has a different CreatedAt :  should still match because we ignore it.
	body := `{"name":"Ada","created_at":"2024-01-01T00:00:00Z"}`
	req := newReqWithBody(t, "POST", "https://example.com/users", body, "application/json")
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 201)

	// Different name :  should NOT match.
	body2 := `{"name":"Grace","created_at":"2024-01-01T00:00:00Z"}`
	req2 := newReqWithBody(t, "POST", "https://example.com/users", body2, "application/json")
	_, err2 := rt.RoundTrip(req2)
	assert.ErrorIs(t, err2, httpmock.ErrNoRouteMatch)
}

func TestBodyJSONNoOpts(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	rt := httpmock.New()
	rt.OnRequest().
		Path("/item").
		Body(httpmock.BodyJSON(Item{ID: 1, Name: "foo"})).
		Respond(httpmock.Status(200))

	req := newReqWithBody(t, "POST", "https://example.com/item", `{"id":1,"name":"foo"}`, "application/json")
	resp, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestBodyMatchersString(t *testing.T) {
	// Verify String() methods return non-empty strings.
	exact := httpmock.BodyExact([]byte("test"))
	assert.True(t, len(exact.String()) > 0)

	contains := httpmock.BodyContains("x")
	assert.True(t, len(contains.String()) > 0)

	fn := httpmock.BodyMatchFn(func([]byte) bool { return true })
	assert.True(t, len(fn.String()) > 0)

	type S struct{ V int }
	jm := httpmock.BodyJSON(S{V: 1})
	assert.True(t, len(jm.String()) > 0)
}
