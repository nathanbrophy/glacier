// SPDX-License-Identifier: Apache-2.0

package figgen_test

import (
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/figgen"
)

func TestRender_RowCount(t *testing.T) {
	t.Parallel()
	rows := figgen.Render("GLACIER")
	assert.Equal(t, 5, len(rows))
}

func TestRender_NonEmpty(t *testing.T) {
	t.Parallel()
	rows := figgen.Render("GLACIER")
	for i, row := range rows {
		assert.True(t, row != "", "row[%d] is empty", i)
	}
}

func TestRender_AllRowsSameWidth(t *testing.T) {
	t.Parallel()
	rows := figgen.Render("GLACIER")
	require.Equal(t, 5, len(rows))
	// All rows in the rendered banner should have the same character width.
	width := len([]rune(rows[0]))
	for i, row := range rows[1:] {
		w := len([]rune(row))
		assert.True(t, w == width,
			"row[%d] has width %d, expected %d", i+1, w, width)
	}
}

func TestRender_EmptyString(t *testing.T) {
	t.Parallel()
	rows := figgen.Render("")
	assert.Equal(t, 5, len(rows))
	// Empty input: each row is an empty string (no glyphs concatenated).
	for i, row := range rows {
		assert.True(t, row == "",
			"row[%d] should be empty for empty input, got %q", i, row)
	}
}

func TestRender_StripUnknownCharacters(t *testing.T) {
	t.Parallel()
	// Characters outside [a-zA-Z0-9 _-] should be stripped.
	rows := figgen.Render("A!B")
	rowsAB := figgen.Render("AB")
	for i := range rows {
		assert.True(t, rows[i] == rowsAB[i],
			"row[%d]: '!' should be stripped; expected %q, got %q", i, rowsAB[i], rows[i])
	}
}

func TestRender_CaseInsensitive(t *testing.T) {
	t.Parallel()
	upper := figgen.Render("GLACIER")
	lower := figgen.Render("glacier")
	for i := range upper {
		assert.True(t, upper[i] == lower[i],
			"row[%d]: lower and upper case should produce same output; upper=%q lower=%q",
			i, upper[i], lower[i])
	}
}

func TestRender_SingleChar(t *testing.T) {
	t.Parallel()
	rows := figgen.Render("A")
	assert.Equal(t, 5, len(rows))
	// Row 0 of 'A' should contain block characters.
	assert.True(t, strings.ContainsAny(rows[0], "█╗╔╝╚║═"),
		"expected block chars in A row[0]: %q", rows[0])
}

func TestRender_SpaceGlyph(t *testing.T) {
	t.Parallel()
	rows := figgen.Render(" ")
	assert.Equal(t, 5, len(rows))
	// Space glyph renders as whitespace only.
	for i, row := range rows {
		assert.True(t, strings.TrimSpace(row) == "",
			"row[%d] should be whitespace only for space glyph, got: %q", i, row)
	}
}

// Example is the canonical package example test.
func Example() {
	rows := figgen.Render("GLACIER")
	_ = len(rows) // 5
}
