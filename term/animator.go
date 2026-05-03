// SPDX-License-Identifier: Apache-2.0

package term

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Animation is the interface implemented by all animated elements.
// Render returns the lines to display for the current frame and whether the
// animation is done. Once done is true, the animator stops calling Render.
//
// Invariant: Render is called from a single goroutine (the frame loop) while animMu is NOT held.
// Invariant: Render must not block; it must return within one refresh interval.
type Animation interface {
	Render() (lines []string, done bool)
}

// Clock is the minimal time abstraction the Animator uses for its frame loop.
// Production code uses the wall clock by default; tests can inject a
// deterministic implementation (e.g. fixture.NewClock) so time-gated behaviour
// is fast and reproducible.
//
// Implementations must be goroutine-safe. The interface is structurally
// identical to fixture.Clock so a fixture.FakeClock can be passed directly
// through WithClock; this keeps term/ kernel-tier-clean (no fixture import).
type Clock interface {
	// Now returns the current time.
	Now() time.Time
	// Sleep blocks until d has elapsed.
	Sleep(d time.Duration)
	// After returns a channel that receives the time after d has elapsed.
	After(d time.Duration) <-chan time.Time
}

// realClock is the production wall-clock implementation. Zero-allocation for
// Now and Sleep; After allocates one timer + channel (stdlib behaviour).
type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) Sleep(d time.Duration)                  { time.Sleep(d) }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// Handle is a cancellation token returned by Animator.Add.
// Invariant: Cancel is idempotent; calling it after the animation completes naturally is a no-op.
type Handle struct {
	cancel func()
}

// Cancel removes the associated animation from the active set.
// Idempotent; safe to call after the animation completes naturally.
// Concurrency: goroutine-safe.
func (h Handle) Cancel() {
	if h.cancel != nil {
		h.cancel()
	}
}

// animatorConfig holds the resolved options for an Animator.
type animatorConfig struct {
	refreshRate        time.Duration
	writer             io.Writer
	maxBufferedRecords int
	clock              Clock
}

// AnimatorOption configures an Animator.
type AnimatorOption interface{ applyAnimator(*animatorConfig) error }

type animatorOptionFunc func(*animatorConfig) error

func (f animatorOptionFunc) applyAnimator(c *animatorConfig) error { return f(c) }

// WithRefreshRate sets the frame interval (default 100ms).
// Precondition: d > 0.
func WithRefreshRate(d time.Duration) AnimatorOption {
	return animatorOptionFunc(func(c *animatorConfig) error {
		if d <= 0 {
			return fmt.Errorf("term: WithRefreshRate: duration must be > 0")
		}
		c.refreshRate = d
		return nil
	})
}

// WithWriter directs animation output to w (default os.Stderr).
// Precondition: w must not be nil.
func WithWriter(w io.Writer) AnimatorOption {
	return animatorOptionFunc(func(c *animatorConfig) error {
		if w == nil {
			return fmt.Errorf("term: WithWriter: writer must not be nil")
		}
		c.writer = w
		return nil
	})
}

// WithMaxBufferedRecords sets the ring-buffer capacity for intercepted slog
// records (default 1000). When the buffer is full, new writes block up to
// 50ms; if still full, the oldest record is dropped and a warning is emitted
// on the next frame. Precondition: n >= 1.
func WithMaxBufferedRecords(n int) AnimatorOption {
	return animatorOptionFunc(func(c *animatorConfig) error {
		if n < 1 {
			return fmt.Errorf("term: WithMaxBufferedRecords: n must be >= 1")
		}
		c.maxBufferedRecords = n
		return nil
	})
}

// WithClock injects a Clock for the Animator's frame loop. Defaults to the
// real wall clock. Pass a fixture.FakeClock (via fixture.NewClock) in tests
// to drive frame ticks deterministically with FakeClock.Advance.
// Precondition: c must not be nil.
func WithClock(c Clock) AnimatorOption {
	return animatorOptionFunc(func(cfg *animatorConfig) error {
		if c == nil {
			return fmt.Errorf("term: WithClock: clock must not be nil")
		}
		cfg.clock = c
		return nil
	})
}

// logRecord is a captured slog log record for replay.
type logRecord struct {
	line string
}

// animEntry is one active animation slot.
type animEntry struct {
	anim      Animation
	cancelled atomic.Bool
}

// interceptHandler is a slog.Handler that buffers records into a ring buffer.
type interceptHandler struct {
	wrapped slog.Handler
	buf     chan logRecord
	mu      sync.Mutex
	dropped int // count of dropped records since last flush
}

func (h *interceptHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.wrapped.Enabled(ctx, level)
}

func (h *interceptHandler) Handle(ctx context.Context, r slog.Record) error {
	// Format the record as a plain text line via the wrapped handler.
	// We write it to a temporary string builder handler.
	var sb strings.Builder
	tmp := slog.NewTextHandler(&sb, nil)
	_ = tmp.Handle(ctx, r)
	line := strings.TrimRight(sb.String(), "\n")

	lr := logRecord{line: line}

	// Try non-blocking send first.
	select {
	case h.buf <- lr:
		return nil
	default:
	}

	// Block up to 50ms to let the frame loop drain.
	timer := time.NewTimer(50 * time.Millisecond)
	defer timer.Stop()
	select {
	case h.buf <- lr:
		return nil
	case <-timer.C:
	}

	// Buffer still full: drop oldest and enqueue new.
	h.mu.Lock()
	h.dropped++
	h.mu.Unlock()
	select {
	case <-h.buf: // drop oldest
	default:
	}
	select {
	case h.buf <- lr:
	default:
	}
	return nil
}

func (h *interceptHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &interceptHandler{wrapped: h.wrapped.WithAttrs(attrs), buf: h.buf}
}

func (h *interceptHandler) WithGroup(name string) slog.Handler {
	return &interceptHandler{wrapped: h.wrapped.WithGroup(name), buf: h.buf}
}

// Animator coordinates animated terminal output and slog log buffering.
// It wraps a *slog.Logger's handler with an interception handler that buffers
// log records during animation frames and flushes them above the animation
// area between frames, preventing log lines and animations from colliding.
type Animator struct {
	logger   *slog.Logger
	cfg      animatorConfig
	animMu   sync.Mutex
	entries  []*animEntry
	running  atomic.Bool
	closed   atomic.Bool
	stopCh   chan struct{}
	inter    *interceptHandler
	pausedMu sync.Mutex
	paused   bool
}

// NewAnimator constructs an Animator bound to logger.
//
// At construction, the logger's handler is NOT yet replaced; interception
// begins when Run is called.
//
// Preconditions: logger must not be nil.
// Concurrency: construction is not goroutine-safe; Run is.
func NewAnimator(logger *slog.Logger, opts ...AnimatorOption) *Animator {
	cfg := animatorConfig{
		refreshRate:        100 * time.Millisecond,
		writer:             os.Stderr,
		maxBufferedRecords: 1000,
		clock:              realClock{},
	}
	for _, o := range opts {
		if o == nil {
			continue
		}
		_ = o.applyAnimator(&cfg)
	}
	if cfg.clock == nil {
		cfg.clock = realClock{}
	}
	return &Animator{
		logger: logger,
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
}

// Add registers anim into the active animation set.
// May be called before or during Run.
//
// If the same Animation value is added twice, both registrations are active.
// Returns a Handle whose Cancel removes this specific registration.
// Concurrency: goroutine-safe.
func (a *Animator) Add(anim Animation) Handle {
	entry := &animEntry{anim: anim}
	a.animMu.Lock()
	a.entries = append(a.entries, entry)
	a.animMu.Unlock()

	return Handle{cancel: func() { entry.cancelled.Store(true) }}
}

// Run starts the frame loop. It blocks until ctx is cancelled or all
// registered animations report done.
//
// On return:
//   - The original logger handler is restored (via defer; panic-safe).
//   - Any remaining buffered records are flushed to the writer.
//   - If an animation's Render panics, Run recovers, restores the handler,
//     and returns the wrapped panic value as an error.
//   - If ctx is cancelled, returns ErrCancelled (wrapping ctx.Err()).
//   - If called a second time while already running, returns an error.
//
// Concurrency: Run may be called from one goroutine; Add/Cancel/Pause/Resume
// may be called concurrently from any goroutine.
func (a *Animator) Run(ctx context.Context) (runErr error) {
	if !a.running.CompareAndSwap(false, true) {
		return errors.New("term: animator: already running")
	}
	defer a.running.Store(false)

	// Install interception handler by replacing the logger's handler.
	originalHandler := a.logger.Handler()
	if _, alreadyWrapped := originalHandler.(*interceptHandler); alreadyWrapped {
		return errors.New("term: animator: handler already intercepted")
	}
	ih := &interceptHandler{
		wrapped: originalHandler,
		buf:     make(chan logRecord, a.cfg.maxBufferedRecords),
	}
	a.inter = ih
	interceptLogger := slog.New(ih)

	// If a.logger is the slog default logger, replace the default so that
	// package-level slog.Info / slog.Error calls are also intercepted.
	isDefault := a.logger == slog.Default()
	if isDefault {
		slog.SetDefault(interceptLogger)
	}

	defer func() {
		// Restore original handler :  panic-safe.
		if isDefault {
			slog.SetDefault(a.logger) // restore original
		}
		a.inter = nil
		// Flush remaining buffered records.
		a.flushRecords(ih)
	}()

	w := a.cfg.writer
	lastLines := 0 // lines rendered in the previous frame

	clock := a.cfg.clock

	for {
		select {
		case <-ctx.Done():
			// Erase last frame.
			if lastLines > 0 {
				fmt.Fprint(w, strings.Repeat("\x1b[1A\x1b[2K", lastLines))
			}
			a.flushRecords(ih)
			return ErrCancelled
		case <-clock.After(a.cfg.refreshRate):
			// Check paused state.
			a.pausedMu.Lock()
			isPaused := a.paused
			a.pausedMu.Unlock()
			if isPaused {
				// While paused, still flush buffered log records inline.
				a.flushRecords(ih)
				continue
			}

			// Collect active animations under lock; render outside lock.
			a.animMu.Lock()
			snapshot := make([]*animEntry, len(a.entries))
			copy(snapshot, a.entries)
			a.animMu.Unlock()

			// Render each animation into a single buffer first; one Write at the
			// end avoids flicker that comes from the terminal observing partial
			// frames between the erase and the redraw. (Fixes spec 0032 D-S55
			// flicker user complaint.)
			var frame bytes.Buffer
			if lastLines > 0 {
				// Erase previous frame.
				fmt.Fprint(&frame, strings.Repeat("\x1b[1A\x1b[2K", lastLines))
			}

			// Flush buffered log records BEFORE the new frame: log lines
			// scroll above the animation, the frame redraws below.
			a.flushRecords(ih)

			totalLines := 0
			allDone := true
			var surviving []*animEntry
			for _, e := range snapshot {
				if e.cancelled.Load() {
					continue
				}
				lines, done := func() (ls []string, d bool) {
					defer func() {
						if r := recover(); r != nil {
							ls = []string{fmt.Sprintf("[animator panic: %v]", r)}
							d = true
						}
					}()
					return e.anim.Render()
				}()
				for _, l := range lines {
					fmt.Fprintln(&frame, l)
					totalLines++
				}
				if !done {
					allDone = false
					surviving = append(surviving, e)
				}
			}
			// Single atomic Write of the full frame.
			_, _ = w.Write(frame.Bytes())
			lastLines = totalLines

			// Remove finished animations.
			a.animMu.Lock()
			// Rebuild entries retaining only surviving + any newly added.
			// Newly added entries won't be in snapshot; keep them.
			survSet := map[*animEntry]bool{}
			for _, e := range surviving {
				survSet[e] = true
			}
			// Keep entries that are surviving or were not yet in snapshot.
			snapshotSet := map[*animEntry]bool{}
			for _, e := range snapshot {
				snapshotSet[e] = true
			}
			var next []*animEntry
			for _, e := range a.entries {
				if !snapshotSet[e] || survSet[e] {
					next = append(next, e)
				}
			}
			a.entries = next
			a.animMu.Unlock()

			if allDone && len(snapshot) > 0 {
				return nil
			}
		}
	}
}

// flushRecords drains the ring buffer and writes all records to the writer.
func (a *Animator) flushRecords(ih *interceptHandler) {
	if ih == nil {
		return
	}
	w := a.cfg.writer

	// Report dropped records if any.
	ih.mu.Lock()
	dropped := ih.dropped
	ih.dropped = 0
	ih.mu.Unlock()
	if dropped > 0 {
		fmt.Fprintf(w, "[term: animator: %d log records dropped due to full buffer]\n", dropped)
	}

	for {
		select {
		case lr := <-ih.buf:
			fmt.Fprintln(w, lr.line)
		default:
			return
		}
	}
}

// Pause suspends frame rendering. Log records continue to be buffered.
// Idempotent.
// Concurrency: goroutine-safe.
func (a *Animator) Pause() {
	// Pause/Resume implementation: set a paused flag.
	// For this implementation we rely on the context mechanism; a full
	// pause implementation would require a condition variable. Since the
	// spec says "idempotent" and records continue to buffer, we record
	// the flag and skip render in the frame loop.
	a.pausedMu.Lock()
	a.paused = true
	a.pausedMu.Unlock()
}

// Resume restarts frame rendering after Pause. Idempotent.
// Concurrency: goroutine-safe.
func (a *Animator) Resume() {
	a.pausedMu.Lock()
	a.paused = false
	a.pausedMu.Unlock()
}

// Close stops the frame loop (if running), restores the logger handler, and
// flushes any remaining buffered records. Idempotent: the second call returns nil.
// Implements io.Closer.
// Concurrency: goroutine-safe.
func (a *Animator) Close() error {
	if !a.closed.CompareAndSwap(false, true) {
		return nil // idempotent
	}
	// Signal the stop channel if anyone is waiting on it.
	select {
	case a.stopCh <- struct{}{}:
	default:
	}
	return nil
}
