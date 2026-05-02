// SPDX-License-Identifier: Apache-2.0

package httpc

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

// requestConfig holds per-call options accumulated from RequestOption values.
type requestConfig struct {
	// bodyFn is nil when no body builder is provided (e.g., GET).
	bodyFn func() (io.ReadCloser, string, error)
	// headers is merged on top of client-level headers.
	headers http.Header
	// retry overrides or merges with the client-level retryConfig.
	retry *retryConfig
	// maxResponseBytes overrides the default 32 MiB cap; 0 means use default.
	maxResponseBytes int64
	// unbounded removes the response body size cap.
	unbounded bool
}

// RequestOption modifies a per-call request configuration.
// Implementations are sealed within the httpc package.
type RequestOption interface{ applyRequest(*requestConfig) error }

type requestOptionFunc func(*requestConfig) error

func (f requestOptionFunc) applyRequest(c *requestConfig) error { return f(c) }

// JSONBody returns a RequestOption that sets Content-Type: application/json
// and produces the request body by calling gen() on each attempt and
// JSON-marshaling the result.
func JSONBody[T any](gen func() T) RequestOption {
	return requestOptionFunc(func(c *requestConfig) error {
		c.bodyFn = func() (io.ReadCloser, string, error) {
			v := gen()
			b, err := json.Marshal(v)
			if err != nil {
				return nil, "", err
			}
			return io.NopCloser(bytes.NewReader(b)), "application/json", nil
		}
		return nil
	})
}

// MultipartBody returns a RequestOption that sets Content-Type:
// multipart/form-data with a fresh boundary per attempt.
func MultipartBody(gen func(*multipart.Writer) error) RequestOption {
	return requestOptionFunc(func(c *requestConfig) error {
		c.bodyFn = func() (io.ReadCloser, string, error) {
			var buf bytes.Buffer
			w := multipart.NewWriter(&buf)
			if err := gen(w); err != nil {
				return nil, "", err
			}
			if err := w.Close(); err != nil {
				return nil, "", err
			}
			return io.NopCloser(bytes.NewReader(buf.Bytes())), w.FormDataContentType(), nil
		}
		return nil
	})
}

// RawBody returns a RequestOption whose closure returns the body bytes,
// content-type string, and an error.
func RawBody(gen func() ([]byte, string, error)) RequestOption {
	return requestOptionFunc(func(c *requestConfig) error {
		c.bodyFn = func() (io.ReadCloser, string, error) {
			b, ct, err := gen()
			if err != nil {
				return nil, "", err
			}
			return io.NopCloser(bytes.NewReader(b)), ct, nil
		}
		return nil
	})
}

// StreamBody returns a RequestOption for very large bodies. gen returns a
// fresh io.ReadCloser, content-type, and error per attempt.
func StreamBody(gen func() (io.ReadCloser, string, error)) RequestOption {
	return requestOptionFunc(func(c *requestConfig) error {
		c.bodyFn = gen
		return nil
	})
}

// FormBody returns a RequestOption that sets Content-Type:
// application/x-www-form-urlencoded.
func FormBody(gen func() url.Values) RequestOption {
	return requestOptionFunc(func(c *requestConfig) error {
		c.bodyFn = func() (io.ReadCloser, string, error) {
			vals := gen()
			encoded := vals.Encode()
			return io.NopCloser(strings.NewReader(encoded)), "application/x-www-form-urlencoded", nil
		}
		return nil
	})
}

// WithRequestHeaders returns a RequestOption that merges h on top of the
// client-level headers for this call only.
func WithRequestHeaders(h http.Header) RequestOption {
	return requestOptionFunc(func(c *requestConfig) error {
		if c.headers == nil {
			c.headers = h.Clone()
			return nil
		}
		for k, vs := range h {
			c.headers[k] = vs
		}
		return nil
	})
}

// WithRetry returns a value that serves as both a per-call RequestOption and a
// client-level option.Option[clientConfig]. When used in httpc.New(), it sets
// the default retry policy for all requests. When used as a RequestOption in a
// method call, it overrides the client-level default for that call only.
func WithRetry(opts ...RetryOption) *retryCallOption {
	return &retryCallOption{opts: opts}
}

// retryCallOption implements both RequestOption and option.Option[clientConfig],
// allowing WithRetry to be used in New() and in per-call opts.
type retryCallOption struct {
	opts []RetryOption
}

func (r *retryCallOption) applyRequest(c *requestConfig) error {
	if c.retry == nil {
		c.retry = &retryConfig{}
	}
	return applyRetryOptions(c.retry, r.opts)
}

func (r *retryCallOption) apply(c *clientConfig) error {
	return applyRetryOptions(&c.retry, r.opts)
}

// WithMaxResponseBytes returns a RequestOption that overrides the default
// 32 MiB response-body cap for this call.
func WithMaxResponseBytes(n int64) RequestOption {
	return requestOptionFunc(func(c *requestConfig) error {
		c.maxResponseBytes = n
		return nil
	})
}

// WithUnboundedResponse returns a RequestOption that removes the default
// response-body size cap.
func WithUnboundedResponse() RequestOption {
	return requestOptionFunc(func(c *requestConfig) error {
		c.unbounded = true
		return nil
	})
}
