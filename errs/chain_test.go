// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"errors"
	"io"
	"io/fs"
	"sync"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/errs"
)

// collectChain collects all errors from errs.Chain(err) into a slice.
func collectChain(err error) []error {
	var out []error
	for e := range errs.Chain(err) {
		out = append(out, e)
	}
	return out
}

// TestChainYieldCount verifies Chain yields the expected number of errors and,
// where applicable, that specific positional elements match.
func TestChainYieldCount(t *testing.T) {
	t.Parallel()

	a := io.EOF
	b := fs.ErrNotExist
	c := fs.ErrPermission

	inner := errs.Wrap(a, "a")
	outer := errs.Wrap(inner, "b")

	joinABC := errors.Join(a, b, c)
	innerJoin := errors.Join(b, c)
	outerJoin := errors.Join(a, innerJoin)

	type tc struct {
		name      string
		err       error
		wantLen   int
		wantFirst error // nil means skip
		wantElems []struct {
			idx int
			val error
		}
	}
	cases := []tc{
		{
			name:      "nil err yields nothing",
			err:       nil,
			wantLen:   0,
			wantFirst: nil,
		},
		{
			name:      "single error yields one element equal to itself",
			err:       a,
			wantLen:   1,
			wantFirst: a,
		},
		{
			name:    "linear Wrap — Wrap(Wrap(e,a),b) yields 3",
			err:     outer,
			wantLen: 3,
			wantElems: []struct {
				idx int
				val error
			}{{0, outer}},
		},
		{
			name:    "errors.Join(a,b,c) yields 4: join then a,b,c",
			err:     joinABC,
			wantLen: 4,
			wantElems: []struct {
				idx int
				val error
			}{{0, joinABC}, {1, a}, {2, b}, {3, c}},
		},
		{
			name:    "nested Join DFS — outer,a,inner,b,c",
			err:     outerJoin,
			wantLen: 5,
			wantElems: []struct {
				idx int
				val error
			}{{0, outerJoin}, {1, a}, {2, innerJoin}, {3, b}, {4, c}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := collectChain(c.err)
			require.Len(t, got, c.wantLen)
			for _, elem := range c.wantElems {
				assert.Equal(t, got[elem.idx], elem.val)
			}
		})
	}
}

// TestChainEarlyTermination: receiver returns false after 2 yields → exactly 2 seen.
func TestChainEarlyTermination(t *testing.T) {
	t.Parallel()
	a := io.EOF
	b := fs.ErrNotExist
	c := fs.ErrPermission
	j := errors.Join(a, b, c)
	count := 0
	for range errs.Chain(j) {
		count++
		if count == 2 {
			break
		}
	}
	assert.Equal(t, count, 2)
}

// TestChainComposesWithFluentTake: simulate bounded iteration via hand-rolled loop.
func TestChainComposesWithFluentTake(t *testing.T) {
	t.Parallel()
	// Build a 10-deep linear chain.
	var err error = io.EOF
	for i := range 9 {
		_ = i
		err = errs.Wrap(err, "layer")
	}
	count := 0
	for range errs.Chain(err) {
		count++
		if count == 5 {
			break
		}
	}
	assert.Equal(t, count, 5)
}

// TestChainNoRaceConcurrent: 100 goroutines iterate Chain(sharedErr) concurrently.
func TestChainNoRaceConcurrent(t *testing.T) {
	shared := errs.Wrap(errs.Wrap(io.EOF, "a"), "b")
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range errs.Chain(shared) {
			}
		}()
	}
	wg.Wait()
}
