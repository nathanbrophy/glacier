// SPDX-License-Identifier: Apache-2.0

package gen

import (
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/tools/go/packages"
)

// scanLoadMode is the minimal set of go/packages flags needed by pkgScanner.
const scanLoadMode = packages.NeedName | packages.NeedModule

// pkgScanner is the production PackageScanner backed by go/packages.
type pkgScanner struct{}

// Scan implements PackageScanner using go/packages.
func (pkgScanner) Scan(pattern, modulePrefix string, _ *slog.Logger) (int, error) {
	cfg := &packages.Config{Mode: scanLoadMode}
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return 0, fmt.Errorf("httpmockgen: load packages %q: %w", pattern, err)
	}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return 0, fmt.Errorf("httpmockgen: package %q has errors: %v", pkg.PkgPath, pkg.Errors[0])
		}
	}
	count := 0
	for _, pkg := range pkgs {
		if modulePrefix != "" && !strings.HasPrefix(pkg.PkgPath, modulePrefix) {
			continue
		}
		count++
	}
	return count, nil
}

// detectModulePrefix detects the module path prefix via go/packages.
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
