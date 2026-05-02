// SPDX-License-Identifier: Apache-2.0

package assert

// mockTB is a hand-rolled testing double that records Errorf / Fatalf /
// FailNow invocations. It is the bootstrap helper that allows testing the
// assert package without recursive self-dependency.
//
// Bootstrap discipline: mockTB is itself verified with bare-if checks only.
// It intentionally satisfies the TB interface so it can be passed to any
// assertion function.
type mockTB struct {
	errorfCalls  int
	fatalfCalls  int
	failNowCalls int
	lastMessage  string
	helperCalls  int
	cleanupFns   []func()
	name         string
}

func (m *mockTB) Helper() { m.helperCalls++ }
func (m *mockTB) Errorf(format string, args ...any) {
	m.errorfCalls++
	m.lastMessage = format
}
func (m *mockTB) Fatalf(format string, args ...any) {
	m.fatalfCalls++
	m.lastMessage = format
}
func (m *mockTB) FailNow() { m.failNowCalls++ }
func (m *mockTB) Cleanup(fn func()) {
	if fn != nil {
		m.cleanupFns = append(m.cleanupFns, fn)
	}
}
func (m *mockTB) Name() string {
	if m.name != "" {
		return m.name
	}
	return "mockTB"
}

// reset clears all recorded state.
func (m *mockTB) reset() {
	m.errorfCalls = 0
	m.fatalfCalls = 0
	m.failNowCalls = 0
	m.lastMessage = ""
	m.helperCalls = 0
	m.cleanupFns = nil
}
