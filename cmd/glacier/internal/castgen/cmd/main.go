// SPDX-License-Identifier: Apache-2.0

// Command castgen records terminal sessions of the SDK and writes both
// asciinema v2 .cast files and self-contained .svg snapshots under
// site/public/casts/.
//
// Usage:
//
//	go run ./cmd/glacier/internal/castgen/cmd
//
// Pre-condition: ./glacier (or glacier.exe on Windows) must be built and
// available at the repo root. Use:
//
//	go build -o glacier ./cmd/glacier
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/nathanbrophy/glacier/cmd/glacier/internal/castgen"
)

func main() {
	binFlag := flag.String("bin", "", "path to the glacier binary; defaults to ./glacier(.exe) at repo root")
	outDir := flag.String("out", "site/public/casts", "output directory for .cast and .svg files")
	flag.Parse()

	bin := *binFlag
	if bin == "" {
		bin = defaultBinPath()
	}
	if _, err := os.Stat(bin); err != nil {
		log.Fatalf("castgen: binary not found at %q: %v\n  build it with: go build -o glacier ./cmd/glacier", bin, err)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("castgen: mkdir %q: %v", *outDir, err)
	}

	scenarios := []castgen.Scenario{
		{
			Name:  "help",
			Title: "glacier --help",
			Bin:   bin,
			Args:  []string{"--help"},
			Cols:  100, Rows: 40,
		},
		{
			Name:  "version",
			Title: "glacier version",
			Bin:   bin,
			Args:  []string{"version"},
			Cols:  60, Rows: 8,
		},
		{
			Name:  "version-json",
			Title: "glacier version --json",
			Bin:   bin,
			Args:  []string{"version", "--json"},
			Cols:  60, Rows: 12,
		},
		{
			Name:  "explain-list",
			Title: "glacier explain --list",
			Bin:   bin,
			Args:  []string{"explain", "--list"},
			Cols:  90, Rows: 40,
		},
		{
			Name:  "explain-exit-66",
			Title: "glacier explain exit:66",
			Bin:   bin,
			Args:  []string{"explain", "exit:66"},
			Cols:  90, Rows: 18,
		},
		{
			Name:  "vibe-static",
			Title: "glacier vibe --ascii",
			Bin:   bin,
			Args:  []string{"vibe", "--ascii"},
			Cols:  60, Rows: 8,
		},
		{
			Name:  "completions-bash",
			Title: "glacier completions bash",
			Bin:   bin,
			Args:  []string{"completions", "bash"},
			Cols:  100, Rows: 30,
		},
		{
			Name:  "version-help",
			Title: "glacier version --help",
			Bin:   bin,
			Args:  []string{"version", "--help"},
			Cols:  90, Rows: 25,
		},
		{
			Name:  "vibe-help",
			Title: "glacier vibe --help",
			Bin:   bin,
			Args:  []string{"vibe", "--help"},
			Cols:  90, Rows: 30,
		},
	}

	for _, scn := range scenarios {
		fmt.Printf("recording %s ...\n", scn.Name)
		c, err := castgen.Record(scn)
		if err != nil {
			log.Printf("  skip: %v", err)
			continue
		}
		if err := writeAll(*outDir, c); err != nil {
			log.Fatalf("castgen: write %s: %v", scn.Name, err)
		}
	}
	fmt.Println("done.")
}

// writeAll writes both <name>.cast and <name>.svg under outDir.
func writeAll(outDir string, c castgen.Cast) error {
	castPath := filepath.Join(outDir, c.Scenario.Name+".cast")
	svgPath := filepath.Join(outDir, c.Scenario.Name+".svg")

	var castBuf, svgBuf bytesWriter
	if err := castgen.WriteCast(&castBuf, c); err != nil {
		return err
	}
	if err := castgen.WriteSVG(&svgBuf, c); err != nil {
		return err
	}

	if err := os.WriteFile(castPath, castBuf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write cast: %w", err)
	}
	if err := os.WriteFile(svgPath, svgBuf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write svg: %w", err)
	}
	fmt.Printf("  wrote %s\n  wrote %s\n", castPath, svgPath)
	return nil
}

// defaultBinPath returns the conventional location of the glacier binary
// for the current platform: ./glacier on Unix, ./glacier.exe on Windows.
// The leading "./" is required on Unix and many Windows shells so exec.Command
// does not require the binary to be on PATH.
func defaultBinPath() string {
	abs, err := filepath.Abs("glacier")
	if err == nil {
		if runtime.GOOS == "windows" {
			return abs + ".exe"
		}
		return abs
	}
	if runtime.GOOS == "windows" {
		return ".\\glacier.exe"
	}
	return "./glacier"
}

// bytesWriter is a small in-memory writer used to buffer file content
// before atomic write. Avoids the bytes.Buffer dependency for the cmd
// package's small surface.
type bytesWriter struct{ data []byte }

func (b *bytesWriter) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *bytesWriter) Bytes() []byte { return b.data }
