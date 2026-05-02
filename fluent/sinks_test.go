// SPDX-License-Identifier: Apache-2.0

package fluent_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/fluent"
)

func TestReduce(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		want int
	}{
		{"empty source returns zero", []int{}, 0},
		{"sum", []int{1, 2, 3, 4, 5}, 15},
		{"single element", []int{42}, 42},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.Reduce(fluent.From(tc.src), 0, func(acc, v int) int { return acc + v })
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestToSlice(t *testing.T) {
	t.Parallel()

	t.Run("empty source returns non-nil empty slice", func(t *testing.T) {
		t.Parallel()
		got := fluent.ToSlice(fluent.From([]int{}))
		assert.True(t, got != nil, "expected non-nil slice")
		assert.Equal(t, len(got), 0)
	})

	t.Run("nil source returns non-nil empty slice", func(t *testing.T) {
		t.Parallel()
		got := fluent.ToSlice(fluent.From[int](nil))
		assert.True(t, got != nil, "expected non-nil slice")
		assert.Equal(t, len(got), 0)
	})

	t.Run("preserves elements", func(t *testing.T) {
		t.Parallel()
		got := fluent.ToSlice(fluent.From([]int{1, 2, 3, 4, 5}))
		assert.Equal(t, got, []int{1, 2, 3, 4, 5})
	})
}

func TestToMap(t *testing.T) {
	t.Parallel()

	t.Run("empty source returns non-nil empty map", func(t *testing.T) {
		t.Parallel()
		got := fluent.ToMap(fluent.FromMap(map[string]int{}))
		assert.True(t, got != nil, "expected non-nil map")
		assert.Equal(t, len(got), 0)
	})

	t.Run("last value wins on duplicate key", func(t *testing.T) {
		t.Parallel()
		kvs := []fluent.KV[string, int]{
			{K: "a", V: 1},
			{K: "a", V: 99},
			{K: "b", V: 2},
		}
		got := fluent.ToMap(fluent.Pairs(fluent.From(kvs)))
		assert.Equal(t, got["a"], 99)
		assert.Equal(t, got["b"], 2)
	})
}

func TestCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		want int
	}{
		{"empty", []int{}, 0},
		{"three elements", []int{1, 2, 3}, 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.Count(fluent.From(tc.src))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestFirst(t *testing.T) {
	t.Parallel()

	t.Run("empty returns zero and false", func(t *testing.T) {
		t.Parallel()
		v, ok := fluent.First(fluent.From([]int{}))
		assert.Equal(t, v, 0)
		assert.Equal(t, ok, false)
	})

	t.Run("returns first element", func(t *testing.T) {
		t.Parallel()
		v, ok := fluent.First(fluent.From([]int{10, 20, 30}))
		assert.Equal(t, v, 10)
		assert.Equal(t, ok, true)
	})
}

func TestLast(t *testing.T) {
	t.Parallel()

	t.Run("empty returns zero and false", func(t *testing.T) {
		t.Parallel()
		v, ok := fluent.Last(fluent.From([]int{}))
		assert.Equal(t, v, 0)
		assert.Equal(t, ok, false)
	})

	t.Run("returns last element", func(t *testing.T) {
		t.Parallel()
		v, ok := fluent.Last(fluent.From([]int{10, 20, 30}))
		assert.Equal(t, v, 30)
		assert.Equal(t, ok, true)
	})
}

func TestAny(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		want bool
	}{
		{"empty returns false", []int{}, false},
		{"no match", []int{1, 3, 5}, false},
		{"one match", []int{1, 2, 3}, true},
		{"all match", []int{2, 4, 6}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.Any(fluent.From(tc.src), func(v int) bool { return v%2 == 0 })
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		want bool
	}{
		{"empty returns true", []int{}, true},
		{"all match", []int{2, 4, 6}, true},
		{"one mismatch", []int{2, 3, 6}, false},
		{"none match", []int{1, 3, 5}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.All(fluent.From(tc.src), func(v int) bool { return v%2 == 0 })
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestSum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		want int
	}{
		{"empty returns zero", []int{}, 0},
		{"positive", []int{1, 2, 3, 4, 5}, 15},
		{"negative values", []int{-1, -2, -3}, -6},
		{"mixed", []int{-1, 2, -3, 4}, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.Sum(fluent.From(tc.src))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestAvg(t *testing.T) {
	t.Parallel()

	t.Run("empty returns 0.0", func(t *testing.T) {
		t.Parallel()
		got := fluent.Avg(fluent.From([]int{}))
		assert.Equal(t, got, 0.0)
	})

	t.Run("integer average", func(t *testing.T) {
		t.Parallel()
		got := fluent.Avg(fluent.From([]int{1, 2, 3, 4, 5}))
		assert.Equal(t, got, 3.0)
	})

	t.Run("float average", func(t *testing.T) {
		t.Parallel()
		got := fluent.Avg(fluent.From([]float64{1.0, 2.0}))
		assert.Equal(t, got, 1.5)
	})
}

func TestMin(t *testing.T) {
	t.Parallel()

	t.Run("empty returns zero and false", func(t *testing.T) {
		t.Parallel()
		v, ok := fluent.Min(fluent.From([]int{}))
		assert.Equal(t, v, 0)
		assert.Equal(t, ok, false)
	})

	tests := []struct {
		name string
		src  []int
		want int
	}{
		{"single", []int{42}, 42},
		{"multiple", []int{3, 1, 4, 1, 5, 9}, 1},
		{"negative", []int{-5, -1, -3}, -5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := fluent.Min(fluent.From(tc.src))
			assert.Equal(t, ok, true)
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestMax(t *testing.T) {
	t.Parallel()

	t.Run("empty returns zero and false", func(t *testing.T) {
		t.Parallel()
		v, ok := fluent.Max(fluent.From([]int{}))
		assert.Equal(t, v, 0)
		assert.Equal(t, ok, false)
	})

	tests := []struct {
		name string
		src  []int
		want int
	}{
		{"single", []int{42}, 42},
		{"multiple", []int{3, 1, 4, 1, 5, 9}, 9},
		{"negative", []int{-5, -1, -3}, -1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := fluent.Max(fluent.From(tc.src))
			assert.Equal(t, ok, true)
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestMinBy(t *testing.T) {
	t.Parallel()

	type item struct {
		name string
		val  int
	}

	t.Run("empty returns zero and false", func(t *testing.T) {
		t.Parallel()
		v, ok := fluent.MinBy(fluent.From([]item{}), func(it item) int { return it.val })
		assert.Equal(t, v, item{})
		assert.Equal(t, ok, false)
	})

	t.Run("returns item with min key", func(t *testing.T) {
		t.Parallel()
		src := []item{{"b", 2}, {"a", 1}, {"c", 3}}
		got, ok := fluent.MinBy(fluent.From(src), func(it item) int { return it.val })
		assert.Equal(t, ok, true)
		assert.Equal(t, got, item{"a", 1})
	})
}

func TestMaxBy(t *testing.T) {
	t.Parallel()

	type item struct {
		name string
		val  int
	}

	t.Run("empty returns zero and false", func(t *testing.T) {
		t.Parallel()
		v, ok := fluent.MaxBy(fluent.From([]item{}), func(it item) int { return it.val })
		assert.Equal(t, v, item{})
		assert.Equal(t, ok, false)
	})

	t.Run("returns item with max key", func(t *testing.T) {
		t.Parallel()
		src := []item{{"b", 2}, {"c", 3}, {"a", 1}}
		got, ok := fluent.MaxBy(fluent.From(src), func(it item) int { return it.val })
		assert.Equal(t, ok, true)
		assert.Equal(t, got, item{"c", 3})
	})
}
