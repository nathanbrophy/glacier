// SPDX-License-Identifier: Apache-2.0

package benchcmp_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/benchcmp"
)

func TestParse_Standard(t *testing.T) {
	t.Parallel()
	input := `goos: linux
goarch: amd64
BenchmarkFoo-8   1000   100.0 ns/op   64 B/op   2 allocs/op
BenchmarkBar-8    500   200.5 ns/op    0 B/op   0 allocs/op
PASS`
	entries := benchcmp.Parse(input)
	assert.Equal(t, len(entries), 2)
	assert.Equal(t, entries[0].Name, "BenchmarkFoo-8")
	assert.Equal(t, entries[0].NsPerOp, 100.0)
	assert.Equal(t, entries[0].AllocsPerOp, int64(2))
	assert.Equal(t, entries[0].BPerOp, int64(64))
	assert.Equal(t, entries[1].Name, "BenchmarkBar-8")
	assert.Equal(t, entries[1].NsPerOp, 200.5)
	assert.Equal(t, entries[1].AllocsPerOp, int64(0))
}

func TestParse_IgnoresNonBenchLines(t *testing.T) {
	t.Parallel()
	input := `PASS
ok github.com/foo/bar 0.123s
--- BENCH: BenchmarkFoo-8`
	// Skips lines without proper ns/op format.
	entries := benchcmp.Parse(input)
	assert.Equal(t, len(entries), 0)
}

func TestParse_Empty(t *testing.T) {
	t.Parallel()
	entries := benchcmp.Parse("")
	assert.Equal(t, len(entries), 0)
}

func TestCompare_NoRegression(t *testing.T) {
	t.Parallel()
	baseline := []benchcmp.BenchEntry{{Name: "BenchmarkFoo-8", NsPerOp: 100}}
	current := []benchcmp.BenchEntry{{Name: "BenchmarkFoo-8", NsPerOp: 104}} // 4% — within threshold
	assert.Equal(t, len(benchcmp.Compare(baseline, current)), 0)
}

func TestCompare_AtThreshold(t *testing.T) {
	t.Parallel()
	baseline := []benchcmp.BenchEntry{{Name: "BenchmarkFoo-8", NsPerOp: 100}}
	current := []benchcmp.BenchEntry{{Name: "BenchmarkFoo-8", NsPerOp: 105}} // exactly 5% — not over
	assert.Equal(t, len(benchcmp.Compare(baseline, current)), 0)
}

func TestCompare_OverThreshold(t *testing.T) {
	t.Parallel()
	baseline := []benchcmp.BenchEntry{{Name: "BenchmarkFoo-8", NsPerOp: 100}}
	current := []benchcmp.BenchEntry{{Name: "BenchmarkFoo-8", NsPerOp: 106}} // 6% — regression
	regressions := benchcmp.Compare(baseline, current)
	assert.Equal(t, len(regressions), 1)
	assert.Equal(t, regressions[0].Name, "BenchmarkFoo-8")
	if regressions[0].DeltaPct <= 5.0 {
		t.Errorf("DeltaPct should be > 5.0, got %f", regressions[0].DeltaPct)
	}
}

func TestCompare_NewBenchNotInBaseline(t *testing.T) {
	t.Parallel()
	// Benchmarks absent from baseline are not flagged as regressions.
	baseline := []benchcmp.BenchEntry{{Name: "BenchmarkOld-8", NsPerOp: 100}}
	current := []benchcmp.BenchEntry{{Name: "BenchmarkNew-8", NsPerOp: 9999}}
	assert.Equal(t, len(benchcmp.Compare(baseline, current)), 0)
}

func TestCompare_Improvement(t *testing.T) {
	t.Parallel()
	baseline := []benchcmp.BenchEntry{{Name: "BenchmarkFoo-8", NsPerOp: 100}}
	current := []benchcmp.BenchEntry{{Name: "BenchmarkFoo-8", NsPerOp: 50}} // 50% faster
	assert.Equal(t, len(benchcmp.Compare(baseline, current)), 0)
}

func TestCompare_MultipleRegressions(t *testing.T) {
	t.Parallel()
	baseline := []benchcmp.BenchEntry{
		{Name: "BenchmarkA-8", NsPerOp: 100},
		{Name: "BenchmarkB-8", NsPerOp: 200},
		{Name: "BenchmarkC-8", NsPerOp: 50},
	}
	current := []benchcmp.BenchEntry{
		{Name: "BenchmarkA-8", NsPerOp: 120}, // 20% regression
		{Name: "BenchmarkB-8", NsPerOp: 202}, // 1% — ok
		{Name: "BenchmarkC-8", NsPerOp: 60},  // 20% regression
	}
	regressions := benchcmp.Compare(baseline, current)
	assert.Equal(t, len(regressions), 2)
}

func TestCompare_ZeroBaselineSkipped(t *testing.T) {
	t.Parallel()
	// A zero baseline ns/op should not produce a regression (division by zero guard).
	baseline := []benchcmp.BenchEntry{{Name: "BenchmarkFoo-8", NsPerOp: 0}}
	current := []benchcmp.BenchEntry{{Name: "BenchmarkFoo-8", NsPerOp: 100}}
	assert.Equal(t, len(benchcmp.Compare(baseline, current)), 0)
}

func TestFormatRegressions_Empty(t *testing.T) {
	t.Parallel()
	assert.Equal(t, benchcmp.FormatRegressions(nil, "baseline.json"), "")
}

func TestFormatRegressions_NonEmpty(t *testing.T) {
	t.Parallel()
	regressions := []benchcmp.Regression{
		{Name: "BenchmarkFoo-8", BaselineNs: 100, CurrentNs: 120, DeltaPct: 20.0},
	}
	out := benchcmp.FormatRegressions(regressions, ".glacier/bench-baseline.json")
	assert.Equal(t, len(out) > 0, true)
	if out == "" {
		t.Error("expected non-empty output")
	}
}

func ExampleParse() {
	input := "BenchmarkFoo-8   1000   100.0 ns/op   64 B/op   2 allocs/op"
	entries := benchcmp.Parse(input)
	_ = entries[0].Name    // "BenchmarkFoo-8"
	_ = entries[0].NsPerOp // 100.0
}
