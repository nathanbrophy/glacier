// SPDX-License-Identifier: Apache-2.0

package assert

// Halt calls t.FailNow, halting the current test goroutine immediately.
// Use as a named alternative to an inline t.FailNow() call, typically
// after a block of assertions:
//
//	if !assert.Equal(t, got, want) {
//	    assert.Halt(t)  // stop here; next assertion would panic on nil
//	}
//
// Preconditions: t is non-nil.
//
// §21.4 F10, E23
func Halt(t TB) {
	t.Helper()
	t.FailNow()
}
