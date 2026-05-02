// SPDX-License-Identifier: Apache-2.0

package term

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// AnsiReset is the SGR reset escape sequence. Provided for callers that
// pre-compute prefix/suffix pairs from Style.Prefix and want to avoid
// re-typing the literal. Most callers should use Render or Sprint instead.
const AnsiReset = "\x1b[0m"

// Style is an immutable terminal style descriptor.
//
// Invariant: every method returns a new Style; the receiver is unchanged.
// Invariant: zero Style (New()) renders text as plain passthrough.
// Invariant: ANSI escapes are emitted only when the target writer supports color;
// Render on a non-color writer returns the plain text unchanged.
type Style struct {
	fg, bg    Color
	hasFG     bool
	hasBG     bool
	bold      bool
	italic    bool
	underline bool
	dim       bool
	strike    bool
}

// escCache maps Style → computed ANSI prefix bytes.
var escCache sync.Map // map[Style]string

// New returns a zero Style with no decorations. Render on a zero Style returns
// the input text unchanged.
// Concurrency: goroutine-safe; Style is a value type.
func New() Style { return Style{} }

// Foreground returns a new Style with the foreground color set to c.
func (s Style) Foreground(c Color) Style {
	s.fg = c
	s.hasFG = true
	return s
}

// Background returns a new Style with the background color set to c.
func (s Style) Background(c Color) Style {
	s.bg = c
	s.hasBG = true
	return s
}

// Bold returns a new Style with bold enabled.
func (s Style) Bold() Style {
	s.bold = true
	return s
}

// Italic returns a new Style with italic enabled.
func (s Style) Italic() Style {
	s.italic = true
	return s
}

// Underline returns a new Style with underline enabled.
func (s Style) Underline() Style {
	s.underline = true
	return s
}

// Dim returns a new Style with dim (faint) enabled.
func (s Style) Dim() Style {
	s.dim = true
	return s
}

// Strike returns a new Style with strikethrough enabled.
func (s Style) Strike() Style {
	s.strike = true
	return s
}

// Render returns text wrapped in the ANSI escape sequences for this style,
// followed by the reset sequence ESC[0m.
//
// If the package-level default writer (os.Stderr) does not support color,
// Render returns text unchanged, with no escapes.
//
// Performance: ANSI escape sequences are pre-computed at first use and cached.
// Target: ≤ 100 ns/op + 1 alloc (the output string) per §23.13.
//
// Concurrency: goroutine-safe; cache uses sync.Map.
func (s Style) Render(text string) string {
	return renderTo(s, text, os.Stderr)
}

// Prefix returns the raw ANSI escape sequence that opens this style. Returns
// "" for the zero Style. No capability detection is performed — the caller
// decides whether color should be emitted at all (typically by consulting
// Capability once and gating subsequent calls on the result).
//
// Most callers should use Render or Sprint. Prefix exists for hot paths that
// pre-compute the prefix once and concatenate with text per call, avoiding
// the per-call capability check (e.g. log handlers that have already
// resolved color at construction time).
//
// Pair with the AnsiReset constant for the matching close sequence.
//
// Concurrency: goroutine-safe; result is read from the package-level cache.
func (s Style) Prefix() string {
	return escapePrefix(s)
}

// Sprint is a convenience wrapper equivalent to s.Render(text).
// Uses the package-level default writer for capability detection.
func Sprint(s Style, text string) string {
	return renderTo(s, text, os.Stderr)
}

// Fprint writes s.Render(text) to w. Capability is detected from w.
func Fprint(w io.Writer, s Style, text string) {
	fmt.Fprint(w, renderTo(s, text, w))
}

// renderTo renders text with style s targeting writer w for capability detection.
// If w is nil, os.Stderr is used as the default.
func renderTo(s Style, text string, w io.Writer) string {
	if w == nil {
		w = os.Stderr
	}
	caps := Capability(w)
	if caps.NoColorEnv || caps.SupportsColor == ColorNone {
		return text
	}
	prefix := escapePrefix(s)
	if prefix == "" {
		return text
	}
	return prefix + text + "\x1b[0m"
}

// escapePrefix returns the ANSI escape sequence prefix for s, cached on first call.
func escapePrefix(s Style) string {
	if v, ok := escCache.Load(s); ok {
		return v.(string)
	}
	esc := computeEscape(s)
	escCache.Store(s, esc)
	return esc
}

// computeEscape builds the ANSI escape string from s's fields.
func computeEscape(s Style) string {
	if !s.hasFG && !s.hasBG && !s.bold && !s.italic && !s.underline && !s.dim && !s.strike {
		return ""
	}
	var buf [64]byte
	b := buf[:0]
	b = append(b, '\x1b', '[')
	first := true
	sep := func() {
		if !first {
			b = append(b, ';')
		}
		first = false
	}
	if s.bold {
		sep()
		b = append(b, '1')
	}
	if s.dim {
		sep()
		b = append(b, '2')
	}
	if s.italic {
		sep()
		b = append(b, '3')
	}
	if s.underline {
		sep()
		b = append(b, '4')
	}
	if s.strike {
		sep()
		b = append(b, '9')
	}
	if s.hasFG {
		sep()
		b = append(b, []byte(fmt.Sprintf("38;2;%d;%d;%d", s.fg.R, s.fg.G, s.fg.B))...)
	}
	if s.hasBG {
		sep()
		b = append(b, []byte(fmt.Sprintf("48;2;%d;%d;%d", s.bg.R, s.bg.G, s.bg.B))...)
	}
	b = append(b, 'm')
	return string(b)
}
