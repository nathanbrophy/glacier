// SPDX-License-Identifier: Apache-2.0

//go:build windows

package fixture

import (
	"fmt"
	"os"
	"sync"

	"github.com/nathanbrophy/glacier/assert"
)

var fdWarnOnce sync.Once

// countFDs is a no-op on Windows. The FD-counting mechanism relies on
// /proc/self/fd (Linux) and /dev/fd (macOS), neither of which exist on
// Windows. Returns -1 to signal "unavailable". Emits a one-time debug
// message to os.Stderr.
func countFDs() int {
	fdWarnOnce.Do(func() {
		fmt.Fprintln(os.Stderr,
			"fixture: GuardLeaks: WatchFDs is a no-op on Windows (requires /proc/self/fd or /dev/fd)")
	})
	return -1
}

func checkFDLeaks(_ assert.TB, _ func(string, ...any), _ int) {
	// No-op on Windows; countFDs emits a one-time debug message.
	countFDs()
}
