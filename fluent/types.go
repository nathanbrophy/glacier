// SPDX-License-Identifier: Apache-2.0

package fluent

// KV is a key/value pair used by Pairs and Entries to bridge between
// Seq[KV[K, V]] and Seq2[K, V].
//
// The zero value is valid (K: zero, V: zero); nil pointers in K or V are
// the caller's concern.
type KV[K, V any] struct {
	K K
	V V
}

// Number is the type constraint for numeric types supported by Sum and Avg.
// It covers all integer and floating-point kinds via their underlying types.
// Exported so callers can declare their own Number-constrained generic
// functions that compose naturally with Sum and Avg.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}
