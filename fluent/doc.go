// SPDX-License-Identifier: Apache-2.0

// Package fluent is Glacier's LINQ-equivalent: chainable, lazy, composable
// operators over Go 1.23+ iterators (iter.Seq[T] and iter.Seq2[K, V]).
//
// A fluent pipeline has three stages:
//
//	Source → Transformers → Sink
//
// Sources produce an iter.Seq[T] or iter.Seq2[K, V] and do no work until
// iterated. Transformers wrap a source and return a new sequence; they are
// lazy :  no element passes through until the sink pulls. Sinks consume the
// sequence and return a concrete value, triggering all upstream work.
//
// The sort family (Sort, SortBy, SortStable, SortDesc) is the only eager
// transformer group: it materializes its input once, sorts it, and returns
// a lazy Seq over the sorted slice.
//
// All functions in this package are goroutine-safe by value: each call
// returns a new closure. Zero external dependencies :  stdlib only.
//
// Spec: 0008-fluent.md
package fluent
