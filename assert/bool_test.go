// SPDX-License-Identifier: Apache-2.0

package assert

import "testing"

// §21.4 F3

func TestTrue(t *testing.T) {
	mt := &mockTB{}
	True(t, True(mt, true), "True(true) = true")
	if mt.errorfCalls != 0 {
		t.Fatalf("True(true): Errorf called %d times, want 0", mt.errorfCalls)
	}
	mt.reset()
	False(t, True(mt, false), "True(false) = false")
	if mt.errorfCalls != 1 {
		t.Fatalf("True(false): Errorf called %d times, want 1", mt.errorfCalls)
	}
}

func TestFalse(t *testing.T) {
	mt := &mockTB{}
	True(t, False(mt, false), "False(false) = true")
	if mt.errorfCalls != 0 {
		t.Fatalf("False(false): Errorf called %d times, want 0", mt.errorfCalls)
	}
	mt.reset()
	False(t, False(mt, true), "False(true) = false")
	if mt.errorfCalls != 1 {
		t.Fatalf("False(true): Errorf called %d times, want 1", mt.errorfCalls)
	}
}

func TestTrueWithMessage(t *testing.T) {
	mt := &mockTB{}
	True(mt, false, "context message")
	// Just verify it doesn't panic.
	True(t, mt.errorfCalls == 1, "Errorf called once")
}
