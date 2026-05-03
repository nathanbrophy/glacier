// SPDX-License-Identifier: Apache-2.0

package gen

import (
	"fmt"
	"log/slog"
)

// Generate runs the httpmock code generator over the packages matching
// opts.Pattern. In v0, httpmock defines no source markers, so Generate
// validates opts, logs a summary, and returns nil. It is a real function
// reference (not a stub) per spec 0032 D-S65.
//
// Generate is goroutine-safe.
func Generate(opts Options) error {
	return GenerateWith(opts, pkgScanner{})
}

// PackageScanner abstracts the package-loading step for testability. The
// production implementation uses go/packages; tests may inject a fake.
//
// +glacier:mock
type PackageScanner interface {
	// Scan loads packages matching pattern and returns the count discovered.
	// Only packages within modulePrefix (if non-empty) are counted.
	Scan(pattern, modulePrefix string, logger *slog.Logger) (int, error)
}

// GenerateWith is the testable entry point. It accepts a PackageScanner so
// unit tests can inject a fake without invoking go/packages. Production code
// should call Generate.
func GenerateWith(opts Options, scanner PackageScanner) error {
	if opts.Pattern == "" {
		return fmt.Errorf("httpmockgen: Generate: Pattern is required")
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	modulePrefix, err := detectModulePrefix(opts.Pattern)
	if err != nil {
		return fmt.Errorf("httpmockgen: Generate: detect module: %w", err)
	}

	n, err := scanner.Scan(opts.Pattern, modulePrefix, opts.Logger)
	if err != nil {
		return fmt.Errorf("httpmockgen: Generate: scan: %w", err)
	}

	opts.Logger.Info("httpmockgen: no +glacier:httpmock markers defined in v0; no files emitted",
		"pattern", opts.Pattern,
		"packages_scanned", n,
	)
	return nil
}
