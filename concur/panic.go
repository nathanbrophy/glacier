// SPDX-License-Identifier: Apache-2.0

package concur

import "fmt"

// PanicError wraps a value recovered from a panicking goroutine inside Group.
type PanicError struct {
	Value any // invariant: never nil
}

// Error implements error.
func (e *PanicError) Error() string {
	return fmt.Sprintf("concur: panic in group goroutine: %v", e.Value)
}

// Unwrap returns a wrapper around Error so errors.Is/errors.As can
// traverse, even though PanicError holds an arbitrary recovered value
// rather than a real error chain.
func (e *PanicError) Unwrap() error {
	return fmt.Errorf("%s", e.Error())
}
