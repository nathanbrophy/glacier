// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/nathanbrophy/glacier/term"
)

func TestBoxRoundedCorners(t *testing.T) {
	t.Parallel()
	got := term.Box("hello", term.WithRoundedCorners())
	// On a UTF-8 capable environment we'd see ╭/╮/╰/╯.
	// On ASCII fallback we'd see +.
	// Either way the box must contain the text.
	if !strings.Contains(got, "hello") {
		t.Errorf("Box output missing content 'hello': %q", got)
	}
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) < 3 {
		t.Errorf("Box: expected at least 3 lines (top, content, bottom), got %d", len(lines))
	}
}

func TestBoxSharpCorners(t *testing.T) {
	t.Parallel()
	got := term.Box("hi", term.WithSharpCorners())
	if !strings.Contains(got, "hi") {
		t.Errorf("Box(sharp) missing content: %q", got)
	}
}

func TestBoxDoubleBorders(t *testing.T) {
	t.Parallel()
	got := term.Box("hi", term.WithDoubleBorders())
	if !strings.Contains(got, "hi") {
		t.Errorf("Box(double) missing content: %q", got)
	}
}

func TestBoxWithTitle(t *testing.T) {
	t.Parallel()
	// Use content wide enough that the title fits in the top border.
	got := term.Box("content to make the box wide enough for the title", term.WithTitle("TITLE"))
	if !strings.Contains(got, "TITLE") {
		t.Errorf("Box(title) missing title in output: %q", got)
	}
	if !strings.Contains(got, "content") {
		t.Errorf("Box(title) missing content: %q", got)
	}
}

func TestBoxWithPadding(t *testing.T) {
	t.Parallel()
	got := term.Box("x", term.WithPadding(1, 2, 1, 2))
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	// With top+bottom padding of 1, content+padding = 5 lines (top border, top pad, content, bottom pad, bottom border).
	if len(lines) < 5 {
		t.Errorf("Box(padding) expected ≥ 5 lines, got %d:\n%s", len(lines), got)
	}
}

func TestBoxWidthExceedsTerminal(t *testing.T) {
	t.Parallel()
	// Generate content wider than 80 chars (default terminal width).
	content := strings.Repeat("A", 200)
	got := term.Box(content)
	lines := strings.Split(got, "\n")
	for _, l := range lines {
		w := utf8.RuneCountInString(l)
		if w > 82 { // 80 content + 2 border
			t.Errorf("Box line too wide: %d runes (limit ~82): %q", w, l)
		}
	}
}

func TestBoxMultiLine(t *testing.T) {
	t.Parallel()
	got := term.Box("line one\nline two\nline three")
	if !strings.Contains(got, "line one") || !strings.Contains(got, "line three") {
		t.Errorf("Box(multiline) missing expected content: %q", got)
	}
}

func TestBoxNoOptions(t *testing.T) {
	t.Parallel()
	got := term.Box("test")
	if got == "" {
		t.Error("Box() returned empty string")
	}
	if !strings.Contains(got, "test") {
		t.Errorf("Box() missing content: %q", got)
	}
}
