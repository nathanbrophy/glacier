// SPDX-License-Identifier: Apache-2.0

package fluent_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/fluent"
)

func BenchmarkMapFilter(b *testing.B) {
	b.ReportAllocs()
	src := make([]int, 1000)
	for i := range src {
		src[i] = i
	}
	b.ResetTimer()
	for range b.N {
		seq := fluent.Filter(fluent.Map(fluent.From(src), func(v int) int { return v * 2 }), func(v int) bool { return v%4 == 0 })
		for range seq {
		}
	}
}

func BenchmarkToSlice(b *testing.B) {
	b.ReportAllocs()
	src := make([]int, 1000)
	for i := range src {
		src[i] = i
	}
	b.ResetTimer()
	for range b.N {
		_ = fluent.ToSlice(fluent.From(src))
	}
}

func BenchmarkDistinct(b *testing.B) {
	b.ReportAllocs()
	// 1000 integers with 500 unique values (each value appears twice).
	src := make([]int, 1000)
	for i := range src {
		src[i] = i % 500
	}
	b.ResetTimer()
	for range b.N {
		_ = fluent.ToSlice(fluent.Distinct(fluent.From(src)))
	}
}
