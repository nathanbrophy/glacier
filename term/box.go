// SPDX-License-Identifier: Apache-2.0

package term

import (
	"os"
	"strings"
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
	// Pre-seed defaults so callers get colored borders/title without having
	// to opt in. Options applied below override these.
	defaultBorderColor, _ := Hex("#46D9FF")  // bright cyan
	defaultTitleColor, _ := Hex("#7FE7FF")   // lighter cyan for the title
	cfg := boxConfig{
		corners:     roundedCorners,
		borderStyle: New().Foreground(defaultBorderColor),
		titleStyle:  New().Foreground(defaultTitleColor).Bold(),
	}
	for _, o := range opts {
		if o == nil {
			continue
		}
		_ = o.applyBox(&cfg)
	}

	caps := Capability(os.Stderr)
	// Default to Unicode borders. Modern Windows Console (post-1903),
	// Windows Terminal, macOS Terminal/iTerm2, Linux terminals: all UTF-8
	// capable. Fall back to ASCII only if a caller explicitly opts in.

	termW := caps.Width
	if termW <= 0 {
		termW = 80
	}

	// Wrap budget: cap content at termW minus borders + padding so very wide
	// content stays within the terminal. visibleWidth ignores ANSI escapes.
	innerW := termW - 2 - cfg.padding[3] - cfg.padding[1]
	if innerW < 1 {
		innerW = 1
	}
	wrapped := Wrap(text, innerW)
	lines := strings.Split(wrapped, "\n")

	// Compute content width as the longest VISIBLE line. ANSI color codes
	// are not counted: a colored "hello" string contains many escape bytes
	// but renders as 5 columns. Without this, a single colored line would
	// stretch the box to 30+ columns and leave a huge dead area to the right.
	contentW := 0
	for _, l := range lines {
		if w := visibleWidth(l); w > contentW {
			contentW = w
		}
	}
	// Also size to fit the title (with its " title " padding) so a long
	// title is not truncated by content that happens to be shorter.
	if cfg.title != "" {
		titleW := visibleWidth(cfg.title) + 4 // "─ title ─"
		if titleW > contentW+cfg.padding[3]+cfg.padding[1] {
			contentW = titleW - cfg.padding[3] - cfg.padding[1]
		}
	}
	if contentW < 1 {
		contentW = 1
	}

	totalInner := contentW + cfg.padding[3] + cfg.padding[1]
	useColor := ShouldColor(os.Stderr)
	border := func(s string) string {
		if !useColor {
			return s
		}
		return renderTo(cfg.borderStyle, s, os.Stderr)
	}
	titleStr := func(s string) string {
		if !useColor {
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
		leftFill := 1                                          // one "─" before space
		rightFill := topFill - leftFill - 2 - visibleWidth(titStr) // remaining; visible width excludes ANSI
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
		// Use visibleWidth (strips ANSI) so colored content lines pad to
		// the same visible column as plain ones. Without this, padding
		// math is off by the byte length of the escape sequences.
		lw := visibleWidth(l)
		fill := contentW - lw
		if fill < 0 {
			fill = 0
		}
		trailing := strings.Repeat(" ", fill)
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

// visibleWidth returns the rune count of s after stripping ANSI escape
// sequences. The Box layout uses this to size and pad lines correctly even
// when content is colored: a 5-character colored "hello" is 5 visible
// columns even though it carries 11+ bytes of ESC[...m sequences.
//
// Recognized escapes:
//   - CSI sequences "\x1b[...<final>" where final is in [@-~]
//   - OSC sequences "\x1b]...BEL" or "\x1b]...\x1b\\"
//   - Two-byte ESC sequences "\x1bX"
//
// Tabs are counted as 1 column (Box callers should expand tabs upstream).
func visibleWidth(s string) int {
	width := 0
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r != '\x1b' {
			width++
			continue
		}
		// Escape sequence: scan until terminating byte.
		i++
		if i >= len(runes) {
			break
		}
		switch runes[i] {
		case '[':
			// CSI: skip until final byte in 0x40..0x7E
			for i++; i < len(runes); i++ {
				c := runes[i]
				if c >= 0x40 && c <= 0x7E {
					break
				}
			}
		case ']':
			// OSC: skip until BEL (0x07) or ESC \
			for i++; i < len(runes); i++ {
				if runes[i] == 0x07 {
					break
				}
				if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '\\' {
					i++
					break
				}
			}
		default:
			// Two-byte escape: ESC + one char already consumed.
		}
	}
	return width
}
