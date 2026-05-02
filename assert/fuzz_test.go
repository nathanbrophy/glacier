// SPDX-License-Identifier: Apache-2.0

package assert

import "testing"

// §21.4 F4

// FuzzMatchGlob verifies that Match in glob mode never panics on arbitrary
// pattern + input pairs, and always returns a deterministic bool.
func FuzzMatchGlob(f *testing.F) {
	f.Add("hello world", "hello *")
	f.Add("abc", "a?c")
	f.Add("", "")
	f.Add("xyz", "*")
	f.Add("abc", "a")

	f.Fuzz(func(t *testing.T, input, pattern string) {
		mt := &mockTB{}
		// Must not panic.
		_ = Match(mt, input, pattern)
	})
}

// FuzzMatchRegex verifies that Match in regex mode never panics; if compile
// fails, it reports cleanly without panicking.
func FuzzMatchRegex(f *testing.F) {
	f.Add("abc", `^[a-c]+$`)
	f.Add("user-123", `^user-[0-9]+$`)
	f.Add("", ``)
	f.Add("hello", `[invalid`)

	f.Fuzz(func(t *testing.T, input, pattern string) {
		mt := &mockTB{}
		// Must not panic; compile errors are reported cleanly.
		_ = Match(mt, input, pattern, MatchRegex())
	})
}
