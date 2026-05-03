// SPDX-License-Identifier: Apache-2.0

package vibetips_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/vibetips"
)

func TestAll_Count(t *testing.T) {
	t.Parallel()
	tips := vibetips.All()
	assert.Equal(t, 12, len(tips))
}

func TestAll_NoEmptyFields(t *testing.T) {
	t.Parallel()
	for i, tip := range vibetips.All() {
		assert.True(t, tip.Category != "", "tip[%d].Category is empty", i)
		assert.True(t, tip.Body != "", "tip[%d].Body is empty", i)
		assert.True(t, tip.SpecRef != "", "tip[%d].SpecRef is empty", i)
	}
}

func TestAll_Immutable(t *testing.T) {
	t.Parallel()
	a := vibetips.All()
	b := vibetips.All()
	a[0].Body = "mutated"
	assert.NotEqual(t, "mutated", b[0].Body)
}

func TestShuffled_Count(t *testing.T) {
	t.Parallel()
	shuffled := vibetips.Shuffled(42)
	assert.Equal(t, 12, len(shuffled))
}

func TestShuffled_Deterministic(t *testing.T) {
	t.Parallel()
	a := vibetips.Shuffled(7)
	b := vibetips.Shuffled(7)
	require.Equal(t, len(a), len(b))
	for i := range a {
		assert.True(t, a[i].Body == b[i].Body,
			"tip[%d].Body differed between calls with same seed: %q vs %q", i, a[i].Body, b[i].Body)
	}
}

func TestShuffled_DifferentSeeds(t *testing.T) {
	t.Parallel()
	a := vibetips.Shuffled(1)
	b := vibetips.Shuffled(2)
	// Different seeds should (almost certainly) produce a different order.
	same := true
	for i := range a {
		if a[i].Body != b[i].Body {
			same = false
			break
		}
	}
	assert.False(t, same, "expected different seeds to produce different orderings")
}

func TestShuffled_ContainsSameTips(t *testing.T) {
	t.Parallel()
	all := vibetips.All()
	shuffled := vibetips.Shuffled(99)
	allSet := make(map[string]bool, len(all))
	for _, tip := range all {
		allSet[tip.Body] = true
	}
	for _, tip := range shuffled {
		assert.True(t, allSet[tip.Body], "shuffled tip not found in All(): %q", tip.Body)
	}
}

func TestAll_UniqueCategories(t *testing.T) {
	t.Parallel()
	// Every tip should document a different framework package.
	seen := make(map[string]bool)
	for _, tip := range vibetips.All() {
		assert.False(t, seen[tip.Category], "duplicate category %q in tip registry", tip.Category)
		seen[tip.Category] = true
	}
}

// Example is the canonical package example test.
func Example() {
	tips := vibetips.All()
	_ = tips[0].Body
}
