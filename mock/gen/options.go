// SPDX-License-Identifier: Apache-2.0

package gen

import "log/slog"

// Options configures a Generate call.
type Options struct {
	// Pattern is the go/packages load pattern (e.g. "./...", "github.com/example/app/...").
	// Precondition: must be a valid go/packages pattern; non-empty.
	Pattern string

	// Check, when true, performs drift detection: Generate writes to an in-memory
	// buffer and diffs against the on-disk file. Returns a non-nil error (containing
	// "stale") if any generated file differs from what is on disk. No files are
	// written in check mode.
	Check bool

	// Lint, when true, upgrades unknown marker warnings to errors.
	Lint bool

	// Logger is used for warning/info messages during generation.
	// Defaults to slog.Default() when nil.
	Logger *slog.Logger
}
