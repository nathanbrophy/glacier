// SPDX-License-Identifier: Apache-2.0

// Package shimmer renders the "GLACIER" wordmark with a scrolling aurora
// gradient per spec 0032 D-S58.
//
// Phase advances by 1 each tick (100 ms). A full cycle takes 6 ticks (600 ms).
// Each character at position i receives color stop (phase+i) % 6.
package shimmer

import (
	"fmt"
	"strings"
)

// wordmark is the fixed string rendered by Wordmark.
const wordmark = "GLACIER"

// stops is the 6-stop aurora gradient: cyan → teal → blue → back to cyan.
// Chosen to evoke an icy ribbon scrolling across the letters.
var stops = [6][3]uint8{
	{140, 235, 255}, // bright cyan
	{110, 225, 255}, // sky cyan
	{80, 210, 255},  // light blue-cyan
	{60, 180, 255},  // mid blue
	{90, 150, 255},  // blue-violet
	{120, 180, 255}, // periwinkle
}

// stop256 maps each 24-bit stop to the closest xterm-256 color index for
// terminals that do not support true color. Precomputed to avoid allocation.
//
// Indices were selected by visual inspection against the xterm-256 palette.
var stops256 = [6]uint8{
	123, // #87ffff :  closest xterm-256 to (140,235,255)
	117, // #87d7ff :  closest to (110,225,255)
	75,  // #5fafff :  closest to (80,210,255)
	69,  // #5f87ff :  closest to (60,180,255)
	63,  // #5f5fff :  closest to (90,150,255)
	69,  // #5f87ff :  closest to (120,180,255)
}

// Wordmark returns the "GLACIER" wordmark string with ANSI color escapes applied
// when color24 or color256 is true. Each character is colored by gradient stop
// (phase+i) % 6 where i is the character's position.
//
// Pass phase = tick % 6, where tick increments once per 100 ms frame.
//
// When color24 is true, 24-bit ESC[38;2;R;G;Bm sequences are used.
// When only color256 is true, ESC[38;5;Nm 256-color sequences are used.
// When both are false, the plain string "GLACIER" is returned.
//
// The visible width of the returned string is always len("GLACIER") == 7.
func Wordmark(phase int, color24, color256 bool) string {
	if !color24 && !color256 {
		return wordmark
	}
	var b strings.Builder
	// Each character gets its own escape + reset pair. A colored 7-char string
	// carries at most 7*(esc_open + 1 + reset) bytes ≈ 160 bytes :  well within
	// a stack-sized buffer on any platform.
	for i, ch := range wordmark {
		idx := (phase + i) % 6
		if color24 {
			r, g, bv := stops[idx][0], stops[idx][1], stops[idx][2]
			fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm%c\x1b[0m", r, g, bv, ch)
		} else {
			fmt.Fprintf(&b, "\x1b[38;5;%dm%c\x1b[0m", stops256[idx], ch)
		}
	}
	return b.String()
}
