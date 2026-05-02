// SPDX-License-Identifier: Apache-2.0

// Command cligen is the Glacier CLI codegen tool. It walks a Go module using
// go/packages, identifies structs that implement the cli.Command interface,
// parses +glacier: markers from their doc comments, and emits committed
// zz_generated_cli.go wiring files. Run via go run ./cmd/cligen or install
// as a standalone binary.
//
// Usage:
//
//	cligen [--check] <pattern>...
//
// With --check, cligen verifies that generated files are up to date and exits
// non-zero if any drift is detected (used by the codegen-drift CI gate).
package main

import (
	// TODO: implement; see specs/0011-cli.md
	_ "github.com/nathanbrophy/glacier/cli/gen"
)

func main() {
	// TODO: implement; see specs/0011-cli.md
}
