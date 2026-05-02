// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/term"
)

func TestRGBConstructor(t *testing.T) {
	t.Parallel()
	c := term.RGB(0x22, 0xD3, 0xEE)
	assert.Equal(t, c.R, uint8(0x22))
	assert.Equal(t, c.G, uint8(0xD3))
	assert.Equal(t, c.B, uint8(0xEE))
}

func TestHexConstructorValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  term.Color
	}{
		{"#22D3EE", term.RGB(0x22, 0xD3, 0xEE)},
		{"22D3EE", term.RGB(0x22, 0xD3, 0xEE)},
		{"#fff", term.RGB(0xFF, 0xFF, 0xFF)},
		{"fff", term.RGB(0xFF, 0xFF, 0xFF)},
		{"#000000", term.RGB(0, 0, 0)},
		{"#aabbcc", term.RGB(0xAA, 0xBB, 0xCC)},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got, err := term.Hex(tc.input)
			require.NoError(t, err, "Hex("+tc.input+")")
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestHexConstructorInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
	}{
		{"not-a-color"},
		{"#gg0000"},
		{"#12345"},   // 5 digits
		{"#1234567"}, // 7 digits
		{""},
		{"#"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			_, err := term.Hex(tc.input)
			require.Error(t, err, "Hex("+tc.input+"): expected error")
			he, ok := err.(*term.HexParseError)
			if !ok {
				t.Fatalf("Hex(%q): expected *HexParseError, got %T", tc.input, err)
			}
			assert.Equal(t, he.Input, tc.input)
			// Error must match ^term: hex:
			assert.True(t, len(err.Error()) >= 10 && err.Error()[:9] == "term: hex",
				"HexParseError.Error() must start with 'term: hex:'")
		})
	}
}

func TestNamedColorsMatchSpec0001(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		got  term.Color
		hex  string
	}{
		{"Cyan", term.Cyan, "#22D3EE"},
		{"Teal", term.Teal, "#2DD4BF"},
		{"Bg", term.Bg, "#0F172A"},
		{"Surface", term.Surface, "#1E293B"},
		{"Surface2", term.Surface2, "#334155"},
		{"Text", term.Text, "#F1F5F9"},
		{"TextMuted", term.TextMuted, "#94A3B8"},
		{"TextFaint", term.TextFaint, "#64748B"},
		{"Success", term.Success, "#4ADE80"},
		{"Warning", term.Warning, "#FACC15"},
		{"Error", term.Error, "#F87171"},
		{"Info", term.Info, "#60A5FA"},
		{"Border", term.Border, "#334155"},
		{"Cyan100", term.Cyan100, "#CFFAFE"},
		{"Cyan300", term.Cyan300, "#67E8F9"},
		{"Cyan500", term.Cyan500, "#06B6D4"},
		{"Cyan700", term.Cyan700, "#0E7490"},
		{"Teal500", term.Teal500, "#14B8A6"},
		{"Teal700", term.Teal700, "#0F766E"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			want, err := term.Hex(tc.hex)
			require.NoError(t, err, "Hex("+tc.hex+")")
			assert.Equal(t, tc.got, want)
		})
	}
}

func TestColorValueType(t *testing.T) {
	t.Parallel()
	a := term.RGB(1, 2, 3)
	b := a
	b.R = 99
	assert.NotEqual(t, a.R, b.R)
}
