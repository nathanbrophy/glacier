// SPDX-License-Identifier: Apache-2.0

package errs_test

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"testing"

	"github.com/nathanbrophy/glacier/errs"
)

// buildErrorTree returns a small set of representative error structures
// for property testing.
func buildErrorTrees() []error {
	leaf := io.EOF
	single := errs.Wrap(leaf, "single: wrap")
	double := errs.Wrap(single, "double: wrap")
	joined := errs.Join(io.EOF, fs.ErrNotExist)
	mixed := errs.Wrap(errs.Join(io.EOF, fs.ErrNotExist), "mixed: wrap")
	return []error{leaf, single, double, joined, mixed}
}

// reachableSet returns the set of all errors reachable from err
// via Unwrap() error and Unwrap() []error, using a simple BFS.
func reachableSet(err error) map[error]struct{} {
	type multiUnwrapper interface {
		Unwrap() []error
	}
	seen := map[error]struct{}{}
	queue := []error{err}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == nil {
			continue
		}
		if _, ok := seen[cur]; ok {
			continue
		}
		seen[cur] = struct{}{}
		if mu, ok := cur.(multiUnwrapper); ok {
			queue = append(queue, mu.Unwrap()...)
		} else if next := errors.Unwrap(cur); next != nil {
			queue = append(queue, next)
		}
	}
	return seen
}

// PropertyChainStartsWithSelf: for any non-nil err, the first yield of
// Chain(err) is err itself.
func PropertyChainStartsWithSelf(t *testing.T) {
	for _, err := range buildErrorTrees() {
		first := true
		for e := range errs.Chain(err) {
			if first {
				if e != err {
					t.Errorf("Chain(%v): first element = %v, want err itself", err, e)
				}
				first = false
			}
			break
		}
	}
}

// TestPropertyChainStartsWithSelf is the test entry point.
func TestPropertyChainStartsWithSelf(t *testing.T) {
	PropertyChainStartsWithSelf(t)
}

// TestPropertyChainContainsAllUnwrapped: Chain(err) contains every error
// reachable from err via Unwrap.
func TestPropertyChainContainsAllUnwrapped(t *testing.T) {
	for _, root := range buildErrorTrees() {
		want := reachableSet(root)
		got := map[error]struct{}{}
		for e := range errs.Chain(root) {
			got[e] = struct{}{}
		}
		for e := range want {
			if _, ok := got[e]; !ok {
				t.Errorf("Chain(%v) missing reachable error %v", root, e)
			}
		}
	}
}

// TestPropertyJoinIdempotent: Join(Join(a, b), c) and Join(a, b, c) are
// semantically equivalent under errs.Chain (same set of leaf errors).
func TestPropertyJoinIdempotent(t *testing.T) {
	a := io.EOF
	b := fs.ErrNotExist
	c := fs.ErrPermission

	j1 := errs.Join(errs.Join(a, b), c)
	j2 := errs.Join(a, b, c)

	leafErrors := func(err error) map[error]struct{} {
		// Collect leaf sentinel errors from chain (skip wrappers/joins themselves).
		set := map[error]struct{}{}
		for e := range errs.Chain(err) {
			if e == a || e == b || e == c {
				set[e] = struct{}{}
			}
		}
		return set
	}

	s1 := leafErrors(j1)
	s2 := leafErrors(j2)
	for e := range s1 {
		if _, ok := s2[e]; !ok {
			t.Errorf("j2 missing error %v that j1 contains", e)
		}
	}
	for e := range s2 {
		if _, ok := s1[e]; !ok {
			t.Errorf("j1 missing error %v that j2 contains", e)
		}
	}
}

// TestPropertyWrapTransparentToErrorsIs: for any sentinel s and prefix p,
// errors.Is(Wrap(s, p), s) == true.
func TestPropertyWrapTransparentToErrorsIs(t *testing.T) {
	sentinels := []error{io.EOF, fs.ErrNotExist, fs.ErrPermission}
	prefixes := []string{"", "x", "pkg: action", "a: b: c", "long prefix string"}
	for _, s := range sentinels {
		for _, p := range prefixes {
			w := errs.Wrap(s, p)
			if !errors.Is(w, s) {
				t.Errorf("errors.Is(Wrap(%v, %q), %v) = false, want true", s, p, s)
			}
		}
	}
}

// TestPropertyMarkRetryableTransparent: for any err,
// errors.Is(MarkRetryable(err), err) == true.
func TestPropertyMarkRetryableTransparent(t *testing.T) {
	errsToTest := []error{io.EOF, fs.ErrNotExist, fmt.Errorf("custom")}
	for _, e := range errsToTest {
		marked := errs.MarkRetryable(e)
		if !errors.Is(marked, e) {
			t.Errorf("errors.Is(MarkRetryable(%v), %v) = false, want true", e, e)
		}
	}
}
