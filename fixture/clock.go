// SPDX-License-Identifier: Apache-2.0

package fixture

import (
	"sort"
	"sync"
	"time"

	"github.com/nathanbrophy/glacier/assert"
)

// Clock is the injectable time interface for production code.
// Implementations must be goroutine-safe.
type Clock interface {
	// Now returns the current time.
	Now() time.Time
	// Sleep blocks until d has elapsed (real) or Advance is called (fake).
	Sleep(d time.Duration)
	// After returns a channel that receives the time after d has elapsed.
	After(d time.Duration) <-chan time.Time
}

// FakeClock extends Clock with deterministic control methods.
// Returned by NewClock. Goroutine-safe.
type FakeClock interface {
	Clock
	// Advance moves the fake clock forward by d, firing all timers whose
	// deadline falls within [now, now+d] in chronological order.
	Advance(d time.Duration)
	// SetTime moves the fake clock to t (may be before current time).
	// Timers whose deadline has passed after the jump fire immediately.
	SetTime(t time.Time)
}

// realClock is the production implementation backed by the standard library.
type realClock struct{}

// Real returns a Clock backed by the real wall clock (time.Now, time.Sleep,
// time.After). Suitable for production use and for tests that do not need
// deterministic time.
func Real() Clock { return realClock{} }

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) Sleep(d time.Duration)                  { time.Sleep(d) }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// pendingTimer is a timer registered via FakeClock.After.
type pendingTimer struct {
	deadline time.Time
	ch       chan time.Time
}

// fakeClock is the deterministic fake clock implementation.
// Internal state is protected by mu; no concur.Mutex is used
// (fixture must not import concur per F4).
type fakeClock struct {
	mu      sync.Mutex
	now     time.Time
	timers  []*pendingTimer
	sleepCh chan struct{} // broadcast-channel trick: close-and-replace to unblock all sleepers
}

// NewClock returns a FakeClock frozen at start. The clock only advances when
// Advance or SetTime is called. Timer channels created via After fire on the
// next Advance that crosses their deadline. NewClock registers a t.Cleanup
// that asserts no pending timer channels remain (to catch misconfigured tests).
func NewClock(t assert.TB, start time.Time) FakeClock {
	fc := &fakeClock{
		now:     start,
		sleepCh: make(chan struct{}),
	}
	t.Cleanup(func() {
		fc.mu.Lock()
		pending := len(fc.timers)
		fc.mu.Unlock()
		if pending > 0 {
			t.Errorf("fixture: NewClock: %d pending timer(s) remain at cleanup; check that Advance was called", pending)
		}
	})
	return fc
}

func (fc *fakeClock) Now() time.Time {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.now
}

// Sleep blocks until d has elapsed on the fake clock (i.e., until Advance
// moves the clock past fc.now+d). Goroutine-safe; multiple concurrent sleepers
// are all unblocked when Advance passes their deadline.
func (fc *fakeClock) Sleep(d time.Duration) {
	if d <= 0 {
		return
	}
	fc.mu.Lock()
	deadline := fc.now.Add(d)
	// Register as a pending timer so Advance fires us.
	ch := make(chan time.Time, 1)
	pt := &pendingTimer{deadline: deadline, ch: ch}
	fc.timers = append(fc.timers, pt)
	sort.Slice(fc.timers, func(i, j int) bool {
		return fc.timers[i].deadline.Before(fc.timers[j].deadline)
	})
	fc.mu.Unlock()

	// Block until the timer fires.
	<-ch
}

// After returns a channel that receives the current fake time when d has
// elapsed (i.e., when Advance crosses the deadline). Multiple timers with
// the same deadline fire in insertion order. Goroutine-safe.
func (fc *fakeClock) After(d time.Duration) <-chan time.Time {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	ch := make(chan time.Time, 1)
	if d <= 0 {
		ch <- fc.now
		return ch
	}
	deadline := fc.now.Add(d)
	pt := &pendingTimer{deadline: deadline, ch: ch}
	fc.timers = append(fc.timers, pt)
	sort.Slice(fc.timers, func(i, j int) bool {
		return fc.timers[i].deadline.Before(fc.timers[j].deadline)
	})
	return ch
}

// Advance moves the fake clock forward by d, firing all timers whose
// deadline falls within [now, now+d] in chronological order.
func (fc *fakeClock) Advance(d time.Duration) {
	if d <= 0 {
		return
	}
	fc.mu.Lock()
	fc.now = fc.now.Add(d)
	fc.fireExpiredLocked()
	fc.mu.Unlock()
}

// SetTime moves the fake clock to t (may be before current time).
// Timers whose deadline has passed after the jump fire immediately.
func (fc *fakeClock) SetTime(t time.Time) {
	fc.mu.Lock()
	fc.now = t
	fc.fireExpiredLocked()
	fc.mu.Unlock()
}

// fireExpiredLocked fires all pending timers whose deadline <= fc.now.
// Must be called with fc.mu held.
func (fc *fakeClock) fireExpiredLocked() {
	newNow := fc.now
	remaining := fc.timers[:0]
	for _, pt := range fc.timers {
		if !pt.deadline.After(newNow) {
			// Fire — non-blocking send (channel is buffered with capacity 1).
			select {
			case pt.ch <- newNow:
			default:
			}
		} else {
			remaining = append(remaining, pt)
		}
	}
	fc.timers = remaining
}
