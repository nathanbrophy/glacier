// SPDX-License-Identifier: Apache-2.0

// explaingen is the build-time tool that renders explain topic files from
// accepted spec sections. Run it from the repo root:
//
//	go run ./cmd/glacier/internal/explaingen/cmd [--check] [--root <repo-root>]
//
// Without --check the tool writes topic files under
// cmd/glacier/internal/explain/topics/. With --check it exits non-zero if any
// on-disk topic file would change (CI freshness gate X17).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nathanbrophy/glacier/cmd/glacier/internal/explaingen"
	"github.com/nathanbrophy/glacier/internal/safefile"
)

func main() {
	check := flag.Bool("check", false, "check freshness; exit non-zero if any topic would change")
	root := flag.String("root", ".", "repo root directory (default: current directory)")
	flag.Parse()

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "explaingen: resolve root: %v\n", err)
		os.Exit(1)
	}

	src := explaingen.NewFileSpecSource(os.DirFS(absRoot))

	if *check {
		topicsFS := os.DirFS(filepath.Join(absRoot, "cmd", "glacier", "internal", "explain"))
		if err := explaingen.Check(src, topicsFS); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, "explaingen: all topic files are up to date.")
		return
	}

	files, err := explaingen.Generate(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "explaingen: generate: %v\n", err)
		os.Exit(1)
	}

	outDir := filepath.Join(absRoot, "cmd", "glacier", "internal", "explain", "topics")
	for name, content := range files {
		if err := safefile.WriteFileAtomic(outDir, name, content, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "explaingen: write %s: %v\n", name, err)
			os.Exit(1)
		}
	}
	fmt.Fprintf(os.Stdout, "explaingen: wrote %d topic files to %s\n", len(files), outDir)
}
