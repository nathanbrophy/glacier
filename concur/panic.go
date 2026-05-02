// SPDX-License-Identifier: Apache-2.0

package concur

import "fmt"

// PanicError wraps a value recovered from a panicking goroutine inside Group.
type PanicError struct {
	Value any // invariant: never nil
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("concur: panic in group goroutine: %v", e.Value)
}

func (e *PanicError) Unwrap() error {
	return fmt.Errorf("%s", e.Error())
}
