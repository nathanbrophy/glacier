// SPDX-License-Identifier: Apache-2.0

package cache_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/cache"
)

// BenchmarkMemHit measures the warm hot-path Get on a populated in-memory cache.
// Budget per spec 0033: ≤ 50 ns/op, 0 B/op, 0 allocs/op.
func BenchmarkMemHit(b *testing.B) {
	c := cache.New[int]()
	c.Set("k", 42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Get("k")
	}
}

// BenchmarkMemSet measures Set on a fresh entry. Budget: ≤ 1 µs/op.
func BenchmarkMemSet(b *testing.B) {
	c := cache.New[int]()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("k", i)
	}
}

// BenchmarkMemMiss measures Get on an empty cache.
func BenchmarkMemMiss(b *testing.B) {
	c := cache.New[int]()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Get("absent")
	}
}

// TestMemZeroAllocOnHit verifies the alloc-free hot-path invariant from the spec.
// Cannot run with t.Parallel() because testing.AllocsPerRun panics when called
// from a parallel test.
func TestMemZeroAllocOnHit(t *testing.T) {
	c := cache.New[int]()
	c.Set("k", 42)
	allocs := testing.AllocsPerRun(100, func() {
		_, _ = c.Get("k")
	})
	if allocs != 0 {
		t.Fatalf("Get hot path allocates %v, want 0", allocs)
	}
}
