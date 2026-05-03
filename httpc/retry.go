// SPDX-License-Identifier: Apache-2.0

package httpc

import (
	"context"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/nathanbrophy/glacier/errs"
)

// defaultRetryStatuses are the HTTP status codes that trigger a retry when no
// RetryOn option is specified.
var defaultRetryStatuses = []int{500, 502, 503, 504, 429}

// retryConfig holds the accumulated retry policy for one request attempt sequence.
type retryConfig struct {
	// maxAttempts >= 1; value 1 means no retry (single attempt).
	maxAttempts int
	// exactly one of exponentialBase or linearDelay is non-zero, or neither.
	exponentialBase time.Duration
	linearDelay     time.Duration
	// jittered applies ±25% to the computed backoff duration.
	jittered bool
	// retryStatuses nil uses the default [500,502,503,504,429].
	retryStatuses []int
	// retryIf combines with retryStatuses: either condition triggers retry.
	retryIf func(*Response, error) bool
	// maxElapsed == 0 means no overall timeout on the retry loop.
	maxElapsed time.Duration
}

// RetryOption modifies a retry policy.
// Implementations are sealed within the httpc package.
type RetryOption interface{ applyRetry(*retryConfig) error }

type retryOptionFunc func(*retryConfig) error

func (f retryOptionFunc) applyRetry(c *retryConfig) error { return f(c) }

// MaxAttempts caps the total number of request attempts (including the first).
// MaxAttempts(1) means one attempt, no retry (the default).
// MaxAttempts(0) is treated identically to MaxAttempts(1).
// Negative values panic (programming error).
func MaxAttempts(n int) RetryOption {
	if n < 0 {
		panic("httpc: MaxAttempts: n must be >= 0")
	}
	return retryOptionFunc(func(c *retryConfig) error {
		if n == 0 {
			c.maxAttempts = 1
		} else {
			c.maxAttempts = n
		}
		return nil
	})
}

// ExponentialBackoff sets the backoff strategy to base * 2^attempt, where
// attempt is zero-indexed (first retry sleeps base, second sleeps 2*base, ...).
// Combine with Jittered for production use.
func ExponentialBackoff(base time.Duration) RetryOption {
	return retryOptionFunc(func(c *retryConfig) error {
		c.exponentialBase = base
		c.linearDelay = 0
		return nil
	})
}

// LinearBackoff sets a fixed delay between attempts.
func LinearBackoff(d time.Duration) RetryOption {
	return retryOptionFunc(func(c *retryConfig) error {
		c.linearDelay = d
		c.exponentialBase = 0
		return nil
	})
}

// Jittered adds ±25% uniform random jitter to the computed backoff duration.
// Apply after ExponentialBackoff or LinearBackoff.
func Jittered() RetryOption {
	return retryOptionFunc(func(c *retryConfig) error {
		c.jittered = true
		return nil
	})
}

// RetryOn specifies HTTP status codes that trigger a retry. Default for
// retry-enabled requests: [500, 502, 503, 504, 429].
// Providing statuses replaces the default (does not append).
func RetryOn(statuses ...int) RetryOption {
	return retryOptionFunc(func(c *retryConfig) error {
		c.retryStatuses = append([]int(nil), statuses...)
		return nil
	})
}

// RetryIf registers a custom predicate. Either RetryOn or RetryIf may trigger
// a retry; both are checked. RetryIf receives the *Response (may be nil on
// transport error) and the error.
func RetryIf(fn func(*Response, error) bool) RetryOption {
	return retryOptionFunc(func(c *retryConfig) error {
		c.retryIf = fn
		return nil
	})
}

// MaxElapsed sets an overall wall-clock budget for the retry loop, including
// all attempt durations and backoff sleeps. When the budget is exceeded, the
// loop is terminated and ErrMaxElapsed is returned.
func MaxElapsed(d time.Duration) RetryOption {
	return retryOptionFunc(func(c *retryConfig) error {
		c.maxElapsed = d
		return nil
	})
}

// applyRetryOptions applies a slice of RetryOption to a retryConfig.
func applyRetryOptions(cfg *retryConfig, opts []RetryOption) error {
	for _, o := range opts {
		if o == nil {
			continue
		}
		if err := o.applyRetry(cfg); err != nil {
			return err
		}
	}
	return nil
}

// mergeRetryConfig merges per-call onto client-level. Per-call scalar fields
// override client defaults when non-zero.
func mergeRetryConfig(base retryConfig, perCall *retryConfig) retryConfig {
	if perCall == nil {
		return base
	}
	merged := base
	if perCall.maxAttempts > 0 {
		merged.maxAttempts = perCall.maxAttempts
	}
	if perCall.exponentialBase != 0 {
		merged.exponentialBase = perCall.exponentialBase
		merged.linearDelay = 0
	}
	if perCall.linearDelay != 0 {
		merged.linearDelay = perCall.linearDelay
		merged.exponentialBase = 0
	}
	if perCall.jittered {
		merged.jittered = true
	}
	if perCall.retryStatuses != nil {
		merged.retryStatuses = perCall.retryStatuses
	}
	if perCall.retryIf != nil {
		merged.retryIf = perCall.retryIf
	}
	if perCall.maxElapsed != 0 {
		merged.maxElapsed = perCall.maxElapsed
	}
	return merged
}

// computeBackoff returns the backoff duration for the given retry attempt
// (zero-indexed: 0 = first retry).
func computeBackoff(cfg *retryConfig, attempt int) time.Duration {
	var d time.Duration
	switch {
	case cfg.exponentialBase != 0:
		d = cfg.exponentialBase
		for i := 0; i < attempt; i++ {
			d *= 2
		}
	case cfg.linearDelay != 0:
		d = cfg.linearDelay
	default:
		return 0
	}
	if cfg.jittered && d > 0 {
		// ±25% uniform random jitter.
		jitter := float64(d) * 0.25
		d = time.Duration(float64(d) + (rand.Float64()*2-1)*jitter)
	}
	return d
}

// shouldRetry reports whether the response/error triggers a retry.
func shouldRetry(cfg *retryConfig, resp *Response, err error) bool {
	// Custom predicate takes priority alongside status check.
	if cfg.retryIf != nil && cfg.retryIf(resp, err) {
		return true
	}
	if resp == nil {
		return false
	}
	statuses := cfg.retryStatuses
	if statuses == nil {
		statuses = defaultRetryStatuses
	}
	for _, s := range statuses {
		if resp.StatusCode == s {
			return true
		}
	}
	return false
}

// sleepWithCtx sleeps for d, returning early if ctx is cancelled.
func sleepWithCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// doWithRetry executes the request, retrying according to cfg. buildBody is
// called once per attempt; a nil buildBody means no request body.
func (c *Client) doWithRetry(
	ctx context.Context,
	req *http.Request,
	cfg *retryConfig,
	buildBody func() (io.ReadCloser, string, error),
) (*Response, error) {
	maxAttempts := cfg.maxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var loopStart time.Time
	if cfg.maxElapsed > 0 {
		loopStart = time.Now()
	}

	var lastErr error
	// prevBody tracks the ReadCloser from the previous attempt for StreamBody semantics.
	var prevBody io.ReadCloser

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check context before each attempt.
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Check MaxElapsed at the top of each attempt.
		if cfg.maxElapsed > 0 && attempt > 0 {
			if time.Since(loopStart) >= cfg.maxElapsed {
				if lastErr != nil {
					return nil, errs.Wrap(lastErr, "httpc: max elapsed")
				}
				return nil, ErrMaxElapsed
			}
		}

		// Clone the request so we can set a fresh body each attempt.
		attemptReq := req.Clone(req.Context())

		if buildBody != nil {
			// Close the previous attempt's body before invoking the closure again
			// (required for StreamBody correctness :  the caller's ReadCloser must
			// be released before a fresh one is created).
			if prevBody != nil {
				_ = prevBody.Close()
				prevBody = nil
			}
			body, ct, err := buildBody()
			if err != nil {
				// Body closure error is not retryable.
				return nil, err
			}
			prevBody = body
			attemptReq.Body = body
			if ct != "" && attemptReq.Header.Get("Content-Type") == "" {
				attemptReq.Header.Set("Content-Type", ct)
			}
		}

		start := time.Now()
		httpResp, err := c.cfg.transport.RoundTrip(attemptReq)
		elapsed := time.Since(start)

		if err != nil {
			lastErr = err
			// Only retry transport errors when more attempts remain AND predicate matches.
			if attempt+1 < maxAttempts && shouldRetry(cfg, nil, err) {
				if sleepErr := c.backoffAndSleep(ctx, cfg, attempt, loopStart); sleepErr != nil {
					return nil, sleepErr
				}
				continue
			}
			break
		}

		wrapped := &Response{
			Response: httpResp,
			Elapsed:  elapsed,
		}

		// Check whether this status code warrants a retry.
		retryable := maxAttempts > 1 && shouldRetry(cfg, wrapped, nil)

		if retryable && attempt+1 < maxAttempts {
			// More attempts remain: drain body and sleep before next attempt.
			_, _ = io.Copy(io.Discard, httpResp.Body)
			_ = httpResp.Body.Close()
			if sleepErr := c.backoffAndSleep(ctx, cfg, attempt, loopStart); sleepErr != nil {
				return nil, sleepErr
			}
			continue
		}

		if retryable {
			// All attempts exhausted with a retryable status on the last attempt.
			_, _ = io.Copy(io.Discard, httpResp.Body)
			_ = httpResp.Body.Close()
			if cfg.maxElapsed > 0 && time.Since(loopStart) >= cfg.maxElapsed {
				return nil, ErrMaxElapsed
			}
			return nil, ErrMaxAttempts
		}

		// Return the response :  caller (do) handles status check.
		return wrapped, nil
	}

	// Loop exited due to transport error exhaustion.
	if cfg.maxElapsed > 0 && time.Since(loopStart) >= cfg.maxElapsed {
		if lastErr != nil {
			return nil, errs.Wrap(lastErr, "httpc: max elapsed")
		}
		return nil, ErrMaxElapsed
	}

	if lastErr != nil {
		return nil, errs.Wrap(lastErr, "httpc: max attempts")
	}
	return nil, ErrMaxAttempts
}

// backoffAndSleep computes and executes the backoff sleep for the given attempt.
func (c *Client) backoffAndSleep(ctx context.Context, cfg *retryConfig, attempt int, loopStart time.Time) error {
	backoff := computeBackoff(cfg, attempt)
	if cfg.maxElapsed > 0 {
		remaining := cfg.maxElapsed - time.Since(loopStart)
		if remaining <= 0 {
			return ErrMaxElapsed
		}
		if backoff > remaining {
			backoff = remaining
		}
	}
	return sleepWithCtx(ctx, backoff)
}
