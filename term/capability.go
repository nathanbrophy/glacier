// SPDX-License-Identifier: Apache-2.0

package term

import (
	"io"
	"os"
	"sync"
)

// ColorSupport is the enum of color depth levels a terminal can render.
// Values are monotonically increasing capability.
type ColorSupport int

const (
	// ColorNone means no color; plain text only.
	ColorNone ColorSupport = iota
	// Color16 means 16-color ANSI (ESC[3xm).
	Color16
	// Color256 means 256-color (ESC[38;5;Nm).
	Color256
	// Color24Bit means 24-bit true color (ESC[38;2;R;G;Bm).
	Color24Bit
)

// Capabilities reports terminal-rendering capabilities for a given writer.
//
// Invariant: if IsTTY is false, SupportsColor == ColorNone and SupportsUTF8 == false.
// Invariant: Width == 0 and Height == 0 when IsTTY is false.
// Invariant: NoColorEnv == true suppresses all color regardless of SupportsColor.
type Capabilities struct {
	IsTTY         bool
	SupportsColor ColorSupport
	SupportsUTF8  bool
	Width, Height int
	// NoColorEnv is set when NO_COLOR or GLACIER_NO_COLOR is present in env.
	// GLACIER_NO_COLOR takes precedence when both are set.
	NoColorEnv bool
}

// capCache caches Capabilities per writer pointer.
var capCache sync.Map // map[io.Writer]Capabilities

// Capability probes w for terminal capabilities. Results are cached per writer
// so repeated calls are zero-allocation after the first.
//
// Preconditions: w must not be nil.
// Postconditions: returns a fully-populated Capabilities struct.
// Concurrency: goroutine-safe; cache uses sync.Map.
func Capability(w io.Writer) Capabilities {
	if v, ok := capCache.Load(w); ok {
		return v.(Capabilities)
	}
	caps := probe(w)
	capCache.Store(w, caps)
	return caps
}

// probe does the actual capability detection. platform-specific helpers
// (isTTY, termSize, colorSupport) are defined in capability_unix.go /
// capability_windows.go.
func probe(w io.Writer) Capabilities {
	noColor := os.Getenv("GLACIER_NO_COLOR") != "" || os.Getenv("NO_COLOR") != ""

	fd, ok := writerFd(w)
	if !ok {
		return Capabilities{NoColorEnv: noColor}
	}

	tty := isTTY(fd)
	if !tty {
		return Capabilities{NoColorEnv: noColor}
	}

	cs := detectColorSupport()
	w2, h := termSize(fd)

	return Capabilities{
		IsTTY:         true,
		SupportsColor: cs,
		SupportsUTF8:  true, // conservative: TTY == UTF-8 capable
		Width:         w2,
		Height:        h,
		NoColorEnv:    noColor,
	}
}

// writerFd tries to extract a file descriptor from w.
func writerFd(w io.Writer) (uintptr, bool) {
	type fder interface{ Fd() uintptr }
	if f, ok := w.(fder); ok {
		return f.Fd(), true
	}
	return 0, false
}

// detectColorSupport reads environment variables to determine color depth.
func detectColorSupport() ColorSupport {
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		return Color24Bit
	}
	term := os.Getenv("TERM")
	if term == "xterm-256color" || term == "screen-256color" || term == "tmux-256color" {
		return Color256
	}
	// Windows Terminal sets WT_SESSION.
	if os.Getenv("WT_SESSION") != "" {
		return Color24Bit
	}
	if term != "" && term != "dumb" {
		return Color16
	}
	return ColorNone
}
