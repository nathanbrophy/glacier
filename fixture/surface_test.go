// SPDX-License-Identifier: Apache-2.0

package fixture_test

import (
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/fixture"
)

// TestSurfaceExports verifies that all expected public symbols are accessible.
// This is a compile-time + runtime surface guard.
func TestSurfaceExports(t *testing.T) {
	// Golden options.
	var _ fixture.GoldenOption = fixture.WithRoot("testdata")
	// LeakOptions.
	var _ fixture.LeakOption = fixture.WatchTempDirs()
	var _ fixture.LeakOption = fixture.WatchGoroutines()
	var _ fixture.LeakOption = fixture.WatchEnv()
	var _ fixture.LeakOption = fixture.WatchFDs()
	var _ fixture.LeakOption = fixture.WatchAll()
	var _ fixture.LeakOption = fixture.StrictLeaks()
	var _ fixture.LeakOption = fixture.WithDrainTimeout(time.Second)

	// Clock interfaces.
	var _ fixture.Clock = fixture.Real()
	var _ fixture.FakeClock = fixture.NewClock(t, time.Now())

	// Constants.
	var _ time.Duration = fixture.Millisecond

	// Sentinel.
	var _ error = fixture.ErrPathRejected
}

// TestClockInterfaceCompliance verifies that realClock and fakeClock satisfy
// both Clock and (fakeClock only) FakeClock.
func TestClockInterfaceCompliance(t *testing.T) {
	var _ fixture.Clock = fixture.Real()

	clk := fixture.NewClock(t, time.Now())
	var _ fixture.Clock = clk
	var _ fixture.FakeClock = clk
}

// TestNewFSInterfaceCompliance verifies that the returned FS satisfies the
// expected interfaces (already covered in mockfs_test.go but stated here for
// the surface matrix).
func TestNewFSInterfaceCompliance(t *testing.T) {
	t.Helper()
	fsys := fixture.NewFS(nil)
	_ = fsys // presence of NewFS is enough; interface checks in mockfs_test.go
}
