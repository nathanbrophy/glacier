// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"errors"
	"fmt"
	"testing"
)

// §21.4 F3

type testError struct{ code int }

func (e *testError) Error() string { return fmt.Sprintf("testError(%d)", e.code) }

var errSentinel = errors.New("sentinel")
var errWrapped = fmt.Errorf("wrapped: %w", errSentinel)

func TestNoError(t *testing.T) {
	mt := &mockTB{}
	True(t, NoError(mt, nil), "NoError(nil) = true")
	Equal(t, mt.errorfCalls, 0)
	mt.reset()
	False(t, NoError(mt, errSentinel), "NoError(err) = false")
	Equal(t, mt.errorfCalls, 1)
}

func TestError(t *testing.T) {
	mt := &mockTB{}
	True(t, Error(mt, errSentinel), "Error(err) = true")
	Equal(t, mt.errorfCalls, 0)
	mt.reset()
	False(t, Error(mt, nil), "Error(nil) = false")
	Equal(t, mt.errorfCalls, 1)
}

func TestErrorIs(t *testing.T) {
	mt := &mockTB{}
	True(t, ErrorIs(mt, errWrapped, errSentinel), "ErrorIs walks chain")
	Equal(t, mt.errorfCalls, 0)
	mt.reset()
	False(t, ErrorIs(mt, errors.New("other"), errSentinel), "ErrorIs mismatch")
	Equal(t, mt.errorfCalls, 1)
}

func TestErrorAs(t *testing.T) {
	errMy := &testError{code: 42}
	wrapped := fmt.Errorf("wrap: %w", errMy)

	mt := &mockTB{}
	var target *testError
	True(t, ErrorAs(mt, wrapped, &target), "ErrorAs: finds target in chain")
	Equal(t, mt.errorfCalls, 0)
	Equal(t, target.code, 42)

	mt.reset()
	target = nil
	False(t, ErrorAs(mt, errors.New("other"), &target), "ErrorAs: not found")
	Equal(t, mt.errorfCalls, 1)
}
