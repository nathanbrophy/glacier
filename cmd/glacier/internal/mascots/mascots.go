// SPDX-License-Identifier: Apache-2.0

// Package mascots provides the curated library of six Glacier mascots. Each
// mascot has a stable ID, a display name, a single-line kaomoji, and a 5-line
// block-character banner. The default mascot is the polar bear.
package mascots

// Mascot is one curated entry from the Glacier mascot library.
type Mascot struct {
	// ID is the stable machine-readable identifier (e.g. "polar_bear").
	ID string
	// Display is the human-readable name (e.g. "Polar Bear").
	Display string
	// Kaomoji is the single-line text art representation.
	Kaomoji string
	// Banner is the 5-line block-character art representation.
	Banner []string
}

// all is the ordered list of curated mascots. Polar bear is first and is the
// default returned by Get when an ID is not found.
var all = []Mascot{
	{
		ID:      "polar_bear",
		Display: "Polar Bear",
		Kaomoji: "ʕ•ᴥ•ʔ",
		Banner: []string{
			"  ▟▀▙   ▟▀▙  ",
			" ▟████████▙ ",
			" █ ● ▼  ● █ ",
			" ▀▀▀▀▀▀▀▀▀▀ ",
			"   ʕ•ᴥ•ʔ   ",
		},
	},
	{
		ID:      "penguin",
		Display: "Penguin",
		Kaomoji: "<(•^•)>",
		Banner: []string{
			"  ▄███▄  ",
			" █ ◉ ◉ █ ",
			" █  ▽  █ ",
			"  ▀▄█▄▀  ",
			" <(•^•)> ",
		},
	},
	{
		ID:      "owl",
		Display: "Owl",
		Kaomoji: "(o,o)",
		Banner: []string{
			" ▄▀▀▀▀▀▄ ",
			"█ ◉   ◉ █",
			"█   ▲   █",
			" ▀▄▄▄▄▄▀ ",
			"  (o,o)  ",
		},
	},
	{
		ID:      "fox",
		Display: "Fox",
		Kaomoji: "^..^",
		Banner: []string{
			"▄▀▄   ▄▀▄",
			"█▀█▄▄▄█▀█",
			" █ ◕◕ █  ",
			"  █  █   ",
			"  ^..^   ",
		},
	},
	{
		ID:      "otter",
		Display: "Otter",
		Kaomoji: "ʕ•˦•ʔ",
		Banner: []string{
			" ▄████▄  ",
			"█ ◉  ◉ █ ",
			"█  ══  █ ",
			" ▀████▀  ",
			" ʕ•˦•ʔ  ",
		},
	},
	{
		ID:      "raccoon",
		Display: "Raccoon",
		Kaomoji: "(^-ω-^)",
		Banner: []string{
			"▄▀▄▀▀▀▄▀▄",
			"█▄ ◕ ◕ ▄█",
			"  █ ω █  ",
			"   ▀▄▀   ",
			" (^-ω-^) ",
		},
	},
}

// All returns all curated mascots in display order.
func All() []Mascot {
	out := make([]Mascot, len(all))
	copy(out, all)
	return out
}

// Get returns the mascot with the given ID. If no mascot matches, Get returns
// the polar bear (the default mascot).
func Get(id string) Mascot {
	for _, m := range all {
		if m.ID == id {
			return m
		}
	}
	return all[0] // polar bear default
}
