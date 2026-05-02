// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/term"
)

// TestCapabilityStdout tests that Capability does not panic on os.Stdout.
func TestCapabilityStdout(t *testing.T) {
	t.Parallel()
	caps := term.Capability(os.Stdout)
	// IsTTY may be true or false depending on the test environment.
	_ = caps
}

// TestCapabilityStderr tests that Capability does not panic on os.Stderr.
func TestCapabilityStderr(t *testing.T) {
	t.Parallel()
	caps := term.Capability(os.Stderr)
	_ = caps
}

// TestCapabilityBytesBufferNonTTY confirms bytes.Buffer is always non-TTY.
func TestCapabilityBytesBufferNonTTY(t *testing.T) {
	t.Parallel()
	caps := term.Capability(&bytes.Buffer{})
	assert.False(t, caps.IsTTY, "bytes.Buffer should never be a TTY")
	assert.Equal(t, caps.SupportsColor, term.ColorNone)
	assert.False(t, caps.SupportsUTF8, "bytes.Buffer should not report SupportsUTF8")
	assert.Equal(t, caps.Width, 0)
	assert.Equal(t, caps.Height, 0)
}

// TestCapabilityNoColorSuppressionInvariant checks: NoColorEnv → no color rendered.
func TestCapabilityNoColorSuppressionInvariant(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv.
	t.Setenv("GLACIER_NO_COLOR", "1")

	var buf bytes.Buffer
	s := term.New().Foreground(term.Cyan).Bold()
	term.Fprint(&buf, s, "text")
	// Even if somehow buf were a TTY (it isn't), NoColorEnv must suppress escapes.
	assert.Equal(t, buf.String(), "text")
}

// TestCapabilityInvariantNonTTY asserts: non-TTY → SupportsColor == ColorNone and SupportsUTF8 == false.
func TestCapabilityInvariantNonTTY(t *testing.T) {
	t.Parallel()
	caps := term.Capability(&bytes.Buffer{})
	if !caps.IsTTY {
		assert.Equal(t, caps.SupportsColor, term.ColorNone)
		assert.False(t, caps.SupportsUTF8, "invariant: IsTTY=false but SupportsUTF8=true")
	}
}
