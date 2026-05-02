// SPDX-License-Identifier: Apache-2.0

package mock

import "log/slog"

// mockOptions holds the resolved configuration for a Mock[T] instance.
// It is built by applying Option values in the order they are passed to Of[T].
type mockOptions struct {
	mode   strictMode
	logger *slog.Logger
	// seqMode is the default SeqExhaustion applied to ReturnSeq calls that
	// do not specify their own exhaustion mode.
	seqMode seqExhaustion
}

// Option is a functional option for Mock[T] construction.
// Options are applied in the order they are passed to Of[T].
type Option interface{ applyMock(*mockOptions) }

// optionFunc is the unexported function adapter that satisfies Option.
type optionFunc func(*mockOptions)

func (f optionFunc) applyMock(o *mockOptions) { f(o) }

// StrictDefault returns an Option that sets the mock to strict mode.
// In strict mode, any call with no matching expectation calls t.Errorf
// (reporting method name, received arguments, and the full list of
// registered expectations) and returns zero values.
//
// Strict is the default; this option exists for explicit documentation.
func StrictDefault() Option {
	return optionFunc(func(o *mockOptions) { o.mode = strictErrorf })
}

// StrictFatal returns an Option that upgrades strict mode to t.Fatalf,
// halting the test goroutine immediately on an unmatched call.
func StrictFatal() Option {
	return optionFunc(func(o *mockOptions) { o.mode = strictFatalf })
}

// LenientMode returns an Option that suppresses failure on unmatched calls.
// Unmatched calls are recorded in UnmatchedCalls() for the test to inspect.
func LenientMode() Option {
	return optionFunc(func(o *mockOptions) { o.mode = lenient })
}

// SeqCycleOpt returns an Option that sets the default sequence exhaustion
// mode to SeqCycle (wrap around). This is the built-in default.
func SeqCycleOpt() Option {
	return optionFunc(func(o *mockOptions) { o.seqMode = SeqCycle })
}

// SeqExhaustOpt returns an Option that sets the default sequence exhaustion
// mode to SeqExhaust (fail test on exhaustion).
func SeqExhaustOpt() Option {
	return optionFunc(func(o *mockOptions) { o.seqMode = SeqExhaust })
}

// WithLogger sets the slog.Logger used by the mock for internal diagnostics.
// If not set, slog.Default() is used.
func WithLogger(l *slog.Logger) Option {
	return optionFunc(func(o *mockOptions) { o.logger = l })
}

// applyMockOptions applies opts over a default mockOptions and returns it.
func applyMockOptions(opts []Option) mockOptions {
	o := mockOptions{
		mode:    strictErrorf,
		seqMode: SeqCycle,
		logger:  slog.Default(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt.applyMock(&o)
		}
	}
	return o
}
