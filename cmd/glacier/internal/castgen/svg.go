// SPDX-License-Identifier: Apache-2.0

package castgen

import (
	"fmt"
	"io"
	"strings"
)

// span holds one run of identically-styled text.
type span struct {
	text  string
	style termStyle
}

// termStyle captures the active SGR state for a span. Default colors map
// to specific palette entries that match the SDK's `term.Capability` 256
// color choices.
type termStyle struct {
	fg        string // hex color "#rrggbb" or "" for default
	bg        string
	bold      bool
	italic    bool
	underline bool
	dim       bool
}

// defaultStyle returns the SVG terminal's "no SGR active" style.
func defaultStyle() termStyle {
	return termStyle{}
}

// svgFG returns the SVG fill color for s's foreground, defaulting to
// the foreground palette color when fg is empty.
func (s termStyle) svgFG(defaultFG string) string {
	if s.fg == "" {
		return defaultFG
	}
	return s.fg
}

// WriteSVG writes a self-contained <svg> element representing the final
// frame of the captured output to w.
//
// Strategy: parse the ANSI byte stream into a sequence of spans annotated
// with style state, lay them out as one <text> per terminal row, with
// <tspan> per styled span. Add a chrome frame (rounded rect + traffic
// lights + title bar) so the SVG works as a standalone graphic on the
// public site.
//
// Width / height are set from the scenario's Cols / Rows.
func WriteSVG(w io.Writer, c Cast) error {
	const (
		cellW    = 8.0  // px per terminal column at 14px font
		cellH    = 18.0 // px per terminal row
		padX     = 12.0
		padY     = 36.0 // top padding leaves room for the title bar
		titleH   = 28.0
		font     = "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace"
		fontSize = 14
		// Glacier dark-mode palette baseline.
		bgColor    = "#1a1a1d"
		fgColor    = "#dde6ee"
		titleBgCol = "#2a2a2f"
		titleFgCol = "#9aa6b2"
	)

	rows := splitRows(string(c.Output))

	// Compute width / height in pixels.
	cols := c.Scenario.Cols
	if cols < 40 {
		cols = 80
	}
	height := titleH + padY/2 + cellH*float64(len(rows)) + padY
	width := cellW*float64(cols) + 2*padX

	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" font-family="%s" font-size="%d">`+"\n",
		width, height, width, height, font, fontSize)

	// Background.
	fmt.Fprintf(w, `<rect x="0" y="0" width="%.0f" height="%.0f" rx="10" fill="%s"/>`+"\n",
		width, height, bgColor)

	// Title bar.
	fmt.Fprintf(w, `<rect x="0" y="0" width="%.0f" height="%.0f" rx="10" fill="%s"/>`+"\n",
		width, titleH, titleBgCol)
	// Mask off the bottom rounded corners of the title bar so the join with the body looks crisp.
	fmt.Fprintf(w, `<rect x="0" y="%.0f" width="%.0f" height="6" fill="%s"/>`+"\n",
		titleH-6, width, titleBgCol)
	// Traffic lights.
	for i, color := range []string{"#ff605c", "#ffbd44", "#00ca4e"} {
		cx := 14.0 + float64(i)*18.0
		fmt.Fprintf(w, `<circle cx="%.1f" cy="%.1f" r="6" fill="%s"/>`+"\n",
			cx, titleH/2, color)
	}
	// Title text (centered).
	title := svgEscape(c.Scenario.Title)
	if title == "" {
		title = "glacier"
	}
	fmt.Fprintf(w, `<text x="%.1f" y="%.1f" text-anchor="middle" fill="%s">%s</text>`+"\n",
		width/2, titleH/2+5, titleFgCol, title)

	// Body.
	fmt.Fprintf(w, `<g transform="translate(%.0f,%.0f)">`+"\n", padX, titleH+padY/2)
	for i, row := range rows {
		y := float64(i+1) * cellH
		fmt.Fprintf(w, `<text x="0" y="%.1f" fill="%s" xml:space="preserve">`, y, fgColor)
		spans := parseANSI(row)
		col := 0
		for _, sp := range spans {
			if sp.text == "" {
				continue
			}
			x := float64(col) * cellW
			// Build attributes from active style.
			attrs := fmt.Sprintf(` x="%.1f" fill="%s"`, x, sp.style.svgFG(fgColor))
			if sp.style.bold {
				attrs += ` font-weight="bold"`
			}
			if sp.style.italic {
				attrs += ` font-style="italic"`
			}
			if sp.style.underline {
				attrs += ` text-decoration="underline"`
			}
			if sp.style.dim {
				attrs += ` opacity="0.6"`
			}
			fmt.Fprintf(w, `<tspan%s>%s</tspan>`, attrs, svgEscape(sp.text))
			col += visibleLen(sp.text)
		}
		fmt.Fprintf(w, "</text>\n")
	}
	fmt.Fprintln(w, `</g>`)
	fmt.Fprintln(w, `</svg>`)
	return nil
}

// splitRows splits the output into terminal rows, normalizing CRLF and
// trimming trailing blank rows.
func splitRows(out string) []string {
	out = strings.ReplaceAll(out, "\r\n", "\n")
	out = strings.ReplaceAll(out, "\r", "")
	out = trimTrailingNewlines(out)
	if out == "" {
		return []string{""}
	}
	return strings.Split(out, "\n")
}

// visibleLen counts the rune count of s after stripping ANSI escapes.
func visibleLen(s string) int {
	n := 0
	in := []rune(s)
	for i := 0; i < len(in); i++ {
		if in[i] == 0x1b && i+1 < len(in) && in[i+1] == '[' {
			i += 2
			for i < len(in) && (in[i] < 0x40 || in[i] > 0x7E) {
				i++
			}
			continue
		}
		n++
	}
	return n
}

// svgEscape escapes the XML-special characters that may appear in a row.
// Runs of spaces are preserved (xml:space="preserve" on the parent text).
func svgEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
	)
	return r.Replace(s)
}

// parseANSI walks a row of bytes and produces a sequence of spans, where
// each span is a contiguous run of characters with one consistent style.
// Recognized escapes:
//   - SGR ("\x1b[<n;m;...>m"): updates the active style
//   - any other CSI: stripped silently
//
// Other escape sequences (cursor movement, DECSET, etc.) are stripped
// verbatim. The terminal model is a flat row of styled glyphs; cursor
// motion within a row is not modelled.
func parseANSI(row string) []span {
	var spans []span
	cur := defaultStyle()
	var buf strings.Builder
	flush := func() {
		if buf.Len() == 0 {
			return
		}
		spans = append(spans, span{text: buf.String(), style: cur})
		buf.Reset()
	}

	runes := []rune(row)
	i := 0
	for i < len(runes) {
		c := runes[i]
		if c != 0x1b {
			buf.WriteRune(c)
			i++
			continue
		}
		flush()
		if i+1 >= len(runes) {
			break
		}
		if runes[i+1] != '[' {
			// Two-byte escape; consume and ignore.
			i += 2
			continue
		}
		// CSI: scan to terminator.
		end := i + 2
		for end < len(runes) && (runes[end] < 0x40 || runes[end] > 0x7E) {
			end++
		}
		if end >= len(runes) {
			break
		}
		final := runes[end]
		if final == 'm' {
			// SGR: parse parameters from runes[i+2:end].
			cur = applySGR(cur, string(runes[i+2:end]))
		}
		// Skip the entire escape including final byte.
		i = end + 1
	}
	flush()
	return spans
}

// applySGR mutates style according to one SGR parameter sequence.
// Recognized codes:
//
//	0    reset
//	1    bold
//	2    dim (faint)
//	3    italic
//	4    underline
//	22   normal intensity (bold/dim off)
//	23   italic off
//	24   underline off
//	30-37   foreground (basic 8 color)
//	38;5;n  foreground (256-color palette)
//	38;2;r;g;b  foreground (24-bit truecolor)
//	39   default foreground
//	40-47   background (basic 8)
//	48;5;n  background (256-color)
//	48;2;r;g;b  background (24-bit)
//	49   default background
//	90-97   bright foreground (8 color)
//	100-107 bright background (8 color)
func applySGR(s termStyle, params string) termStyle {
	if params == "" || params == "0" {
		return defaultStyle()
	}
	parts := strings.Split(params, ";")
	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "0":
			s = defaultStyle()
		case "1":
			s.bold = true
		case "2":
			s.dim = true
		case "3":
			s.italic = true
		case "4":
			s.underline = true
		case "22":
			s.bold = false
			s.dim = false
		case "23":
			s.italic = false
		case "24":
			s.underline = false
		case "39":
			s.fg = ""
		case "49":
			s.bg = ""
		case "38":
			if i+1 < len(parts) {
				switch parts[i+1] {
				case "5":
					if i+2 < len(parts) {
						s.fg = palette256(parts[i+2])
						i += 2
					}
				case "2":
					if i+4 < len(parts) {
						s.fg = rgbHex(parts[i+2], parts[i+3], parts[i+4])
						i += 4
					}
				}
			}
		case "48":
			if i+1 < len(parts) {
				switch parts[i+1] {
				case "5":
					if i+2 < len(parts) {
						s.bg = palette256(parts[i+2])
						i += 2
					}
				case "2":
					if i+4 < len(parts) {
						s.bg = rgbHex(parts[i+2], parts[i+3], parts[i+4])
						i += 4
					}
				}
			}
		default:
			// Basic 8-color foregrounds and backgrounds + bright variants.
			if c := basicColor(parts[i]); c != "" {
				s.fg = c
			}
		}
	}
	return s
}

// rgbHex turns a "r;g;b" triple of decimal strings into a "#rrggbb" hex.
func rgbHex(r, g, b string) string {
	return fmt.Sprintf("#%02x%02x%02x", atoi(r)&0xff, atoi(g)&0xff, atoi(b)&0xff)
}

// atoi is a tiny strconv.Atoi clone that returns 0 on parse failure.
func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// basicColor maps the 30-37, 90-97 SGR codes to hex colors.
func basicColor(code string) string {
	switch code {
	case "30":
		return "#3c3c3c"
	case "31":
		return "#cc6666"
	case "32":
		return "#9ec07c"
	case "33":
		return "#e3c46e"
	case "34":
		return "#7daea3"
	case "35":
		return "#d49bd2"
	case "36":
		return "#7fdfe7"
	case "37":
		return "#d6dde6"
	case "90":
		return "#5a5a5a"
	case "91":
		return "#ec7878"
	case "92":
		return "#b8d99a"
	case "93":
		return "#f1d68b"
	case "94":
		return "#9ac8be"
	case "95":
		return "#e7b6e3"
	case "96":
		return "#a4ecf3"
	case "97":
		return "#f1f5fa"
	}
	return ""
}

// palette256 returns the hex color for a 256-color palette index.
// Indices 0-15 reuse basicColor; 16-231 are the 6x6x6 cube; 232-255 are
// grayscale.
func palette256(idx string) string {
	n := atoi(idx)
	if n < 16 {
		// Map standard 8 to basic; bright 8 to 90+ range.
		if n < 8 {
			return basicColor(fmt.Sprintf("%d", 30+n))
		}
		return basicColor(fmt.Sprintf("%d", 90+n-8))
	}
	if n >= 232 {
		// Grayscale ramp: 0x08 .. 0xee in 24 steps.
		step := n - 232
		v := 8 + step*10
		if v > 0xee {
			v = 0xee
		}
		return fmt.Sprintf("#%02x%02x%02x", v, v, v)
	}
	// 6x6x6 color cube starting at 16.
	cube := n - 16
	r := cube / 36
	g := (cube / 6) % 6
	b := cube % 6
	cv := func(v int) int {
		if v == 0 {
			return 0
		}
		return 55 + v*40
	}
	return fmt.Sprintf("#%02x%02x%02x", cv(r), cv(g), cv(b))
}
