// SPDX-License-Identifier: Apache-2.0

package mock

import "time"

// Call records one invocation of a mocked method.
//
// invariant: Args is a defensive copy; mutations by the caller after the call
// do not affect the recorded slice.
// invariant: At is set once at call time and never mutated.
// invariant: Matched is true iff an Expectation was found for this call.
type Call struct {
	// Method is the name of the called interface method.
	Method string
	// Args is a defensive copy of the arguments at call time.
	Args []any
	// At is the time the call was recorded (monotonic clock).
	At time.Time
	// Matched reports whether an Expectation was found for this call.
	Matched bool
}
