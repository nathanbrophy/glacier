// SPDX-License-Identifier: Apache-2.0

package httpc

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/nathanbrophy/glacier/internal/safejson"
)

const (
	// defaultMaxResponseBytes is the default response body size cap (32 MiB).
	defaultMaxResponseBytes int64 = 32 * 1024 * 1024
	// maxURLBytes is the maximum allowed URL length (8 KiB).
	maxURLBytes = 8 * 1024
	// maxHeaderBytes is the maximum total response header bytes (8 KiB).
	maxHeaderBytes = 8 * 1024
	// bodySnippetSize is the maximum body bytes captured in BodyParseError.
	bodySnippetSize = 1024
)

// byteSliceType is the reflect.Type for []byte, used for the T == []byte special case.
var byteSliceType = reflect.TypeOf([]byte(nil))

// Package-level typed methods delegate to Default.

// Get sends an HTTP GET request to rawURL and decodes the response body into T.
// When T is []byte, the raw body bytes are returned without decoding.
func Get[T any](ctx context.Context, rawURL string, opts ...RequestOption) (T, *Response, error) {
	return clientGet[T](Default, ctx, rawURL, opts...)
}

// Head sends an HTTP HEAD request. No body is read or returned.
func Head(ctx context.Context, rawURL string, opts ...RequestOption) (*Response, error) {
	return Default.head(ctx, rawURL, opts...)
}

// Post sends an HTTP POST request and decodes the response body into T.
func Post[T any](ctx context.Context, rawURL string, opts ...RequestOption) (T, *Response, error) {
	return clientMethod[T](Default, ctx, http.MethodPost, rawURL, opts...)
}

// Put sends an HTTP PUT request and decodes the response body into T.
func Put[T any](ctx context.Context, rawURL string, opts ...RequestOption) (T, *Response, error) {
	return clientMethod[T](Default, ctx, http.MethodPut, rawURL, opts...)
}

// Patch sends an HTTP PATCH request and decodes the response body into T.
func Patch[T any](ctx context.Context, rawURL string, opts ...RequestOption) (T, *Response, error) {
	return clientMethod[T](Default, ctx, http.MethodPatch, rawURL, opts...)
}

// Delete sends an HTTP DELETE request and decodes the response body into T.
func Delete[T any](ctx context.Context, rawURL string, opts ...RequestOption) (T, *Response, error) {
	return clientMethod[T](Default, ctx, http.MethodDelete, rawURL, opts...)
}

// Do sends a raw *http.Request. No body is auto-read or decoded.
func Do(ctx context.Context, req *http.Request) (*Response, error) {
	return Default.Do(ctx, req)
}

// GetWith sends an HTTP GET request using the supplied *Client and decodes
// the response body into T. This enables typed GET on an explicit client,
// working around Go's restriction on generic methods.
func GetWith[T any](c *Client, ctx context.Context, rawURL string, opts ...RequestOption) (T, *Response, error) {
	return clientGet[T](c, ctx, rawURL, opts...)
}

// PostWith sends an HTTP POST request using the supplied *Client and decodes
// the response body into T.
func PostWith[T any](c *Client, ctx context.Context, rawURL string, opts ...RequestOption) (T, *Response, error) {
	return clientMethod[T](c, ctx, http.MethodPost, rawURL, opts...)
}

// PutWith sends an HTTP PUT request using the supplied *Client and decodes
// the response body into T.
func PutWith[T any](c *Client, ctx context.Context, rawURL string, opts ...RequestOption) (T, *Response, error) {
	return clientMethod[T](c, ctx, http.MethodPut, rawURL, opts...)
}

// PatchWith sends an HTTP PATCH request using the supplied *Client and decodes
// the response body into T.
func PatchWith[T any](c *Client, ctx context.Context, rawURL string, opts ...RequestOption) (T, *Response, error) {
	return clientMethod[T](c, ctx, http.MethodPatch, rawURL, opts...)
}

// DeleteWith sends an HTTP DELETE request using the supplied *Client and decodes
// the response body into T.
func DeleteWith[T any](c *Client, ctx context.Context, rawURL string, opts ...RequestOption) (T, *Response, error) {
	return clientMethod[T](c, ctx, http.MethodDelete, rawURL, opts...)
}

// Methods on *Client.

// Get sends an HTTP GET request to rawURL and decodes the response body into T.
func (c *Client) Get(ctx context.Context, rawURL string, opts ...RequestOption) (*Response, error) {
	_, resp, err := clientGet[struct{}](c, ctx, rawURL, opts...)
	return resp, err
}

// GetTyped is a convenience wrapper :  external callers use the package-level Get[T].
// Internal use: clientGet[T](c, ctx, url, opts...) directly.

// Head sends an HTTP HEAD request.
func (c *Client) Head(ctx context.Context, rawURL string, opts ...RequestOption) (*Response, error) {
	return c.head(ctx, rawURL, opts...)
}

// Post sends an HTTP POST request and decodes the response body into T.
func (c *Client) Post(ctx context.Context, rawURL string, opts ...RequestOption) (*Response, error) {
	_, resp, err := clientMethod[struct{}](c, ctx, http.MethodPost, rawURL, opts...)
	return resp, err
}

// Put sends an HTTP PUT request and decodes the response body into T.
func (c *Client) Put(ctx context.Context, rawURL string, opts ...RequestOption) (*Response, error) {
	_, resp, err := clientMethod[struct{}](c, ctx, http.MethodPut, rawURL, opts...)
	return resp, err
}

// Patch sends an HTTP PATCH request and decodes the response body into T.
func (c *Client) Patch(ctx context.Context, rawURL string, opts ...RequestOption) (*Response, error) {
	_, resp, err := clientMethod[struct{}](c, ctx, http.MethodPatch, rawURL, opts...)
	return resp, err
}

// Delete sends an HTTP DELETE request and decodes the response body into T.
func (c *Client) Delete(ctx context.Context, rawURL string, opts ...RequestOption) (*Response, error) {
	_, resp, err := clientMethod[struct{}](c, ctx, http.MethodDelete, rawURL, opts...)
	return resp, err
}

// Do sends a raw *http.Request. No body is auto-read or decoded.
// Retry, dry-run, base URL joining, and client headers do NOT apply.
func (c *Client) Do(ctx context.Context, req *http.Request) (*Response, error) {
	start := time.Now()
	httpResp, err := c.cfg.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	return &Response{
		Response: httpResp,
		Elapsed:  time.Since(start),
	}, nil
}

// head is the unexported HEAD implementation.
func (c *Client) head(ctx context.Context, rawURL string, opts ...RequestOption) (*Response, error) {
	_, resp, err := clientMethod[struct{}](c, ctx, http.MethodHead, rawURL, opts...)
	return resp, err
}

// clientGet is the typed GET entry point, callable with any T.
func clientGet[T any](c *Client, ctx context.Context, rawURL string, opts ...RequestOption) (T, *Response, error) {
	return clientMethod[T](c, ctx, http.MethodGet, rawURL, opts...)
}

// clientMethod is the core typed method implementation.
func clientMethod[T any](c *Client, ctx context.Context, method, rawURL string, opts ...RequestOption) (T, *Response, error) {
	var zero T

	// Apply per-call request options.
	var rcfg requestConfig
	for _, o := range opts {
		if o == nil {
			continue
		}
		if err := o.applyRequest(&rcfg); err != nil {
			return zero, nil, err
		}
	}

	// Resolve effective response byte cap.
	maxBytes := defaultMaxResponseBytes
	if rcfg.unbounded {
		maxBytes = 0
	} else if rcfg.maxResponseBytes > 0 {
		maxBytes = rcfg.maxResponseBytes
	}

	// Dry-run path.
	if dryCfg, ok := ctx.Value(dryRunKey{}).(*dryRunConfig); ok {
		return handleDryRun[T](c, ctx, method, rawURL, &rcfg, dryCfg)
	}

	// Build the effective URL.
	resolvedURL, err := c.resolveURL(rawURL)
	if err != nil {
		return zero, nil, err
	}

	// Build the base request.
	req, err := http.NewRequestWithContext(ctx, method, resolvedURL, http.NoBody)
	if err != nil {
		return zero, nil, fmt.Errorf("httpc: build request: %w", err)
	}

	// Apply client-level headers.
	for k, vs := range c.cfg.headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	// Apply per-call headers (merge on top).
	for k, vs := range rcfg.headers {
		req.Header[k] = vs
	}

	// Apply per-request timeout via context.
	reqCtx := ctx
	if c.cfg.timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, c.cfg.timeout)
		defer cancel()
		req = req.WithContext(reqCtx)
	}

	// Merge retry config.
	retryCfg := mergeRetryConfig(c.cfg.retry, rcfg.retry)

	// Execute with retry.
	resp, err := c.doWithRetry(reqCtx, req, &retryCfg, rcfg.bodyFn)
	if err != nil {
		return zero, resp, err
	}

	// HEAD: no body to read.
	if method == http.MethodHead {
		return zero, resp, nil
	}

	// Validate response headers.
	if err := validateResponseHeaders(resp.Response); err != nil {
		_ = resp.Drain()
		return zero, resp, err
	}

	// Read and cap the body.
	body, err := readBody(resp.Response, maxBytes)
	if err != nil {
		_ = resp.Drain()
		if errors.Is(err, ErrBodyTooLarge) {
			ct := resp.Response.Header.Get("Content-Type")
			return zero, resp, &BodyParseError{Cause: err, ContentType: ct}
		}
		return zero, resp, err
	}
	resp.Body = body

	// Check status after reading body (so Body field is populated in StatusError).
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return zero, resp, &StatusError{
			Status: resp.StatusCode,
			Body:   body,
		}
	}

	// T == []byte special case: return raw bytes.
	if reflect.TypeOf(zero) == byteSliceType {
		result := reflect.New(reflect.TypeOf(zero)).Elem()
		result.Set(reflect.ValueOf(body))
		return result.Interface().(T), resp, nil
	}

	// JSON decode via safejson.
	ct := resp.Response.Header.Get("Content-Type")
	var result T
	decErr := safejson.Decode(strings.NewReader(string(body)), &result)
	if decErr != nil {
		snippet := body
		if len(snippet) > bodySnippetSize {
			snippet = snippet[:bodySnippetSize]
		}
		return zero, resp, &BodyParseError{
			Cause:       decErr,
			Body:        snippet,
			ContentType: ct,
		}
	}
	return result, resp, nil
}

// handleDryRun emits a RequestPlan and returns without making a network call.
func handleDryRun[T any](c *Client, ctx context.Context, method, rawURL string, rcfg *requestConfig, dryCfg *dryRunConfig) (T, *Response, error) {
	var zero T

	resolvedURL, err := c.resolveURL(rawURL)
	if err != nil {
		return zero, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, resolvedURL, nil)
	if err != nil {
		return zero, nil, fmt.Errorf("httpc: dry run build request: %w", err)
	}

	// Apply headers.
	for k, vs := range c.cfg.headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	for k, vs := range rcfg.headers {
		req.Header[k] = vs
	}

	// Scrub headers unless secrets are included.
	if !dryCfg.includeSecrets {
		req.Header = scrubHeaders(req.Header)
	}

	// Capture body bytes if body closure is present.
	var bodyBytes []byte
	if rcfg.bodyFn != nil {
		rc, _, bErr := rcfg.bodyFn()
		if bErr == nil && rc != nil {
			bodyBytes, _ = io.ReadAll(rc)
			_ = rc.Close()
		}
	}

	retryCfg := mergeRetryConfig(c.cfg.retry, rcfg.retry)

	plan := &RequestPlan{
		Request: req,
		Body:    bodyBytes,
		Retry:   retryCfg,
		Timeout: c.cfg.timeout,
	}

	sink := dryCfg.sink
	if sink == nil {
		sink = defaultPlanSink
	}
	sink(plan)

	if dryCfg.returnErrors {
		return zero, nil, ErrDryRun
	}
	return zero, nil, nil
}

// resolveURL applies the base URL (if any), rejects userinfo, and checks length.
func (c *Client) resolveURL(rawURL string) (string, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("httpc: parse url: %w", err)
	}

	if c.cfg.baseURL != "" {
		base, err := url.Parse(c.cfg.baseURL)
		if err != nil {
			return "", fmt.Errorf("httpc: parse base url: %w", err)
		}
		target = base.ResolveReference(target)
	}

	if target.User != nil {
		return "", fmt.Errorf("httpc: url contains userinfo (credentials in url are not allowed)")
	}

	resolved := target.String()
	if len(resolved) > maxURLBytes {
		return "", fmt.Errorf("httpc: url exceeds 8 kib limit (%d bytes)", len(resolved))
	}
	return resolved, nil
}

// validateResponseHeaders checks for security-relevant header violations.
func validateResponseHeaders(resp *http.Response) error {
	totalBytes := 0
	for name, vals := range resp.Header {
		totalBytes += len(name)
		for _, v := range vals {
			totalBytes += len(v)
			if strings.ContainsRune(v, 0) {
				return fmt.Errorf("httpc: response header %q contains nul byte", name)
			}
			if strings.ContainsAny(v, "\r\n") {
				return fmt.Errorf("httpc: response header %q contains crlf", name)
			}
		}
	}
	if totalBytes > maxHeaderBytes {
		return fmt.Errorf("httpc: response headers exceed 8 kib (%d bytes)", totalBytes)
	}
	return nil
}

// readBody reads the response body, handling gzip and applying the size cap.
func readBody(resp *http.Response, maxBytes int64) ([]byte, error) {
	body := resp.Body
	if body == nil {
		return nil, nil
	}
	defer body.Close()

	var reader io.Reader = body

	// Apply pre-decompression cap (cap + 1 so we can detect overflow).
	if maxBytes > 0 {
		reader = io.LimitReader(reader, maxBytes+1)
	}

	// Handle gzip content encoding.
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("httpc: gzip reader: %w", err)
		}
		defer gz.Close()
		// Apply post-decompression cap.
		if maxBytes > 0 {
			reader = io.LimitReader(gz, maxBytes+1)
		} else {
			reader = gz
		}
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("httpc: read body: %w", err)
	}

	if maxBytes > 0 && int64(len(data)) > maxBytes {
		return nil, ErrBodyTooLarge
	}

	return data, nil
}
