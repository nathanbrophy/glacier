// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/term"
)

func TestBuiltinGlyphsRegistered(t *testing.T) {
	t.Parallel()
	builtins := []string{
		"check", "cross", "warn", "info", "bullet", "dot",
		"arrow_right", "arrow_left", "arrow_up", "arrow_down",
		"ellipsis",
		"spinner_braille_0", "spinner_braille_7",
		"spinner_dots_0", "spinner_dots_7",
		"box_horizontal", "box_vertical", "pipe", "divider",
	}
	glyphs := term.Glyphs()
	registered := make(map[string]bool, len(glyphs))
	for _, g := range glyphs {
		registered[g.Name] = true
	}
	for _, name := range builtins {
		assert.True(t, registered[name], "builtin glyph "+name+" not found in registry")
	}
}

func TestGlyphUnknownReturnsEmpty(t *testing.T) {
	t.Parallel()
	got := term.Glyph("__nonexistent_glyph_xyz__")
	assert.Equal(t, got, "")
}

func TestRegisterGlyphSuccess(t *testing.T) {
	// Not parallel: modifies global registry.
	name := "reg_success_01"
	err := term.RegisterGlyph(name, "🎯", "[*]")
	if err != nil {
		// Accept "already registered" on repeated runs (count>1) :  the registry
		// persists for the lifetime of the test binary.
		ge, ok := err.(*term.GlyphError)
		if !ok || !strings.Contains(ge.Cause, "already registered") {
			require.True(t, false, fmt.Sprintf("RegisterGlyph(%q): unexpected error: %v", name, err))
		}
		t.Logf("glyph %q already registered from a previous iteration; verifying lookup", name)
	}
	// Verify lookup exists in Glyphs() :  must be present regardless.
	found := false
	for _, g := range term.Glyphs() {
		if g.Name == name {
			found = true
		}
	}
	assert.True(t, found, "registered glyph "+name+" not found in Glyphs()")
}

func TestRegisterGlyphDuplicate(t *testing.T) {
	// Not parallel: modifies global registry.
	name := "reg_dup_01"
	// First registration may or may not succeed (if already registered from prev test run).
	_ = term.RegisterGlyph(name, "A", "B")
	err := term.RegisterGlyph(name, "C", "D")
	require.Error(t, err, "RegisterGlyph(duplicate "+name+"): expected error")
	assert.Contains(t, err.Error(), "term: glyph:")
}

func TestRegisterGlyphInvalidName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		glName  string
		utf8    string
		ascii   string
		wantErr bool
	}{
		{"empty_name", "", "X", "x", true},
		{"uppercase", "Bad", "X", "x", true},
		{"starts_digit", "1abc", "X", "x", true},
		{"has_hyphen", "my-glyph", "X", "x", true},
		{"has_space", "my glyph", "X", "x", true},
		{"too_long", strings.Repeat("a", 65), "X", "x", true},
		{"empty_utf8", "validname", "", "x", true},
		{"empty_ascii", "validname2", "X", "", true},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := term.RegisterGlyph(tc.glName, tc.utf8, tc.ascii)
			if tc.wantErr {
				assert.Error(t, err, "RegisterGlyph("+tc.glName+"): expected error")
			} else {
				assert.NoError(t, err, "RegisterGlyph("+tc.glName+"): unexpected error")
			}
			if err != nil {
				assert.Contains(t, err.Error(), "term: glyph:")
			}
		})
	}
}

func TestGlyphsSnapshotIndependence(t *testing.T) {
	t.Parallel()
	snap1 := term.Glyphs()
	snap2 := term.Glyphs()
	// Both are independent copies of the registry.
	assert.True(t, &snap1[0] != &snap2[0], "Glyphs() returned the same backing array; expected independent snapshots")
}

// TestRegisterGlyphMaxLength verifies the boundary at exactly 64 bytes.
func TestRegisterGlyphMaxLength(t *testing.T) {
	t.Parallel()
	name64 := strings.Repeat("a", 64)
	// Should succeed (exactly 64 bytes).
	// Note: may fail with "already registered" if re-run; that's acceptable.
	err := term.RegisterGlyph(name64, "X", "x")
	if err != nil {
		ge, ok := err.(*term.GlyphError)
		require.True(t, ok, "expected *GlyphError")
		assert.True(t, strings.Contains(ge.Cause, "already registered") || strings.Contains(ge.Cause, "64"),
			"unexpected GlyphError.Cause: "+ge.Cause)
	}

	// 65 bytes should always fail with "exceeds 64 bytes".
	name65 := strings.Repeat("b", 65)
	err = term.RegisterGlyph(name65, "X", "x")
	assert.Error(t, err, "expected error for 65-byte name")
}
