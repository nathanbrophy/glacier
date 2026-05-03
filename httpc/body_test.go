// SPDX-License-Identifier: Apache-2.0

package httpc_test

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpc"
	"github.com/nathanbrophy/glacier/httpmock"
)

func TestJSONBodyClosurePerAttempt(t *testing.T) {
	t.Parallel()
	var calls int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(503, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	httpc.PostWith[struct{}](c, ctx, "http://example.com/",
		httpc.JSONBody(func() testUser {
			calls++
			return testUser{}
		}),
		httpc.WithRetry(httpc.MaxAttempts(3)),
	)
	assert.Equal(t, 3, calls)
}

func TestMultipartBody(t *testing.T) {
	t.Parallel()
	rt := httpmock.NewWithT(t)
	rt.OnRequest().Method("POST").Path("/upload").
		Respond(httpmock.JSON(200, map[string]string{"ok": "true"}))

	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, _, err := httpc.PostWith[map[string]string](c, ctx, "http://example.com/upload",
		httpc.MultipartBody(func(w *multipart.Writer) error {
			return w.WriteField("key", "value")
		}),
	)
	assert.NoError(t, err)
}

func TestMultipartBodyClosurePerAttempt(t *testing.T) {
	t.Parallel()
	var calls int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(503, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	httpc.PostWith[struct{}](c, ctx, "http://example.com/",
		httpc.MultipartBody(func(w *multipart.Writer) error {
			calls++
			return w.WriteField("k", "v")
		}),
		httpc.WithRetry(httpc.MaxAttempts(3)),
	)
	assert.Equal(t, 3, calls)
}

func TestRawBody(t *testing.T) {
	t.Parallel()
	var gotContentType string
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotContentType = r.Header.Get("Content-Type")
		return jsonResponse(200, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, err := c.Post(ctx, "http://example.com/",
		httpc.RawBody(func() ([]byte, string, error) {
			return []byte("raw-data"), "text/plain", nil
		}),
	)
	assert.NoError(t, err)
	assert.Equal(t, "text/plain", gotContentType)
}

func TestStreamBody(t *testing.T) {
	t.Parallel()
	var bodyData string
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(r.Body)
		bodyData = string(b)
		return jsonResponse(200, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, err := c.Post(ctx, "http://example.com/",
		httpc.StreamBody(func() (io.ReadCloser, string, error) {
			return io.NopCloser(strings.NewReader("stream-data")), "application/octet-stream", nil
		}),
	)
	assert.NoError(t, err)
	assert.Equal(t, "stream-data", bodyData)
}

func TestStreamBodyCloserClosed(t *testing.T) {
	t.Parallel()
	var closedCount int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(503, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()

	httpc.PostWith[struct{}](c, ctx, "http://example.com/",
		httpc.StreamBody(func() (io.ReadCloser, string, error) {
			return &trackingReadCloser{
				Reader:  strings.NewReader("data"),
				onClose: func() { closedCount++ },
			}, "application/octet-stream", nil
		}),
		httpc.WithRetry(httpc.MaxAttempts(3)),
	)
	// Each attempt's ReadCloser should be closed before the next.
	assert.True(t, closedCount >= 2)
}

type trackingReadCloser struct {
	io.Reader
	onClose func()
}

func (t *trackingReadCloser) Close() error {
	if t.onClose != nil {
		t.onClose()
	}
	return nil
}

func TestFormBody(t *testing.T) {
	t.Parallel()
	var gotContentType string
	var gotBody string
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		return jsonResponse(200, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, err := c.Post(ctx, "http://example.com/",
		httpc.FormBody(func() url.Values {
			return url.Values{"name": []string{"Ada"}}
		}),
	)
	assert.NoError(t, err)
	assert.Equal(t, "application/x-www-form-urlencoded", gotContentType)
	assert.True(t, strings.Contains(gotBody, "Ada"))
}

func TestBodyClosureReturnsError(t *testing.T) {
	t.Parallel()
	var transportCalled bool
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		transportCalled = true
		return jsonResponse(200, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, err := c.Post(ctx, "http://example.com/",
		httpc.RawBody(func() ([]byte, string, error) {
			return nil, "", io.ErrUnexpectedEOF
		}),
	)
	assert.Error(t, err)
	assert.False(t, transportCalled)
}

func TestWithRequestHeaders(t *testing.T) {
	t.Parallel()
	var gotHeader string
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotHeader = r.Header.Get("X-Request-Id")
		return jsonResponse(200, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, err := c.Get(ctx, "http://example.com/",
		httpc.WithRequestHeaders(http.Header{"X-Request-Id": []string{"req-123"}}),
	)
	assert.NoError(t, err)
	assert.Equal(t, "req-123", gotHeader)
}
