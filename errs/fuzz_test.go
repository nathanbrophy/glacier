// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/errs"
)

// FuzzSentinelRegister fuzzes the validRegister path via Sentinel.
// Invariants checked:
//   - If text contains ASCII uppercase A–Z, Sentinel panics.
//   - If text ends with '.', Sentinel panics.
//   - If text is empty, Sentinel panics.
//   - If text contains no ':', Sentinel panics.
//   - Otherwise, Sentinel does not panic.
func FuzzSentinelRegister(f *testing.F) {
	// Seed corpus from spec examples.
	f.Add("pkg: cause")
	f.Add("cli: cancelled")
	f.Add("cli: unknown flag")
	f.Add("pkg:")
	f.Add("")
	f.Add("Pkg: cause")
	f.Add("pkg: cause.")
	f.Add("nocolon")
	f.Add("pkg: Ünicode")
	f.Add("A")
	f.Add("z: lowercase only")
	f.Add("errs: something")

	f.Fuzz(func(t *testing.T, text string) {
		// Determine whether we expect a panic.
		expectPanic := false
		if text == "" {
			expectPanic = true
		}
		// ASCII uppercase check.
		for i := range len(text) {
			c := text[i]
			if c >= 'A' && c <= 'Z' {
				expectPanic = true
				break
			}
		}
		if !strings.Contains(text, ":") {
			expectPanic = true
		}
		if strings.HasSuffix(text, ".") {
			expectPanic = true
		}

		panicked := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			_ = errs.Sentinel(text)
		}()

		if expectPanic {
			assert.True(t, panicked, "Sentinel("+text+"): expected panic, got none")
		} else {
			assert.False(t, panicked, "Sentinel("+text+"): unexpected panic")
		}
	})
}
