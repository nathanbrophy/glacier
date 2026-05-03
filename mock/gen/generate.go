// SPDX-License-Identifier: Apache-2.0

package gen

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/tools/go/packages"
)

// dirMuMap serializes concurrent Generate calls targeting the same directory.
var dirMuMap sync.Map // map[string]*sync.Mutex

func dirLock(dir string) func() {
	v, _ := dirMuMap.LoadOrStore(dir, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// Generate discovers interface types annotated with +glacier:mock in the
// packages matching opts.Pattern, then emits or checks zz_generated_mocks.go
// under each package directory that contains at least one such interface.
//
// Generate is goroutine-safe for distinct opts.Pattern values.
func Generate(opts Options) error {
	return GenerateWith(opts, pkgDiscoverer{})
}

// GenerateWith is the testable core: it accepts a Discoverer so unit tests
// can inject a fake without invoking go/packages. Production code should call
// Generate; this is exported for testing only.
func GenerateWith(opts Options, disc Discoverer) error {
	if opts.Pattern == "" {
		return fmt.Errorf("mockgen: Generate: Pattern is required")
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	modulePrefix, err := detectModulePrefix(opts.Pattern)
	if err != nil {
		return fmt.Errorf("mockgen: Generate: detect module: %w", err)
	}

	ifaces, err := disc.Discover(opts.Pattern, modulePrefix, opts.Logger)
	if err != nil {
		return fmt.Errorf("mockgen: Generate: discover: %w", err)
	}

	if len(ifaces) == 0 {
		opts.Logger.Info("mockgen: no +glacier:mock interfaces discovered", "pattern", opts.Pattern)
		return nil
	}

	// Group interfaces by package directory.
	type pkgGroup struct {
		pkgDir  string
		pkgName string
		ifaces  []DiscoveredInterface
	}
	pkgMap := make(map[string]*pkgGroup)
	for _, iface := range ifaces {
		key := iface.PkgDir
		if _, ok := pkgMap[key]; !ok {
			pkgMap[key] = &pkgGroup{pkgDir: iface.PkgDir, pkgName: iface.PkgName}
		}
		pkgMap[key].ifaces = append(pkgMap[key].ifaces, iface)
	}

	for dir, grp := range pkgMap {
		unlock := dirLock(dir)

		var buf bytes.Buffer
		if err := emitMocks(&buf, grp.pkgName, grp.ifaces); err != nil {
			unlock()
			return fmt.Errorf("mockgen: emit for %q: %w", dir, err)
		}

		generated := buf.Bytes()
		outPath := filepath.Join(dir, defaultOutput)

		if opts.Check {
			if err := checkDrift(outPath, generated); err != nil {
				unlock()
				return err
			}
			unlock()
			continue
		}

		if err := writeAtomic(outPath, generated); err != nil {
			unlock()
			return fmt.Errorf("mockgen: write %q: %w", outPath, err)
		}
		unlock()
	}

	return nil
}

// detectModulePrefix detects the module path prefix by loading package metadata.
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

// writeAtomic writes data to path using a temp-file + rename strategy.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".mockgen-*")
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
