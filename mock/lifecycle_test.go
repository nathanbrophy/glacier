// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/mock"
)

func TestVerifyIdempotent(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	// No calls → first Verify reports error.
	m.Verify()
	firstErrorCount := len(ft.errors)
	// Second Verify must be a no-op.
	m.Verify()
	if len(ft.errors) != firstErrorCount {
		t.Fatal("second Verify should not add more errors")
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	_ = m.Close()
	firstCount := len(ft.errors)
	_ = m.Close()
	if len(ft.errors) != firstCount {
		t.Fatal("second Close() should not re-run Verify")
	}
}

func TestCloseAlwaysReturnsNil(t *testing.T) {
	m := mock.Of[Greeter](t)
	m.OnCall("Greet").Return("hi").AnyTimes()
	if err := m.Close(); err != nil {
		t.Fatalf("Close() returned non-nil: %v", err)
	}
}

func TestCleanupPreemptsVerify(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	// Manual Close preempts the t.Cleanup Verify.
	_ = m.Close()
	errorCountAfterClose := len(ft.errors)
	ft.runCleanup()
	if len(ft.errors) != errorCountAfterClose {
		t.Fatal("t.Cleanup after manual Close() should be a no-op")
	}
}
