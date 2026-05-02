// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/nathanbrophy/glacier/term"
)

func TestCapabilityIODiscard(t *testing.T) {
	t.Parallel()
	caps := term.Capability(io.Discard)
	if caps.IsTTY {
		t.Errorf("IsTTY: got true, want false for io.Discard")
	}
	if caps.SupportsColor != term.ColorNone {
		t.Errorf("SupportsColor: got %d, want ColorNone", caps.SupportsColor)
	}
	if caps.Width != 0 || caps.Height != 0 {
		t.Errorf("Width/Height: got %d/%d, want 0/0", caps.Width, caps.Height)
	}
	if caps.SupportsUTF8 {
		t.Errorf("SupportsUTF8: got true, want false for io.Discard")
	}
}

func TestCapabilityBytesBuffer(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	caps := term.Capability(buf)
	if caps.IsTTY {
		t.Errorf("IsTTY: got true, want false for bytes.Buffer")
	}
	if caps.SupportsColor != term.ColorNone {
		t.Errorf("SupportsColor: got %d, want ColorNone", caps.SupportsColor)
	}
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
	if caps.SupportsColor != term.ColorNone {
		t.Errorf("SupportsColor: got %d, want ColorNone for non-TTY", caps.SupportsColor)
	}
}

func TestCapabilityNoColor(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv.
	t.Setenv("NO_COLOR", "1")
	t.Setenv("GLACIER_NO_COLOR", "")
	caps := term.Capability(&bytes.Buffer{})
	if !caps.NoColorEnv {
		t.Errorf("NoColorEnv: got false, want true when NO_COLOR=1")
	}
}

func TestCapabilityGlacierNoColor(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv.
	t.Setenv("GLACIER_NO_COLOR", "1")
	caps := term.Capability(&bytes.Buffer{})
	if !caps.NoColorEnv {
		t.Errorf("NoColorEnv: got false, want true when GLACIER_NO_COLOR=1")
	}
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
	if allocs != 0 {
		t.Errorf("expected 0 allocs on cache hit, got %v", allocs)
	}
}

func TestCapabilityNoBothNoColor(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv.
	t.Setenv("NO_COLOR", "")
	t.Setenv("GLACIER_NO_COLOR", "")
	caps := term.Capability(&bytes.Buffer{})
	if caps.NoColorEnv {
		t.Errorf("NoColorEnv: got true, want false when neither NO_COLOR env is set")
	}
}

// TestColorSupportConstants verifies monotonic ordering.
func TestColorSupportConstants(t *testing.T) {
	t.Parallel()
	if !(term.ColorNone < term.Color16) {
		t.Error("ColorNone must be < Color16")
	}
	if !(term.Color16 < term.Color256) {
		t.Error("Color16 must be < Color256")
	}
	if !(term.Color256 < term.Color24Bit) {
		t.Error("Color256 must be < Color24Bit")
	}
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
			if int(tc.cs) != tc.want {
				t.Errorf("ColorSupport(%s) = %d, want %d", tc.name, int(tc.cs), tc.want)
			}
		})
	}
}
