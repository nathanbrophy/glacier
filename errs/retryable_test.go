// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/nathanbrophy/glacier/errs"
)

// retryableImpl is a custom error type that implements Retryable() bool.
type retryableImpl struct{ val bool }

func (r *retryableImpl) Error() string   { return "retryable-impl" }
func (r *retryableImpl) Retryable() bool { return r.val }
func (r *retryableImpl) Unwrap() error   { return nil }

// TestRetryable covers all Retryable() and MarkRetryable() behaviours.
func TestRetryable(t *testing.T) {
	t.Parallel()
	type tc struct {
		name string
		err  func() error // builder so each subtest gets a fresh value
		want bool
	}
	cases := []tc{
		{
			name: "MarkRetryable round-trip returns true",
			err:  func() error { return errs.MarkRetryable(io.EOF) },
			want: true,
		},
		{
			name: "plain error without marker returns false",
			err:  func() error { return io.EOF },
			want: false,
		},
		{
			name: "nil err returns false",
			err:  func() error { return nil },
			want: false,
		},
		{
			name: "custom type Retryable()=true detected without MarkRetryable",
			err:  func() error { return &retryableImpl{val: true} },
			want: true,
		},
		{
			name: "custom type Retryable()=false returns false",
			err:  func() error { return &retryableImpl{val: false} },
			want: false,
		},
		{
			name: "traverses fmt.Errorf wrap chain",
			err:  func() error { return fmt.Errorf("x: %w", errs.MarkRetryable(io.EOF)) },
			want: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := errs.Retryable(c.err())
			if got != c.want {
				t.Fatalf("Retryable() = %v, want %v", got, c.want)
			}
		})
	}
}

// TestMarkRetryableNil: MarkRetryable(nil) == nil.
func TestMarkRetryableNil(t *testing.T) {
	t.Parallel()
	if got := errs.MarkRetryable(nil); got != nil {
		t.Fatalf("MarkRetryable(nil) = %v, want nil", got)
	}
}

// TestRetryableMarkPreservesUnwrap: errors.Is(MarkRetryable(io.EOF), io.EOF) == true.
func TestRetryableMarkPreservesUnwrap(t *testing.T) {
	t.Parallel()
	marked := errs.MarkRetryable(io.EOF)
	if !errors.Is(marked, io.EOF) {
		t.Fatal("errors.Is(MarkRetryable(io.EOF), io.EOF) = false, want true")
	}
}

// TestRetryableNoRaceConcurrent: 100 goroutines call Retryable(sharedErr) concurrently.
func TestRetryableNoRaceConcurrent(t *testing.T) {
	shared := errs.MarkRetryable(errs.Wrap(io.EOF, "x"))
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = errs.Retryable(shared)
		}()
	}
	wg.Wait()
}
