// SPDX-License-Identifier: Apache-2.0

package gen

import "log/slog"

// Options configures a Generate call. The shape mirrors cli/gen.Options and
// mock/gen.Options so that the glacier SDK's generate command can treat all
// generators uniformly.
type Options struct {
	// Pattern is the go/packages load pattern (e.g. "./...", "github.com/example/app/...").
	// Precondition: must be a valid go/packages pattern; non-empty.
	Pattern string

	// Check, when true, performs drift detection. In v0, httpmock/gen emits no
	// files, so Check always returns nil when Pattern resolves to zero emitted
	// files.
	Check bool

	// Lint, when true, upgrades unknown marker warnings to errors.
	Lint bool

	// Logger is used for info/warning messages during generation.
	// Defaults to slog.Default() when nil.
	Logger *slog.Logger
}
