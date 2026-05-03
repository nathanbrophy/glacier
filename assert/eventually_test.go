// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"testing"
	"time"
)

// §21.4 E20

func TestEventuallyPassesEarly(t *testing.T) {
	mt := &mockTB{}
	count := 0
	ok := Eventually(mt, func() bool {
		count++
		return count >= 1
	}, 100*time.Millisecond, 10*time.Millisecond)
	True(t, ok, "Eventually: fn returns true on first poll")
	Equal(t, mt.errorfCalls, 0)
}

func TestEventuallyTimeout(t *testing.T) {
	mt := &mockTB{}
	ok := Eventually(mt, func() bool { return false },
		50*time.Millisecond, 10*time.Millisecond)
	False(t, ok, "Eventually: fn never returns true → timeout")
	Equal(t, mt.errorfCalls, 1)
}

func TestEventuallyHonorsInterval(t *testing.T) {
	mt := &mockTB{}
	count := 0
	start := time.Now()
	Eventually(mt, func() bool {
		count++
		return count >= 3
	}, 500*time.Millisecond, 20*time.Millisecond)
	elapsed := time.Since(start)
	True(t, elapsed >= 20*time.Millisecond, "Eventually: at least one interval elapsed")
}

// L-add-7: Eventually where fn panics :  propagated (not treated as condition-false).
func TestEventuallyFnPanics(t *testing.T) {
	mt := &mockTB{}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic to propagate")
		}
	}()
	Eventually(mt, func() bool {
		panic("test panic")
	}, 100*time.Millisecond, 10*time.Millisecond)
}
