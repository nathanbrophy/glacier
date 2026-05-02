// SPDX-License-Identifier: Apache-2.0

package fluent_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/fluent"
)

func TestSort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		want []int
	}{
		{"empty source", []int{}, []int{}},
		{"already sorted", []int{1, 2, 3, 4}, []int{1, 2, 3, 4}},
		{"reverse order", []int{5, 3, 1, 4, 2}, []int{1, 2, 3, 4, 5}},
		{"duplicates", []int{3, 1, 2, 1, 3}, []int{1, 1, 2, 3, 3}},
		{"single", []int{42}, []int{42}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.Sort(fluent.From(tc.src)))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestSortDesc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  []int
		want []int
	}{
		{"empty source", []int{}, []int{}},
		{"ascending input", []int{1, 2, 3, 4, 5}, []int{5, 4, 3, 2, 1}},
		{"already descending", []int{5, 4, 3, 2, 1}, []int{5, 4, 3, 2, 1}},
		{"duplicates", []int{3, 1, 2, 1, 3}, []int{3, 3, 2, 1, 1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fluent.ToSlice(fluent.SortDesc(fluent.From(tc.src)))
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestSortBy(t *testing.T) {
	t.Parallel()

	type item struct {
		name string
		rank int
	}

	t.Run("empty source", func(t *testing.T) {
		t.Parallel()
		got := fluent.ToSlice(fluent.SortBy(fluent.From([]item{}), func(v item) int { return v.rank }))
		assert.Equal(t, got, []item{})
	})

	t.Run("sort by rank ascending", func(t *testing.T) {
		t.Parallel()
		src := []item{{"c", 3}, {"a", 1}, {"b", 2}}
		got := fluent.ToSlice(fluent.SortBy(fluent.From(src), func(v item) int { return v.rank }))
		assert.Equal(t, got, []item{{"a", 1}, {"b", 2}, {"c", 3}})
	})

	t.Run("sort strings by length", func(t *testing.T) {
		t.Parallel()
		src := []string{"banana", "fig", "apple", "kiwi"}
		got := fluent.ToSlice(fluent.SortBy(fluent.From(src), func(s string) int { return len(s) }))
		assert.Equal(t, got, []string{"fig", "kiwi", "apple", "banana"})
	})
}

func TestSortStable(t *testing.T) {
	t.Parallel()

	type item struct {
		name string
		rank int
	}

	t.Run("empty source", func(t *testing.T) {
		t.Parallel()
		got := fluent.ToSlice(fluent.SortStable(fluent.From([]item{}), func(a, b item) int {
			if a.rank < b.rank {
				return -1
			}
			if a.rank > b.rank {
				return 1
			}
			return 0
		}))
		assert.Equal(t, got, []item{})
	})

	t.Run("equal elements preserve original order", func(t *testing.T) {
		t.Parallel()
		// Items with the same rank should stay in their original relative order.
		src := []item{{"first", 1}, {"second", 2}, {"third", 1}, {"fourth", 2}}
		cmp := func(a, b item) int {
			if a.rank < b.rank {
				return -1
			}
			if a.rank > b.rank {
				return 1
			}
			return 0
		}
		got := fluent.ToSlice(fluent.SortStable(fluent.From(src), cmp))
		// rank-1 items: "first" before "third"; rank-2 items: "second" before "fourth".
		assert.Equal(t, got, []item{{"first", 1}, {"third", 1}, {"second", 2}, {"fourth", 2}})
	})
}
