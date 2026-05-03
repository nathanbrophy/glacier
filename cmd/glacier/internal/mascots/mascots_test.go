// SPDX-License-Identifier: Apache-2.0

package mascots_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/mascots"
)

func TestAll_Count(t *testing.T) {
	t.Parallel()
	all := mascots.All()
	assert.Equal(t, 6, len(all))
}

func TestAll_IDs(t *testing.T) {
	t.Parallel()
	wantIDs := []string{"polar_bear", "penguin", "owl", "fox", "otter", "raccoon"}
	got := mascots.All()
	for i, want := range wantIDs {
		assert.True(t, got[i].ID == want,
			"mascot[%d].ID: got %q, want %q", i, got[i].ID, want)
	}
}

func TestAll_NoEmptyFields(t *testing.T) {
	t.Parallel()
	for _, m := range mascots.All() {
		assert.True(t, m.ID != "", "mascot %q has empty ID", m.Display)
		assert.True(t, m.Display != "", "mascot %q has empty Display", m.ID)
		assert.True(t, m.Kaomoji != "", "mascot %q has empty Kaomoji", m.ID)
		assert.True(t, len(m.Banner) == 5,
			"mascot %q Banner should have 5 rows, got %d", m.ID, len(m.Banner))
		for i, row := range m.Banner {
			assert.True(t, row != "", "mascot %q Banner[%d] is empty", m.ID, i)
		}
	}
}

func TestAll_Immutable(t *testing.T) {
	t.Parallel()
	a := mascots.All()
	b := mascots.All()
	a[0].Kaomoji = "mutated"
	assert.NotEqual(t, "mutated", b[0].Kaomoji)
}

func TestGet_KnownIDs(t *testing.T) {
	t.Parallel()
	ids := []string{"polar_bear", "penguin", "owl", "fox", "otter", "raccoon"}
	for _, id := range ids {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			m := mascots.Get(id)
			assert.Equal(t, id, m.ID)
		})
	}
}

func TestGet_UnknownID_ReturnsPolarBear(t *testing.T) {
	t.Parallel()
	m := mascots.Get("does-not-exist")
	assert.True(t, m.ID == "polar_bear",
		"unknown ID should return polar bear default, got %q", m.ID)
}

func TestGet_EmptyID_ReturnsPolarBear(t *testing.T) {
	t.Parallel()
	m := mascots.Get("")
	assert.Equal(t, "polar_bear", m.ID)
}

func TestGet_PolarBearKaomoji(t *testing.T) {
	t.Parallel()
	m := mascots.Get("polar_bear")
	assert.Equal(t, "ʕ•ᴥ•ʔ", m.Kaomoji)
}

// Example is the canonical package example test.
func Example() {
	m := mascots.Get("polar_bear")
	_ = m.Kaomoji // "ʕ•ᴥ•ʔ"
}
