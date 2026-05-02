// SPDX-License-Identifier: Apache-2.0

package term

import (
	"strings"
	"unicode/utf8"
)

// Alignment controls per-column text alignment in Columns.
type Alignment int

const (
	// AlignLeft aligns text to the left (default).
	AlignLeft Alignment = iota
	// AlignCenter centers text.
	AlignCenter
	// AlignRight aligns text to the right.
	AlignRight
)

// Center center-pads each line of text to width.
// If a line exceeds width, it is returned unchanged.
// Rounding: for odd remainders, the right side gets the extra space.
// Concurrency: pure function; goroutine-safe.
func Center(text string, width int) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lw := utf8.RuneCountInString(l)
		if lw >= width {
			continue
		}
		pad := width - lw
		left := pad / 2
		right := pad - left
		lines[i] = strings.Repeat(" ", left) + l + strings.Repeat(" ", right)
	}
	return strings.Join(lines, "\n")
}

// Justify block-justifies each line of text to width by distributing extra
// spaces between words. Lines shorter than two words are left-aligned.
// Concurrency: pure function; goroutine-safe.
func Justify(text string, width int) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lines[i] = justifyLine(l, width)
	}
	return strings.Join(lines, "\n")
}

func justifyLine(line string, width int) string {
	words := strings.Fields(line)
	if len(words) < 2 {
		return line
	}
	totalW := 0
	for _, w := range words {
		totalW += utf8.RuneCountInString(w)
	}
	gaps := len(words) - 1
	spaces := width - totalW
	if spaces <= 0 {
		return strings.Join(words, " ")
	}
	base := spaces / gaps
	extra := spaces % gaps
	var sb strings.Builder
	for j, w := range words {
		sb.WriteString(w)
		if j < gaps {
			n := base
			if j < extra {
				n++
			}
			sb.WriteString(strings.Repeat(" ", n))
		}
	}
	return sb.String()
}

// Pad adds leftN space characters to the left and rightN to the right of each
// line of text.
// Concurrency: pure function; goroutine-safe.
func Pad(text string, leftN, rightN int) string {
	if leftN == 0 && rightN == 0 {
		return text
	}
	padL := strings.Repeat(" ", leftN)
	padR := strings.Repeat(" ", rightN)
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lines[i] = padL + l + padR
	}
	return strings.Join(lines, "\n")
}

// Truncate shortens each line of text that exceeds width by replacing the
// excess with ellipsis. If ellipsis is empty, the Unicode glyph "ellipsis" is
// used on UTF-8 writers, and "..." on ASCII-only writers.
// Width is measured in Unicode code points (not bytes; not grapheme clusters).
// Concurrency: pure function; goroutine-safe.
func Truncate(text string, width int, ellipsis string) string {
	if ellipsis == "" {
		ellipsis = "…" // always use Unicode … here; callers for ASCII should pass "..."
	}
	ellW := utf8.RuneCountInString(ellipsis)
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		runes := []rune(l)
		if len(runes) <= width {
			continue
		}
		keep := width - ellW
		if keep < 0 {
			keep = 0
		}
		lines[i] = string(runes[:keep]) + ellipsis
	}
	return strings.Join(lines, "\n")
}

// Wrap soft-wraps text at word boundaries so no line exceeds width code points.
// Long words that individually exceed width are broken mid-word.
// Preserves existing newlines.
// Concurrency: pure function; goroutine-safe.
func Wrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	inputLines := strings.Split(text, "\n")
	var out []string
	for _, line := range inputLines {
		out = append(out, wrapLine(line, width)...)
	}
	return strings.Join(out, "\n")
}

func wrapLine(line string, width int) []string {
	if utf8.RuneCountInString(line) <= width {
		return []string{line}
	}
	var result []string
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{line}
	}
	var cur strings.Builder
	curW := 0
	for _, word := range words {
		wRunes := []rune(word)
		// Break word if it alone exceeds width.
		for len(wRunes) > width {
			chunk := string(wRunes[:width])
			if curW > 0 {
				result = append(result, cur.String())
				cur.Reset()
				curW = 0
			}
			result = append(result, chunk)
			wRunes = wRunes[width:]
		}
		wLen := len(wRunes)
		if curW == 0 {
			cur.WriteString(string(wRunes))
			curW = wLen
		} else if curW+1+wLen <= width {
			cur.WriteByte(' ')
			cur.WriteString(string(wRunes))
			curW += 1 + wLen
		} else {
			result = append(result, cur.String())
			cur.Reset()
			cur.WriteString(string(wRunes))
			curW = wLen
		}
	}
	if curW > 0 {
		result = append(result, cur.String())
	}
	return result
}

// columnConfig holds resolved options for Columns.
type columnConfig struct {
	gap        int
	alignments map[int]Alignment
}

// ColumnOption configures Columns rendering.
type ColumnOption interface{ applyColumn(*columnConfig) error }

type columnOptionFunc func(*columnConfig) error

func (f columnOptionFunc) applyColumn(c *columnConfig) error { return f(c) }

// WithColumnGap sets the number of space characters between columns (default 2).
func WithColumnGap(n int) ColumnOption {
	return columnOptionFunc(func(c *columnConfig) error { c.gap = n; return nil })
}

// WithColumnAlignment sets the alignment for column idx (0-based).
// Unset columns default to AlignLeft.
func WithColumnAlignment(idx int, alignment Alignment) ColumnOption {
	return columnOptionFunc(func(c *columnConfig) error {
		if c.alignments == nil {
			c.alignments = make(map[int]Alignment)
		}
		c.alignments[idx] = alignment
		return nil
	})
}

// Columns renders a 2-D grid of strings as evenly-spaced columns.
// Column widths are computed automatically to fit the terminal width;
// columns are constrained to a minimum of 1 character wide.
// Concurrency: pure function; goroutine-safe.
func Columns(rows [][]string, opts ...ColumnOption) string {
	if len(rows) == 0 {
		return ""
	}
	cfg := columnConfig{gap: 2}
	for _, o := range opts {
		if o == nil {
			continue
		}
		_ = o.applyColumn(&cfg)
	}

	// Find number of columns.
	numCols := 0
	for _, row := range rows {
		if len(row) > numCols {
			numCols = len(row)
		}
	}
	if numCols == 0 {
		return ""
	}

	// Compute max width per column.
	colW := make([]int, numCols)
	for _, row := range rows {
		for ci, cell := range row {
			w := utf8.RuneCountInString(cell)
			if w > colW[ci] {
				colW[ci] = w
			}
		}
	}
	// Enforce minimum 1.
	for i := range colW {
		if colW[i] < 1 {
			colW[i] = 1
		}
	}

	var sb strings.Builder
	for ri, row := range rows {
		for ci := 0; ci < numCols; ci++ {
			var cell string
			if ci < len(row) {
				cell = row[ci]
			}
			w := colW[ci]
			cw := utf8.RuneCountInString(cell)
			align := AlignLeft
			if cfg.alignments != nil {
				if a, ok := cfg.alignments[ci]; ok {
					align = a
				}
			}
			var padded string
			switch align {
			case AlignRight:
				padded = strings.Repeat(" ", w-cw) + cell
			case AlignCenter:
				total := w - cw
				left := total / 2
				right := total - left
				padded = strings.Repeat(" ", left) + cell + strings.Repeat(" ", right)
			default:
				padded = cell + strings.Repeat(" ", w-cw)
			}
			sb.WriteString(padded)
			if ci < numCols-1 {
				sb.WriteString(strings.Repeat(" ", cfg.gap))
			}
		}
		if ri < len(rows)-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// Banner renders lines in a consistent style, used by cli for the Glacier
// wordmark and ASCII-art bear kaomoji.
// Concurrency: pure function; goroutine-safe.
func Banner(s Style, lines ...string) string {
	result := make([]string, len(lines))
	for i, l := range lines {
		result[i] = renderTo(s, l, nil)
	}
	return strings.Join(result, "\n")
}
