// SPDX-License-Identifier: Apache-2.0

package require

import (
	"cmp"
	"time"

	"github.com/nathanbrophy/glacier/assert"
)

// Equal reports whether got equals want using Glacier's smart-equal algorithm.
// If the assertion fails, t.FailNow() is called to halt the test immediately.
//
// §21.4 F22, E24; §23.17
func Equal[T any](t assert.TB, got, want T, opts ...assert.EqualOption) bool {
	t.Helper()
	ok := assert.Equal(t, got, want, opts...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// NotEqual reports whether got does not equal want.
// Halts the test on failure.
//
// §21.4 F22; §23.17
func NotEqual[T any](t assert.TB, got, want T, opts ...assert.EqualOption) bool {
	t.Helper()
	ok := assert.NotEqual(t, got, want, opts...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// True reports whether cond is true. Halts on failure.
//
// §21.4 F22
func True(t assert.TB, cond bool, msg ...any) bool {
	t.Helper()
	ok := assert.True(t, cond, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// False reports whether cond is false. Halts on failure.
//
// §21.4 F22
func False(t assert.TB, cond bool, msg ...any) bool {
	t.Helper()
	ok := assert.False(t, cond, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// Nil reports whether v is nil (typed-nil-aware). Halts on failure.
//
// §21.4 F22
func Nil(t assert.TB, v any, msg ...any) bool {
	t.Helper()
	ok := assert.Nil(t, v, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// NotNil reports whether v is non-nil. Halts on failure.
//
// §21.4 F22
func NotNil(t assert.TB, v any, msg ...any) bool {
	t.Helper()
	ok := assert.NotNil(t, v, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// NoError reports whether err is nil. Halts on failure.
//
// §21.4 F22
func NoError(t assert.TB, err error, msg ...any) bool {
	t.Helper()
	ok := assert.NoError(t, err, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// Error reports whether err is non-nil. Halts on failure.
//
// §21.4 F22
func Error(t assert.TB, err error, msg ...any) bool {
	t.Helper()
	ok := assert.Error(t, err, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// ErrorIs reports whether errors.Is(err, target) is true. Halts on failure.
//
// §21.4 F22
func ErrorIs(t assert.TB, err, target error, msg ...any) bool {
	t.Helper()
	ok := assert.ErrorIs(t, err, target, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// ErrorAs reports whether errors.As(err, target) is true. Halts on failure.
//
// §21.4 F22
func ErrorAs(t assert.TB, err error, target any, msg ...any) bool {
	t.Helper()
	ok := assert.ErrorAs(t, err, target, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// Contains reports whether haystack contains needle. Halts on failure.
//
// §21.4 F22
func Contains(t assert.TB, haystack, needle any, opts ...assert.EqualOption) bool {
	t.Helper()
	ok := assert.Contains(t, haystack, needle, opts...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// Len reports whether the length of container equals want. Halts on failure.
//
// §21.4 F22
func Len(t assert.TB, container any, want int, msg ...any) bool {
	t.Helper()
	ok := assert.Len(t, container, want, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// Eventually polls fn at interval until it returns true or timeout elapses.
// Halts on timeout failure.
//
// §21.4 F22
func Eventually(t assert.TB, fn func() bool, timeout, interval time.Duration, msg ...any) bool {
	t.Helper()
	ok := assert.Eventually(t, fn, timeout, interval, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// Match reports whether got matches pattern. Halts on failure.
//
// §21.4 F22
func Match(t assert.TB, got, pattern string, opts ...assert.MatchOption) bool {
	t.Helper()
	ok := assert.Match(t, got, pattern, opts...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// Greater reports whether got > threshold. Halts on failure.
//
// §21.4 F22; §23.17
func Greater[T cmp.Ordered](t assert.TB, got, threshold T, msg ...any) bool {
	t.Helper()
	ok := assert.Greater(t, got, threshold, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// Less reports whether got < threshold. Halts on failure.
//
// §21.4 F22; §23.17
func Less[T cmp.Ordered](t assert.TB, got, threshold T, msg ...any) bool {
	t.Helper()
	ok := assert.Less(t, got, threshold, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// GreaterOrEqual reports whether got >= threshold. Halts on failure.
//
// §21.4 F22; §23.17
func GreaterOrEqual[T cmp.Ordered](t assert.TB, got, threshold T, msg ...any) bool {
	t.Helper()
	ok := assert.GreaterOrEqual(t, got, threshold, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// LessOrEqual reports whether got <= threshold. Halts on failure.
//
// §21.4 F22; §23.17
func LessOrEqual[T cmp.Ordered](t assert.TB, got, threshold T, msg ...any) bool {
	t.Helper()
	ok := assert.LessOrEqual(t, got, threshold, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// InDelta reports whether |got - want| <= delta. Halts on failure.
//
// §21.4 F22; §23.17
func InDelta[T ~float32 | ~float64](t assert.TB, got, want, delta T, msg ...any) bool {
	t.Helper()
	ok := assert.InDelta(t, got, want, delta, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// JSONEq parses got and want as JSON and reports deep equality. Halts on failure.
//
// §21.4 F22
func JSONEq(t assert.TB, got, want []byte, opts ...assert.EqualOption) bool {
	t.Helper()
	ok := assert.JSONEq(t, got, want, opts...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// BytesEq reports whether got and want are byte-for-byte equal. Halts on failure.
//
// §21.4 F22
func BytesEq(t assert.TB, got, want []byte, msg ...any) bool {
	t.Helper()
	ok := assert.BytesEq(t, got, want, msg...)
	if !ok {
		t.FailNow()
	}
	return ok
}

// Subset reports whether every element of want appears in got. Halts on failure.
//
// §21.4 F22; §23.17
func Subset[T any](t assert.TB, got, want []T, opts ...assert.EqualOption) bool {
	t.Helper()
	ok := assert.Subset(t, got, want, opts...)
	if !ok {
		t.FailNow()
	}
	return ok
}
