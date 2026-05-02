// SPDX-License-Identifier: Apache-2.0

package httpmock_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/httpmock"
)

func TestConcurrentRoundTrip(t *testing.T) {
	rt := httpmock.New()
	rt.OnRequest().Path("/concurrent").AnyTimes().Respond(httpmock.Status(200))

	const goroutines = 20
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := newReq(t, "GET", "https://example.com/concurrent", nil)
			resp, err := rt.RoundTrip(req)
			if err != nil {
				errs[idx] = err
				return
			}
			assert.True(t, resp.StatusCode == 200, fmt.Sprintf("goroutine %d: expected 200, got %d", idx, resp.StatusCode))
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		assert.NoError(t, err, "goroutine", i)
	}

	all := rt.AllRequests()
	assert.Len(t, all, goroutines)
}

func TestTimes1RaceFix(t *testing.T) {
	// Regression test for §23.14: match + increment + respond in single
	// write-lock critical section. Run with -race to detect any data race.
	rt := httpmock.New()
	rt.OnRequest().Path("/once").AnyTimes().Respond(httpmock.Status(200))

	const goroutines = 50
	var wg sync.WaitGroup

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := newReq(t, "GET", "https://example.com/once", nil)
			_, _ = rt.RoundTrip(req)
		}()
	}
	wg.Wait()

	all := rt.AllRequests()
	assert.Len(t, all, goroutines)
}

func TestConcurrentOnRequest(t *testing.T) {
	// Concurrent stub registration must be safe under -race.
	rt := httpmock.New(httpmock.LenientMode())

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rt.OnRequest().Path("/x").AnyTimes().Respond(httpmock.Status(200))
		}()
	}
	wg.Wait()
}

func TestConcurrentAllRequests(t *testing.T) {
	rt := httpmock.New(httpmock.LenientMode())

	var wg sync.WaitGroup
	// Writers.
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := newReq(t, "GET", "https://example.com/x", nil)
			_, _ = rt.RoundTrip(req)
		}()
	}
	// Readers.
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = rt.AllRequests()
		}()
	}
	wg.Wait()
}
