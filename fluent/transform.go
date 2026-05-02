// SPDX-License-Identifier: Apache-2.0

package fluent

import "iter"

// Map returns a lazy sequence that applies f to each element of src.
func Map[T, U any](src iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range src {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// Filter returns a lazy sequence of elements from src for which pred returns true.
func Filter[T any](src iter.Seq[T], pred func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range src {
			if pred(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Take returns a lazy sequence of at most n elements from src.
func Take[T any](src iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		i := 0
		for v := range src {
			if i >= n {
				return
			}
			if !yield(v) {
				return
			}
			i++
		}
	}
}

// Drop returns a lazy sequence that skips the first n elements of src.
func Drop[T any](src iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		i := 0
		for v := range src {
			if i < n {
				i++
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

// Window returns a lazy sequence of overlapping slices of length size over src.
// Each yielded slice is a fresh copy; callers may mutate it safely.
// Panics with "fluent: Window: size must be positive" if size <= 0.
func Window[T any](src iter.Seq[T], size int) iter.Seq[[]T] {
	if size <= 0 {
		panic("fluent: Window: size must be positive")
	}
	return func(yield func([]T) bool) {
		buf := make([]T, 0, size)
		for v := range src {
			buf = append(buf, v)
			if len(buf) == size {
				cp := make([]T, size)
				copy(cp, buf)
				if !yield(cp) {
					return
				}
				// slide window by one
				buf = buf[1:]
			}
		}
	}
}

// Chunk returns a lazy sequence of non-overlapping slices of length size over src.
// The last chunk may be shorter than size. Each yielded slice is a fresh copy.
// Panics with "fluent: Chunk: size must be positive" if size <= 0.
func Chunk[T any](src iter.Seq[T], size int) iter.Seq[[]T] {
	if size <= 0 {
		panic("fluent: Chunk: size must be positive")
	}
	return func(yield func([]T) bool) {
		buf := make([]T, 0, size)
		for v := range src {
			buf = append(buf, v)
			if len(buf) == size {
				cp := make([]T, size)
				copy(cp, buf)
				if !yield(cp) {
					return
				}
				buf = buf[:0]
			}
		}
		if len(buf) > 0 {
			cp := make([]T, len(buf))
			copy(cp, buf)
			yield(cp)
		}
	}
}

// Distinct returns a lazy sequence with duplicate elements removed.
// First occurrence of each element is preserved; subsequent duplicates are dropped.
// Allocates a map[T]struct{} on first iteration.
func Distinct[T comparable](src iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for v := range src {
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			if !yield(v) {
				return
			}
		}
	}
}

// Zip returns a lazy Seq2 that pairs elements from a and b until either is exhausted.
func Zip[A, B any](a iter.Seq[A], b iter.Seq[B]) iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		nextB, stopB := iter.Pull(b)
		defer stopB()
		for va := range a {
			vb, ok := nextB()
			if !ok {
				return
			}
			if !yield(va, vb) {
				return
			}
		}
	}
}

// GroupBy materializes src and returns a Seq2 of (key, group) pairs.
// Internally eager: all elements are read before any pair is yielded.
// Order of groups follows first-occurrence order of each key.
func GroupBy[T any, K comparable](src iter.Seq[T], key func(T) K) iter.Seq2[K, []T] {
	return func(yield func(K, []T) bool) {
		order := make([]K, 0)
		groups := make(map[K][]T)
		for v := range src {
			k := key(v)
			if _, exists := groups[k]; !exists {
				order = append(order, k)
			}
			groups[k] = append(groups[k], v)
		}
		for _, k := range order {
			if !yield(k, groups[k]) {
				return
			}
		}
	}
}

// Join returns a lazy inner-join Seq2 of matching (A, B) pairs.
// Materializes b into a map[K][]B on first pull; elements of a with no match in b are dropped.
func Join[A, B any, K comparable](a iter.Seq[A], b iter.Seq[B], keyA func(A) K, keyB func(B) K) iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		bMap := make(map[K][]B)
		for vb := range b {
			k := keyB(vb)
			bMap[k] = append(bMap[k], vb)
		}
		for va := range a {
			k := keyA(va)
			for _, vb := range bMap[k] {
				if !yield(va, vb) {
					return
				}
			}
		}
	}
}

// LeftJoin returns a lazy left-join Seq2.
// Every element of a is yielded; if no matching element exists in b, the zero value of B is used.
// Materializes b into a map[K][]B on first pull.
func LeftJoin[A, B any, K comparable](a iter.Seq[A], b iter.Seq[B], keyA func(A) K, keyB func(B) K) iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		bMap := make(map[K][]B)
		for vb := range b {
			k := keyB(vb)
			bMap[k] = append(bMap[k], vb)
		}
		for va := range a {
			k := keyA(va)
			matches := bMap[k]
			if len(matches) == 0 {
				var zero B
				if !yield(va, zero) {
					return
				}
			} else {
				for _, vb := range matches {
					if !yield(va, vb) {
						return
					}
				}
			}
		}
	}
}

// Union returns a lazy sequence of distinct elements from a followed by distinct elements of b
// that did not appear in a. Allocates a map[T]struct{} on first iteration.
func Union[T comparable](a, b iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for v := range a {
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			if !yield(v) {
				return
			}
		}
		for v := range b {
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			if !yield(v) {
				return
			}
		}
	}
}

// Intersect returns a lazy sequence of elements that appear in both a and b.
// Materializes b into a set on first iteration; preserves a's order.
func Intersect[T comparable](a, b iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		bSet := make(map[T]struct{})
		for v := range b {
			bSet[v] = struct{}{}
		}
		seen := make(map[T]struct{})
		for v := range a {
			if _, inB := bSet[v]; !inB {
				continue
			}
			if _, already := seen[v]; already {
				continue
			}
			seen[v] = struct{}{}
			if !yield(v) {
				return
			}
		}
	}
}

// Except returns a lazy sequence of elements in a that are not in b.
// Materializes b into a set on first iteration; preserves a's order.
func Except[T comparable](a, b iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		bSet := make(map[T]struct{})
		for v := range b {
			bSet[v] = struct{}{}
		}
		for v := range a {
			if _, inB := bSet[v]; inB {
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}
