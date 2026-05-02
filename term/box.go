// SPDX-License-Identifier: Apache-2.0

package term

import (
	"os"
	"strings"
	"unicode/utf8"
)

// cornerSet describes the eight characters used to draw a box border.
type cornerSet struct {
	TL, TR, BL, BR string // corners
	H, V           string // horizontal/vertical edges
	TRight, TLeft  string // T-junctions
}

var (
	roundedCorners = cornerSet{TL: "╭", TR: "╮", BL: "╰", BR: "╯", H: "─", V: "│", TRight: "├", TLeft: "┤"}
	sharpCorners   = cornerSet{TL: "┌", TR: "┐", BL: "└", BR: "┘", H: "─", V: "│", TRight: "├", TLeft: "┤"}
	doubleCorners  = cornerSet{TL: "╔", TR: "╗", BL: "╚", BR: "╝", H: "═", V: "║", TRight: "╠", TLeft: "╣"}
	asciiCorners   = cornerSet{TL: "+", TR: "+", BL: "+", BR: "+", H: "-", V: "|", TRight: "+", TLeft: "+"}
)

// boxConfig holds options for Box rendering.
type boxConfig struct {
	corners     cornerSet
	borderStyle Style
	padding     [4]int // top, right, bottom, left
	title       string
	titleStyle  Style
}

// BoxOption configures Box rendering.
type BoxOption interface{ applyBox(*boxConfig) error }

type boxOptionFunc func(*boxConfig) error

func (f boxOptionFunc) applyBox(c *boxConfig) error { return f(c) }

// WithRoundedCorners selects ╭╮╰╯ corner characters (default).
func WithRoundedCorners() BoxOption {
	return boxOptionFunc(func(c *boxConfig) error { c.corners = roundedCorners; return nil })
}

// WithSharpCorners selects ┌┐└┘ corner characters.
func WithSharpCorners() BoxOption {
	return boxOptionFunc(func(c *boxConfig) error { c.corners = sharpCorners; return nil })
}

// WithDoubleBorders selects ╔╗╚╝ / ═║ border characters.
func WithDoubleBorders() BoxOption {
	return boxOptionFunc(func(c *boxConfig) error { c.corners = doubleCorners; return nil })
}

// WithBorderStyle applies s to the box border characters.
func WithBorderStyle(s Style) BoxOption {
	return boxOptionFunc(func(c *boxConfig) error { c.borderStyle = s; return nil })
}

// WithPadding adds whitespace inside the box (top, right, bottom, left cell counts).
func WithPadding(top, right, bottom, left int) BoxOption {
	return boxOptionFunc(func(c *boxConfig) error {
		c.padding = [4]int{top, right, bottom, left}
		return nil
	})
}

// WithTitle places title in the top-border line, formatted as "─ title ─".
// If title exceeds box width, it is truncated with the ellipsis glyph.
func WithTitle(s string) BoxOption {
	return boxOptionFunc(func(c *boxConfig) error { c.title = s; return nil })
}

// WithTitleStyle applies s to the title text within the border.
func WithTitleStyle(s Style) BoxOption {
	return boxOptionFunc(func(c *boxConfig) error { c.titleStyle = s; return nil })
}

// Box renders text inside a Unicode box border.
//
// Default: rounded corners, no padding, no title, no border style.
// Content lines that exceed the terminal width are pre-wrapped via Wrap before boxing.
// On non-UTF-8 writers, ASCII border characters are used automatically.
// Concurrency: pure function; goroutine-safe.
func Box(text string, opts ...BoxOption) string {
	cfg := boxConfig{corners: roundedCorners}
	for _, o := range opts {
		if o == nil {
			continue
		}
		_ = o.applyBox(&cfg)
	}

	caps := Capability(os.Stderr)
	if !caps.SupportsUTF8 {
		cfg.corners = asciiCorners
	}

	termW := caps.Width
	if termW <= 0 {
		termW = 80
	}

	// Pre-wrap content so no content line exceeds termW - 2 (border) - padding.
	innerW := termW - 2 - cfg.padding[3] - cfg.padding[1]
	if innerW < 1 {
		innerW = 1
	}
	wrapped := Wrap(text, innerW)
	lines := strings.Split(wrapped, "\n")

	// Compute content width.
	contentW := 0
	for _, l := range lines {
		if w := utf8.RuneCountInString(l); w > contentW {
			contentW = w
		}
	}
	if contentW < 1 {
		contentW = 1
	}

	totalInner := contentW + cfg.padding[3] + cfg.padding[1]
	border := func(s string) string {
		if caps.NoColorEnv || caps.SupportsColor == ColorNone {
			return s
		}
		return renderTo(cfg.borderStyle, s, os.Stderr)
	}
	titleStr := func(s string) string {
		if caps.NoColorEnv || caps.SupportsColor == ColorNone {
			return s
		}
		return renderTo(cfg.titleStyle, s, os.Stderr)
	}

	var sb strings.Builder

	// Top border
	topFill := totalInner
	if cfg.title != "" {
		// "─ title ─"
		titRunes := []rune(cfg.title)
		avail := topFill - 4 // "─ " + " ─"
		if avail < 0 {
			avail = 0
		}
		if len(titRunes) > avail {
			titRunes = titRunes[:avail]
			if avail > 0 {
				titRunes[len(titRunes)-1] = '…'
			}
		}
		titStr := string(titRunes)
		leftFill := 1                                                        // one "─" before space
		rightFill := topFill - leftFill - 2 - utf8.RuneCountInString(titStr) // remaining
		if rightFill < 0 {
			rightFill = 0
		}
		sb.WriteString(border(cfg.corners.TL))
		sb.WriteString(border(strings.Repeat(cfg.corners.H, leftFill)))
		sb.WriteString(border(" "))
		sb.WriteString(titleStr(titStr))
		sb.WriteString(border(" "))
		sb.WriteString(border(strings.Repeat(cfg.corners.H, rightFill)))
		sb.WriteString(border(cfg.corners.TR))
	} else {
		sb.WriteString(border(cfg.corners.TL))
		sb.WriteString(border(strings.Repeat(cfg.corners.H, topFill)))
		sb.WriteString(border(cfg.corners.TR))
	}
	sb.WriteByte('\n')

	// Top padding
	for i := 0; i < cfg.padding[0]; i++ {
		sb.WriteString(border(cfg.corners.V))
		sb.WriteString(strings.Repeat(" ", totalInner))
		sb.WriteString(border(cfg.corners.V))
		sb.WriteByte('\n')
	}

	// Content lines
	padL := strings.Repeat(" ", cfg.padding[3])
	padR := strings.Repeat(" ", cfg.padding[1])
	for _, l := range lines {
		lw := utf8.RuneCountInString(l)
		trailing := strings.Repeat(" ", contentW-lw)
		sb.WriteString(border(cfg.corners.V))
		sb.WriteString(padL)
		sb.WriteString(l)
		sb.WriteString(trailing)
		sb.WriteString(padR)
		sb.WriteString(border(cfg.corners.V))
		sb.WriteByte('\n')
	}

	// Bottom padding
	for i := 0; i < cfg.padding[2]; i++ {
		sb.WriteString(border(cfg.corners.V))
		sb.WriteString(strings.Repeat(" ", totalInner))
		sb.WriteString(border(cfg.corners.V))
		sb.WriteByte('\n')
	}

	// Bottom border
	sb.WriteString(border(cfg.corners.BL))
	sb.WriteString(border(strings.Repeat(cfg.corners.H, totalInner)))
	sb.WriteString(border(cfg.corners.BR))

	return sb.String()
}
