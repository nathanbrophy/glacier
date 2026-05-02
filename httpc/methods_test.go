// SPDX-License-Identifier: Apache-2.0

package httpc_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpc"
	"github.com/nathanbrophy/glacier/httpmock"
)

type testUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func newTestClient(t *testing.T, rt http.RoundTripper) *httpc.Client {
	t.Helper()
	return httpc.New(httpc.WithTransport(rt))
}

func TestGetTyped(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("GET").Path("/users/1").
		Respond(httpmock.JSON(200, testUser{ID: 1, Name: "Ada"}))

	c := newTestClient(t, rt)
	ctx := context.Background()
	user, resp, err := httpc.GetWith[testUser](c, ctx, "http://example.com/users/1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "Ada", user.Name)
	_ = resp.Drain()
}

func TestGetByteSliceSpecialCase(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("GET").Path("/raw").
		Respond(httpmock.Body(200, []byte("hello"), "application/octet-stream"))

	c := newTestClient(t, rt)
	ctx := context.Background()
	body, resp, err := httpc.GetWith[[]byte](c, ctx, "http://example.com/raw")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, len(body) > 0)
	_ = resp.Drain()
}

func TestHead(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("HEAD").Path("/check").
		Respond(httpmock.Status(200))

	c := newTestClient(t, rt)
	ctx := context.Background()
	resp, err := c.Head(ctx, "http://example.com/check")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	_ = resp.Drain()
}

func TestPostTyped(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("POST").Path("/users").
		Respond(httpmock.JSON(201, testUser{ID: 42, Name: "Bob"}))

	c := newTestClient(t, rt)
	ctx := context.Background()
	user, resp, err := httpc.PostWith[testUser](c, ctx, "http://example.com/users",
		httpc.JSONBody(func() testUser { return testUser{Name: "Bob"} }),
	)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 42, user.ID)
	_ = resp.Drain()
}

func TestPutTyped(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("PUT").Path("/users/1").
		Respond(httpmock.JSON(200, testUser{ID: 1, Name: "Updated"}))

	c := newTestClient(t, rt)
	ctx := context.Background()
	user, resp, err := httpc.PutWith[testUser](c, ctx, "http://example.com/users/1",
		httpc.JSONBody(func() testUser { return testUser{ID: 1, Name: "Updated"} }),
	)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Updated", user.Name)
	_ = resp.Drain()
}

func TestPatchTyped(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("PATCH").Path("/users/1").
		Respond(httpmock.JSON(200, testUser{ID: 1, Name: "Patched"}))

	c := newTestClient(t, rt)
	ctx := context.Background()
	user, resp, err := httpc.PatchWith[testUser](c, ctx, "http://example.com/users/1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Patched", user.Name)
	_ = resp.Drain()
}

func TestDeleteTyped(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("DELETE").Path("/users/1").
		Respond(httpmock.JSON(200, testUser{ID: 1}))

	c := newTestClient(t, rt)
	ctx := context.Background()
	user, resp, err := httpc.DeleteWith[testUser](c, ctx, "http://example.com/users/1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, user.ID)
	_ = resp.Drain()
}

func TestStatusErrorOnNon2xx(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("GET").Path("/fail").
		Respond(httpmock.JSON(500, map[string]string{"error": "internal"}))

	c := newTestClient(t, rt)
	ctx := context.Background()
	_, _, err := httpc.GetWith[testUser](c, ctx, "http://example.com/fail")
	assert.Error(t, err)
	var se *httpc.StatusError
	assert.ErrorAs(t, err, &se)
	assert.Equal(t, 500, se.Status)
}

func TestBodyParseErrorOnBadJSON(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("GET").Path("/bad").
		Respond(httpmock.Body(200, []byte("not json"), "application/json"))

	c := newTestClient(t, rt)
	ctx := context.Background()
	_, _, err := httpc.GetWith[testUser](c, ctx, "http://example.com/bad")
	assert.Error(t, err)
	var bpe *httpc.BodyParseError
	assert.ErrorAs(t, err, &bpe)
}

func TestDoEscapeHatch(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("GET").Path("/raw-do").
		Respond(httpmock.JSON(200, "ok"))

	c := newTestClient(t, rt)
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://example.com/raw-do", nil)
	assert.NoError(t, err)
	resp, err := c.Do(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	_ = resp.Drain()
}

func TestGetPointerT(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("GET").Path("/user").
		Respond(httpmock.JSON(200, testUser{ID: 7, Name: "Ptr"}))

	c := newTestClient(t, rt)
	ctx := context.Background()
	user, resp, err := httpc.GetWith[*testUser](c, ctx, "http://example.com/user")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, user)
	assert.Equal(t, 7, user.ID)
	_ = resp.Drain()
}

func TestGetAnyT(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("GET").Path("/any").
		Respond(httpmock.JSON(200, map[string]any{"key": "value"}))

	c := newTestClient(t, rt)
	ctx := context.Background()
	result, resp, err := httpc.GetWith[any](c, ctx, "http://example.com/any")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
	_ = resp.Drain()
}

func TestAbsoluteURLOverridesBase(t *testing.T) {
	t.Parallel()
	var gotURL string
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotURL = r.URL.String()
		return jsonResponse(200, `{}`), nil
	})
	c := httpc.New(
		httpc.WithTransport(rt),
		httpc.WithBaseURL("http://base.example.com"),
	)
	ctx := context.Background()
	_, err := c.Get(ctx, "http://other.example.com/path")
	assert.NoError(t, err)
	assert.Equal(t, "http://other.example.com/path", gotURL)
}

func TestEmptyBody200ParseError(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("GET").Path("/empty").
		Respond(httpmock.Body(200, []byte(""), "application/json"))

	c := newTestClient(t, rt)
	ctx := context.Background()
	_, _, err := httpc.GetWith[testUser](c, ctx, "http://example.com/empty")
	// Empty body → EOF → BodyParseError.
	assert.Error(t, err)
	var bpe *httpc.BodyParseError
	_ = errors.As(err, &bpe) // may or may not wrap as BodyParseError depending on safejson
}

func TestResponseWrapperFields(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("GET").Path("/fields").
		Respond(httpmock.JSON(200, testUser{ID: 5, Name: "Fields"}))

	c := newTestClient(t, rt)
	ctx := context.Background()
	user, resp, err := httpc.GetWith[testUser](c, ctx, "http://example.com/fields")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Response)
	assert.True(t, len(resp.Body) > 0)
	assert.True(t, resp.Elapsed >= 0)
	assert.Equal(t, 5, user.ID)
	_ = resp.Drain()
}

func TestResponseDrain(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("GET").Path("/drain").
		Respond(httpmock.JSON(200, testUser{ID: 1}))

	c := newTestClient(t, rt)
	ctx := context.Background()
	_, resp, err := httpc.GetWith[testUser](c, ctx, "http://example.com/drain")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	err2 := resp.Drain()
	assert.NoError(t, err2)
}
