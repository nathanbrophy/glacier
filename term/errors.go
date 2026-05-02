// SPDX-License-Identifier: Apache-2.0

package term

import "github.com/nathanbrophy/glacier/errs"

// Sentinel errors returned by term functions. All match the library-register
// format: lowercase, no trailing period, prefix "term: ".
var (
	// ErrCancelled is returned when the user cancels a prompt (Ctrl-C / EOF)
	// or a context is cancelled.
	ErrCancelled = errs.Sentinel("term: cancelled")

	// ErrNotInteractive is returned when a prompt that requires a TTY is called
	// on a non-TTY writer (e.g. piped stdin).
	ErrNotInteractive = errs.Sentinel("term: not interactive")

	// ErrTimeout is returned when a WithTimeout deadline is exceeded.
	ErrTimeout = errs.Sentinel("term: timeout")

	// ErrTooManyAttempts is returned when WithMaxAttempts is set and the
	// validator rejects input that many times.
	ErrTooManyAttempts = errs.Sentinel("term: too many attempts")
)

// HexParseError is returned by Hex when the input is not a valid CSS hex color.
// Error() matches ^term: hex: .+$ per the library register.
type HexParseError struct {
	// Input is the string that failed to parse.
	Input string
	// Cause is the underlying parse failure.
	Cause error
}

// Error implements error.
func (e *HexParseError) Error() string { return "term: hex: " + e.Cause.Error() }

// Unwrap returns the underlying cause, enabling errors.Is / errors.As traversal.
func (e *HexParseError) Unwrap() error { return e.Cause }

// GlyphError is returned by RegisterGlyph for constraint violations.
// Error() matches ^term: glyph: .+$ per the library register.
type GlyphError struct {
	// Name is the glyph name that triggered the error.
	Name string
	// Cause describes the constraint violation.
	Cause string
}

// Error implements error.
func (e *GlyphError) Error() string { return "term: glyph: " + e.Cause }
