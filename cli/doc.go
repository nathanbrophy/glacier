// SPDX-License-Identifier: Apache-2.0

// Package cli is a first-class, gold-standard CLI builder — designed for any
// consumer building a CLI in Go, not as plumbing for Glacier's own SDK binary.
// Delivered in two halves that fit together seamlessly: a runtime library that
// parses argv, dispatches to typed handlers, and renders banner/help/version
// output; and a companion codegen toolchain (cli/gen) that walks the
// consumer's module, identifies structs implementing the Command interface,
// parses kubebuilder-style +glacier: markers in their doc comments, and emits
// committed zz_generated_cli.go wiring files. The result: write a Go struct
// with a Run(ctx) error method plus comments-as-help-text plus markers for
// defaults/short/env — everything else is auto-wired. Full API in
// specs/0011-cli.md.
package cli
