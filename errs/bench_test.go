// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"errors"
	"io"
	"io/fs"
	"testing"

	"github.com/nathanbrophy/glacier/errs"
)

// BenchmarkWrapNoStack: Wrap(io.EOF, "x") ≤ 1 alloc/op.
func BenchmarkWrapNoStack(b *testing.B) {
	b.ReportAllocs()
	var sink error
	for range b.N {
		sink = errs.Wrap(io.EOF, "x")
	}
	_ = sink
	// Validate alloc target outside benchmark loop.
	allocs := testing.AllocsPerRun(100, func() {
		sink = errs.Wrap(io.EOF, "x")
	})
	_ = sink
	if allocs > 1 {
		b.Errorf("BenchmarkWrapNoStack: %.0f allocs/op, want ≤ 1", allocs)
	}
}

// BenchmarkWrapWithStackTrace: records ns/op and allocs; documents the cost.
func BenchmarkWrapWithStackTrace(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = errs.Wrap(io.EOF, "x").WithStackTrace()
	}
}

// BenchmarkChainLinear: walk a 10-deep linear chain.
func BenchmarkChainLinear(b *testing.B) {
	var err error = io.EOF
	for range 9 {
		err = errs.Wrap(err, "layer")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		for range errs.Chain(err) {
		}
	}
}

// BenchmarkChainOverJoin: walk a 1-level join with 10 children.
func BenchmarkChainOverJoin(b *testing.B) {
	children := make([]error, 10)
	for i := range children {
		children[i] = errors.New("child")
	}
	j := errors.Join(children...)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		for range errs.Chain(j) {
		}
	}
}

// BenchmarkSentinel: one-time construction cost.
func BenchmarkSentinel(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = errs.Sentinel("pkg: benchmark")
	}
}

// BenchmarkJoinSingleCollapse: Join(nil, e, nil) → 0 alloc (returns e directly).
func BenchmarkJoinSingleCollapse(b *testing.B) {
	e := io.EOF
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = errs.Join(nil, e, nil)
	}
	allocs := testing.AllocsPerRun(100, func() {
		_ = errs.Join(nil, e, nil)
	})
	if allocs != 0 {
		b.Errorf("BenchmarkJoinSingleCollapse: %.0f allocs/op, want 0", allocs)
	}
}

// BenchmarkJoinMultiple: Join(e1, e2, e3) allocs equivalent to errors.Join.
func BenchmarkJoinMultiple(b *testing.B) {
	e1 := io.EOF
	e2 := fs.ErrNotExist
	e3 := fs.ErrPermission
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = errs.Join(e1, e2, e3)
	}
}

// BenchmarkRetryableWalk: 5-deep chain, last link retryable.
func BenchmarkRetryableWalk(b *testing.B) {
	var err error = errs.MarkRetryable(io.EOF)
	for range 4 {
		err = errs.Wrap(err, "layer")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = errs.Retryable(err)
	}
}

// BenchmarkCode: 5-deep chain, last link Coded.
func BenchmarkCode(b *testing.B) {
	inner := &codedError{msg: "inner", code: "E_CODE"}
	var err error = inner
	for range 4 {
		err = errs.Wrap(err, "layer")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = errs.Code(err)
	}
}
