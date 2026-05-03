// SPDX-License-Identifier: Apache-2.0

package cache_test

import (
	"context"
	"fmt"
	"time"

	"github.com/nathanbrophy/glacier/cache"
)

// Example demonstrates basic in-memory caching with a default TTL.
func Example() {
	c := cache.New[string](cache.WithDefaultTTL(5 * time.Minute))
	c.Set("greeting", "hello")
	v, ok := c.Get("greeting")
	fmt.Println(v, ok)
	// Output: hello true
}

// ExampleCache_GetOrLoad shows the singleflight load pattern.
func ExampleCache_GetOrLoad() {
	c := cache.New[int]()

	loaderCalls := 0
	loader := func(_ context.Context) (int, error) {
		loaderCalls++
		return 42, nil
	}

	// First call: miss, loader runs.
	v, _ := c.GetOrLoad(context.Background(), "answer", loader)
	// Second call: hit, loader skipped.
	v, _ = c.GetOrLoad(context.Background(), "answer", loader)
	fmt.Println(v, loaderCalls)
	// Output: 42 1
}
