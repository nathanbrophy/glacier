// SPDX-License-Identifier: Apache-2.0

// Package option provides the canonical functional-options framework
// used by every Glacier package configurable at construction.
//
// The package centers on a single generic interface, Option[T], which
// every per-package WithX constructor returns. A package's New
// constructor calls Apply to fold options into a zero-valued config
// struct, then optionally calls Validate to check invariants.
//
// Options may fail: Option's apply method returns an error. By default,
// Apply short-circuits at the first failing option; pass Strict() to
// apply every option and join all collected errors.
package option
