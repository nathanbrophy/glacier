// SPDX-License-Identifier: Apache-2.0

// Package gen is the codegen library for the Glacier mock package. It walks a
// consumer's Go module using go/packages, identifies interface declarations
// annotated with the +glacier:mock marker, and emits a committed
// zz_generated_mocks.go file containing typed wrapper structs. Each wrapper
// provides per-method OnFoo(...) expectation builders, giving full IDE
// autocomplete with compile-time argument types. The library is consumed by the
// glacier SDK's generate command and is importable independently for
// programmatic use. Full API in specs/0012-mock.md.
package gen
