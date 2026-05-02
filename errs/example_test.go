// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"errors"
	"fmt"
	"io"

	"github.com/nathanbrophy/glacier/errs"
)

// ExampleWrap demonstrates chain-preserving wrapping with an opt-in stack trace.
func ExampleWrap() {
	err := errs.Wrap(io.EOF, "cli: read")
	fmt.Println(err)
	fmt.Println(errors.Is(err, io.EOF))
	// Output:
	// cli: read: EOF
	// true
}

// ExampleSentinel demonstrates declaring and using a library-register sentinel.
func ExampleSentinel() {
	var ErrCancelled = errs.Sentinel("cli: cancelled")
	wrapped := errs.Wrap(ErrCancelled, "runner: run")
	fmt.Println(errors.Is(wrapped, ErrCancelled))
	// Output:
	// true
}

// ExampleChain demonstrates walking the full error tree.
func ExampleChain() {
	a := errs.Sentinel("pkg: a")
	b := errs.Sentinel("pkg: b")
	joined := errors.Join(a, b)
	top := errs.Wrap(joined, "top: op")

	// Count only the sentinel leaves themselves (identity check).
	var found int
	for e := range errs.Chain(top) {
		if e == a || e == b {
			found++
		}
	}
	fmt.Println(found)
	// Output:
	// 2
}

// ExampleMarkRetryable demonstrates marking and checking a retryable error.
func ExampleMarkRetryable() {
	err := errs.MarkRetryable(errs.Wrap(io.EOF, "client: do"))
	fmt.Println(errs.Retryable(err))
	fmt.Println(errors.Is(err, io.EOF))
	// Output:
	// true
	// true
}

// ExampleIsAny demonstrates multi-target error matching.
func ExampleIsAny() {
	err := errs.Wrap(io.EOF, "store: read")
	fmt.Println(errs.IsAny(err, io.EOF))
	fmt.Println(errs.IsAny(err, errors.New("unrelated")))
	// Output:
	// true
	// false
}

// ExampleJoin demonstrates nil-dropping, single-survivor joining.
func ExampleJoin() {
	// All nil → nil.
	fmt.Println(errs.Join(nil, nil) == nil)

	// Single non-nil → returned directly.
	e := io.EOF
	got := errs.Join(nil, e, nil)
	fmt.Println(got == e)

	// Output:
	// true
	// true
}
