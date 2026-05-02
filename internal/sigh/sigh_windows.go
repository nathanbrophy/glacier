// SPDX-License-Identifier: Apache-2.0

//go:build windows

package sigh

import (
	"context"
	"os"
	"os/signal"
)

// Notify returns a derived context that is cancelled when os.Interrupt is
// received (CTRL+C or CTRL_BREAK_EVENT on Windows). The stop function must
// be called to release signal resources; defer it immediately after calling Notify.
//
// The first signal cancels the context exactly once. Subsequent signals are
// ignored after the first cancellation.
//
// Concurrency: goroutine-safe; the returned cancel is safe to call multiple times.
func Notify(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(ch)
		close(ch)
	}()
	return ctx, cancel
}
