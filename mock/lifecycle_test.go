// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
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
	assert.True(t, len(ft.errors) == firstErrorCount, "second Verify should not add more errors")
}

func TestCloseIsIdempotent(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	_ = m.Close()
	firstCount := len(ft.errors)
	_ = m.Close()
	assert.True(t, len(ft.errors) == firstCount, "second Close() should not re-run Verify")
}

func TestCloseAlwaysReturnsNil(t *testing.T) {
	m := mock.Of[Greeter](t)
	m.OnCall("Greet").Return("hi").AnyTimes()
	require.NoError(t, m.Close(), "Close() returned non-nil")
}

func TestCleanupPreemptsVerify(t *testing.T) {
	ft := newFakeT()
	m := mock.Of[Greeter](ft)
	m.OnCall("Greet").Return("hi").Times(1)
	// Manual Close preempts the t.Cleanup Verify.
	_ = m.Close()
	errorCountAfterClose := len(ft.errors)
	ft.runCleanup()
	assert.True(t, len(ft.errors) == errorCountAfterClose, "t.Cleanup after manual Close() should be a no-op")
}
