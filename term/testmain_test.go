// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"os"
	"testing"

	"github.com/nathanbrophy/glacier/term"
)

// TestMain runs all tests under ModeAuto so tests against bytes.Buffer (a
// non-TTY writer) get the capability-based "no color" behavior they were
// written to expect. The package's runtime default is ModeAlways (color on);
// resetting at TestMain keeps the test suite hermetic.
func TestMain(m *testing.M) {
	prev := term.GetColorMode()
	term.SetColorMode(term.ModeAuto)
	code := m.Run()
	term.SetColorMode(prev)
	os.Exit(code)
}
