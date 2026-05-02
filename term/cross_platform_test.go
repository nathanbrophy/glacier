// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"bytes"
	"os"
	"testing"

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
	if caps.IsTTY {
		t.Error("bytes.Buffer should never be a TTY")
	}
	if caps.SupportsColor != term.ColorNone {
		t.Errorf("bytes.Buffer SupportsColor = %d, want ColorNone", caps.SupportsColor)
	}
	if caps.SupportsUTF8 {
		t.Error("bytes.Buffer should not report SupportsUTF8")
	}
	if caps.Width != 0 || caps.Height != 0 {
		t.Errorf("bytes.Buffer Width/Height = %d/%d, want 0/0", caps.Width, caps.Height)
	}
}

// TestCapabilityNoColorSuppressionInvariant checks: NoColorEnv → no color rendered.
func TestCapabilityNoColorSuppressionInvariant(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv.
	t.Setenv("GLACIER_NO_COLOR", "1")

	var buf bytes.Buffer
	s := term.New().Foreground(term.Cyan).Bold()
	term.Fprint(&buf, s, "text")
	// Even if somehow buf were a TTY (it isn't), NoColorEnv must suppress escapes.
	got := buf.String()
	if got != "text" {
		t.Errorf("NoColorEnv=1 but Fprint produced %q, want plain 'text'", got)
	}
}

// TestCapabilityInvariantNonTTY asserts: non-TTY → SupportsColor == ColorNone and SupportsUTF8 == false.
func TestCapabilityInvariantNonTTY(t *testing.T) {
	t.Parallel()
	caps := term.Capability(&bytes.Buffer{})
	if !caps.IsTTY && caps.SupportsColor != term.ColorNone {
		t.Errorf("invariant violated: IsTTY=false but SupportsColor=%d", caps.SupportsColor)
	}
	if !caps.IsTTY && caps.SupportsUTF8 {
		t.Errorf("invariant violated: IsTTY=false but SupportsUTF8=true")
	}
}
