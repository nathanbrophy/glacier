// SPDX-License-Identifier: Apache-2.0

package commands

import "context"

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
}

// Run shows help when the root command is invoked with no subcommand.
// The cli package shows the banner automatically on bare invocation.
func (c *GlacierCmd) Run(_ context.Context) error {
	if err := c.validateVerbosity(); err != nil {
		return err
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

func (mutexVerbosityError) Error() string {
	return "cli: -q/--quiet cannot be combined with -V/--verbose or --very-verbose"
}
