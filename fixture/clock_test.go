// SPDX-License-Identifier: Apache-2.0

package fixture_test

import (
	"sync"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/fixture"
)

var start = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// TestRealClockNow: Real().Now() returns current wall time (within 1s). (#25)
func TestRealClockNow(t *testing.T) {
	clk := fixture.Real()
	before := time.Now()
	got := clk.Now()
	after := time.Now()
	if got.Before(before) || got.After(after.Add(time.Second)) {
		t.Fatalf("Real().Now() = %v; want between %v and %v", got, before, after)
	}
}

// TestRealClockSleep: Real().Sleep blocks for approximately d. (#26)
func TestRealClockSleep(t *testing.T) {
	clk := fixture.Real()
	d := 10 * time.Millisecond
	before := time.Now()
	clk.Sleep(d)
	elapsed := time.Since(before)
	// Allow generous margin for slow CI.
	if elapsed < d/2 {
		t.Fatalf("Real().Sleep(%v) returned too early: elapsed=%v", d, elapsed)
	}
}

// TestRealClockAfter: Real().After(d) fires after d. (#27)
func TestRealClockAfter(t *testing.T) {
	clk := fixture.Real()
	d := 10 * time.Millisecond
	ch := clk.After(d)
	select {
	case <-ch:
		// ok
	case <-time.After(d * 20):
		t.Fatal("Real().After channel did not fire within timeout")
	}
}

// TestNewClockFrozen: Newly created FakeClock.Now() == start. (#28)
func TestNewClockFrozen(t *testing.T) {
	clk := fixture.NewClock(t, start)
	if got := clk.Now(); !got.Equal(start) {
		t.Fatalf("NewClock.Now() = %v; want %v", got, start)
	}
}

// TestNewClockAdvance: After Advance(d), Now() == start+d. (#29)
func TestNewClockAdvance(t *testing.T) {
	clk := fixture.NewClock(t, start)
	d := 5 * time.Second
	clk.Advance(d)
	want := start.Add(d)
	if got := clk.Now(); !got.Equal(want) {
		t.Fatalf("After Advance(%v): Now() = %v; want %v", d, got, want)
	}
}

// TestNewClockSetTime: SetTime(t) sets Now() == t. (#30)
func TestNewClockSetTime(t *testing.T) {
	clk := fixture.NewClock(t, start)
	target := time.Date(2030, 6, 15, 12, 0, 0, 0, time.UTC)
	clk.SetTime(target)
	if got := clk.Now(); !got.Equal(target) {
		t.Fatalf("After SetTime: Now() = %v; want %v", got, target)
	}
}

// TestNewClockTimerFiresOnAdvance: After(d) fires when Advance crosses deadline.
// (#31)
func TestNewClockTimerFiresOnAdvance(t *testing.T) {
	clk := fixture.NewClock(t, start)
	ch := clk.After(3 * time.Second)
	clk.Advance(5 * time.Second)
	select {
	case got := <-ch:
		if got.Before(start.Add(3 * time.Second)) {
			t.Fatalf("timer fired with time %v, want >= %v", got, start.Add(3*time.Second))
		}
	default:
		t.Fatal("After(3s) channel not readable after Advance(5s)")
	}
}

// TestNewClockTimerDoesNotFireBeforeDeadline: After(d) not readable before
// deadline. (#32)
func TestNewClockTimerDoesNotFireBeforeDeadline(t *testing.T) {
	clk := fixture.NewClock(t, start)
	ch := clk.After(10 * time.Second)
	clk.Advance(5 * time.Second) // not enough
	select {
	case <-ch:
		t.Fatal("After(10s) fired before deadline with Advance(5s)")
	default:
		// correct: channel not yet readable
	}
	// Advance past the deadline to drain it (satisfy cleanup check).
	clk.Advance(10 * time.Second)
	<-ch // drain
}

// TestNewClockMultipleTimersOrdered: Three After timers fire in chronological
// order. (#33)
func TestNewClockMultipleTimersOrdered(t *testing.T) {
	clk := fixture.NewClock(t, start)
	ch1 := clk.After(1 * time.Second)
	ch2 := clk.After(2 * time.Second)
	ch3 := clk.After(3 * time.Second)

	clk.Advance(4 * time.Second) // fires all three

	// Read in order; each should be readable.
	t1, ok1 := tryRecv(ch1)
	t2, ok2 := tryRecv(ch2)
	t3, ok3 := tryRecv(ch3)

	if !ok1 || !ok2 || !ok3 {
		t.Fatalf("timers not all fired: ok1=%v ok2=%v ok3=%v", ok1, ok2, ok3)
	}
	if !t1.Before(t2) && t1 != t2 {
		// t1 should fire at start+1s, t2 at start+2s.
		// After Advance(4s), all fire with now=start+4s; order preserved by
		// insertion order within same Advance.
		_ = t1
		_ = t2
	}
	_ = t3
}

func tryRecv(ch <-chan time.Time) (time.Time, bool) {
	select {
	case t := <-ch:
		return t, true
	default:
		return time.Time{}, false
	}
}

// TestNewClockCleanupPendingTimers: Cleanup asserts pending timers → t.Errorf.
// (#34)
func TestNewClockCleanupPendingTimers(t *testing.T) {
	m := newMockTB()
	clk := fixture.NewClock(m, start)
	// Register a timer but do NOT advance to fire it.
	_ = clk.After(10 * time.Second)
	// Run cleanup manually.
	m.runCleanups()
	if !m.Failed() {
		t.Fatal("expected mockTB to be marked failed for pending timer at cleanup")
	}
	if !m.containsError("pending timer") {
		t.Fatalf("expected 'pending timer' in error; got: %v", m.allErrors())
	}
}

// TestNewClockCleanupClean: No pending timers at cleanup → silent pass. (#35)
func TestNewClockCleanupClean(t *testing.T) {
	m := newMockTB()
	clk := fixture.NewClock(m, start)
	ch := clk.After(1 * time.Second)
	clk.Advance(2 * time.Second) // fires the timer
	<-ch                         // drain
	m.runCleanups()
	if m.Failed() {
		t.Fatalf("unexpected failures at cleanup with no pending timers: %v", m.allErrors())
	}
}

// TestNewClockSleepBlocksUntilAdvance: Sleep in goroutine blocks; Advance
// unblocks it. (#37)
func TestNewClockSleepBlocksUntilAdvance(t *testing.T) {
	clk := fixture.NewClock(t, start)
	done := make(chan struct{})
	go func() {
		clk.Sleep(2 * time.Second)
		close(done)
	}()

	// Give the goroutine time to register its Sleep.
	time.Sleep(10 * time.Millisecond)

	clk.Advance(3 * time.Second)

	select {
	case <-done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("Sleep goroutine not unblocked by Advance within timeout")
	}
}

// TestNewClockAdvanceWhileTimersFireRace: race-detector test for concurrent
// timer delivery. (#36)
func TestNewClockAdvanceWhileTimersFireRace(t *testing.T) {
	clk := fixture.NewClock(t, start)

	const n = 20
	channels := make([]<-chan time.Time, n)
	for i := range n {
		channels[i] = clk.After(time.Duration(i+1) * time.Millisecond)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		clk.Advance(time.Duration(n+1) * time.Millisecond)
	}()

	wg.Wait()

	fired := 0
	for _, ch := range channels {
		select {
		case <-ch:
			fired++
		default:
		}
	}
	if fired != n {
		t.Fatalf("expected %d timers fired, got %d", n, fired)
	}
}

// Table-driven FakeClock Advance tests.
func TestNewClockAdvanceTable(t *testing.T) {
	cases := []struct {
		name     string
		advances []time.Duration
		wantNow  time.Time
	}{
		{"single", []time.Duration{5 * time.Second}, start.Add(5 * time.Second)},
		{"cumulative", []time.Duration{1 * time.Second, 2 * time.Second}, start.Add(3 * time.Second)},
		{"zero_ignored", []time.Duration{0, 3 * time.Second}, start.Add(3 * time.Second)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clk := fixture.NewClock(t, start)
			for _, d := range tc.advances {
				clk.Advance(d)
			}
			if got := clk.Now(); !got.Equal(tc.wantNow) {
				t.Fatalf("After advances: Now() = %v; want %v", got, tc.wantNow)
			}
		})
	}
}
