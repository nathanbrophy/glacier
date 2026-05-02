// SPDX-License-Identifier: Apache-2.0

package fluent

import "iter"

// Map2 returns a lazy Seq2 that applies f to each (K, V) pair of src,
// producing a new (K2, V2) pair.
func Map2[K, V, K2, V2 any](src iter.Seq2[K, V], f func(K, V) (K2, V2)) iter.Seq2[K2, V2] {
	return func(yield func(K2, V2) bool) {
		for k, v := range src {
			k2, v2 := f(k, v)
			if !yield(k2, v2) {
				return
			}
		}
	}
}

// Filter2 returns a lazy Seq2 of pairs from src for which pred returns true.
func Filter2[K, V any](src iter.Seq2[K, V], pred func(K, V) bool) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range src {
			if pred(k, v) {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// KeysOf returns a lazy Seq of the keys from src.
func KeysOf[K, V any](src iter.Seq2[K, V]) iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range src {
			if !yield(k) {
				return
			}
		}
	}
}

// ValuesOf returns a lazy Seq of the values from src.
func ValuesOf[K, V any](src iter.Seq2[K, V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range src {
			if !yield(v) {
				return
			}
		}
	}
}

// Entries converts a Seq2[K, V] into a Seq[KV[K, V]].
// Entries is the inverse of Pairs: Entries(Pairs(seq)) round-trips.
func Entries[K, V any](src iter.Seq2[K, V]) iter.Seq[KV[K, V]] {
	return func(yield func(KV[K, V]) bool) {
		for k, v := range src {
			if !yield(KV[K, V]{K: k, V: v}) {
				return
			}
		}
	}
}
