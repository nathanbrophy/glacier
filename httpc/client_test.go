// SPDX-License-Identifier: Apache-2.0

package httpc_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpc"
	"github.com/nathanbrophy/glacier/httpmock"
)

// roundTripFunc is a convenience adapter for inline transport implementations.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestNewDefaultClient(t *testing.T) {
	t.Parallel()
	c := httpc.New()
	assert.NotNil(t, c)
}

func TestDefaultPackageVar(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, httpc.Default)
}

func TestWithTransport(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("GET").Path("/ping").
		Respond(httpmock.JSON(200, map[string]string{"ok": "yes"}))

	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	resp, err := c.Get(ctx, "http://example.com/ping")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	_ = resp.Drain()
}

func TestWithBaseURL(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("GET").Path("/users/1").
		Respond(httpmock.JSON(200, map[string]string{"id": "1"}))

	c := httpc.New(
		httpc.WithTransport(rt),
		httpc.WithBaseURL("http://api.example.com"),
	)
	ctx := context.Background()
	resp, err := c.Get(ctx, "/users/1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	_ = resp.Drain()
}

func TestWithBaseURLJoinSafety(t *testing.T) {
	t.Parallel()
	c := httpc.New(
		httpc.WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, nil
		})),
		httpc.WithBaseURL("http://example.com"),
	)
	ctx := context.Background()
	// Build a URL that is >8 KiB after joining.
	long := "/" + strings.Repeat("a", 9000)
	resp, err := c.Get(ctx, long)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestWithBaseURLRejectsUserinfo(t *testing.T) {
	t.Parallel()
	c := httpc.New(
		httpc.WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, nil
		})),
	)
	ctx := context.Background()
	resp, err := c.Get(ctx, "http://user:pass@example.com/path")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestWithHeaders(t *testing.T) {
	t.Parallel()
	var gotHeader string
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotHeader = r.Header.Get("X-Custom")
		return jsonResponse(200, `{}`), nil
	})
	c := httpc.New(
		httpc.WithTransport(rt),
		httpc.WithHeaders(http.Header{"X-Custom": []string{"glacier"}}),
	)
	ctx := context.Background()
	resp, err := c.Get(ctx, "http://example.com/")
	assert.NoError(t, err)
	_ = resp.Drain()
	assert.Equal(t, "glacier", gotHeader)
}

func TestWithLogger(t *testing.T) {
	t.Parallel()
	// Verify that a valid logger can be set without panic.
	logger := slog.Default()
	c := httpc.New(httpc.WithLogger(logger))
	assert.NotNil(t, c)
}
