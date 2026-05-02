// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"io"
	"io/fs"
	"testing"

	"github.com/nathanbrophy/glacier/errs"
)

// TestIsAny covers all IsAny scenarios: match, no-match, zero-targets, nil-err,
// and wrapped chain traversal.
func TestIsAny(t *testing.T) {
	t.Parallel()
	type tc struct {
		name    string
		err     error
		targets []error
		want    bool
	}
	cases := []tc{
		{
			name:    "matches when target is in the list",
			err:     io.EOF,
			targets: []error{fs.ErrNotExist, io.EOF},
			want:    true,
		},
		{
			name:    "no match when target is absent",
			err:     io.EOF,
			targets: []error{fs.ErrNotExist},
			want:    false,
		},
		{
			name:    "zero targets always returns false",
			err:     io.EOF,
			targets: []error{},
			want:    false,
		},
		{
			name:    "nil err always returns false",
			err:     nil,
			targets: []error{io.EOF, fs.ErrNotExist},
			want:    false,
		},
		{
			name:    "traverses wrapped chain",
			err:     errs.Wrap(io.EOF, "x"),
			targets: []error{io.EOF},
			want:    true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := errs.IsAny(c.err, c.targets...)
			if got != c.want {
				t.Fatalf("IsAny(%v, %v) = %v, want %v", c.err, c.targets, got, c.want)
			}
		})
	}
}
