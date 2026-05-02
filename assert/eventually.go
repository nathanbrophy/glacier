// SPDX-License-Identifier: Apache-2.0

package assert

import "time"

// Eventually polls fn at interval until it returns true or timeout elapses.
// On success returns true. On timeout reports via t.Errorf and returns false.
// The failure message includes the configured timeout.
//
// Preconditions: fn is non-nil; timeout > 0; interval > 0; interval <= timeout.
// Postconditions: fn is called at least once.
// Concurrency: goroutine-safe. fn is called from the same goroutine as the
// caller.
//
// §21.4 E20
func Eventually(t TB, fn func() bool, timeout, interval time.Duration, msg ...any) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if fn() {
			return true
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(interval)
		if time.Now().After(deadline) {
			break
		}
	}
	suffix := fmtMsg(msg)
	t.Errorf("Eventually failed: condition not met within %s.%s", timeout, suffix)
	return false
}
