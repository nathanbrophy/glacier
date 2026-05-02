// SPDX-License-Identifier: Apache-2.0

package fluent

import "iter"

// MapErr applies f to each element of src and yields (result, nil) on success
// or (zero, err) on failure. The caller may break the loop to short-circuit
// on the first error.
func MapErr[T, U any](src iter.Seq[T], f func(T) (U, error)) iter.Seq2[U, error] {
	return func(yield func(U, error) bool) {
		for v := range src {
			u, err := f(v)
			if !yield(u, err) {
				return
			}
		}
	}
}

// FilterErr applies pred to each element of src.
// If pred returns (true, nil), the element is yielded as (elem, nil).
// If pred returns (false, nil), the element is skipped.
// If pred returns (_, err), the element is yielded as (elem, err).
// The caller may break the loop to short-circuit on the first error.
func FilterErr[T any](src iter.Seq[T], pred func(T) (bool, error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for v := range src {
			ok, err := pred(v)
			if err != nil {
				if !yield(v, err) {
					return
				}
				continue
			}
			if ok {
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}
