// SPDX-License-Identifier: Apache-2.0

package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"
)

// Options configures a Generate call.
type Options struct {
	// Pattern is the go/packages load pattern (e.g. "./...", "github.com/example/app/cmd/...").
	// Precondition: must be a valid go/packages pattern; non-empty.
	Pattern string

	// OutputName is the filename written under each discovered package directory.
	// Defaults to "zz_generated_cli.go".
	// Precondition: must be a simple filename (no path separators, no leading dots).
	OutputName string

	// AppName selects the named App to target in the generated code.
	// Defaults to "cli.Default". The +glacier:command app= marker overrides per-command.
	// Precondition: when non-empty, must be a valid Go identifier.
	AppName string

	// Check, when true, performs drift detection: Generate writes to an in-memory
	// buffer and diffs against the on-disk file. Returns a non-nil error with a
	// human-readable diff if stale. No files are written in check mode.
	Check bool

	// Lint, when true, upgrades unknown marker warnings to errors.
	Lint bool

	// Logger is used for warning/info messages during generation.
	// Defaults to slog.Default() when nil.
	Logger *slog.Logger
}

// dirMu serializes concurrent Generate calls targeting the same directory.
var (
	dirMuMap sync.Map // map[string]*sync.Mutex
)

func dirLock(dir string) func() {
	v, _ := dirMuMap.LoadOrStore(dir, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// Generate discovers cli.Command implementers matching opts.Pattern, parses
// their +glacier:* markers, and emits or checks the generated registration file.
//
// Generate is goroutine-safe for distinct opts.Pattern values.
func Generate(opts Options) error {
	if opts.Pattern == "" {
		return fmt.Errorf("gen: Generate: Pattern is required")
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	outputName := opts.OutputName
	if outputName == "" {
		outputName = defaultOutput
	}
	// Validate outputName: no separators, no leading dot.
	if strings.ContainsAny(outputName, `/\`) || strings.HasPrefix(outputName, ".") {
		return fmt.Errorf("gen: Generate: OutputName %q is invalid (no path separators or leading dots)", outputName)
	}

	// Determine module prefix from go/packages.
	modulePrefix, err := detectModulePrefix(opts.Pattern)
	if err != nil {
		return fmt.Errorf("gen: Generate: detect module: %w", err)
	}

	commands, err := discoverCommands(opts.Pattern, modulePrefix, opts.Logger)
	if err != nil {
		return fmt.Errorf("gen: Generate: discover: %w", err)
	}

	if len(commands) == 0 {
		opts.Logger.Info("gen: no commands discovered", "pattern", opts.Pattern)
		return nil
	}

	// Handle unknown markers.
	for _, cmd := range commands {
		for _, unk := range cmd.ParseResult.Unknown {
			msg := fmt.Sprintf("gen: unknown marker %q on type %s", unk, cmd.TypeName)
			if opts.Lint {
				return fmt.Errorf("%s", msg)
			}
			opts.Logger.Warn(msg)
		}
	}

	// Group commands by package directory and package name.
	type pkgGroup struct {
		pkgDir  string
		pkgName string
		cmds    []DiscoveredCommand
	}
	pkgMap := make(map[string]*pkgGroup)
	for _, cmd := range commands {
		key := cmd.PkgDir
		if _, ok := pkgMap[key]; !ok {
			pkgMap[key] = &pkgGroup{pkgDir: cmd.PkgDir}
		}
		pkgMap[key].cmds = append(pkgMap[key].cmds, cmd)
	}

	// Fill package names.
	for _, grp := range pkgMap {
		if len(grp.cmds) > 0 {
			// Derive package name from PkgPath last segment.
			parts := strings.Split(grp.cmds[0].PkgPath, "/")
			grp.pkgName = parts[len(parts)-1]
		}
	}

	// Emit per package.
	for dir, grp := range pkgMap {
		unlock := dirLock(dir)

		var buf bytes.Buffer
		if err := emitRegistrations(&buf, grp.pkgName, grp.cmds, opts.AppName); err != nil {
			unlock()
			return fmt.Errorf("gen: emit for %q: %w", dir, err)
		}

		generated := buf.Bytes()
		outPath := filepath.Join(dir, outputName)

		if opts.Check {
			if err := checkDrift(outPath, generated); err != nil {
				unlock()
				return err
			}
			unlock()
			continue
		}

		// Write atomically.
		if err := writeAtomic(outPath, generated); err != nil {
			unlock()
			return fmt.Errorf("gen: write %q: %w", outPath, err)
		}
		unlock()
	}

	return nil
}

// detectModulePrefix detects the module path prefix for the pattern's module.
func detectModulePrefix(pattern string) (string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedModule,
	}
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return "", err
	}
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Path != "" {
			return pkg.Module.Path, nil
		}
	}
	return "", nil
}

// writeAtomic writes data to path using a temp file + rename strategy.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".glaciergen-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

// formatSource applies gofmt to src, returning the formatted result.
func formatSource(src []byte) ([]byte, error) {
	return format.Source(src)
}
