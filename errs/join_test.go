// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"errors"
	"io"
	"io/fs"
	"testing"

	"github.com/nathanbrophy/glacier/errs"
)

// TestJoinNilAndEmptyCases covers Join with zero or all-nil inputs.
func TestJoinNilAndEmptyCases(t *testing.T) {
	t.Parallel()
	type tc struct {
		name string
		args []error
	}
	cases := []tc{
		{name: "zero args returns nil", args: []error{}},
		{name: "all nil args returns nil", args: []error{nil, nil}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := errs.Join(c.args...); got != nil {
				t.Fatalf("Join(%v) = %v, want nil", c.args, got)
			}
		})
	}
}

// TestJoinSingleNonNilCollapses: Join(nil, e, nil) == e (identity-equal).
func TestJoinSingleNonNilCollapses(t *testing.T) {
	t.Parallel()
	e := io.EOF
	got := errs.Join(nil, e, nil)
	if got != e {
		t.Fatalf("Join(nil,e,nil) = %v, want identity-equal to e", got)
	}
}

// TestJoinMultipleNonNil covers the two-or-more-error cases: errors.Is semantics
// and Unwrap() []error implementation.
func TestJoinMultipleNonNil(t *testing.T) {
	t.Parallel()
	type tc struct {
		name      string
		args      []error
		wantCount int // expected len of Unwrap() []error
		targets   []error
	}
	cases := []tc{
		{
			name:      "two non-nil errors — both reachable via errors.Is",
			args:      []error{io.EOF, fs.ErrNotExist},
			wantCount: 2,
			targets:   []error{io.EOF, fs.ErrNotExist},
		},
		{
			name:      "nil dropped before stdlib call — two-error join from three args",
			args:      []error{io.EOF, nil, fs.ErrNotExist},
			wantCount: 2,
			targets:   []error{io.EOF, fs.ErrNotExist},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			j := errs.Join(c.args...)
			if j == nil {
				t.Fatal("Join returned nil for non-nil inputs")
			}
			for _, tgt := range c.targets {
				if !errors.Is(j, tgt) {
					t.Errorf("errors.Is(j, %v) = false, want true", tgt)
				}
			}
			type multiUnwrapper interface{ Unwrap() []error }
			mu, ok := j.(multiUnwrapper)
			if !ok {
				t.Fatal("result does not implement Unwrap() []error")
			}
			if got := len(mu.Unwrap()); got != c.wantCount {
				t.Fatalf("Unwrap() []error len = %d, want %d", got, c.wantCount)
			}
		})
	}
}
