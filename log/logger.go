// SPDX-License-Identifier: Apache-2.0

package log

import (
	"context"
	"log/slog"
)

// Default returns slog.Default(). Provided for symmetry with From.
//
// Most programs should set the default once at startup via SetDefault and
// then use the package-level slog functions (slog.Info, slog.Error, etc.)
// directly.
func Default() *slog.Logger {
	return slog.Default()
}

// SetDefault sets slog's global default logger. Programs typically call this
// once in main with a logger constructed from NewHandler or NewJSONHandler.
//
//	log.SetDefault(slog.New(log.NewHandler(os.Stderr, log.WithLevel(log.LevelInfo))))
//
// After this call, package-level slog functions (slog.Info, etc.) use the
// provided logger.
func SetDefault(l *slog.Logger) {
	slog.SetDefault(l)
}

// From returns the logger associated with ctx via Inject, or slog.Default()
// if none has been injected. Never returns nil.
//
// From is the consumer-side accessor used by code that wants the current
// scoped logger. If no logger has been injected via Inject, the global
// default is returned — making From safe to call in any context.
//
//	l := log.From(ctx)
//	l.Info("handler started")
func From(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if l, ok := ctx.Value(ctxKeyLogger).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}

// Inject returns a new context carrying l, retrievable via From. Used by
// middleware and request handlers that want to scope a logger to a request
// lifetime without threading the logger through every function signature.
//
// If l is nil, subsequent From calls on the returned context return
// slog.Default() — nil injection is treated as "remove the injected logger."
//
//	ctx = log.Inject(ctx, log.From(ctx).With("request_id", id))
func Inject(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, l)
}
