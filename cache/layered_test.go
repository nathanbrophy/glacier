// SPDX-License-Identifier: Apache-2.0

package cache_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/cache"
)

func TestLayeredHitPrimary(t *testing.T) {
	t.Parallel()
	primary := cache.New[string]()
	backing := cache.New[string]()
	primary.Set("k", "from-primary")
	backing.Set("k", "from-backing")

	c := cache.NewLayered(primary, backing)
	v, ok := c.Get("k")
	require.True(t, ok)
	assert.Equal(t, "from-primary", v)
}

func TestLayeredHitBackingPopulatesPrimary(t *testing.T) {
	t.Parallel()
	primary := cache.New[string]()
	backing := cache.New[string]()
	backing.Set("k", "v")

	c := cache.NewLayered(primary, backing)
	v, ok := c.Get("k")
	require.True(t, ok)
	assert.Equal(t, "v", v)

	// Subsequent direct read from primary now returns the value.
	v2, ok := primary.Get("k")
	require.True(t, ok)
	assert.Equal(t, "v", v2)
}

func TestLayeredSetWritesBoth(t *testing.T) {
	t.Parallel()
	primary := cache.New[int]()
	backing := cache.New[int]()
	c := cache.NewLayered(primary, backing)
	c.Set("k", 5)

	if v, ok := primary.Get("k"); !ok || v != 5 {
		t.Fatalf("primary missing")
	}
	if v, ok := backing.Get("k"); !ok || v != 5 {
		t.Fatalf("backing missing")
	}
}

func TestLayeredDeleteRemovesBoth(t *testing.T) {
	t.Parallel()
	primary := cache.New[int]()
	backing := cache.New[int]()
	c := cache.NewLayered(primary, backing)
	c.Set("k", 1)
	c.Delete("k")
	_, ok1 := primary.Get("k")
	_, ok2 := backing.Get("k")
	assert.False(t, ok1)
	assert.False(t, ok2)
}

func TestLayeredGetOrLoadCollapses(t *testing.T) {
	t.Parallel()
	primary := cache.New[int]()
	backing := cache.New[int]()
	c := cache.NewLayered(primary, backing)

	var calls atomic.Int32
	loader := func(_ context.Context) (int, error) {
		calls.Add(1)
		return 99, nil
	}

	const N = 50
	results := make(chan int, N)
	for i := 0; i < N; i++ {
		go func() {
			v, _ := c.GetOrLoad(context.Background(), "shared", loader)
			results <- v
		}()
	}
	for i := 0; i < N; i++ {
		<-results
	}
	// Loader called once due to singleflight.
	assert.Equal(t, int32(1), calls.Load())
}
