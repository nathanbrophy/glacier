// SPDX-License-Identifier: Apache-2.0

package fluent

import (
	"cmp"
	"iter"
	"slices"
)

// Sort returns a lazy sequence of elements from src sorted in ascending order.
// Materializes src once, sorts it, then yields elements lazily.
func Sort[T cmp.Ordered](src iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		s := slices.Collect(src)
		slices.Sort(s)
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// SortBy returns a lazy sequence of elements from src sorted ascending by
// the value returned by key.
func SortBy[T any, K cmp.Ordered](src iter.Seq[T], key func(T) K) iter.Seq[T] {
	return func(yield func(T) bool) {
		s := slices.Collect(src)
		slices.SortFunc(s, func(a, b T) int {
			return cmp.Compare(key(a), key(b))
		})
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// SortStable returns a lazy sequence of elements from src sorted by less,
// preserving the original order of equal elements.
// less must return a negative int when a < b, zero when a == b, positive when a > b.
func SortStable[T any](src iter.Seq[T], less func(a, b T) int) iter.Seq[T] {
	return func(yield func(T) bool) {
		s := slices.Collect(src)
		slices.SortStableFunc(s, less)
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// SortDesc returns a lazy sequence of elements from src sorted in descending order.
func SortDesc[T cmp.Ordered](src iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		s := slices.Collect(src)
		slices.SortFunc(s, func(a, b T) int {
			return cmp.Compare(b, a)
		})
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}
