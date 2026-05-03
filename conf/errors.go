// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"fmt"

	"github.com/nathanbrophy/glacier/errs"
)

// ErrLayerConflict is returned when incompatible types are found for the same
// field across configuration layers.
var ErrLayerConflict = errs.Sentinel("conf: layer conflict: incompatible types for field")

// ErrFileTooLarge is returned when a JSON configuration file exceeds 1 MiB.
var ErrFileTooLarge = errs.Sentinel("conf: file too large: maximum size is 1 mib")

// ErrDepthExceeded is returned when JSON configuration nesting exceeds 32 levels.
var ErrDepthExceeded = errs.Sentinel("conf: depth exceeded: maximum nesting depth is 32")

// ErrLoaderClosed is returned when Load is called on a closed Loader.
var ErrLoaderClosed = errs.Sentinel("conf: loader closed")

// DecodeError is returned for all configuration decode failures.
type DecodeError struct {
	Path  string // dot-separated field path; "" for top-level failures
	Cause error  // underlying error; always non-nil
	Layer string // source layer label; informational
}

// Error implements error.
func (e *DecodeError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("conf: decode: %v", e.Cause)
	}
	return fmt.Sprintf("conf: decode %s: %v", e.Path, e.Cause)
}

// Unwrap allows errors.Is and errors.As to traverse through DecodeError.
func (e *DecodeError) Unwrap() error { return e.Cause }
