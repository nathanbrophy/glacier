// SPDX-License-Identifier: Apache-2.0

// Package gen is the codegen library for the Glacier CLI builder. It walks a
// consumer's Go module using go/packages, identifies structs that implement
// the cli.Command interface, parses +glacier: kubebuilder-style markers from
// their doc comments, and emits committed zz_generated_cli.go wiring files
// that register commands with the cli.App. The library is consumed by the
// cmd/cligen binary and is importable independently for programmatic use.
// All marker parsing applies strict per-marker character allowlists and
// identifier validation per the security contract in specs/0002-framework-shape.md
// §23.8. Full API in specs/0011-cli.md.
package gen
