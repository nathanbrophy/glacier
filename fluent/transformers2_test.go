// SPDX-License-Identifier: Apache-2.0

package fluent_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/fluent"
)

func TestMap2(t *testing.T) {
	t.Parallel()

	t.Run("empty source", func(t *testing.T) {
		t.Parallel()
		src := fluent.FromMap(map[string]int{})
		got := fluent.ToMap(fluent.Map2(src, func(k string, v int) (string, int) {
			return k + "!", v * 10
		}))
		assert.Equal(t, len(got), 0)
	})

	t.Run("transforms keys and values", func(t *testing.T) {
		t.Parallel()
		// Use a deterministic single-entry map.
		src := fluent.FromMap(map[string]int{"a": 1})
		got := fluent.ToMap(fluent.Map2(src, func(k string, v int) (string, int) {
			return k + "!", v * 10
		}))
		assert.Equal(t, got, map[string]int{"a!": 10})
	})
}

func TestFilter2(t *testing.T) {
	t.Parallel()

	t.Run("empty source", func(t *testing.T) {
		t.Parallel()
		src := fluent.FromMap(map[string]int{})
		got := fluent.ToMap(fluent.Filter2(src, func(k string, v int) bool { return v > 0 }))
		assert.Equal(t, len(got), 0)
	})

	t.Run("keeps matching pairs", func(t *testing.T) {
		t.Parallel()
		// Build a Seq2 from a stable slice of KV pairs.
		kvs := []fluent.KV[string, int]{
			{K: "a", V: 1},
			{K: "b", V: -2},
			{K: "c", V: 3},
		}
		src := fluent.Pairs(fluent.From(kvs))
		got := fluent.ToMap(fluent.Filter2(src, func(_ string, v int) bool { return v > 0 }))
		assert.Equal(t, got, map[string]int{"a": 1, "c": 3})
	})
}

func TestKeysOf(t *testing.T) {
	t.Parallel()

	t.Run("empty source", func(t *testing.T) {
		t.Parallel()
		got := fluent.ToSlice(fluent.KeysOf(fluent.FromMap(map[string]int{})))
		assert.Equal(t, got, []string{})
	})

	t.Run("extracts keys in order", func(t *testing.T) {
		t.Parallel()
		kvs := []fluent.KV[int, string]{
			{K: 1, V: "a"},
			{K: 2, V: "b"},
			{K: 3, V: "c"},
		}
		src := fluent.Pairs(fluent.From(kvs))
		got := fluent.ToSlice(fluent.KeysOf(src))
		assert.Equal(t, got, []int{1, 2, 3})
	})
}

func TestValuesOf(t *testing.T) {
	t.Parallel()

	t.Run("empty source", func(t *testing.T) {
		t.Parallel()
		got := fluent.ToSlice(fluent.ValuesOf(fluent.FromMap(map[string]int{})))
		assert.Equal(t, got, []int{})
	})

	t.Run("extracts values in order", func(t *testing.T) {
		t.Parallel()
		kvs := []fluent.KV[int, string]{
			{K: 1, V: "x"},
			{K: 2, V: "y"},
			{K: 3, V: "z"},
		}
		src := fluent.Pairs(fluent.From(kvs))
		got := fluent.ToSlice(fluent.ValuesOf(src))
		assert.Equal(t, got, []string{"x", "y", "z"})
	})
}

func TestEntries(t *testing.T) {
	t.Parallel()

	t.Run("empty source", func(t *testing.T) {
		t.Parallel()
		got := fluent.ToSlice(fluent.Entries(fluent.FromMap(map[string]int{})))
		assert.Equal(t, got, []fluent.KV[string, int]{})
	})

	t.Run("entries round-trip via Pairs", func(t *testing.T) {
		t.Parallel()
		kvs := []fluent.KV[string, int]{
			{K: "a", V: 1},
			{K: "b", V: 2},
			{K: "c", V: 3},
		}
		seq := fluent.From(kvs)
		got := fluent.ToSlice(fluent.Entries(fluent.Pairs(seq)))
		assert.Equal(t, got, kvs)
	})
}
