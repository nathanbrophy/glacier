// SPDX-License-Identifier: Apache-2.0

package assert

import "errors"

// NoError reports whether err is nil. On failure reports the error value
// via t.Errorf. Preferred over Nil(t, err) because the failure message
// includes err.Error().
//
// Concurrency: goroutine-safe.
//
// §21.4 F3
func NoError(t TB, err error, msg ...any) bool {
	t.Helper()
	if err == nil {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("NoError failed: unexpected error: %s.%s", err.Error(), suffix)
	return false
}

// Error reports whether err is non-nil. On failure (err == nil) reports
// via t.Errorf.
//
// Concurrency: goroutine-safe.
//
// §21.4 F3
func Error(t TB, err error, msg ...any) bool {
	t.Helper()
	if err != nil {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("Error failed: expected an error but got nil.%s", suffix)
	return false
}

// ErrorIs reports whether errors.Is(err, target) is true. On failure
// reports the full error chain via t.Errorf.
//
// Preconditions: target is non-nil.
// Concurrency: goroutine-safe.
//
// §21.4 F3
func ErrorIs(t TB, err, target error, msg ...any) bool {
	t.Helper()
	if errors.Is(err, target) {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("ErrorIs failed: error chain does not include target.\n  err:    %v\n  target: %v.%s",
		err, target, suffix)
	return false
}

// ErrorAs reports whether errors.As(err, target) is true. target must be
// a non-nil pointer to either a type that implements error, or to any
// interface type. On failure reports via t.Errorf.
//
// Preconditions: target is a non-nil pointer per errors.As contract.
// Concurrency: goroutine-safe.
//
// §21.4 F3
func ErrorAs(t TB, err error, target any, msg ...any) bool {
	t.Helper()
	if errors.As(err, target) {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("ErrorAs failed: error chain does not match target type.\n  err: %v.%s", err, suffix)
	return false
}
