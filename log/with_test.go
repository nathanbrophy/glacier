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

// T#L7 With accumulates attrs: each With call extends the previous set.
func TestWithAccumulates(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = log.With(ctx, slog.String("a", "1"))
	ctx = log.With(ctx, slog.String("b", "2"))

	var buf bytes.Buffer
	h := log.NewHandler(&buf, log.WithColor(log.ColorNever))
	l := slog.New(h)
	l.InfoContext(ctx, "msg")

	out := buf.String()
	assert.Contains(t, out, "a=1")
	assert.Contains(t, out, "b=2")
}

// T#L8 With(ctx) with no attrs returns the same ctx unchanged.
func TestWithNoAttrs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	got := log.With(ctx)
	assert.Equal(t, got, ctx)
}

// T#L9 ctx-attached attrs appear in log output when using NewHandler.
func TestWithAttrsInOutput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = log.With(ctx, slog.String("req_id", "abc123"))

	var buf bytes.Buffer
	h := log.NewHandler(&buf, log.WithColor(log.ColorNever))
	l := slog.New(h)
	l.InfoContext(ctx, "request handled")

	assert.Contains(t, buf.String(), "req_id=abc123")
}

// T#L10 Parent context is not mutated by a child With call.
func TestWithDoesNotMutateParent(t *testing.T) {
	t.Parallel()
	parent := context.Background()
	parent = log.With(parent, slog.String("parent_key", "pval"))

	child := log.With(parent, slog.String("child_key", "cval"))

	// Parent output must not contain child_key.
	var parentBuf bytes.Buffer
	ph := log.NewHandler(&parentBuf, log.WithColor(log.ColorNever))
	pl := slog.New(ph)
	pl.InfoContext(parent, "parent")
	parentOut := parentBuf.String()
	assert.Contains(t, parentOut, "parent_key=pval")
	assert.True(t, !containsSubstring(parentOut, "child_key"), "parent ctx should not contain child_key")

	// Child output must contain both.
	var childBuf bytes.Buffer
	ch := log.NewHandler(&childBuf, log.WithColor(log.ColorNever))
	cl := slog.New(ch)
	cl.InfoContext(child, "child")
	childOut := childBuf.String()
	assert.Contains(t, childOut, "parent_key=pval")
	assert.Contains(t, childOut, "child_key=cval")
}

// T#L11 With(ctx, /* no attrs */) does not panic; returns same ctx.
func TestWithNoAttrsNoPanic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// Empty attrs: must return ctx as-is without panicking.
	got := log.With(ctx)
	assert.Equal(t, got, ctx)
}

// containsSubstring is a helper that avoids importing strings in the test file.
func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
