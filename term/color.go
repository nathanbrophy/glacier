// SPDX-License-Identifier: Apache-2.0

package term

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// Color holds a 24-bit RGB color value.
// Invariant: R, G, B are in [0, 255].
type Color struct {
	R, G, B uint8
}

// RGB constructs a 24-bit Color from its component channels.
// Concurrency: goroutine-safe; Color is a value type.
func RGB(r, g, b uint8) Color { return Color{R: r, G: g, B: b} }

// Hex parses a CSS-style hex color string (with or without leading '#').
// Accepts 3-digit (#RGB) and 6-digit (#RRGGBB) forms.
// Returns HexParseError on invalid input.
//
// Examples: "#22D3EE", "22D3EE", "#fff", "fff".
func Hex(s string) (Color, error) {
	orig := s
	s = strings.TrimPrefix(s, "#")
	switch len(s) {
	case 3:
		// Expand short form: #RGB → #RRGGBB
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	case 6:
		// Already full form.
	default:
		return Color{}, &HexParseError{
			Input: orig,
			Cause: fmt.Errorf("invalid length %d; want 3 or 6 hex digits", len(s)),
		}
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return Color{}, &HexParseError{Input: orig, Cause: errors.New("invalid hex digits")}
	}
	return Color{R: b[0], G: b[1], B: b[2]}, nil
}

// Spec-0001 palette tokens. Package-level vars initialized at package init;
// never mutated after init.
var (
	// Core palette
	Cyan      Color // #22D3EE
	Teal      Color // #2DD4BF
	Bg        Color // #0F172A
	Surface   Color // #1E293B
	Surface2  Color // #334155
	Text      Color // #F1F5F9
	TextMuted Color // #94A3B8
	TextFaint Color // #64748B

	// Semantic
	Success Color // #4ADE80
	Warning Color // #FACC15
	Error   Color // #F87171
	Info    Color // #60A5FA
	Border  Color // #334155

	// Gradient stops (Glacier brand :  D42)
	Cyan100 Color // #CFFAFE
	Cyan300 Color // #67E8F9
	Cyan500 Color // #06B6D4
	Cyan700 Color // #0E7490
	Teal500 Color // #14B8A6
	Teal700 Color // #0F766E
)

func init() {
	mustHex := func(s string) Color {
		c, err := Hex(s)
		if err != nil {
			//glacier:nolint=panic-in-library programmer error: palette literals are constants, malformed values surface at init.
			panic("term: bad palette hex: " + s)
		}
		return c
	}

	Cyan = mustHex("#22D3EE")
	Teal = mustHex("#2DD4BF")
	Bg = mustHex("#0F172A")
	Surface = mustHex("#1E293B")
	Surface2 = mustHex("#334155")
	Text = mustHex("#F1F5F9")
	TextMuted = mustHex("#94A3B8")
	TextFaint = mustHex("#64748B")

	Success = mustHex("#4ADE80")
	Warning = mustHex("#FACC15")
	Error = mustHex("#F87171")
	Info = mustHex("#60A5FA")
	Border = mustHex("#334155")

	Cyan100 = mustHex("#CFFAFE")
	Cyan300 = mustHex("#67E8F9")
	Cyan500 = mustHex("#06B6D4")
	Cyan700 = mustHex("#0E7490")
	Teal500 = mustHex("#14B8A6")
	Teal700 = mustHex("#0F766E")
}
