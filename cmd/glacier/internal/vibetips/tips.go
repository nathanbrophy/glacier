// SPDX-License-Identifier: Apache-2.0

// Package vibetips provides the 12-tip registry shown by glacier vibe.
// Each tip covers one Glacier framework package. Tips can be retrieved in
// registration order or in a deterministically shuffled order by seed.
package vibetips

import (
	"math/rand"
	"time"
)

// Tip is a single vibe tip entry.
type Tip struct {
	// Category is the Glacier package the tip refers to (e.g. "cli", "term").
	Category string
	// Body is the human-readable tip text.
	Body string
	// SpecRef is the spec ID that documents the described behaviour.
	SpecRef string
}

// registry holds the 12 canonical tips in registration order.
var registry = []Tip{
	{
		Category: "cli",
		Body:     "cli.Default.Register accepts any struct with a Run(ctx) error method. No embedding, no code generation required for prototyping.",
		SpecRef:  "0003",
	},
	{
		Category: "term",
		Body:     "term.Box with WithRoundedCorners() and WithTitle() is the standard way to frame a summary block in any Glacier command.",
		SpecRef:  "0001",
	},
	{
		Category: "log",
		Body:     "Kaomoji belong in report.Status, not in log records. Logs are for machines; report.Status lines are for humans at command boundaries.",
		SpecRef:  "0004",
	},
	{
		Category: "conf",
		Body:     "conf.Register returns a zero-allocation accessor: call cfg() anywhere, the atomic pointer gives you the latest loaded config.",
		SpecRef:  "0007",
	},
	{
		Category: "errs",
		Body:     "errs.Sentinel rejects messages that don't conform to the library register. A per-package unit test enforces the format at build time.",
		SpecRef:  "0005",
	},
	{
		Category: "assert",
		Body:     "assert.Equal uses a field-by-field diff. assert/require.Equal stops the test immediately, assert.Equal keeps going.",
		SpecRef:  "0008",
	},
	{
		Category: "fixture",
		Body:     "fixture.Golden updates expected output with -update; CI runs without it. One flag, zero drift.",
		SpecRef:  "0009",
	},
	{
		Category: "mock",
		Body:     "mock.Of[T]() generates a full mock at compile time via the +glacier:mock marker. No reflect at test time.",
		SpecRef:  "0010",
	},
	{
		Category: "httpmock",
		Body:     "httpmock.NewRouter() intercepts any http.Client whose transport is replaced with httpmock.Transport(router).",
		SpecRef:  "0013",
	},
	{
		Category: "httpc",
		Body:     "httpc.Default has retry, backoff, tracing, and timeout pre-wired. Pass httpc.WithMaxAttempts(1) to disable retry for tests.",
		SpecRef:  "0012",
	},
	{
		Category: "option",
		Body:     "option.Apply is goroutine-safe and zero-allocation when the option slice is nil or empty.",
		SpecRef:  "0006",
	},
	{
		Category: "concur",
		Body:     "concur.Group collects goroutine errors; the first non-nil error cancels the group context automatically.",
		SpecRef:  "0014",
	},
}

// All returns all 12 tips in registration order.
func All() []Tip {
	out := make([]Tip, len(registry))
	copy(out, registry)
	return out
}

// Shuffled returns a deterministically shuffled copy of the tip registry using
// the given seed. When seed is 0, time.Now().UnixNano() is used as the seed so
// each call produces a different order.
func Shuffled(seed int64) []Tip {
	out := All()
	if seed == 0 {
		// Use a time-based seed for non-deterministic shuffle when seed == 0.
		// Import is avoided by using the default global source reset.
		seed = time.Now().UnixNano()
	}
	//nolint:gosec // deterministic shuffle; not security-sensitive
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}
