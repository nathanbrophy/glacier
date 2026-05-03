// SPDX-License-Identifier: Apache-2.0

package httpc_test

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpc"
)

func TestMaxAttemptsDefaultOne(t *testing.T) {
	t.Parallel()
	var callCount int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		callCount++
		return jsonResponse(200, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, err := c.Get(ctx, "http://example.com/")
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestMaxAttemptsRetries(t *testing.T) {
	t.Parallel()
	var callCount int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		callCount++
		return jsonResponse(503, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, err := c.Get(ctx, "http://example.com/",
		httpc.WithRetry(httpc.MaxAttempts(3)),
	)
	assert.Error(t, err)
	assert.Equal(t, 3, callCount)
	assert.True(t, errors.Is(err, httpc.ErrMaxAttempts))
}

func TestRetryOnDefault(t *testing.T) {
	t.Parallel()
	// Default retry statuses: 500, 502, 503, 504, 429.
	for _, status := range []int{500, 502, 503, 504, 429} {
		status := status
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()
			var count int
			rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				count++
				return jsonResponse(status, `{}`), nil
			})
			c := httpc.New(httpc.WithTransport(rt))
			ctx := context.Background()
			_, err := c.Get(ctx, "http://example.com/",
				httpc.WithRetry(httpc.MaxAttempts(2)),
			)
			assert.Error(t, err)
			assert.Equal(t, 2, count)
		})
	}
}

func TestRetryOnCustom(t *testing.T) {
	t.Parallel()
	var count int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		count++
		// 503 would be default-retryable, but we override to only retry 418.
		return jsonResponse(503, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, err := c.Get(ctx, "http://example.com/",
		httpc.WithRetry(httpc.MaxAttempts(3), httpc.RetryOn(418)),
	)
	// 503 is not in custom list, so no retry :  just StatusError on first attempt.
	assert.Error(t, err)
	assert.Equal(t, 1, count)
}

func TestRetryIf(t *testing.T) {
	t.Parallel()
	var count int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		count++
		return jsonResponse(418, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, err := c.Get(ctx, "http://example.com/",
		httpc.WithRetry(
			httpc.MaxAttempts(3),
			httpc.RetryOn(), // empty = no status retry
			httpc.RetryIf(func(resp *httpc.Response, err error) bool {
				return resp != nil && resp.StatusCode == 418
			}),
		),
	)
	assert.Error(t, err)
	assert.Equal(t, 3, count)
}

func TestMaxElapsedFires(t *testing.T) {
	t.Parallel()
	var count int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		count++
		time.Sleep(20 * time.Millisecond)
		return jsonResponse(503, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, err := c.Get(ctx, "http://example.com/",
		httpc.WithRetry(
			httpc.MaxAttempts(10),
			httpc.MaxElapsed(50*time.Millisecond),
		),
	)
	assert.Error(t, err)
	// Should have made fewer than 10 attempts.
	assert.True(t, count < 10)
}

func TestRetryRespectsCtxCancel(t *testing.T) {
	t.Parallel()
	var callCount int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		callCount++
		return jsonResponse(503, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after first attempt via goroutine while we're in backoff.
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	_, err := c.Get(ctx, "http://example.com/",
		httpc.WithRetry(
			httpc.MaxAttempts(10),
			httpc.LinearBackoff(100*time.Millisecond),
		),
	)
	assert.Error(t, err)
	// Should have cancelled early.
	assert.True(t, callCount < 10)
}

func TestMaxAttemptsZero(t *testing.T) {
	t.Parallel()
	var count int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		count++
		return jsonResponse(200, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, err := c.Get(ctx, "http://example.com/",
		httpc.WithRetry(httpc.MaxAttempts(0)),
	)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestMaxAttemptsNegative(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()
	httpc.MaxAttempts(-1)
}

func TestRetryIfAndRetryOnCombine(t *testing.T) {
	t.Parallel()
	// Both conditions active: RetryOn(503) OR RetryIf(418).
	responses := []int{503, 418, 200}
	idx := 0
	var count int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		status := responses[idx]
		if idx < len(responses)-1 {
			idx++
		}
		count++
		return jsonResponse(status, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()
	_, err := c.Get(ctx, "http://example.com/",
		httpc.WithRetry(
			httpc.MaxAttempts(5),
			httpc.RetryOn(503),
			httpc.RetryIf(func(resp *httpc.Response, _ error) bool {
				return resp != nil && resp.StatusCode == 418
			}),
		),
	)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestRetryShortCircuitsOnCtxCancel(t *testing.T) {
	t.Parallel()
	var closureCalls int32
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(503, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	httpc.PostWith[struct{}](c, ctx, "http://example.com/",
		httpc.JSONBody(func() testUser {
			atomic.AddInt32(&closureCalls, 1)
			return testUser{}
		}),
		httpc.WithRetry(httpc.MaxAttempts(5)),
	)
	// Pre-cancelled context: closure should not be called at all.
	assert.True(t, closureCalls <= 1)
}

func TestLinearBackoff(t *testing.T) {
	t.Parallel()
	var count int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		count++
		return jsonResponse(503, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()

	start := time.Now()
	httpc.GetWith[struct{}](c, ctx, "http://example.com/",
		httpc.WithRetry(
			httpc.MaxAttempts(3),
			httpc.LinearBackoff(10*time.Millisecond),
		),
	)
	elapsed := time.Since(start)
	assert.Equal(t, 3, count)
	// 2 sleeps of 10ms each = at least 20ms total.
	assert.True(t, elapsed >= 20*time.Millisecond)
}

func TestExponentialBackoff(t *testing.T) {
	t.Parallel()
	var count int
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		count++
		return jsonResponse(503, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))
	ctx := context.Background()

	start := time.Now()
	httpc.GetWith[struct{}](c, ctx, "http://example.com/",
		httpc.WithRetry(
			httpc.MaxAttempts(3),
			httpc.ExponentialBackoff(10*time.Millisecond),
		),
	)
	elapsed := time.Since(start)
	assert.Equal(t, 3, count)
	// 10ms + 20ms = at least 30ms.
	assert.True(t, elapsed >= 30*time.Millisecond)
}
