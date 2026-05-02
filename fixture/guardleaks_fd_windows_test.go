// SPDX-License-Identifier: Apache-2.0

//go:build windows

package fixture_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/fixture"
)

// TestGuardLeaksFDsNoOpOnWindows: WatchFDs is a no-op on Windows. (#59)
func TestGuardLeaksFDsNoOpOnWindows(t *testing.T) {
	// Capture the debug log message emitted by countFDs on Windows.
	stdout, stderr := fixture.Capture(t, func() {
		m := newMockTB()
		fixture.GuardLeaks(m, fixture.WatchFDs())
		m.runCleanups()
		if m.Failed() {
			t.Errorf("WatchFDs reported failure on Windows; should be a no-op")
		}
	})
	// The debug message may appear on stderr or stdout; just verify no failure.
	_ = stdout
	_ = stderr
}
