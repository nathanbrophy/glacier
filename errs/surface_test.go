// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/errs"
)

// TestSurfaceClosed_ErrsPackage verifies the public API surface of the errs package.
// Expected exports: Wrap, Wrapper (+ Error, Unwrap, WithStackTrace methods),
// StackOf, Join, Chain, Sentinel, IsAny, MarkRetryable, Retryable, Coded (interface), Code.
// This test exercises every exported symbol to confirm its existence and basic type.
func TestSurfaceClosed_ErrsPackage(t *testing.T) {
	// Wrap returns *Wrapper.
	var _ *errs.Wrapper = errs.Wrap(nil, "")

	// *Wrapper methods: Error, Unwrap, WithStackTrace.
	var w *errs.Wrapper
	var _ string = w.Error()
	var _ error = w.Unwrap()
	var _ *errs.Wrapper = w.WithStackTrace()

	// StackOf.
	var _ interface{} = errs.StackOf(nil) // returns []runtime.Frame or nil

	// Join.
	var _ error = errs.Join()

	// Chain.
	_ = errs.Chain(nil)

	// Sentinel.
	s := errs.Sentinel("surface: test")
	assert.NotNil(t, s, "Sentinel returned nil")

	// IsAny.
	var _ bool = errs.IsAny(nil)

	// MarkRetryable.
	var _ error = errs.MarkRetryable(nil)

	// Retryable.
	var _ bool = errs.Retryable(nil)

	// Coded interface :  confirm our codedError satisfies it.
	var _ errs.Coded = &codedError{msg: "x", code: "E_X"}

	// Code.
	var _ string = errs.Code(nil)
}
