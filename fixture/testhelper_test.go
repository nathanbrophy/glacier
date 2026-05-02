// SPDX-License-Identifier: Apache-2.0

package fixture_test

import (
	"fmt"
	"strings"
	"sync"
)

// mockTB is a test-double for assert.TB that records errors and fatals.
// It is used in fixture tests where the real *testing.T would be failed by the
// code under test, and we want to assert on the failure message itself.
type mockTB struct {
	mu      sync.Mutex
	errors  []string
	fatals  []string
	failed  bool
	fataled bool

	cleanups []func()
}

func newMockTB() *mockTB { return &mockTB{} }

func (m *mockTB) Helper() {}

func (m *mockTB) Errorf(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	msg := fmt.Sprintf(format, args...)
	m.errors = append(m.errors, msg)
	m.failed = true
}

func (m *mockTB) Fatalf(format string, args ...any) {
	m.mu.Lock()
	msg := fmt.Sprintf(format, args...)
	m.fatals = append(m.fatals, msg)
	m.failed = true
	m.fataled = true
	m.mu.Unlock()
	panic(fatalPanic{msg: msg})
}

func (m *mockTB) FailNow() {
	m.mu.Lock()
	m.fataled = true
	m.failed = true
	m.mu.Unlock()
	panic(fatalPanic{})
}

func (m *mockTB) Cleanup(fn func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanups = append(m.cleanups, fn)
}

func (m *mockTB) Name() string { return "mockTB" }

// runCleanups executes registered cleanup functions in LIFO order.
func (m *mockTB) runCleanups() {
	m.mu.Lock()
	cs := make([]func(), len(m.cleanups))
	copy(cs, m.cleanups)
	m.mu.Unlock()
	for i := len(cs) - 1; i >= 0; i-- {
		cs[i]()
	}
}

// Failed reports whether any Errorf or Fatalf was called.
func (m *mockTB) Failed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.failed
}

// containsError reports whether any recorded error contains substr.
func (m *mockTB) containsError(substr string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range append(m.errors, m.fatals...) {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}

// allErrors returns all recorded errors and fatals concatenated.
func (m *mockTB) allErrors() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, 0, len(m.errors)+len(m.fatals))
	out = append(out, m.errors...)
	out = append(out, m.fatals...)
	return out
}

// fatalPanic is the sentinel panic value used by mockTB.Fatalf to simulate
// testing.T's behavior of stopping the goroutine.
type fatalPanic struct{ msg string }

// callAndRecover calls fn, recovering from any fatalPanic, and returns
// whether fn panicked with a fatalPanic.
func callAndRecover(fn func()) (panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(fatalPanic); ok {
				panicked = true
			} else {
				panic(r) // re-panic non-fatal panics
			}
		}
	}()
	fn()
	return false
}
