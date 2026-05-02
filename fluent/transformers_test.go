// SPDX-License-Identifier: Apache-2.0

package fluent_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/fluent"
)

func TestMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		want []int
	}{
		{"empty source", []int{}, []int{}},
		{"double each", []int{1, 2, 3}, []int{2, 4, 6}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Map(fluent.From(tc.src), func(v int) int { return v * 2 }))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		want []int
	}{
		{"empty source", []int{}, []int{}},
		{"evens only", []int{1, 2, 3, 4, 5, 6}, []int{2, 4, 6}},
		{"none match", []int{1, 3, 5}, []int{}},
		{"all match", []int{2, 4, 6}, []int{2, 4, 6}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Filter(fluent.From(tc.src), func(v int) bool { return v%2 == 0 }))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestTake(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		n    int
		want []int
	}{
		{"empty source", []int{}, 3, []int{}},
		{"take zero", []int{1, 2, 3}, 0, []int{}},
		{"take less than length", []int{1, 2, 3, 4, 5}, 3, []int{1, 2, 3}},
		{"take more than length", []int{1, 2, 3}, 10, []int{1, 2, 3}},
		{"take exact length", []int{1, 2, 3}, 3, []int{1, 2, 3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Take(fluent.From(tc.src), tc.n))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestDrop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		n    int
		want []int
	}{
		{"empty source", []int{}, 3, []int{}},
		{"drop zero", []int{1, 2, 3}, 0, []int{1, 2, 3}},
		{"drop some", []int{1, 2, 3, 4, 5}, 2, []int{3, 4, 5}},
		{"drop all", []int{1, 2, 3}, 3, []int{}},
		{"drop more than length", []int{1, 2, 3}, 10, []int{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Drop(fluent.From(tc.src), tc.n))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestWindow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		size int
		want [][]int
	}{
		{"empty source", []int{}, 2, [][]int{}},
		{"size larger than src", []int{1, 2}, 3, [][]int{}},
		{"exact size", []int{1, 2, 3}, 3, [][]int{{1, 2, 3}}},
		{"sliding window", []int{1, 2, 3, 4}, 2, [][]int{{1, 2}, {2, 3}, {3, 4}}},
		{"size 1", []int{1, 2, 3}, 1, [][]int{{1}, {2}, {3}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Window(fluent.From(tc.src), tc.size))
			assert.Equal(t, got, tc.want)
		})
	}

	t.Run("panics on size zero", func(t *testing.T) {
		t.Parallel()
		defer func() {
			r := recover()
			assert.True(t, r != nil, "expected panic on size 0")
		}()
		fluent.Window(fluent.From([]int{1, 2, 3}), 0)
	})

	t.Run("panics on negative size", func(t *testing.T) {
		t.Parallel()
		defer func() {
			r := recover()
			assert.True(t, r != nil, "expected panic on negative size")
		}()
		fluent.Window(fluent.From([]int{1, 2, 3}), -1)
	})

	t.Run("yielded slices are independent copies", func(t *testing.T) {
		t.Parallel()
		windows := fluent.ToSlice(fluent.Window(fluent.From([]int{1, 2, 3}), 2))
		assert.Equal(t, len(windows), 2)
		// Mutate first window and confirm second is unaffected.
		windows[0][0] = 999
		assert.Equal(t, windows[1][0], 2)
	})
}

func TestChunk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		size int
		want [][]int
	}{
		{"empty source", []int{}, 2, [][]int{}},
		{"even chunks", []int{1, 2, 3, 4}, 2, [][]int{{1, 2}, {3, 4}}},
		{"last chunk partial", []int{1, 2, 3, 4, 5}, 2, [][]int{{1, 2}, {3, 4}, {5}}},
		{"size larger than src", []int{1, 2}, 5, [][]int{{1, 2}}},
		{"size 1", []int{1, 2, 3}, 1, [][]int{{1}, {2}, {3}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Chunk(fluent.From(tc.src), tc.size))
			assert.Equal(t, got, tc.want)
		})
	}

	t.Run("panics on size zero", func(t *testing.T) {
		t.Parallel()
		defer func() {
			r := recover()
			assert.True(t, r != nil, "expected panic on size 0")
		}()
		fluent.Chunk(fluent.From([]int{1, 2, 3}), 0)
	})

	t.Run("panics on negative size", func(t *testing.T) {
		t.Parallel()
		defer func() {
			r := recover()
			assert.True(t, r != nil, "expected panic on negative size")
		}()
		fluent.Chunk(fluent.From([]int{1, 2, 3}), -1)
	})
}

func TestDistinct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		want []int
	}{
		{"empty source", []int{}, []int{}},
		{"no duplicates", []int{1, 2, 3}, []int{1, 2, 3}},
		{"all same", []int{1, 1, 1}, []int{1}},
		{"preserves first-occurrence order", []int{3, 1, 2, 1, 3, 2}, []int{3, 1, 2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Distinct(fluent.From(tc.src)))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestZip(t *testing.T) {
	t.Parallel()

	t.Run("equal length", func(t *testing.T) {
		t.Parallel()
		a := fluent.From([]int{1, 2, 3})
		b := fluent.From([]string{"a", "b", "c"})
		var keys []int
		var vals []string
		for k, v := range fluent.Zip(a, b) {
			keys = append(keys, k)
			vals = append(vals, v)
		}
		assert.Equal(t, keys, []int{1, 2, 3})
		assert.Equal(t, vals, []string{"a", "b", "c"})
	})

	t.Run("a shorter", func(t *testing.T) {
		t.Parallel()
		a := fluent.From([]int{1, 2})
		b := fluent.From([]int{10, 20, 30})
		count := fluent.Count(fluent.KeysOf(fluent.Zip(a, b)))
		assert.Equal(t, count, 2)
	})

	t.Run("b shorter", func(t *testing.T) {
		t.Parallel()
		a := fluent.From([]int{1, 2, 3})
		b := fluent.From([]int{10})
		count := fluent.Count(fluent.KeysOf(fluent.Zip(a, b)))
		assert.Equal(t, count, 1)
	})

	t.Run("empty a", func(t *testing.T) {
		t.Parallel()
		a := fluent.From([]int{})
		b := fluent.From([]int{1, 2, 3})
		count := fluent.Count(fluent.KeysOf(fluent.Zip(a, b)))
		assert.Equal(t, count, 0)
	})
}

func TestGroupBy(t *testing.T) {
	t.Parallel()

	t.Run("empty source", func(t *testing.T) {
		t.Parallel()
		m := fluent.ToMap(fluent.GroupBy(fluent.From([]int{}), func(v int) int { return v % 2 }))
		assert.Equal(t, len(m), 0)
	})

	t.Run("groups by parity", func(t *testing.T) {
		t.Parallel()
		m := fluent.ToMap(fluent.GroupBy(fluent.From([]int{1, 2, 3, 4, 5, 6}), func(v int) int { return v % 2 }))
		assert.Equal(t, m[0], []int{2, 4, 6})
		assert.Equal(t, m[1], []int{1, 3, 5})
	})

	t.Run("preserves first-occurrence group order", func(t *testing.T) {
		t.Parallel()
		var order []int
		for k := range fluent.GroupBy(fluent.From([]int{3, 1, 2, 3, 1}), func(v int) int { return v }) {
			order = append(order, k)
		}
		assert.Equal(t, order, []int{3, 1, 2})
	})
}

func TestJoin(t *testing.T) {
	t.Parallel()

	t.Run("inner join drops unmatched a", func(t *testing.T) {
		t.Parallel()
		a := fluent.From([]int{1, 2, 3})
		b := fluent.From([]int{2, 3, 4})
		var pairs [][2]int
		for va, vb := range fluent.Join(a, b, func(v int) int { return v }, func(v int) int { return v }) {
			pairs = append(pairs, [2]int{va, vb})
		}
		assert.Equal(t, pairs, [][2]int{{2, 2}, {3, 3}})
	})

	t.Run("empty a yields nothing", func(t *testing.T) {
		t.Parallel()
		a := fluent.From([]int{})
		b := fluent.From([]int{1, 2, 3})
		count := 0
		for range fluent.Join(a, b, func(v int) int { return v }, func(v int) int { return v }) {
			count++
		}
		assert.Equal(t, count, 0)
	})

	t.Run("empty b yields nothing", func(t *testing.T) {
		t.Parallel()
		a := fluent.From([]int{1, 2, 3})
		b := fluent.From([]int{})
		count := 0
		for range fluent.Join(a, b, func(v int) int { return v }, func(v int) int { return v }) {
			count++
		}
		assert.Equal(t, count, 0)
	})
}

func TestLeftJoin(t *testing.T) {
	t.Parallel()

	t.Run("all a elements present, unmatched get zero B", func(t *testing.T) {
		t.Parallel()
		a := fluent.From([]int{1, 2, 3})
		b := fluent.From([]int{2, 3})
		matched := make(map[int]int)
		for va, vb := range fluent.LeftJoin(a, b, func(v int) int { return v }, func(v int) int { return v }) {
			matched[va] = vb
		}
		assert.Equal(t, len(matched), 3)
		assert.Equal(t, matched[1], 0) // zero value of int
		assert.Equal(t, matched[2], 2)
		assert.Equal(t, matched[3], 3)
	})

	t.Run("empty a yields nothing", func(t *testing.T) {
		t.Parallel()
		a := fluent.From([]int{})
		b := fluent.From([]int{1, 2})
		count := 0
		for range fluent.LeftJoin(a, b, func(v int) int { return v }, func(v int) int { return v }) {
			count++
		}
		assert.Equal(t, count, 0)
	})

	t.Run("empty b all a get zero value", func(t *testing.T) {
		t.Parallel()
		a := fluent.From([]int{1, 2, 3})
		b := fluent.From([]int{})
		count := 0
		for _, vb := range fluent.LeftJoin(a, b, func(v int) int { return v }, func(v int) int { return v }) {
			assert.Equal(t, vb, 0)
			count++
		}
		assert.Equal(t, count, 3)
	})
}

func TestUnion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b []int
		want []int
	}{
		{"both empty", []int{}, []int{}, []int{}},
		{"disjoint", []int{1, 2}, []int{3, 4}, []int{1, 2, 3, 4}},
		{"overlapping", []int{1, 2, 3}, []int{2, 3, 4}, []int{1, 2, 3, 4}},
		{"duplicates within a", []int{1, 1, 2}, []int{3}, []int{1, 2, 3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Union(fluent.From(tc.a), fluent.From(tc.b)))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestIntersect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b []int
		want []int
	}{
		{"both empty", []int{}, []int{}, []int{}},
		{"no common", []int{1, 2}, []int{3, 4}, []int{}},
		{"some common", []int{1, 2, 3}, []int{2, 3, 4}, []int{2, 3}},
		{"duplicates in a produces unique results", []int{1, 2, 2, 3}, []int{2, 3}, []int{2, 3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Intersect(fluent.From(tc.a), fluent.From(tc.b)))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestExcept(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b []int
		want []int
	}{
		{"both empty", []int{}, []int{}, []int{}},
		{"empty b returns all a", []int{1, 2, 3}, []int{}, []int{1, 2, 3}},
		{"remove some", []int{1, 2, 3, 4}, []int{2, 4}, []int{1, 3}},
		{"remove all", []int{1, 2, 3}, []int{1, 2, 3}, []int{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Except(fluent.From(tc.a), fluent.From(tc.b)))
			assert.Equal(t, got, tc.want)
		})
	}
}
