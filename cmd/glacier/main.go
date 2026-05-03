// SPDX-License-Identifier: Apache-2.0

// Command glacier is the Glacier SDK binary.
// It is the framework's longest-running integration test and public face
// for new developers. See spec 0032 for design and decisions.
package main

import (
	"context"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/nathanbrophy/glacier/cli"
	_ "github.com/nathanbrophy/glacier/cmd/glacier/commands" // register all commands via init()
	"github.com/nathanbrophy/glacier/log"
	"github.com/nathanbrophy/glacier/obs"
	"github.com/nathanbrophy/glacier/term"
)

// main is the SDK binary entry point. It performs three setup steps before
// dispatching to cli.Default.Main():
//
//  1. Configure slog via the framework's log package (spec 0005). This routes
//     all log records through Glacier's TTY-aware handler with palette/color
//     auto-detection. The handler honors NO_COLOR and GLACIER_NO_COLOR.
//  2. If OTEL_EXPORTER_OTLP_ENDPOINT is set, initialize the obs Provider so
//     per-command spans and counters (including cache.hits/misses from the
//     new caching package) are emitted. When the env var is unset, obs.Init
//     returns a no-op provider with zero overhead per spec 0032 D-S25.
//  3. Hand off to cli.Default.Main(), which inspects the returned error
//     chain for any cli.ExitCoder and propagates the embedded code to
//     os.Exit per spec 0032 D-S27.
func main() {
	// Color is on by default. Pre-parse os.Args BEFORE the cli framework
	// renders its banner so the color decision is made before any output.
	configureColor()
	configureLogging()
	shutdown := configureTelemetry()
	defer shutdown()

	cli.Default.Main()
}

// configureColor sets the global term.ColorMode based on, in order of
// precedence (highest first):
//
//  1. --no-color flag, NO_COLOR env, GLACIER_NO_COLOR env: ModeNever
//  2. --force-color flag, FORCE_COLOR env, GLACIER_FORCE_COLOR env: ModeAlways
//  3. default: ModeAlways (color on by default)
//
// term.ShouldColor(w) is the single point of truth at every emission site
// (kaomoji status lines, banner, help, box borders, lint findings, etc.).
func configureColor() {
	args := os.Args[1:]
	noColor := slices.Contains(args, "--no-color") ||
		os.Getenv("NO_COLOR") != "" ||
		os.Getenv("GLACIER_NO_COLOR") != ""
	forceColor := slices.Contains(args, "--force-color") ||
		os.Getenv("FORCE_COLOR") != "" ||
		os.Getenv("GLACIER_FORCE_COLOR") != ""

	switch {
	case noColor:
		term.SetColorMode(term.ModeNever)
	case forceColor:
		term.SetColorMode(term.ModeAlways)
	default:
		term.SetColorMode(term.ModeAlways) // default: color on
	}
}

// configureLogging sets slog.Default to a Glacier-style handler and pins the
// process logger via the log package's SetDefault. Subsequent calls to
// slog.Default()/log.Default() return the configured logger.
//
// The handler is bound to log.WithDynamicLevel so the threshold can be
// updated by GlacierCmd.ApplyRoot once the cli framework has parsed the
// persistent --quiet / --verbose / --very-verbose flags. The initial level
// here is Warn (per spec 0032 D-S31) unless GLACIER_DEBUG / GLACIER_VERBOSE
// is set, in which case env vars seed a more chatty default that the
// flags can still raise but never silently lower.
func configureLogging() {
	level := slog.LevelWarn
	switch {
	case os.Getenv("GLACIER_DEBUG") != "":
		level = log.LevelTrace
	case os.Getenv("GLACIER_VERBOSE") != "":
		level = slog.LevelDebug
	}
	log.SetDefaultLevel(level)
	handler := log.NewHandler(os.Stderr, log.WithDynamicLevel())
	logger := slog.New(handler)
	slog.SetDefault(logger)
	log.SetDefault(logger)
}

// configureTelemetry initialises an obs.Provider from OTEL_EXPORTER_OTLP_ENDPOINT
// when set, and returns a Shutdown closure to flush exporters at process exit.
// When the env var is unset, the provider is a no-op and the closure is also
// a no-op; both paths are zero-allocation per spec 0032 D-S25.
func configureTelemetry() func() {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		// No endpoint: skip provider construction entirely. obs.Counter calls
		// from any package (including cache) become no-ops with no allocation.
		return func() {}
	}

	ctx := context.Background()
	provider, err := obs.Init(ctx,
		obs.WithEndpoint(endpoint),
		obs.WithLogger(slog.Default()),
	)
	if err != nil {
		// Failure to wire telemetry must not break the binary.
		slog.Warn("obs init failed; continuing without telemetry", "err", err)
		return func() {}
	}
	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = provider.Shutdown(shutdownCtx)
	}
}
