// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"strings"
	"testing"
	"unicode/utf8"

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
			if got != tc.want {
				t.Errorf("Center(%q, %d) = %q, want %q", tc.text, tc.width, got, tc.want)
			}
		})
	}
}

func TestCenterMultiLine(t *testing.T) {
	t.Parallel()
	got := term.Center("ab\ncd", 6)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	for _, l := range lines {
		if utf8.RuneCountInString(l) != 6 {
			t.Errorf("line %q is not width 6", l)
		}
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
				if gotW != tc.width {
					t.Errorf("Justify(%q, %d) = %q (width %d), want width %d", tc.text, tc.width, got, gotW, tc.width)
				}
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
			if got != tc.want {
				t.Errorf("Pad(%q, %d, %d) = %q, want %q", tc.text, tc.left, tc.right, got, tc.want)
			}
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
			if got != tc.want {
				t.Errorf("Truncate(%q, %d, %q) = %q, want %q", tc.text, tc.width, tc.ellipsis, got, tc.want)
			}
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
				if w > tc.width {
					t.Errorf("Wrap(%q, %d): line %q has width %d > %d", tc.text, tc.width, line, w, tc.width)
				}
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
	if got == "" {
		t.Error("Columns() returned empty string")
	}
	// Must contain all cell values.
	for _, row := range rows {
		for _, cell := range row {
			if !strings.Contains(got, cell) {
				t.Errorf("Columns() missing cell %q", cell)
			}
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
	if got != "" {
		t.Errorf("Columns(nil) = %q, want empty string", got)
	}
}

func TestBanner(t *testing.T) {
	t.Parallel()
	s := term.New().Bold()
	got := term.Banner(s, "Line 1", "Line 2")
	if !strings.Contains(got, "Line 1") || !strings.Contains(got, "Line 2") {
		t.Errorf("Banner() missing lines: %q", got)
	}
}

func TestAlignmentConstants(t *testing.T) {
	t.Parallel()
	if term.AlignLeft != 0 {
		t.Errorf("AlignLeft = %d, want 0", term.AlignLeft)
	}
	if term.AlignCenter != 1 {
		t.Errorf("AlignCenter = %d, want 1", term.AlignCenter)
	}
	if term.AlignRight != 2 {
		t.Errorf("AlignRight = %d, want 2", term.AlignRight)
	}
}
