// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/term"
)

// TestAnimatorConcurrentLogWrites exercises the animator under concurrent log
// writes from many goroutines (satisfies NF3, NF5, §23.14).
func TestAnimatorConcurrentLogWrites(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := term.NewAnimator(logger,
		term.WithWriter(io.Discard),
		term.WithRefreshRate(10*time.Millisecond),
		term.WithMaxBufferedRecords(200),
	)
	a.Add(&neverDoneAnimation{})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	// Start Run in the background.
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = a.Run(ctx)
	}()

	// 100 goroutines each emit 10 log records.
	var logWg sync.WaitGroup
	for i := 0; i < 100; i++ {
		i := i
		logWg.Add(1)
		go func() {
			defer logWg.Done()
			for j := 0; j < 10; j++ {
				logger.Info("log", "goroutine", i, "iter", j)
			}
		}()
	}
	logWg.Wait()
	wg.Wait()
}

// TestProgressConcurrentSetIncrement exercises atomic ops under concurrent writes.
func TestProgressConcurrentSetIncrement(t *testing.T) {
	t.Parallel()
	p := term.NewProgress(10000)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				p.Increment(1)
			}
		}()
	}
	// Concurrent Render calls.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				p.Render()
			}
		}()
	}
	wg.Wait()
}

// TestStatusBarConcurrentSetRemove exercises the StatusBar's RWMutex under
// concurrent mutations and renders.
func TestStatusBarConcurrentSetRemove(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := "k"
			for j := 0; j < 50; j++ {
				if j%3 == 0 {
					sb.Remove(key)
				} else {
					sb.SetSection(key, "val")
				}
				_ = i
			}
		}()
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				sb.Render()
			}
		}()
	}
	wg.Wait()
}

// TestAnimatorConcurrentAddDuringRun exercises Add called concurrently with Run.
func TestAnimatorConcurrentAddDuringRun(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := term.NewAnimator(logger,
		term.WithWriter(io.Discard),
		term.WithRefreshRate(10*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = a.Run(ctx)
	}()

	// Add animations concurrently while Run is active.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			h := a.Add(&neverDoneAnimation{})
			time.Sleep(10 * time.Millisecond)
			h.Cancel()
		}()
	}
	wg.Wait()
}

// TestDownloadProgressConcurrentRead verifies thread-safe reads on DownloadProgress.
func TestDownloadProgressConcurrentRead(t *testing.T) {
	t.Parallel()
	// We can't call Read concurrently on the same DownloadProgress since
	// io.Reader is typically not concurrent-safe. Instead we test that
	// Increment and Render are safe concurrently.
	p := term.NewProgress(1000)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				p.Increment(5)
				_, _ = p.Render()
			}
		}()
	}
	wg.Wait()
}
