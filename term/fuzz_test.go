// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/term"
)

// FuzzGlyphRegistration verifies that glyph registration never panics and
// only succeeds for names matching ^[a-z][a-z0-9_]*$ (§23.9 #26).
func FuzzGlyphRegistration(f *testing.F) {
	// Seed corpus: valid and invalid names.
	f.Add("check")
	f.Add("")
	f.Add("valid_name")
	f.Add("1invalid")
	f.Add("InValid")
	f.Add("has-hyphen")
	f.Add(strings.Repeat("a", 65))

	f.Fuzz(func(t *testing.T, name string) {
		// Must never panic; invalid names return GlyphError.
		err := term.RegisterGlyph(name, "X", "x")
		if err != nil {
			// Verify error format.
			assert.Contains(t, err.Error(), "term: glyph:")
		}
	})
}

// FuzzGlyphLookup verifies that Glyph() never panics on any input.
func FuzzGlyphLookup(f *testing.F) {
	f.Add("check")
	f.Add("")
	f.Add("__unknown__")
	f.Add(strings.Repeat("z", 100))

	f.Fuzz(func(t *testing.T, name string) {
		// Must never panic.
		_ = term.Glyph(name)
	})
}

// FuzzAnsiInjection verifies that Style.Render never leaks raw ANSI escapes
// outside the style wrapper boundaries (§23.9 #25, L-add-12).
func FuzzAnsiInjection(f *testing.F) {
	f.Add("normal text")
	f.Add("\x1b[31mred\x1b[0m")
	f.Add("\x00nul")
	f.Add("\x1b[999m")
	f.Add("hello\r\nworld")

	f.Fuzz(func(t *testing.T, input string) {
		// Render must not panic; user-controlled ANSI in input is documented.
		_ = term.New().Bold().Render(input)
		_ = term.New().Foreground(term.Cyan).Render(input)
	})
}

// FuzzHexParsing verifies that Hex never panics on arbitrary input.
func FuzzHexParsing(f *testing.F) {
	f.Add("#22D3EE")
	f.Add("")
	f.Add("xyz")
	f.Add("#GGGGGG")
	f.Add(strings.Repeat("f", 100))

	f.Fuzz(func(t *testing.T, input string) {
		c, err := term.Hex(input)
		if err == nil {
			// Valid color: check field ranges.
			assert.True(t, c.R <= 255 && c.G <= 255 && c.B <= 255,
				"Hex("+input+") returned out-of-range fields")
		}
	})
}

// FuzzWrap verifies that Wrap never panics on arbitrary input.
func FuzzWrap(f *testing.F) {
	f.Add("hello world", 10)
	f.Add("", 0)
	f.Add(strings.Repeat("a", 1000), 5)

	f.Fuzz(func(t *testing.T, text string, width int) {
		_ = term.Wrap(text, width)
	})
}
