// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"context"
	"log/slog"

	"github.com/nathanbrophy/glacier/log"
	"github.com/nathanbrophy/glacier/term"
)

// GlacierCmd is the root command. It holds global flags inherited by all commands.
//
// +glacier:command name=glacier
// +glacier:root
type GlacierCmd struct {
	// Quiet lowers log level to Warn; suppresses animations; keeps final summary.
	//
	// +glacier:short q
	// +glacier:default false
	Quiet bool

	// Verbose raises log level to Debug.
	//
	// +glacier:short V
	// +glacier:default false
	Verbose bool

	// VeryVerbose raises log level to Trace level (custom).
	//
	// +glacier:default false
	VeryVerbose bool

	// NoAnimate disables all animations even on a TTY.
	//
	// +glacier:default false
	NoAnimate bool

	// NoBanner suppresses the banner on --help.
	//
	// +glacier:default false
	NoBanner bool

	// Profile writes pprof files to <Profile>.cpu/.heap/.goroutine.
	Profile string

	// OtelEndpoint overrides OTEL_EXPORTER_OTLP_ENDPOINT for this invocation.
	//
	// +glacier:env OTEL_EXPORTER_OTLP_ENDPOINT
	OtelEndpoint string

	// NoColor disables ANSI color output. Equivalent to setting NO_COLOR
	// in the environment.
	//
	// +glacier:default false
	NoColor bool

	// ForceColor forces ANSI color emission even when output is not a TTY.
	// Useful for piping into less -R or capturing colored logs to a file.
	//
	// +glacier:default false
	ForceColor bool
}

// Run shows help when the root command is invoked with no subcommand.
// The cli package shows the banner automatically on bare invocation. All
// global-flag side effects (log level, color mode) happen in ApplyRoot,
// which the cli package invokes for every dispatch including subcommands.
func (c *GlacierCmd) Run(_ context.Context) error {
	return nil
}

// ApplyRoot implements cli.RootApplier. The cli package invokes it after
// parsing the root command's persistent flags and before resolving the
// active command, so subcommand handlers see the requested log level and
// color mode in place.
//
// Precedence: explicit flags win over env vars (env vars seeded the
// initial level in main.configureLogging; here we override only when a
// flag was actually set). Mutual exclusion of --quiet and --verbose /
// --very-verbose is enforced per spec 0032 D-S31; combining them returns
// an exit-2 usage error.
func (c *GlacierCmd) ApplyRoot(_ context.Context) error {
	if err := c.validateVerbosity(); err != nil {
		return err
	}
	switch {
	case c.VeryVerbose:
		log.SetDefaultLevel(log.LevelTrace)
	case c.Verbose:
		log.SetDefaultLevel(slog.LevelDebug)
	case c.Quiet:
		log.SetDefaultLevel(slog.LevelWarn)
	}
	switch {
	case c.NoColor:
		term.SetColorMode(term.ModeNever)
	case c.ForceColor:
		term.SetColorMode(term.ModeAlways)
	}
	return nil
}

// validateVerbosity enforces D-S31: -q/--quiet is mutually exclusive with
// -V/--verbose and --very-verbose. Combining them is a usage error (exit 2).
// Called by Run() on the root and by every subcommand's Run() that imports
// the global flags.
func (c *GlacierCmd) validateVerbosity() error {
	if c.Quiet && (c.Verbose || c.VeryVerbose) {
		return &exitCodeError{
			code:  exitUsage,
			cause: errMutuallyExclusiveVerbosity,
		}
	}
	return nil
}

// errMutuallyExclusiveVerbosity is the static error returned by
// validateVerbosity. Defined as a var so tests can compare with errors.Is.
var errMutuallyExclusiveVerbosity = mutexVerbosityError{}

type mutexVerbosityError struct{}

// Error implements error.
func (mutexVerbosityError) Error() string {
	return "cli: -q/--quiet cannot be combined with -V/--verbose or --very-verbose"
}
