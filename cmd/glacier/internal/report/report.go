// SPDX-License-Identifier: Apache-2.0

// Package report writes kaomoji-prefixed status lines to a configurable writer.
// It implements the spec 0001 D45 kaomoji scope: kaomoji belong at command
// boundaries in human-facing output, not embedded in structured log records.
//
// Output is colored via ANSI when the writer is a TTY that advertises color
// support and neither NO_COLOR nor GLACIER_NO_COLOR is set in the environment
// (spec 0032 D-S11). On non-capable terminals the color codes are dropped
// and only the plain "<kaomoji> <message>" form is written.
package report

import (
	"io"
	"os"
	"sync"

	"github.com/nathanbrophy/glacier/term"
)

// Level indicates the emotional tone of a status line, mapped to a specific
// kaomoji from spec 0001 D45 and to an ANSI color from the Glacier palette.
type Level int

const (
	// Calm is the default, friendly state. Kaomoji: ʕ•ᴥ•ʔ
	Calm Level = iota
	// Confident signals a successful or assertive outcome. Kaomoji: ʕ⌐■-■ʔ
	Confident
	// Thinking signals an in-progress or uncertain state. Kaomoji: ʕ•_•ʔ
	Thinking
	// Alarmed signals a warning or unexpected situation. Kaomoji: ʕ◉_◉ʔ
	Alarmed
	// Err signals a failure. Kaomoji: ʕ× ×ʔ
	Err
)

// kaomoji maps each Level to its spec 0001 D45 glyph.
var kaomoji = map[Level]string{
	Calm:      "ʕ•ᴥ•ʔ",
	Confident: "ʕ⌐■-■ʔ",
	Thinking:  "ʕ•_•ʔ",
	Alarmed:   "ʕ◉_◉ʔ",
	Err:       "ʕ× ×ʔ",
}

// ansiColor maps each Level to its ANSI 8-bit color code (used inside the
// `\x1b[38;5;<n>m` 256-color sequence). Choices follow spec 0001's palette:
//   - Calm:      cyan (87)
//   - Confident: green (84)
//   - Thinking:  yellow (228)
//   - Alarmed:   orange (215)
//   - Err:       red (203)
var ansiColor = map[Level]int{
	Calm:      87,
	Confident: 84,
	Thinking:  228,
	Alarmed:   215,
	Err:       203,
}

const (
	// ansiReset clears all ANSI styling.
	ansiReset = "\x1b[0m"
	// ansiBold turns on bold.
	ansiBold = "\x1b[1m"
)

var (
	mu     sync.Mutex
	writer io.Writer = os.Stderr
)

// SetWriter changes the destination for subsequent Status calls.
// Passing nil restores the default destination (os.Stderr).
// Concurrency: goroutine-safe.
func SetWriter(w io.Writer) {
	mu.Lock()
	if w == nil {
		writer = os.Stderr
	} else {
		writer = w
	}
	mu.Unlock()
}

// Status writes "<kaomoji> <message>" plus a trailing newline to the
// configured writer. When the writer is a color-capable TTY, the kaomoji is
// bolded and tinted with the level's palette color; the message itself is
// written in the level's color without bold. On non-capable writers (no TTY,
// NO_COLOR / GLACIER_NO_COLOR, or no color support) the line is plain.
//
// If level has no registered kaomoji, the Calm glyph is used as a fallback.
// Concurrency: goroutine-safe.
func Status(level Level, message string) {
	glyph, ok := kaomoji[level]
	if !ok {
		glyph = kaomoji[Calm]
	}

	mu.Lock()
	w := writer
	mu.Unlock()

	line := renderLine(level, glyph, message, w)

	mu.Lock()
	_, _ = io.WriteString(w, line)
	mu.Unlock()
}

// renderLine returns the rendered line for the given level + glyph + message,
// honoring the writer's color capability. Exposed for testing.
func renderLine(level Level, glyph, message string, w io.Writer) string {
	if !shouldColor(w) {
		return glyph + " " + message + "\n"
	}
	color, ok := ansiColor[level]
	if !ok {
		color = ansiColor[Calm]
	}
	colorStart := "\x1b[38;5;" + itoa(color) + "m"
	// Bolded kaomoji + colored message + reset.
	return ansiBold + colorStart + glyph + ansiReset +
		" " + colorStart + message + ansiReset + "\n"
}

// shouldColor reports whether ANSI color codes should be emitted for w.
// Returns false for non-TTY writers, when NO_COLOR / GLACIER_NO_COLOR is set,
// or when term.Capability reports no color support.
func shouldColor(w io.Writer) bool {
	caps := term.Capability(w)
	if caps.NoColorEnv {
		return false
	}
	if !caps.IsTTY {
		return false
	}
	return caps.SupportsColor != term.ColorNone
}

// itoa is a small base-10 integer-to-string helper that avoids the strconv
// import for what is only ever a 3-digit color code.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [4]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
