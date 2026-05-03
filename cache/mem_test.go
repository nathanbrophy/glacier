// SPDX-License-Identifier: Apache-2.0

package cache_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/cache"
)

// fakeClock is a deterministic clock for expiry tests.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)}
}

func (f *fakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *fakeClock) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
}

func TestMemHit(t *testing.T) {
	t.Parallel()
	c := cache.New[string]()
	c.Set("k", "v")
	v, ok := c.Get("k")
	require.True(t, ok)
	assert.Equal(t, "v", v)
}

func TestMemMiss(t *testing.T) {
	t.Parallel()
	c := cache.New[string]()
	v, ok := c.Get("absent")
	assert.False(t, ok)
	assert.Equal(t, "", v)
}

func TestMemSetOverwrites(t *testing.T) {
	t.Parallel()
	c := cache.New[int]()
	c.Set("k", 1)
	c.Set("k", 2)
	v, _ := c.Get("k")
	assert.Equal(t, 2, v)
}

func TestMemDelete(t *testing.T) {
	t.Parallel()
	c := cache.New[int]()
	c.Set("k", 1)
	c.Delete("k")
	_, ok := c.Get("k")
	assert.False(t, ok)
}

func TestMemExpiry(t *testing.T) {
	t.Parallel()
	clock := newFakeClock()
	c := cache.New[int](cache.WithClock(clock.Now))
	c.SetWithTTL("k", 1, 100*time.Millisecond)

	// Still alive at 50ms.
	clock.Advance(50 * time.Millisecond)
	_, ok := c.Get("k")
	require.True(t, ok)

	// Expired at 150ms.
	clock.Advance(101 * time.Millisecond)
	_, ok = c.Get("k")
	assert.False(t, ok)
}

func TestMemDefaultTTL(t *testing.T) {
	t.Parallel()
	clock := newFakeClock()
	c := cache.New[int](
		cache.WithClock(clock.Now),
		cache.WithDefaultTTL(1*time.Hour),
	)
	c.Set("k", 1) // uses default TTL
	clock.Advance(30 * time.Minute)
	_, ok := c.Get("k")
	assert.True(t, ok)
	clock.Advance(31 * time.Minute)
	_, ok = c.Get("k")
	assert.False(t, ok)
}

func TestMemZeroTTLNoExpiry(t *testing.T) {
	t.Parallel()
	clock := newFakeClock()
	c := cache.New[int](cache.WithClock(clock.Now))
	c.Set("k", 1) // default TTL is 0 → no expiry
	clock.Advance(100 * 365 * 24 * time.Hour)
	_, ok := c.Get("k")
	assert.True(t, ok)
}

func TestMemConcurrencyRace(t *testing.T) {
	t.Parallel()
	c := cache.New[int]()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(i int) { defer wg.Done(); c.Set("k", i) }(i)
		go func() { defer wg.Done(); _, _ = c.Get("k") }()
	}
	wg.Wait()
}

func TestGetOrLoadMissCallsLoader(t *testing.T) {
	t.Parallel()
	c := cache.New[int]()
	calls := 0
	v, err := c.GetOrLoad(context.Background(), "k", func(_ context.Context) (int, error) {
		calls++
		return 7, nil
	})
	require.Nil(t, err)
	assert.Equal(t, 7, v)
	assert.Equal(t, 1, calls)

	// Second call hits.
	v, err = c.GetOrLoad(context.Background(), "k", func(_ context.Context) (int, error) {
		calls++
		return 99, nil
	})
	require.Nil(t, err)
	assert.Equal(t, 7, v)
	assert.Equal(t, 1, calls)
}

func TestGetOrLoadCollapsesConcurrent(t *testing.T) {
	t.Parallel()
	c := cache.New[int]()
	var calls atomic.Int32
	start := make(chan struct{})

	loader := func(_ context.Context) (int, error) {
		<-start
		calls.Add(1)
		return 42, nil
	}

	const N = 100
	results := make(chan int, N)
	for i := 0; i < N; i++ {
		go func() {
			v, _ := c.GetOrLoad(context.Background(), "shared", loader)
			results <- v
		}()
	}
	// Allow all goroutines to enter the singleflight call.
	time.Sleep(50 * time.Millisecond)
	close(start)

	for i := 0; i < N; i++ {
		assert.Equal(t, 42, <-results)
	}
	assert.Equal(t, int32(1), calls.Load())
}

func TestGetOrLoadErrorNotCached(t *testing.T) {
	t.Parallel()
	c := cache.New[int]()
	want := errors.New("boom")
	calls := 0
	loader := func(_ context.Context) (int, error) {
		calls++
		return 0, want
	}
	_, err1 := c.GetOrLoad(context.Background(), "k", loader)
	assert.True(t, errors.Is(err1, want))
	_, err2 := c.GetOrLoad(context.Background(), "k", loader)
	assert.True(t, errors.Is(err2, want))
	// Loader was called twice (errors are not cached).
	assert.Equal(t, 2, calls)
}

func TestGetOrLoadContextCancel(t *testing.T) {
	t.Parallel()
	c := cache.New[int]()

	// Block first loader so ctx cancellation actually happens before it returns.
	block := make(chan struct{})
	loaderStarted := make(chan struct{}, 1)
	go func() {
		_, _ = c.GetOrLoad(context.Background(), "k", func(_ context.Context) (int, error) {
			loaderStarted <- struct{}{}
			<-block
			return 1, nil
		})
	}()
	<-loaderStarted

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.GetOrLoad(ctx, "k", func(_ context.Context) (int, error) { return 0, nil })
	assert.True(t, errors.Is(err, context.Canceled))
	close(block)
}
