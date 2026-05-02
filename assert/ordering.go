// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"cmp"
	"math"
)

// Greater reports whether got > threshold. T is constrained to cmp.Ordered
// (integers, floats, strings). On failure reports both values via t.Errorf.
//
// Preconditions: t is non-nil.
// Concurrency: goroutine-safe.
//
// §21.4 F5
func Greater[T cmp.Ordered](t TB, got, threshold T, msg ...any) bool {
	t.Helper()
	if got > threshold {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("Greater failed: got=%v is not > threshold=%v.%s", got, threshold, suffix)
	return false
}

// Less reports whether got < threshold.
//
// Concurrency: goroutine-safe.
//
// §21.4 F5
func Less[T cmp.Ordered](t TB, got, threshold T, msg ...any) bool {
	t.Helper()
	if got < threshold {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("Less failed: got=%v is not < threshold=%v.%s", got, threshold, suffix)
	return false
}

// GreaterOrEqual reports whether got >= threshold.
//
// Concurrency: goroutine-safe.
//
// §21.4 F5
func GreaterOrEqual[T cmp.Ordered](t TB, got, threshold T, msg ...any) bool {
	t.Helper()
	if got >= threshold {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("GreaterOrEqual failed: got=%v is not >= threshold=%v.%s", got, threshold, suffix)
	return false
}

// LessOrEqual reports whether got <= threshold.
//
// Concurrency: goroutine-safe.
//
// §21.4 F5
func LessOrEqual[T cmp.Ordered](t TB, got, threshold T, msg ...any) bool {
	t.Helper()
	if got <= threshold {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("LessOrEqual failed: got=%v is not <= threshold=%v.%s", got, threshold, suffix)
	return false
}

// InDelta reports whether |got - want| <= delta. T is constrained to
// ~float32 | ~float64, including user-defined float types. NaN values are
// never within any delta (NaN != NaN per Go semantics).
//
// Preconditions: delta >= 0. Negative delta: reports via t.Errorf and
// returns false.
// Concurrency: goroutine-safe.
//
// §21.4 F6, E13
func InDelta[T ~float32 | ~float64](t TB, got, want, delta T, msg ...any) bool {
	t.Helper()
	if delta < 0 {
		t.Errorf("InDelta failed: delta must be >= 0; got delta=%v.", delta)
		return false
	}
	gf := float64(got)
	wf := float64(want)
	df := float64(delta)
	if math.IsNaN(gf) || math.IsNaN(wf) {
		suffix := fmtMsg(msg)
		t.Errorf("InDelta failed: NaN values are never equal.%s", suffix)
		return false
	}
	if math.Abs(gf-wf) <= df {
		return true
	}
	suffix := fmtMsg(msg)
	t.Errorf("InDelta failed: |got - want| = %v > delta = %v (got=%v, want=%v).%s",
		math.Abs(gf-wf), df, got, want, suffix)
	return false
}
