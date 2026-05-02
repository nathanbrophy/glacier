// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"reflect"
	"testing"
)

// §21.4 NF1; §23.5, §23.13

// BenchmarkEqualPrimitive benchmarks the primitive fast path for int equality.
// Target: ≤ 50 ns/op, 0 allocs/op.
func BenchmarkEqualPrimitive(b *testing.B) {
	mt := &mockTB{}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Equal(mt, 42, 42)
	}
}

// TestPrimitiveFastPathBypass verifies the fast path is taken for comparable
// equal types, and NOT taken for non-comparable types.
// §23.5
func TestPrimitiveFastPathBypass(t *testing.T) {
	// Non-comparable: slice must NOT be fast-pathed (returns false).
	notFast := !primitiveEqual([]int{1}, []int{1})
	True(t, notFast, "non-comparable type bypasses fast path")

	// Comparable int must be fast-pathed.
	fast := primitiveEqual(42, 42)
	True(t, fast, "comparable equal ints take fast path")
}

// BenchmarkEqualPrimitiveFastPath is the D35 §23.13 benchmark.
func BenchmarkEqualPrimitiveFastPath(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		primitiveEqual(42, 42)
	}
}

// BenchmarkEqualSmallStruct benchmarks smart-equal on a small struct (slow path).
// Target: ≤ 200 ns/op.
func BenchmarkEqualSmallStruct(b *testing.B) {
	type S struct{ A, B, C, D, E int }
	mt := &mockTB{}
	got := S{1, 2, 3, 4, 5}
	want := S{1, 2, 3, 4, 5}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Equal(mt, got, want)
	}
}

// BenchmarkSmartEqualSlowPath benchmarks the slow path directly.
// §23.13: target ≤ 200 ns/op.
func BenchmarkSmartEqualSlowPath(b *testing.B) {
	type S struct{ A, B, C int }
	got := S{1, 2, 3}
	want := S{1, 2, 3}
	cfg := applyEqualOptions(nil)
	gv := reflect.ValueOf(got)
	wv := reflect.ValueOf(want)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		smartEqual(gv, wv, &cfg, nil)
	}
}

// BenchmarkEqualLargeSlice benchmarks a 1000-element int slice.
func BenchmarkEqualLargeSlice(b *testing.B) {
	mt := &mockTB{}
	n := 1000
	got := make([]int, n)
	want := make([]int, n)
	for i := range n {
		got[i] = i
		want[i] = i
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Equal(mt, got, want)
	}
}

// BenchmarkEqualLargeMap benchmarks a 1000-entry map.
func BenchmarkEqualLargeMap(b *testing.B) {
	mt := &mockTB{}
	n := 1000
	got := make(map[int]int, n)
	want := make(map[int]int, n)
	for i := range n {
		got[i] = i
		want[i] = i
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Equal(mt, got, want)
	}
}

// BenchmarkMatchGlob benchmarks glob matching.
func BenchmarkMatchGlob(b *testing.B) {
	mt := &mockTB{}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Match(mt, "user-12345", "user-*")
	}
}

// BenchmarkMatchRegex benchmarks regex matching (no cache).
func BenchmarkMatchRegex(b *testing.B) {
	mt := &mockTB{}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Match(mt, "user-123", `^user-[0-9]+$`, MatchRegex())
	}
}

// BenchmarkContainsSlice benchmarks Contains on a 100-element slice.
func BenchmarkContainsSlice(b *testing.B) {
	mt := &mockTB{}
	s := make([]int, 100)
	for i := range 100 {
		s[i] = i
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Contains(mt, s, 99)
	}
}

// BenchmarkContainsString benchmarks Contains on a string.
func BenchmarkContainsString(b *testing.B) {
	mt := &mockTB{}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Contains(mt, "hello world foo bar", "foo")
	}
}

// BenchmarkJSONEq benchmarks JSONEq on a small JSON payload.
func BenchmarkJSONEq(b *testing.B) {
	mt := &mockTB{}
	got := []byte(`{"name":"Ada","age":36,"tags":["math","logic","cs"],"active":true}`)
	want := []byte(`{"age":36,"name":"Ada","tags":["math","logic","cs"],"active":true}`)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		JSONEq(mt, got, want)
	}
}
