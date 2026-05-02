// SPDX-License-Identifier: Apache-2.0

// Command cligen is the Glacier CLI code generator. It discovers all types
// in the target Go module that implement cli.Command (Run(ctx) error), parses
// their +glacier:* doc-comment markers, and emits zz_generated_cli.go.
//
// Usage:
//
//	go run github.com/nathanbrophy/glacier/cmd/cligen [--check] [--lint] [./...]
package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/nathanbrophy/glacier/cli/gen"
)

func main() {
	check := flag.Bool("check", false, "check mode: detect drift without writing")
	lint := flag.Bool("lint", false, "upgrade unknown marker warnings to errors")
	flag.Parse()

	pattern := "./..."
	if args := flag.Args(); len(args) > 0 {
		pattern = args[0]
	}

	if err := gen.Generate(gen.Options{
		Pattern: pattern,
		Check:   *check,
		Lint:    *lint,
		Logger:  slog.Default(),
	}); err != nil {
		slog.Error("cligen failed", "error", err)
		os.Exit(1)
	}
}
