// SPDX-License-Identifier: Apache-2.0

package gen_test

import (
	"github.com/nathanbrophy/glacier/cli/gen"
)

// ExampleGenerate shows a typical glaciergen invocation: pass a Go package
// pattern, and Generate scans it for cli.Command implementations carrying
// +glacier:* markers, then emits or refreshes the generated registration
// file (zz_generated_cli.go).
//
// In Check mode, Generate compares the buffered output to the file on disk
// and returns an error if they diverge: useful for CI gates.
func ExampleGenerate() {
	_ = gen.Generate(gen.Options{
		Pattern: "github.com/example/app/cmd/...",
		Check:   true,
	})
}
