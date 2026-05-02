// SPDX-License-Identifier: Apache-2.0

package httpc

import (
	"fmt"

	"github.com/nathanbrophy/glacier/errs"
)

var (
	// ErrDryRun is returned by typed methods when WithDryRunErrors() is set and
	// the context carries a dry-run attribute. Callers check with errors.Is.
	ErrDryRun = errs.Sentinel("httpc: dry run")

	// ErrMaxAttempts is returned when the retry loop exhausts its attempt budget.
	// The final response error is wrapped via errs.Wrap.
	ErrMaxAttempts = errs.Sentinel("httpc: max attempts")

	// ErrMaxElapsed is returned when the retry loop exceeds its overall time budget.
	// The final response error is wrapped via errs.Wrap.
	ErrMaxElapsed = errs.Sentinel("httpc: max elapsed")

	// ErrBodyTooLarge is returned when the response body exceeds the configured
	// byte cap (default 32 MiB). Callers check with errors.Is.
	ErrBodyTooLarge = errs.Sentinel("httpc: response body too large")
)

// StatusError is returned by typed methods when the server responds with a
// non-2xx status code and the retry policy (if any) is exhausted.
type StatusError struct {
	// Status is the HTTP status code (e.g., 404, 500).
	Status int
	// Body holds the raw response body bytes. Accessible for explicit inspection;
	// NOT included in the Error() string.
	Body []byte
	// Cause wraps an underlying error, if present.
	Cause error
}

// Error implements the error interface. The string includes the status code
// but never includes body bytes.
func (e *StatusError) Error() string {
	return fmt.Sprintf("httpc: status %d", e.Status)
}

// Unwrap enables errors.Is / errors.As chaining on Cause.
func (e *StatusError) Unwrap() error { return e.Cause }

// BodyParseError is returned by typed methods when the response body cannot
// be decoded into T (e.g., malformed JSON, depth exceeded, UTF-8 violation).
// Named BodyParseError (not ParseError) to avoid collision with
// cli.FlagParseError.
type BodyParseError struct {
	// Cause is the underlying decode error. Never nil.
	Cause error
	// Body holds the first 1 KiB of the response body for debugging.
	// NOT included in the Error() string.
	Body []byte
	// ContentType is the response Content-Type header value.
	ContentType string
}

// Error implements the error interface. The string includes the ContentType
// and Cause message but never includes Body bytes.
func (e *BodyParseError) Error() string {
	return fmt.Sprintf("httpc: body parse (%s): %s", e.ContentType, e.Cause.Error())
}

// Unwrap enables errors.Is / errors.As chaining on Cause.
func (e *BodyParseError) Unwrap() error { return e.Cause }
