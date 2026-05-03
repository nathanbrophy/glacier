// SPDX-License-Identifier: Apache-2.0

package httpc

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"
	"time"
)

// dryRunKey is the unexported context key for dry-run configuration.
type dryRunKey struct{}

// dryRunConfig carries the options set by WithDryRun.
type dryRunConfig struct {
	// sink is never nil; defaults to a slog debug-level emitter.
	sink func(*RequestPlan)
	// returnErrors causes ErrDryRun to be returned instead of nil.
	returnErrors bool
	// includeSecrets disables header redaction when true.
	includeSecrets bool
}

// DryRunOption modifies the dry-run configuration.
// Implementations are sealed within the httpc package.
type DryRunOption interface{ applyDryRun(*dryRunConfig) error }

type dryRunOptionFunc func(*dryRunConfig) error

func (f dryRunOptionFunc) applyDryRun(c *dryRunConfig) error { return f(c) }

// WithDryRun derives a context that makes every subsequent httpc call in that
// context skip the network. Instead of sending a request, the call emits a
// *RequestPlan to the configured sink and returns immediately.
func WithDryRun(ctx context.Context, opts ...DryRunOption) context.Context {
	cfg := &dryRunConfig{
		sink: defaultPlanSink,
	}
	for _, o := range opts {
		if o == nil {
			continue
		}
		_ = o.applyDryRun(cfg)
	}
	return context.WithValue(ctx, dryRunKey{}, cfg)
}

// WithPlanSink registers a function that receives every *RequestPlan emitted
// during dry-run. The default sink emits a slog debug event.
func WithPlanSink(fn func(*RequestPlan)) DryRunOption {
	return dryRunOptionFunc(func(c *dryRunConfig) error {
		c.sink = fn
		return nil
	})
}

// WithDryRunErrors causes typed methods in dry-run mode to return ErrDryRun
// instead of nil error.
func WithDryRunErrors() DryRunOption {
	return dryRunOptionFunc(func(c *dryRunConfig) error {
		c.returnErrors = true
		return nil
	})
}

// WithPlanIncludeSecrets opts into raw header values in the *RequestPlan.
// By default, headers matching the redaction list are replaced with "[REDACTED]".
func WithPlanIncludeSecrets() DryRunOption {
	return dryRunOptionFunc(func(c *dryRunConfig) error {
		c.includeSecrets = true
		return nil
	})
}

// IsDryRun reports whether ctx carries a dry-run attribute.
func IsDryRun(ctx context.Context) bool {
	_, ok := ctx.Value(dryRunKey{}).(*dryRunConfig)
	return ok
}

// RequestPlan is the audit record produced during dry-run mode.
type RequestPlan struct {
	// Request is the fully-prepared *http.Request (URL joined, headers applied).
	Request *http.Request
	// Body holds the bytes the body closure would have produced. Nil for HEAD/GET.
	Body []byte
	// Retry is a copy of the effective retry policy for this request.
	Retry retryConfig
	// Timeout is the effective per-request timeout (0 means none).
	Timeout time.Duration
}

// defaultPlanSink emits a slog debug event with structured plan fields.
func defaultPlanSink(p *RequestPlan) {
	slog.Default().Debug("httpc: dry run",
		"method", p.Request.Method,
		"url", p.Request.URL.String(),
		"timeout", p.Timeout,
		"retry.maxAttempts", p.Retry.maxAttempts,
	)
}

// sensitiveHeaderNames lists exact canonical header names to redact.
var sensitiveHeaderNames = map[string]bool{
	"Authorization":       true,
	"Cookie":              true,
	"Set-Cookie":          true,
	"X-Api-Key":           true,
	"X-Auth-Token":        true,
	"Proxy-Authorization": true,
}

// sensitiveHeaderRegex matches header names containing auth/key/token/cookie/secret patterns.
var sensitiveHeaderRegex = regexp.MustCompile(`(?i)auth|key|token|cookie|secret`)

// scrubHeaders returns a copy of h with sensitive values replaced by "[REDACTED]".
func scrubHeaders(h http.Header) http.Header {
	out := h.Clone()
	for name := range out {
		if sensitiveHeaderNames[name] || sensitiveHeaderRegex.MatchString(name) {
			out[name] = []string{"[REDACTED]"}
		}
	}
	return out
}
