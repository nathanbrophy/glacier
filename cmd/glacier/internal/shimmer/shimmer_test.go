// SPDX-License-Identifier: Apache-2.0

package shimmer_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/cmd/glacier/internal/shimmer"
)

// ExampleWordmark_plain shows plain-mode usage; output is always "GLACIER".
func ExampleWordmark_plain() {
	out := shimmer.Wordmark(0, false, false)
	fmt.Println(out)
	// Output: GLACIER
}

// TestWordmarkPlain confirms the no-color path returns the bare string.
func TestWordmarkPlain(t *testing.T) {
	t.Parallel()
	for phase := 0; phase < 6; phase++ {
		got := shimmer.Wordmark(phase, false, false)
		if got != "GLACIER" {
			t.Errorf("phase %d: got %q, want \"GLACIER\"", phase, got)
		}
	}
}

// TestWordmarkCycleLength confirms phases 0..5 each produce a unique string
// when color24 is enabled.
func TestWordmarkCycleLength(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for phase := 0; phase < 6; phase++ {
		s := shimmer.Wordmark(phase, true, false)
		seen[s] = true
	}
	if len(seen) != 6 {
		t.Errorf("expected 6 distinct frames, got %d", len(seen))
	}
}

// TestWordmarkWrapAround confirms shimmer.Wordmark(n) == shimmer.Wordmark(n+6) for all n.
func TestWordmarkWrapAround(t *testing.T) {
	t.Parallel()
	for n := 0; n < 6; n++ {
		a := shimmer.Wordmark(n, true, false)
		b := shimmer.Wordmark(n+6, true, false)
		if a != b {
			t.Errorf("Wordmark(%d) != Wordmark(%d): wrap-around broken", n, n+6)
		}
	}
}

// TestWordmarkVisibleWidth confirms every phase and color mode yields a
// visible width of 7 (len("GLACIER")).
func TestWordmarkVisibleWidth(t *testing.T) {
	t.Parallel()
	type mode struct{ c24, c256 bool }
	modes := []mode{{false, false}, {true, false}, {false, true}}
	for _, m := range modes {
		for phase := 0; phase < 6; phase++ {
			s := shimmer.Wordmark(phase, m.c24, m.c256)
			w := visibleLen(s)
			if w != 7 {
				t.Errorf("Wordmark(phase=%d, c24=%v, c256=%v): visible width = %d, want 7",
					phase, m.c24, m.c256, w)
			}
		}
	}
}

// TestWordmark256Escapes confirms that 256-color mode uses ESC[38;5;Nm sequences.
func TestWordmark256Escapes(t *testing.T) {
	t.Parallel()
	s := shimmer.Wordmark(0, false, true)
	if !strings.Contains(s, "\x1b[38;5;") {
		t.Errorf("256-color Wordmark does not contain ESC[38;5; prefix: %q", s)
	}
}

// TestWordmark24BitEscapes confirms that 24-bit mode uses ESC[38;2;R;G;Bm sequences.
func TestWordmark24BitEscapes(t *testing.T) {
	t.Parallel()
	s := shimmer.Wordmark(0, true, false)
	if !strings.Contains(s, "\x1b[38;2;") {
		t.Errorf("24-bit Wordmark does not contain ESC[38;2; prefix: %q", s)
	}
}

// visibleLen strips ANSI escapes and returns the rune count of visible text.
func visibleLen(s string) int {
	width := 0
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r != '\x1b' {
			width++
			continue
		}
		i++
		if i >= len(runes) {
			break
		}
		switch runes[i] {
		case '[':
			for i++; i < len(runes); i++ {
				c := runes[i]
				if c >= 0x40 && c <= 0x7E {
					break
				}
			}
		case ']':
			for i++; i < len(runes); i++ {
				if runes[i] == 0x07 {
					break
				}
				if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '\\' {
					i++
					break
				}
			}
		}
	}
	return width
}
