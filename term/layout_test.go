// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/term"
)

func TestCenterPadsToWidth(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text  string
		width int
		want  string
	}{
		{"hi", 6, "  hi  "},
		{"hi", 5, " hi  "},        // odd: right gets extra
		{"hello", 5, "hello"},     // exactly width
		{"toolong", 4, "toolong"}, // longer than width, unchanged
		{"a", 4, " a  "},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()
			got := term.Center(tc.text, tc.width)
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestCenterMultiLine(t *testing.T) {
	t.Parallel()
	got := term.Center("ab\ncd", 6)
	lines := strings.Split(got, "\n")
	require.Equal(t, len(lines), 2)
	for _, l := range lines {
		assert.Equal(t, utf8.RuneCountInString(l), 6)
	}
}

func TestJustify(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text  string
		width int
	}{
		{"hello world", 20},
		{"one two three", 20},
		{"single", 10}, // single word: unchanged
		{"", 10},       // empty: unchanged
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()
			got := term.Justify(tc.text, tc.width)
			words := strings.Fields(tc.text)
			if len(words) >= 2 {
				gotW := utf8.RuneCountInString(got)
				assert.Equal(t, gotW, tc.width)
			}
		})
	}
}

func TestPad(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text  string
		left  int
		right int
		want  string
	}{
		{"hi", 2, 3, "  hi   "},
		{"hi", 0, 0, "hi"},
		{"hi", 1, 0, " hi"},
		{"hi", 0, 1, "hi "},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()
			got := term.Pad(tc.text, tc.left, tc.right)
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text     string
		width    int
		ellipsis string
		want     string
	}{
		{"hello", 10, "...", "hello"}, // shorter than width
		{"hello", 5, "...", "hello"},  // exactly width
		{"hello world", 8, "...", "hello..."},
		{"hello world", 8, "…", "hello w…"}, // width 8 total: 7 chars + 1-char ellipsis
		{"abc", 2, "...", "..."},            // keep=0: just ellipsis
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()
			got := term.Truncate(tc.text, tc.width, tc.ellipsis)
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text  string
		width int
	}{
		{"hello world foo bar", 10},
		{"averylongwordthatexceedsthewidth", 5},
		{"short", 80},
		{"", 10},
		{"multi\nline\ninput", 20},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()
			got := term.Wrap(tc.text, tc.width)
			for _, line := range strings.Split(got, "\n") {
				w := utf8.RuneCountInString(line)
				assert.True(t, w <= tc.width,
					fmt.Sprintf("Wrap(%q, %d): line %q has width %d > %d", tc.text, tc.width, line, w, tc.width))
			}
		})
	}
}

func TestColumns(t *testing.T) {
	t.Parallel()
	rows := [][]string{
		{"Name", "Age", "City"},
		{"Alice", "30", "NYC"},
		{"Bob", "25", "LA"},
	}
	got := term.Columns(rows)
	assert.NotEqual(t, got, "")
	// Must contain all cell values.
	for _, row := range rows {
		for _, cell := range row {
			assert.Contains(t, got, cell)
		}
	}
}

func TestColumnsAlignment(t *testing.T) {
	t.Parallel()
	rows := [][]string{
		{"A", "B"},
		{"foo", "bar"},
	}
	_ = term.Columns(rows, term.WithColumnAlignment(0, term.AlignRight), term.WithColumnGap(1))
}

func TestColumnsEmpty(t *testing.T) {
	t.Parallel()
	got := term.Columns(nil)
	assert.Equal(t, got, "")
}

func TestBanner(t *testing.T) {
	t.Parallel()
	s := term.New().Bold()
	got := term.Banner(s, "Line 1", "Line 2")
	assert.Contains(t, got, "Line 1")
	assert.Contains(t, got, "Line 2")
}

func TestAlignmentConstants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, int(term.AlignLeft), 0)
	assert.Equal(t, int(term.AlignCenter), 1)
	assert.Equal(t, int(term.AlignRight), 2)
}
