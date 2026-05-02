// SPDX-License-Identifier: Apache-2.0

package httpc_test

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpc"
)

func TestConcurrentGet(t *testing.T) {
	var callCount int64
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		atomic.AddInt64(&callCount, 1)
		return jsonResponse(200, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))

	const goroutines = 100
	var wg sync.WaitGroup
	var errCount int64
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			_, err := c.Get(ctx, "http://example.com/")
			if err != nil {
				atomic.AddInt64(&errCount, 1)
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, int64(goroutines), callCount)
	assert.Equal(t, int64(0), errCount)
}

func TestConcurrentPostBodyClosures(t *testing.T) {
	// Body closures for the same request must not be called concurrently.
	// Different requests run concurrently; each has its own closure counter.
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{}`), nil
	})
	c := httpc.New(httpc.WithTransport(rt))

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	var totalClosureCalls int64
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			httpc.PostWith[struct{}](c, ctx, "http://example.com/",
				httpc.JSONBody(func() testUser {
					atomic.AddInt64(&totalClosureCalls, 1)
					return testUser{ID: 1}
				}),
			)
		}()
	}
	wg.Wait()
	// Each goroutine's single attempt calls the closure once.
	assert.Equal(t, int64(goroutines), totalClosureCalls)
}
