// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"io"
	"testing"

	"github.com/nathanbrophy/glacier/errs"
)

// codedError is a test error type that implements errs.Coded.
type codedError struct {
	msg  string
	code string
	next error
}

func (e *codedError) Error() string { return e.msg }
func (e *codedError) Code() string  { return e.code }
func (e *codedError) Unwrap() error { return e.next }

// TestCode covers all Code() extraction scenarios.
func TestCode(t *testing.T) {
	t.Parallel()
	type tc struct {
		name     string
		err      error
		wantCode string
	}
	cases := []tc{
		{
			name:     "extracts code from Coded error",
			err:      &codedError{msg: "coded: oops", code: "E_OOPS"},
			wantCode: "E_OOPS",
		},
		{
			name:     "returns empty string for non-Coded error",
			err:      io.EOF,
			wantCode: "",
		},
		{
			name:     "returns empty string for nil error",
			err:      nil,
			wantCode: "",
		},
		{
			name:     "outermost Coded wins in a chain (errors.As semantics)",
			err:      &codedError{msg: "outer", code: "E_OUTER", next: &codedError{msg: "inner", code: "E_INNER"}},
			wantCode: "E_OUTER",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := errs.Code(c.err)
			if got != c.wantCode {
				t.Fatalf("Code(%v) = %q, want %q", c.err, got, c.wantCode)
			}
		})
	}
}
