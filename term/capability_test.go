// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/term"
)

func TestCapabilityIODiscard(t *testing.T) {
	t.Parallel()
	caps := term.Capability(io.Discard)
	assert.False(t, caps.IsTTY, "IsTTY: got true, want false for io.Discard")
	assert.Equal(t, caps.SupportsColor, term.ColorNone)
	assert.Equal(t, caps.Width, 0)
	assert.Equal(t, caps.Height, 0)
	assert.False(t, caps.SupportsUTF8, "SupportsUTF8: got true, want false for io.Discard")
}

func TestCapabilityBytesBuffer(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	caps := term.Capability(buf)
	assert.False(t, caps.IsTTY, "IsTTY: got true, want false for bytes.Buffer")
	assert.Equal(t, caps.SupportsColor, term.ColorNone)
}

func TestCapabilityCOLORTERMTruecolor(t *testing.T) {
	// Note: non-TTY writers always return ColorNone regardless of COLORTERM.
	// This tests that the env parsing works when probed with a real TTY fd.
	// We test the colorSupport function indirectly via detectColorSupport.
	// Cannot use t.Parallel() with t.Setenv.
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TERM", "")
	// Using bytes.Buffer so IsTTY = false; color detection path not reached.
	// We verify detectColorSupport via a real TTY test in cross_platform_test.go.
	caps := term.Capability(&bytes.Buffer{})
	if caps.IsTTY {
		t.Skip("expected non-TTY writer")
	}
	// Non-TTY always ColorNone regardless of env.
	assert.Equal(t, caps.SupportsColor, term.ColorNone)
}

func TestCapabilityNoColor(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv.
	t.Setenv("NO_COLOR", "1")
	t.Setenv("GLACIER_NO_COLOR", "")
	caps := term.Capability(&bytes.Buffer{})
	assert.True(t, caps.NoColorEnv, "NoColorEnv: got false, want true when NO_COLOR=1")
}

func TestCapabilityGlacierNoColor(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv.
	t.Setenv("GLACIER_NO_COLOR", "1")
	caps := term.Capability(&bytes.Buffer{})
	assert.True(t, caps.NoColorEnv, "NoColorEnv: got false, want true when GLACIER_NO_COLOR=1")
}

func TestCapabilityCachedPerWriter(t *testing.T) {
	// AllocsPerRun cannot be used in a parallel test.
	w := &bytes.Buffer{}
	// Warm the cache.
	_ = term.Capability(w)
	// Second call must be zero-allocation.
	allocs := testing.AllocsPerRun(100, func() {
		_ = term.Capability(w)
	})
	assert.Equal(t, allocs, float64(0))
}

func TestCapabilityNoBothNoColor(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv.
	t.Setenv("NO_COLOR", "")
	t.Setenv("GLACIER_NO_COLOR", "")
	caps := term.Capability(&bytes.Buffer{})
	assert.False(t, caps.NoColorEnv, "NoColorEnv: got true, want false when neither NO_COLOR env is set")
}

// TestColorSupportConstants verifies monotonic ordering.
func TestColorSupportConstants(t *testing.T) {
	t.Parallel()
	assert.True(t, term.ColorNone < term.Color16, "ColorNone must be < Color16")
	assert.True(t, term.Color16 < term.Color256, "Color16 must be < Color256")
	assert.True(t, term.Color256 < term.Color24Bit, "Color256 must be < Color24Bit")
}

// Table-driven tests for ColorSupport values.
func TestColorSupportValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cs   term.ColorSupport
		want int
	}{
		{"none", term.ColorNone, 0},
		{"16", term.Color16, 1},
		{"256", term.Color256, 2},
		{"24bit", term.Color24Bit, 3},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, int(tc.cs), tc.want)
		})
	}
}
