// SPDX-License-Identifier: Apache-2.0

package fluent_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/fluent"
)

func TestPropertyToSliceRoundTrip(t *testing.T) {
	t.Parallel()

	cases := [][]int{
		{},
		{1},
		{1, 2, 3},
		{5, 4, 3, 2, 1},
	}
	for _, s := range cases {
		got := fluent.ToSlice(fluent.From(s))
		assert.Equal(t, got, s)
	}
}

func TestPropertyCountMatchesLen(t *testing.T) {
	t.Parallel()

	cases := [][]int{
		{},
		{1},
		{1, 2, 3, 4, 5},
	}
	for _, s := range cases {
		got := fluent.Count(fluent.From(s))
		assert.Equal(t, got, len(s))
	}
}

func TestPropertyReduceSum(t *testing.T) {
	t.Parallel()

	cases := []struct {
		src  []int
		want int
	}{
		{[]int{}, 0},
		{[]int{1, 2, 3}, 6},
		{[]int{-1, 1}, 0},
	}
	for _, tc := range cases {
		got := fluent.Reduce(fluent.From(tc.src), 0, func(acc, v int) int { return acc + v })
		assert.Equal(t, got, tc.want)
	}
}

func TestPropertyTakeSlice(t *testing.T) {
	t.Parallel()

	s := []int{1, 2, 3, 4, 5}
	cases := []int{0, 1, 3, 5, 10}
	for _, n := range cases {
		got := fluent.ToSlice(fluent.Take(fluent.From(s), n))
		limit := n
		if limit > len(s) {
			limit = len(s)
		}
		assert.Equal(t, got, s[:limit])
	}
}

func TestPropertyDropAll(t *testing.T) {
	t.Parallel()

	s := []int{1, 2, 3, 4, 5}
	got := fluent.ToSlice(fluent.Drop(fluent.From(s), len(s)))
	assert.Equal(t, got, []int{})
}

func TestPropertyDropMoreThanLen(t *testing.T) {
	t.Parallel()

	s := []int{1, 2, 3}
	got := fluent.ToSlice(fluent.Drop(fluent.From(s), 100))
	assert.Equal(t, got, []int{})
}

func TestPropertyEntriesPairsRoundTrip(t *testing.T) {
	t.Parallel()

	kvs := []fluent.KV[string, int]{
		{K: "a", V: 1},
		{K: "b", V: 2},
		{K: "c", V: 3},
	}
	got := fluent.ToSlice(fluent.Entries(fluent.Pairs(fluent.From(kvs))))
	assert.Equal(t, got, kvs)
}

func TestPropertyDistinct(t *testing.T) {
	t.Parallel()

	src := []int{1, 1, 2}
	got := fluent.ToSlice(fluent.Distinct(fluent.From(src)))
	assert.Equal(t, got, []int{1, 2})
}
