// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/fixture"
	"github.com/nathanbrophy/glacier/term"
)

// epoch is a fixed instant used as the start time for fake clocks.
// 2026-05-01 00:00:00 UTC keeps tests independent of the wall clock.
var epoch = time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)

// mockAnimation is a simple test animation that counts Render calls.
type mockAnimation struct {
	renders atomic.Int32
	doneAt  int32 // 0 = never done
}

func (m *mockAnimation) Render() ([]string, bool) {
	n := m.renders.Add(1)
	if m.doneAt > 0 && n >= m.doneAt {
		return []string{"done"}, true
	}
	return []string{string(rune('A' + int(n)%26))}, false
}

// neverDoneAnimation never reports done — used with ctx cancellation.
type neverDoneAnimation struct{}

func (n *neverDoneAnimation) Render() ([]string, bool) {
	return []string{"."}, false
}

// runAsync starts a.Run in a goroutine and returns a function that waits for
// it to exit and returns its error.
func runAsync(a *term.Animator, ctx context.Context) func() error {
	errCh := make(chan error, 1)
	go func() { errCh <- a.Run(ctx) }()
	return func() error { return <-errCh }
}

// tick fires exactly one frame on the fake clock. It first waits for the
// frame loop to register its After timer, then advances past the deadline.
// This eliminates the race between test and animator-goroutine scheduling.
func tick(fc fixture.FakeClock, refresh time.Duration) {
	fc.BlockUntilTimers(1)
	fc.Advance(refresh)
}

// drain advances the fake clock past any straggling After timers registered by
// the frame loop after cancellation. Without this, NewClock's t.Cleanup would
// flag the dangling timer.
func drain(fc fixture.FakeClock, refresh time.Duration) {
	fc.Advance(refresh * 4)
}

func TestNewAnimatorOptions(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var buf bytes.Buffer
	a := term.NewAnimator(logger,
		term.WithRefreshRate(50*time.Millisecond),
		term.WithWriter(&buf),
		term.WithMaxBufferedRecords(10),
	)
	assert.NotNil(t, a, "NewAnimator should return non-nil")
}

func TestAnimatorRunCancelled(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var buf bytes.Buffer
	fc := fixture.NewClock(t, epoch)
	const refresh = 10 * time.Millisecond
	a := term.NewAnimator(logger,
		term.WithWriter(&buf),
		term.WithRefreshRate(refresh),
		term.WithClock(fc),
	)
	a.Add(&neverDoneAnimation{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before Run starts

	err := a.Run(ctx)
	assert.True(t, errors.Is(err, term.ErrCancelled),
		"Run(cancelled) = %v, want ErrCancelled", err)

	// Drain any residual After timer registered before ctx.Done was selected.
	drain(fc, refresh)
}

func TestAnimatorRunAlreadyRunning(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var buf bytes.Buffer
	fc := fixture.NewClock(t, epoch)
	const refresh = 200 * time.Millisecond
	a := term.NewAnimator(logger,
		term.WithWriter(&buf),
		term.WithRefreshRate(refresh),
		term.WithClock(fc),
	)
	a.Add(&neverDoneAnimation{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wait := runAsync(a, ctx)
	// Drive one frame so the loop is definitely "running".
	tick(fc, refresh)

	err := a.Run(ctx)
	assert.NotNil(t, err, "second Run should return error")
	if err != nil {
		assert.Contains(t, err.Error(), "already running")
	}

	cancel()
	_ = wait()
	drain(fc, refresh)
}

func TestAnimatorRestoresHandlerOnExit(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	original := slog.NewTextHandler(&buf, nil)
	logger := slog.New(original)

	fc := fixture.NewClock(t, epoch)
	const refresh = 10 * time.Millisecond
	a := term.NewAnimator(logger,
		term.WithWriter(io.Discard),
		term.WithRefreshRate(refresh),
		term.WithClock(fc),
	)

	// Animation that completes after 2 renders.
	m := &mockAnimation{doneAt: 2}
	a.Add(m)

	wait := runAsync(a, context.Background())
	// Drive frames until the animation reports done. doneAt=2 means
	// after 2 frames, allDone=true and Run returns nil.
	tick(fc, refresh)
	tick(fc, refresh)

	assert.NoError(t, wait(), "Run should exit cleanly when animation reports done")

	// After Run, the logger's handler should be the original.
	logger.Info("post-run message")
	assert.Contains(t, buf.String(), "post-run message")
}

func TestAnimatorHandleCancel(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var buf bytes.Buffer
	fc := fixture.NewClock(t, epoch)
	const refresh = 20 * time.Millisecond
	a := term.NewAnimator(logger,
		term.WithWriter(&buf),
		term.WithRefreshRate(refresh),
		term.WithClock(fc),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handle := a.Add(&neverDoneAnimation{})
	handle.Cancel() // cancel before Run starts

	wait := runAsync(a, ctx)
	// Cancelled entries are skipped; loop continues until ctx is cancelled.
	tick(fc, refresh)
	cancel()
	err := wait()
	if err != nil {
		assert.True(t, errors.Is(err, term.ErrCancelled),
			"Run = %v, want nil or ErrCancelled", err)
	}
	drain(fc, refresh)
}

func TestAnimatorCloseIdempotent(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := term.NewAnimator(logger)
	assert.NoError(t, a.Close(), "Close() first call")
	assert.NoError(t, a.Close(), "Close() second call (idempotent)")
}

func TestAnimatorBuffersLogsDuringFrame(t *testing.T) {
	// Uses slog.SetDefault; cannot run in parallel with other slog tests.
	var buf bytes.Buffer
	fc := fixture.NewClock(t, epoch)
	const refresh = 50 * time.Millisecond
	a := term.NewAnimator(slog.Default(),
		term.WithWriter(&buf),
		term.WithRefreshRate(refresh),
		term.WithClock(fc),
	)

	m := &mockAnimation{doneAt: 3}
	a.Add(m)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wait := runAsync(a, ctx)

	// Wait for the frame loop's first After registration; that guarantees
	// the intercept handler is installed (it happens earlier in Run).
	fc.BlockUntilTimers(1)
	slog.Info("mid-animation log")

	// Drive frames until the animation completes (doneAt: 3).
	tick(fc, refresh)
	tick(fc, refresh)
	tick(fc, refresh)

	assert.NoError(t, wait(), "Run")

	out := buf.String()
	assert.Contains(t, out, "mid-animation log")
}

func TestAnimatorPauseResume(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var buf bytes.Buffer
	fc := fixture.NewClock(t, epoch)
	const refresh = 20 * time.Millisecond
	a := term.NewAnimator(logger,
		term.WithWriter(&buf),
		term.WithRefreshRate(refresh),
		term.WithClock(fc),
	)
	a.Add(&neverDoneAnimation{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wait := runAsync(a, ctx)
	tick(fc, refresh)

	// Pause/Resume should not panic.
	a.Pause()
	a.Pause() // idempotent
	tick(fc, refresh)
	a.Resume()
	a.Resume() // idempotent
	cancel()
	_ = wait()
	drain(fc, refresh)
}

func TestWithRefreshRateInvalidReturnsError(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// A zero refresh rate produces an error from the option, but NewAnimator
	// silently swallows option errors. Verify only that the constructor still
	// returns non-nil.
	a := term.NewAnimator(logger, term.WithRefreshRate(0))
	assert.NotNil(t, a, "NewAnimator should return non-nil even on bad option")
}

func TestWithClockNilReturnsErrorOption(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Nil clock option errors out, but NewAnimator silently swallows option
	// errors and falls back to the default real clock.
	a := term.NewAnimator(logger, term.WithClock(nil))
	assert.NotNil(t, a, "NewAnimator should return non-nil")
}

func TestAnimatorAddMultiple(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fc := fixture.NewClock(t, epoch)
	const refresh = 10 * time.Millisecond
	a := term.NewAnimator(logger,
		term.WithWriter(io.Discard),
		term.WithRefreshRate(refresh),
		term.WithClock(fc),
	)

	m1 := &mockAnimation{doneAt: 2}
	m2 := &mockAnimation{doneAt: 2}
	a.Add(m1)
	a.Add(m2)

	wait := runAsync(a, context.Background())
	tick(fc, refresh)
	tick(fc, refresh)
	assert.NoError(t, wait(), "Run")
	assert.GreaterOrEqual(t, m1.renders.Load(), int32(2),
		"m1 should have rendered ≥ 2 frames")
}

// Ensure fixture's FakeClock structurally satisfies term.Clock — the whole
// dogfooding story rests on this. Compile-time assertion only.
var _ term.Clock = (fixture.FakeClock)(nil)
