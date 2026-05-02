// SPDX-License-Identifier: Apache-2.0

package assert

import "testing"

// §21.4 F10, E23

func TestHaltCallsFailNow(t *testing.T) {
	mt := &mockTB{}
	Halt(mt)
	if mt.failNowCalls != 1 {
		t.Fatalf("Halt: FailNow called %d times, want 1", mt.failNowCalls)
	}
}
