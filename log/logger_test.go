// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/log"
)

// T#L3 From(nil) returns slog.Default().
func TestFromNilCtx(t *testing.T) {
	t.Parallel()
	got := log.From(nil)
	assert.Equal(t, got, slog.Default())
}

// T#L4 From(context with no logger) returns slog.Default().
func TestFromCtxNoLogger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	got := log.From(ctx)
	assert.Equal(t, got, slog.Default())
}

// T#L5 Inject + From round-trip: From(Inject(ctx, l)) == l.
func TestInjectFromRoundTrip(t *testing.T) {
	t.Parallel()
	l := slog.New(slog.NewTextHandler(nil, nil))
	ctx := log.Inject(context.Background(), l)
	got := log.From(ctx)
	assert.Equal(t, got, l)
}

// T#L6 Inject(ctx, nil) → From returns slog.Default().
func TestInjectNilLogger(t *testing.T) {
	t.Parallel()
	ctx := log.Inject(context.Background(), nil)
	got := log.From(ctx)
	assert.Equal(t, got, slog.Default())
}

// Default() returns slog.Default().
func TestDefault(t *testing.T) {
	t.Parallel()
	assert.Equal(t, log.Default(), slog.Default())
}
