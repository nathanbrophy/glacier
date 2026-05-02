// SPDX-License-Identifier: Apache-2.0

package assert

// TB is the testing-target interface satisfied by *testing.T, *testing.B,
// and *testing.F. All assertion functions accept TB rather than a concrete
// testing type so they are usable in benchmarks and fuzz tests without
// casting.
type TB interface {
	Helper()
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	FailNow()
	Cleanup(fn func())
	Name() string
}
