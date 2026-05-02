// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/nathanbrophy/glacier/log"
)

// BenchmarkInfoText logs one Info record through NewHandler to io.Discard.
func BenchmarkInfoText(b *testing.B) {
	h := log.NewHandler(io.Discard, log.WithColor(log.ColorNever))
	l := slog.New(h)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		l.Info("benchmark message", "key", "value")
	}
}

// BenchmarkInfoJSON logs one Info record through NewJSONHandler to io.Discard.
func BenchmarkInfoJSON(b *testing.B) {
	h := log.NewJSONHandler(io.Discard, log.WithLevel(log.LevelInfo))
	l := slog.New(h)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		l.Info("benchmark message", "key", "value")
	}
}

// BenchmarkWithCtx attaches 3 attrs via With then logs one record.
func BenchmarkWithCtx(b *testing.B) {
	h := log.NewHandler(io.Discard, log.WithColor(log.ColorNever))
	l := slog.New(h)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		ctx := log.With(context.Background(),
			slog.String("a", "1"),
			slog.String("b", "2"),
			slog.String("c", "3"),
		)
		l.InfoContext(ctx, "benchmark with ctx")
	}
}
