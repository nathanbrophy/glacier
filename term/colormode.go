// SPDX-License-Identifier: Apache-2.0

package term

import (
	"io"
	"os"
	"sync"
	"sync/atomic"
)

// Mode is the global color emission policy.
//
// The default for the SDK is ModeAlways: every writer gets ANSI color unless
// the user opts out via --no-color, NO_COLOR, or GLACIER_NO_COLOR. This
// matches the spec 0032 user direction "color on by default, with a toggle
// off flag" and avoids the false-negative problem of TTY-based gating
// (Windows Console, MinTTY, IDE consoles, etc. that lie about their TTY-ness).
//
// Callers that want true TTY auto-detection (suppress color when piped to
// a file) can set ModeAuto explicitly; ModeAuto consults Capability(w).
type Mode int32

const (
	// ModeAlways emits color regardless of TTY detection. This is the
	// default for the SDK; turn off via SetColorMode(ModeNever) or env.
	ModeAlways Mode = iota

	// ModeNever suppresses color on every writer. Use this for --no-color,
	// NO_COLOR, GLACIER_NO_COLOR, or pipelines that need plain text.
	ModeNever

	// ModeAuto emits color only when the writer reports a color-capable
	// TTY. Useful for tools that must produce plain output when piped.
	ModeAuto
)

// colorMode is the package-wide color emission policy. Defaults to ModeAlways
// (color ON by default per spec 0032 user direction). Callers override via
// SetColorMode based on --no-color / --force-color / NO_COLOR.
var colorMode atomic.Int32 // stores Mode; zero value is ModeAlways

// SetColorMode sets the global color emission policy. Goroutine-safe.
// All subsequent calls to ShouldColor honor the new mode.
func SetColorMode(m Mode) {
	colorMode.Store(int32(m))
}

// GetColorMode returns the current global color mode.
func GetColorMode() Mode {
	return Mode(colorMode.Load())
}

// envChecked guards a one-time read of the NO_COLOR / FORCE_COLOR env vars.
// Reading os.Getenv allocates, so the result is cached. Process-level state
// is fine: env vars do not change during a process's lifetime in any
// realistic SDK usage.
var (
	envCheckOnce sync.Once
	envForceOff  bool // true if NO_COLOR / GLACIER_NO_COLOR is set
	envForceOn   bool // true if FORCE_COLOR / GLACIER_FORCE_COLOR is set
)

func resolveColorEnv() (forceOff, forceOn bool) {
	envCheckOnce.Do(func() {
		envForceOff = os.Getenv("GLACIER_NO_COLOR") != "" || os.Getenv("NO_COLOR") != ""
		envForceOn = os.Getenv("GLACIER_FORCE_COLOR") != "" || os.Getenv("FORCE_COLOR") != ""
	})
	return envForceOff, envForceOn
}

// ShouldColor reports whether ANSI color should be emitted for w under the
// current color mode and environment.
//
// Decision rules, in order:
//  1. NO_COLOR or GLACIER_NO_COLOR set in env → false (industry convention)
//  2. ModeNever → false
//  3. FORCE_COLOR or GLACIER_FORCE_COLOR set in env → true
//  4. ModeAlways (default) → true
//  5. ModeAuto → true only if Capability(w).IsTTY and color is supported
//
// The env lookups happen once per process; subsequent calls hit a cached
// result so the hot path is allocation-free under ModeAuto.
func ShouldColor(w io.Writer) bool {
	off, on := resolveColorEnv()
	if off {
		return false
	}
	mode := GetColorMode()
	if mode == ModeNever {
		return false
	}
	if on {
		return true
	}
	if mode == ModeAlways {
		return true
	}
	// ModeAuto.
	caps := Capability(w)
	return caps.IsTTY && caps.SupportsColor != ColorNone
}
