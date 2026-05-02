// SPDX-License-Identifier: Apache-2.0

//go:build linux || darwin

package fixture

import (
	"os"
	"strings"

	"github.com/nathanbrophy/glacier/assert"
)

// fdProcDir returns the directory listing open file descriptors on the
// current OS: /proc/self/fd on Linux, /dev/fd on macOS.
func fdProcDir() string {
	if _, err := os.Stat("/proc/self/fd"); err == nil {
		return "/proc/self/fd"
	}
	return "/dev/fd"
}

// countFDs counts open file descriptors by listing the pseudo-filesystem.
// Returns -1 if the directory is unreadable.
func countFDs() int {
	entries, err := os.ReadDir(fdProcDir())
	if err != nil {
		return -1
	}
	return len(entries)
}

func checkFDLeaks(t assert.TB, report func(string, ...any), base int) {
	t.Helper()
	if base < 0 {
		return // baseline unavailable; skip silently
	}
	current := countFDs()
	if current < 0 {
		return
	}
	// The ReadDir call itself opens FDs; allow a small buffer.
	const tolerance = 2
	if current > base+tolerance {
		entries, _ := os.ReadDir(fdProcDir())
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		report("fixture: GuardLeaks: WatchFDs detected %d new file descriptor(s) (baseline %d, current %d): [%s]",
			current-base, base, current, strings.Join(names, ", "))
	}
}
