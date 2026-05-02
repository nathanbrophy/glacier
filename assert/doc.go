// SPDX-License-Identifier: Apache-2.0

// Package assert is the Glacier assertion suite. It has three faces in two
// packages. The assert package provides test assertions that report failures
// via t.Errorf and return bool, letting tests stack multiple assertions and
// see all failures in one run. It includes a smart deep-equality comparator
// that goes well beyond reflect.DeepEqual (pointer-aware, map-order-insensitive,
// slice-order-optional, custom Equal() method honored), wildcard/regex matching,
// ordering and tolerance assertions, JSON-equivalent and subset checks, and a
// constellation of helpers (Nil, ErrorIs, Contains, Len, Eventually). The
// assert/require sub-package mirrors every assertion with t.Fatal semantics for
// the test-halt case. The Must, Must2, and Mustf runtime helpers panic on
// failure and are intended only for initialization-time invariants in non-test
// code.
package assert
