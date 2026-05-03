// SPDX-License-Identifier: Apache-2.0

package httpc

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/nathanbrophy/glacier/errs"
	"github.com/nathanbrophy/glacier/option"
)

// clientConfig holds the mutable configuration built by option.Option[clientConfig] functions.
type clientConfig struct {
	// transport is never nil after New(); defaults to http.DefaultTransport.
	transport http.RoundTripper
	// timeout == 0 means no per-request timeout.
	timeout time.Duration
	// baseURL is either empty or a valid absolute URL.
	baseURL string
	// headers is never nil after New(); may be empty.
	headers http.Header
	// retry is the client-level default retry policy.
	retry retryConfig
	// logger is never nil after New(); defaults to slog.Default().
	logger *slog.Logger
	// ownsTransport is true only when New() constructs the transport itself.
	ownsTransport bool
}

// Client is a configured HTTP client. The zero value is not usable; construct
// via New. A single Client is goroutine-safe: concurrent calls share the
// underlying transport.
type Client struct {
	cfg       clientConfig
	closeOnce sync.Once
}

// Default is the package-level shared Client, equivalent to New() with all
// defaults. Package-level functions (Get, Post, etc.) delegate to Default.
// Replace Default in tests by setting httpc.Default = httpc.New(httpc.WithTransport(rt)).
var Default = New()

// New constructs a Client from the given options. If no WithTransport option
// is provided, New uses http.DefaultTransport and owns the transport for
// Close purposes. All options are applied in order; later options override
// earlier ones for scalar fields.
//
// Panics if an option returns an error (programming error).
func New(opts ...option.Option[clientConfig]) *Client {
	cfg, err := option.Apply(opts)
	if err != nil {
		//glacier:nolint=panic-in-library programmer error: option misuse surfaces at construction; documented in func doc.
		panic("httpc: New: " + err.Error())
	}
	// Apply defaults.
	if cfg.transport == nil {
		cfg.transport = http.DefaultTransport
		cfg.ownsTransport = true
	}
	if cfg.headers == nil {
		cfg.headers = make(http.Header)
	}
	if cfg.logger == nil {
		cfg.logger = slog.Default()
	}
	if cfg.retry.maxAttempts == 0 {
		cfg.retry.maxAttempts = 1
	}
	return &Client{cfg: cfg}
}

// Close releases resources held by the client. If the client owns its
// transport (i.e., no WithTransport option was provided to New), Close closes
// that transport. Close is idempotent.
func (c *Client) Close() error {
	var closeErr error
	c.closeOnce.Do(func() {
		if c.cfg.ownsTransport {
			type closer interface{ CloseIdleConnections() }
			type ioCloser interface{ Close() error }
			if cl, ok := c.cfg.transport.(ioCloser); ok {
				closeErr = errs.Join(closeErr, cl.Close())
			}
		}
	})
	return closeErr
}

// WithTransport sets the HTTP transport for the client.
func WithTransport(rt http.RoundTripper) option.Option[clientConfig] {
	return option.OptionFunc[clientConfig](func(c *clientConfig) error {
		c.transport = rt
		c.ownsTransport = false
		return nil
	})
}

// WithTimeout sets the per-request deadline, applied via context.WithTimeout
// before each attempt.
func WithTimeout(d time.Duration) option.Option[clientConfig] {
	return option.OptionFunc[clientConfig](func(c *clientConfig) error {
		c.timeout = d
		return nil
	})
}

// WithBaseURL sets a base URL prepended to relative URLs in method calls.
func WithBaseURL(rawURL string) option.Option[clientConfig] {
	return option.OptionFunc[clientConfig](func(c *clientConfig) error {
		c.baseURL = rawURL
		return nil
	})
}

// WithHeaders sets headers sent on every request.
func WithHeaders(h http.Header) option.Option[clientConfig] {
	return option.OptionFunc[clientConfig](func(c *clientConfig) error {
		c.headers = h.Clone()
		return nil
	})
}

// WithLogger sets the structured logger for lifecycle events.
func WithLogger(l *slog.Logger) option.Option[clientConfig] {
	return option.OptionFunc[clientConfig](func(c *clientConfig) error {
		c.logger = l
		return nil
	})
}
