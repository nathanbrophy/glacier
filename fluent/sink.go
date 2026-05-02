// SPDX-License-Identifier: Apache-2.0

package fluent

import (
	"cmp"
	"iter"
)

// Reduce folds src into a single value R by applying f(accumulator, element)
// for each element, starting from zero.
func Reduce[T, R any](src iter.Seq[T], zero R, f func(R, T) R) R {
	acc := zero
	for v := range src {
		acc = f(acc, v)
	}
	return acc
}

// ToSlice collects src into a []T. Always returns a non-nil slice;
// returns an empty (non-nil) slice for an empty src.
func ToSlice[T any](src iter.Seq[T]) []T {
	s := make([]T, 0)
	for v := range src {
		s = append(s, v)
	}
	return s
}

// ToMap collects a Seq2[K, V] into a map[K]V. Always returns a non-nil map.
// If duplicate keys are present, the last value wins.
func ToMap[K comparable, V any](src iter.Seq2[K, V]) map[K]V {
	m := make(map[K]V)
	for k, v := range src {
		m[k] = v
	}
	return m
}

// Count returns the number of elements in src.
func Count[T any](src iter.Seq[T]) int {
	n := 0
	for range src {
		n++
	}
	return n
}

// First returns the first element of src and true, or the zero value and false
// if src is empty.
func First[T any](src iter.Seq[T]) (T, bool) {
	for v := range src {
		return v, true
	}
	var zero T
	return zero, false
}

// Last returns the last element of src and true, or the zero value and false
// if src is empty.
func Last[T any](src iter.Seq[T]) (T, bool) {
	var last T
	found := false
	for v := range src {
		last = v
		found = true
	}
	return last, found
}

// Any returns true if pred returns true for at least one element of src.
func Any[T any](src iter.Seq[T], pred func(T) bool) bool {
	for v := range src {
		if pred(v) {
			return true
		}
	}
	return false
}

// All returns true if pred returns true for every element of src.
// Returns true for an empty src.
func All[T any](src iter.Seq[T], pred func(T) bool) bool {
	for v := range src {
		if !pred(v) {
			return false
		}
	}
	return true
}

// Sum returns the arithmetic sum of all elements in src.
// Returns the zero value of T for an empty src.
func Sum[T Number](src iter.Seq[T]) T {
	var total T
	for v := range src {
		total += v
	}
	return total
}

// Avg returns the arithmetic mean of all elements in src as float64.
// Returns 0.0 for an empty src.
func Avg[T Number](src iter.Seq[T]) float64 {
	var total float64
	n := 0
	for v := range src {
		total += float64(v)
		n++
	}
	if n == 0 {
		return 0.0
	}
	return total / float64(n)
}

// Min returns the smallest element in src and true, or the zero value and false
// if src is empty.
func Min[T cmp.Ordered](src iter.Seq[T]) (T, bool) {
	var min T
	found := false
	for v := range src {
		if !found || v < min {
			min = v
			found = true
		}
	}
	return min, found
}

// Max returns the largest element in src and true, or the zero value and false
// if src is empty.
func Max[T cmp.Ordered](src iter.Seq[T]) (T, bool) {
	var max T
	found := false
	for v := range src {
		if !found || v > max {
			max = v
			found = true
		}
	}
	return max, found
}

// MinBy returns the element of src for which key returns the smallest value,
// and true. Returns the zero value and false if src is empty.
func MinBy[T any, K cmp.Ordered](src iter.Seq[T], key func(T) K) (T, bool) {
	var minElem T
	var minKey K
	found := false
	for v := range src {
		k := key(v)
		if !found || k < minKey {
			minElem = v
			minKey = k
			found = true
		}
	}
	return minElem, found
}

// MaxBy returns the element of src for which key returns the largest value,
// and true. Returns the zero value and false if src is empty.
func MaxBy[T any, K cmp.Ordered](src iter.Seq[T], key func(T) K) (T, bool) {
	var maxElem T
	var maxKey K
	found := false
	for v := range src {
		k := key(v)
		if !found || k > maxKey {
			maxElem = v
			maxKey = k
			found = true
		}
	}
	return maxElem, found
}
