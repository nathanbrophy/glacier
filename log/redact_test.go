// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/log"
)

// T#L20 Redact(v).LogValue() returns slog.StringValue("[REDACTED]").
func TestRedactLogValue(t *testing.T) {
	t.Parallel()
	rv := log.Redact("super-secret")
	got := rv.LogValue()
	assert.Equal(t, got, slog.StringValue("[REDACTED]"))
}

// T#L21 Redact(nil) does not panic; formats as [REDACTED].
func TestRedactNil(t *testing.T) {
	t.Parallel()
	rv := log.Redact(nil)
	got := rv.LogValue()
	assert.Equal(t, got, slog.StringValue("[REDACTED]"))
}

// T#L22 Redact value used as slog attribute: text handler emits [REDACTED].
func TestRedactTextHandler(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := log.NewHandler(&buf, log.WithColor(log.ColorNever))
	l := slog.New(h)
	l.Info("auth", "token", log.Redact("s3cr3t"))

	out := buf.String()
	assert.Contains(t, out, "[REDACTED]")
	assert.True(t, !containsSubstring(out, "s3cr3t"), "raw secret must not appear in output")
}

// T#L23 Redact value used as slog attribute: JSON handler emits "[REDACTED]".
func TestRedactJSONHandler(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := log.NewJSONHandler(&buf, log.WithLevel(log.LevelInfo))
	l := slog.New(h)
	l.Info("auth", "token", log.Redact("s3cr3t"))

	out := buf.String()
	assert.Contains(t, out, `"[REDACTED]"`)
	assert.True(t, !containsSubstring(out, "s3cr3t"), "raw secret must not appear in JSON output")
}
