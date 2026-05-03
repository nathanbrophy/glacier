// SPDX-License-Identifier: Apache-2.0

package log

import (
	"context"
	"log/slog"
)

// ctxKey is the unexported key type for context storage of log attrs and loggers.
// Two distinct key values prevent collision between attr storage and logger storage.
type ctxKey int

const (
	ctxKeyAttrs  ctxKey = iota // stores []slog.Attr
	ctxKeyLogger               // stores *slog.Logger
)

// With returns a new context that carries the supplied attrs in addition to
// any already attached. When code logs through a Glacier handler using this
// ctx (via slog.InfoContext(ctx, ...) or l.InfoContext(ctx, ...)), the carried
// attrs are appended to the record automatically.
//
// Multiple With calls accumulate: each call's attrs are appended to the
// previous set. Order is preserved. The accumulation is an immutable append :
// no caller's context is mutated.
//
//	ctx = log.With(ctx, slog.String("request_id", id))
//	ctx = log.With(ctx, slog.String("user_id", uid))
//	slog.InfoContext(ctx, "handled")
//	// record contains both request_id and user_id
//
// Important: ctx-attached attrs are appended by Glacier handlers only.
// Stdlib handlers (e.g., slog.NewTextHandler) do not inspect the context
// for attrs. See the package FAQ for details.
func With(ctx context.Context, attrs ...slog.Attr) context.Context {
	if len(attrs) == 0 {
		return ctx
	}
	prev := ctxAttrs(ctx)
	// Immutable append: allocate a new slice; never mutate the parent's slice.
	next := make([]slog.Attr, len(prev)+len(attrs))
	copy(next, prev)
	copy(next[len(prev):], attrs)
	return context.WithValue(ctx, ctxKeyAttrs, next)
}

// ctxAttrs returns the attrs stored in ctx, or nil if none have been attached.
func ctxAttrs(ctx context.Context) []slog.Attr {
	if ctx == nil {
		return nil
	}
	if v, ok := ctx.Value(ctxKeyAttrs).([]slog.Attr); ok {
		return v
	}
	return nil
}
