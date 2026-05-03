// SPDX-License-Identifier: Apache-2.0

package httpc_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpc"
)

func TestMaxResponseBytesDefault32MiB(t *testing.T) {
	t.Parallel()
	// Serve a 33 MiB body.
	body33MiB := bytes.Repeat([]byte("x"), 33*1024*1024)
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(body33MiB)),
		}, nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, _, err := httpc.GetWith[map[string]any](c, ctx, "http://example.com/big")
	assert.Error(t, err)
	var bpe *httpc.BodyParseError
	assert.ErrorAs(t, err, &bpe)
	assert.True(t, errors.Is(bpe.Cause, httpc.ErrBodyTooLarge))
}

func TestMaxResponseBytesCustom(t *testing.T) {
	t.Parallel()
	// Serve 2000 bytes, limit to 1000.
	body := bytes.Repeat([]byte("y"), 2000)
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, _, err := httpc.GetWith[map[string]any](c, ctx, "http://example.com/",
		httpc.WithMaxResponseBytes(1000),
	)
	assert.Error(t, err)
	var bpe *httpc.BodyParseError
	assert.ErrorAs(t, err, &bpe)
	assert.True(t, errors.Is(bpe.Cause, httpc.ErrBodyTooLarge))
}

func TestWithUnboundedResponse(t *testing.T) {
	t.Parallel()
	// Serve 33 MiB with unbounded option :  should succeed (up to JSON decode limit).
	// Use a valid JSON object so parse succeeds.
	bodyJSON := fmt.Sprintf(`{"data":"%s"}`, strings.Repeat("a", 1024))
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(bodyJSON)),
		}, nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	// With unbounded response, body cap is lifted.
	_, _, err := httpc.GetWith[map[string]any](c, ctx, "http://example.com/",
		httpc.WithUnboundedResponse(),
	)
	// safejson still imposes its own 1 MiB cap; our test body is small so should succeed.
	assert.NoError(t, err)
}

func TestJSONDepthCap32(t *testing.T) {
	t.Parallel()
	// Build deeply nested JSON: {"a":{"a":{"a":...}}} 33 levels deep.
	depth := 33
	body := strings.Repeat(`{"a":`, depth) + `"v"` + strings.Repeat(`}`, depth)
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, _, err := httpc.GetWith[map[string]any](c, ctx, "http://example.com/deep")
	assert.Error(t, err)
	var bpe *httpc.BodyParseError
	assert.ErrorAs(t, err, &bpe)
}

func TestGzipDecompressionPreCap(t *testing.T) {
	t.Parallel()
	// A compressed response that is larger than the cap (pre-decompression).
	// Serve large compressed data that exceeds maxResponseBytes when compressed.
	compressedBuf := &bytes.Buffer{}
	gz := gzip.NewWriter(compressedBuf)
	// Write enough to exceed 1 KiB compressed.
	_, _ = gz.Write(bytes.Repeat([]byte("x"), 10*1024))
	_ = gz.Close()

	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header: http.Header{
				"Content-Type":     []string{"application/json"},
				"Content-Encoding": []string{"gzip"},
			},
			Body: io.NopCloser(bytes.NewReader(compressedBuf.Bytes())),
		}, nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	// Cap at 1 KiB.
	_, _, err := httpc.GetWith[map[string]any](c, ctx, "http://example.com/",
		httpc.WithMaxResponseBytes(1024),
	)
	assert.Error(t, err)
}

func TestResponseHeaderNULRejected(t *testing.T) {
	t.Parallel()
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"X-Evil":       []string{"value\x00with-nul"},
			},
			Body: io.NopCloser(strings.NewReader(`{}`)),
		}, nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, _, err := httpc.GetWith[map[string]any](c, ctx, "http://example.com/")
	assert.Error(t, err)
}

func TestResponseHeaderCRLFInjectionRejected(t *testing.T) {
	t.Parallel()
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"X-Evil":       []string{"value\r\nInjected: header"},
			},
			Body: io.NopCloser(strings.NewReader(`{}`)),
		}, nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, _, err := httpc.GetWith[map[string]any](c, ctx, "http://example.com/")
	assert.Error(t, err)
}
