// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/log"
)

// T#L24 ColorNever → no ANSI escapes in output.
func TestColorNever(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := log.NewHandler(&buf, log.WithColor(log.ColorNever))
	l := slog.New(h)
	l.Info("no color please")

	out := buf.String()
	assert.True(t, !containsSubstring(out, "\x1b"), "ColorNever must produce no ANSI escapes")
}

// T#L25 ColorAlways → ANSI escapes present when writing to *bytes.Buffer.
func TestColorAlways(t *testing.T) {
	// Defend against environments that pre-set the no-color env vars.
	// (t.Setenv is incompatible with t.Parallel.)
	t.Setenv("GLACIER_NO_COLOR", "")
	t.Setenv("NO_COLOR", "")

	var buf bytes.Buffer
	h := log.NewHandler(&buf, log.WithColor(log.ColorAlways))
	l := slog.New(h)
	l.Info("with color")

	// ColorAlways forces color regardless of TTY status. The handler injects
	// the per-level term.Style escape via ReplaceAttr, so the level label is
	// wrapped in the cyan (INFO) prefix and the term.AnsiReset suffix.
	out := buf.String()
	assert.True(t, len(out) > 0, "ColorAlways handler should emit output")
	assert.True(t, containsSubstring(out, "\x1b["), "ColorAlways must emit ANSI escapes")
	assert.True(t, containsSubstring(out, "\x1b[0m"), "ColorAlways must emit reset sequence")
	// The colored level field should appear adjacent to ` level=` and the
	// label INFO should be wrapped (not the raw uncolored form).
	assert.True(t, containsSubstring(out, " level=\x1b["), "color escape must wrap the level field")
	assert.True(t, !containsSubstring(out, " level=INFO "), "uncolored INFO must not appear")
}

// T#L26 ColorAuto + non-TTY writer (bytes.Buffer) → no color (IsTTY=false).
func TestColorAutoNonTTY(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	// ColorAuto should detect that *bytes.Buffer is not a TTY and disable color.
	h := log.NewHandler(&buf, log.WithColor(log.ColorAuto))
	l := slog.New(h)
	l.Info("auto color check")

	out := buf.String()
	assert.True(t, len(out) > 0, "ColorAuto handler should emit output")
	// A bytes.Buffer is not a TTY, so no ANSI escapes should appear.
	assert.True(t, !containsSubstring(out, "\x1b"), "ColorAuto with non-TTY must produce no ANSI escapes")
}
