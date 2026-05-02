// SPDX-License-Identifier: Apache-2.0

package httpmock

import (
	"fmt"

	"github.com/nathanbrophy/glacier/errs"
)

// ErrNoRouteMatch is returned by RoundTrip when strict mode is active and
// no registered stub matches the incoming request.
var ErrNoRouteMatch = errs.Sentinel("httpmock: no route match")

// ScriptError is the typed error returned by RoundTrip when a matched stub
// has an invalid configuration (e.g., missing Responder).
// Library register format: "httpmock: script step <n>: <cause>".
type ScriptError struct {
	// Step is the zero-based registration index of the misconfigured stub.
	Step int
	// Cause is the underlying configuration error.
	Cause error
}

// Error implements error.
func (e *ScriptError) Error() string {
	return fmt.Sprintf("httpmock: script step %d: %s", e.Step, e.Cause)
}

// Unwrap returns the underlying cause.
func (e *ScriptError) Unwrap() error { return e.Cause }
