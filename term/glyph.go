// SPDX-License-Identifier: Apache-2.0

package term

import (
	"log/slog"
	"os"
	"regexp"
	"sync"
)

// GlyphInfo describes one entry in the glyph registry.
//
// Invariant: Name matches ^[a-z][a-z0-9_]*$ and len(Name) <= 64.
// Invariant: UTF8 and ASCII are both non-empty.
type GlyphInfo struct {
	Name  string
	UTF8  string
	ASCII string
}

var (
	glyphMu       sync.RWMutex
	glyphRegistry = map[string]GlyphInfo{}
	glyphNameRE   = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

// Glyph returns the registered glyph's UTF-8 form if the active output writer
// supports UTF-8, otherwise its ASCII fallback.
//
// If name is not in the registry, Glyph returns "" and emits a debug-level log
// warning to slog.Default().
//
// Concurrency: goroutine-safe; registry uses sync.RWMutex.
func Glyph(name string) string {
	glyphMu.RLock()
	info, ok := glyphRegistry[name]
	glyphMu.RUnlock()
	if !ok {
		slog.Default().Debug("term: glyph: unknown glyph name", "name", name)
		return ""
	}
	caps := Capability(os.Stderr)
	if caps.SupportsUTF8 {
		return info.UTF8
	}
	return info.ASCII
}

// RegisterGlyph adds a custom glyph to the registry.
//
// Preconditions:
//   - name must match ^[a-z][a-z0-9_]*$ and be ≤ 64 bytes (§23.9 #26).
//   - utf8 and ascii must both be non-empty.
//   - name must not already be registered.
//
// Returns GlyphError on any violation.
// Concurrency: goroutine-safe.
func RegisterGlyph(name, utf8, ascii string) error {
	if len(name) > 64 {
		return &GlyphError{Name: name, Cause: "name exceeds 64 bytes"}
	}
	if !glyphNameRE.MatchString(name) {
		return &GlyphError{Name: name, Cause: "name must match ^[a-z][a-z0-9_]*$"}
	}
	if utf8 == "" {
		return &GlyphError{Name: name, Cause: "utf8 must not be empty"}
	}
	if ascii == "" {
		return &GlyphError{Name: name, Cause: "ascii must not be empty"}
	}

	glyphMu.Lock()
	defer glyphMu.Unlock()
	if _, exists := glyphRegistry[name]; exists {
		return &GlyphError{Name: name, Cause: "glyph already registered: " + name}
	}
	glyphRegistry[name] = GlyphInfo{Name: name, UTF8: utf8, ASCII: ascii}
	return nil
}

// Glyphs returns a snapshot of all registered glyphs (built-in + custom).
// The returned slice is a copy; mutations do not affect the registry.
// Concurrency: goroutine-safe.
func Glyphs() []GlyphInfo {
	glyphMu.RLock()
	defer glyphMu.RUnlock()
	out := make([]GlyphInfo, 0, len(glyphRegistry))
	for _, g := range glyphRegistry {
		out = append(out, g)
	}
	return out
}

// builtinGlyphs defines the ~30 pre-registered glyphs per spec F11.
var builtinGlyphs = []GlyphInfo{
	{Name: "check", UTF8: "✓", ASCII: "[OK]"},
	{Name: "cross", UTF8: "✗", ASCII: "[X]"},
	{Name: "warn", UTF8: "⚠", ASCII: "[!]"},
	{Name: "info", UTF8: "ℹ", ASCII: "[i]"},
	{Name: "bullet", UTF8: "•", ASCII: "*"},
	{Name: "dot", UTF8: "·", ASCII: "."},
	{Name: "arrow_right", UTF8: "→", ASCII: "->"},
	{Name: "arrow_left", UTF8: "←", ASCII: "<-"},
	{Name: "arrow_up", UTF8: "↑", ASCII: "^"},
	{Name: "arrow_down", UTF8: "↓", ASCII: "v"},
	{Name: "ellipsis", UTF8: "…", ASCII: "..."},
	{Name: "spinner_braille_0", UTF8: "⠋", ASCII: "-"},
	{Name: "spinner_braille_1", UTF8: "⠙", ASCII: `\`},
	{Name: "spinner_braille_2", UTF8: "⠹", ASCII: "|"},
	{Name: "spinner_braille_3", UTF8: "⠸", ASCII: "/"},
	{Name: "spinner_braille_4", UTF8: "⠼", ASCII: "-"},
	{Name: "spinner_braille_5", UTF8: "⠴", ASCII: `\`},
	{Name: "spinner_braille_6", UTF8: "⠦", ASCII: "|"},
	{Name: "spinner_braille_7", UTF8: "⠇", ASCII: "/"},
	{Name: "spinner_dots_0", UTF8: "⣾", ASCII: "-"},
	{Name: "spinner_dots_1", UTF8: "⣽", ASCII: `\`},
	{Name: "spinner_dots_2", UTF8: "⣻", ASCII: "|"},
	{Name: "spinner_dots_3", UTF8: "⢿", ASCII: "/"},
	{Name: "spinner_dots_4", UTF8: "⡿", ASCII: "-"},
	{Name: "spinner_dots_5", UTF8: "⣟", ASCII: `\`},
	{Name: "spinner_dots_6", UTF8: "⣯", ASCII: "|"},
	{Name: "spinner_dots_7", UTF8: "⣷", ASCII: "/"},
	{Name: "box_horizontal", UTF8: "─", ASCII: "-"},
	{Name: "box_vertical", UTF8: "│", ASCII: "|"},
	{Name: "pipe", UTF8: "│", ASCII: "|"},
	{Name: "divider", UTF8: "━", ASCII: "="},
}

func init() {
	for _, g := range builtinGlyphs {
		glyphRegistry[g.Name] = g
	}
}
