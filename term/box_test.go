// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/term"
)

func TestBoxRoundedCorners(t *testing.T) {
	t.Parallel()
	got := term.Box("hello", term.WithRoundedCorners())
	// On a UTF-8 capable environment we'd see ╭/╮/╰/╯.
	// On ASCII fallback we'd see +.
	// Either way the box must contain the text.
	assert.Contains(t, got, "hello")
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	assert.GreaterOrEqual(t, len(lines), 3, "Box: expected at least 3 lines (top, content, bottom)")
}

func TestBoxSharpCorners(t *testing.T) {
	t.Parallel()
	got := term.Box("hi", term.WithSharpCorners())
	assert.Contains(t, got, "hi")
}

func TestBoxDoubleBorders(t *testing.T) {
	t.Parallel()
	got := term.Box("hi", term.WithDoubleBorders())
	assert.Contains(t, got, "hi")
}

func TestBoxWithTitle(t *testing.T) {
	t.Parallel()
	// Use content wide enough that the title fits in the top border.
	got := term.Box("content to make the box wide enough for the title", term.WithTitle("TITLE"))
	assert.Contains(t, got, "TITLE")
	assert.Contains(t, got, "content")
}

func TestBoxWithPadding(t *testing.T) {
	t.Parallel()
	got := term.Box("x", term.WithPadding(1, 2, 1, 2))
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	// With top+bottom padding of 1, content+padding = 5 lines (top border, top pad, content, bottom pad, bottom border).
	assert.GreaterOrEqual(t, len(lines), 5, "Box(padding) expected ≥ 5 lines")
}

func TestBoxWidthExceedsTerminal(t *testing.T) {
	t.Parallel()
	// Generate content wider than 80 chars (default terminal width).
	content := strings.Repeat("A", 200)
	got := term.Box(content)
	lines := strings.Split(got, "\n")
	for _, l := range lines {
		w := utf8.RuneCountInString(l)
		assert.True(t, w <= 82, "Box line too wide: "+fmt.Sprintf("%d runes (limit ~82): %q", w, l))
	}
}

func TestBoxMultiLine(t *testing.T) {
	t.Parallel()
	got := term.Box("line one\nline two\nline three")
	assert.Contains(t, got, "line one")
	assert.Contains(t, got, "line three")
}

func TestBoxNoOptions(t *testing.T) {
	t.Parallel()
	got := term.Box("test")
	assert.NotEqual(t, got, "")
	assert.Contains(t, got, "test")
}
