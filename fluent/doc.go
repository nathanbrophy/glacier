// SPDX-License-Identifier: Apache-2.0

// Package fluent is the Glacier answer to "I want to chain operations over a
// sequence the way LINQ does, but in idiomatic Go." Built on Go 1.23+
// iter.Seq[T] and iter.Seq2[K, V]. Lazy by construction — only sinks
// materialize. Top-level functions only: Go generics do not compose
// method-chained pipelines across T → U transitions cleanly, and explicit
// nested calls survive go doc. Covers the LINQ pattern set with Glacier
// discipline: source builders, lazy transformers, sorting, generic and
// specialized aggregations, set operations, joins, string and generator
// helpers, and fallible MapErr / FilterErr for I/O-shaped pipelines that
// surface per-element errors as Seq2[U, error].
package fluent
