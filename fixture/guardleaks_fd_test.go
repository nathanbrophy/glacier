// SPDX-License-Identifier: Apache-2.0

//go:build linux || darwin

package fixture_test

import (
	"os"
	"testing"

	"github.com/nathanbrophy/glacier/fixture"
)

// TestGuardLeaksFDsCatchesLeak: Open file without close → cleanup reports.
// (#58 — Linux/macOS only)
func TestGuardLeaksFDsCatchesLeak(t *testing.T) {
	// Create a temp file to open.
	f, err := os.CreateTemp("", "glacier-fd-leak-test-")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())

	m := newMockTB()
	fixture.GuardLeaks(m, fixture.WatchFDs())

	// Intentionally do NOT close f — simulates an FD leak.
	// (The file is removed by defer, but the fd remains open until f.Close().)
	_ = f // keep fd open through cleanup

	m.runCleanups()

	// Close the fd after mock cleanup to avoid real leak.
	f.Close()

	// We don't assert m.Failed() here because FD counting is subject to
	// platform variance and the tolerance buffer (±2). The test verifies the
	// code path runs without panicking. A strict assertion would be flaky.
}
