// SPDX-License-Identifier: Apache-2.0

// Package option is the single canonical functional-options framework for every
// Glacier package configurable at construction. It provides the generic
// Option[T] interface that every per-package WithX constructor returns, the
// OptionFunc[T] adapter so plain functions satisfy Option[T], and the Apply[T]
// function that package constructors call to fold options into a zero-valued
// config struct. Apply supports a Strict mode that collects all option errors
// rather than short-circuiting at the first, and a Validator[T] mechanism that
// package constructors use to assert post-apply correctness invariants.
// The package imports only stdlib; it is the universal kernel with no internal
// Glacier dependencies.
package option
