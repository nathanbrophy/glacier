// SPDX-License-Identifier: Apache-2.0

// Package report writes kaomoji-prefixed status lines to a configurable writer.
// It implements the spec 0001 D45 kaomoji scope: kaomoji belong at command
// boundaries in human-facing output, not embedded in structured log records.
package report

import (
	"io"
	"os"
	"sync"
)

// Level indicates the emotional tone of a status line, mapped to a specific
// kaomoji from spec 0001 D45.
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

// Status writes "<kaomoji> <message>\n" to the configured writer.
// If level has no registered kaomoji, the Calm glyph is used as a fallback.
// Concurrency: goroutine-safe.
func Status(level Level, message string) {
	glyph, ok := kaomoji[level]
	if !ok {
		glyph = kaomoji[Calm]
	}
	line := glyph + " " + message + "\n"
	mu.Lock()
	_, _ = io.WriteString(writer, line)
	mu.Unlock()
}
