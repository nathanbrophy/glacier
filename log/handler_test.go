// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/log"
)

// T#L12 Records at or above configured level are emitted.
func TestHandlerEmitsAtLevel(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := log.NewHandler(&buf,
		log.WithLevel(log.LevelWarn),
		log.WithColor(log.ColorNever),
	)
	l := slog.New(h)
	l.Warn("warn message")
	l.Error("error message")

	out := buf.String()
	assert.Contains(t, out, "warn message")
	assert.Contains(t, out, "error message")
}

// T#L13 Records below configured level are discarded.
func TestHandlerDiscardsBeforeLevel(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := log.NewHandler(&buf,
		log.WithLevel(log.LevelWarn),
		log.WithColor(log.ColorNever),
	)
	l := slog.New(h)
	l.Info("info message")
	l.Debug("debug message")

	out := buf.String()
	assert.True(t, !containsSubstring(out, "info message"), "info should be discarded")
	assert.True(t, !containsSubstring(out, "debug message"), "debug should be discarded")
}

// T#L14 NewHandler defaults to LevelInfo when WithLevel not supplied.
func TestHandlerDefaultsToInfo(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := log.NewHandler(&buf, log.WithColor(log.ColorNever))
	l := slog.New(h)

	l.Debug("should be discarded")
	l.Info("should appear")

	out := buf.String()
	assert.True(t, !containsSubstring(out, "should be discarded"), "debug should be discarded at default Info level")
	assert.Contains(t, out, "should appear")
}

// T#L15 ctx-attached attrs appear in handler output.
func TestHandlerCtxAttrs(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := log.NewHandler(&buf, log.WithColor(log.ColorNever))
	l := slog.New(h)

	ctx := log.With(context.Background(), slog.String("trace_id", "xyz789"))
	l.InfoContext(ctx, "traced")

	assert.Contains(t, buf.String(), "trace_id=xyz789")
}

// T#L16 ANSI escapes are sanitized when ColorNever and string attr contains \x1b.
func TestHandlerSanitizesANSI(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := log.NewHandler(&buf, log.WithColor(log.ColorNever))
	l := slog.New(h)

	// Inject an ANSI escape in an attribute value.
	l.Info("msg", "injected", "\x1b[31mred\x1b[0m")

	out := buf.String()
	// The raw \x1b byte must not appear.
	assert.True(t, !containsSubstring(out, "\x1b"), "raw ESC byte must not appear in sanitized output")
	// The sanitized form should be present.
	assert.Contains(t, out, "<ESC>")
}
