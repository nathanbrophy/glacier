// SPDX-License-Identifier: Apache-2.0

// Package benchcmp parses Go benchmark output and compares results against a
// stored baseline file. The baseline file format is a JSON array of BenchEntry
// values written by WriteBaseline.
package benchcmp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// BenchEntry holds measured values for one benchmark.
type BenchEntry struct {
	// Name is the benchmark function name, e.g. "BenchmarkFoo/bar-8".
	Name string `json:"name"`
	// NsPerOp is nanoseconds per operation.
	NsPerOp float64 `json:"ns_per_op"`
	// AllocsPerOp is heap allocations per operation.
	AllocsPerOp int64 `json:"allocs_per_op"`
	// BPerOp is bytes allocated per operation.
	BPerOp int64 `json:"b_per_op"`
}

// Regression describes one benchmark that exceeds the regression threshold.
type Regression struct {
	// Name is the benchmark name.
	Name string
	// BaselineNs is the baseline ns/op.
	BaselineNs float64
	// CurrentNs is the current run's ns/op.
	CurrentNs float64
	// DeltaPct is the percentage slowdown (positive = slower).
	DeltaPct float64
}

// RegressionThreshold is the maximum allowed slowdown percentage (5%).
const RegressionThreshold = 5.0

// benchLineRe matches a Go benchmark output line:
// BenchmarkFoo-8   1000   1234.5 ns/op   64 B/op   2 allocs/op
var benchLineRe = regexp.MustCompile(
	`^(Benchmark\S+)\s+\d+\s+([\d.]+)\s+ns/op(?:\s+([\d.]+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?`,
)

// Parse extracts BenchEntry values from go test -bench output lines.
// Lines that do not match the benchmark format are silently skipped.
func Parse(output string) []BenchEntry {
	var entries []BenchEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		m := benchLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		e := BenchEntry{Name: m[1]}
		e.NsPerOp, _ = strconv.ParseFloat(m[2], 64)
		if m[3] != "" {
			v, _ := strconv.ParseFloat(m[3], 64)
			e.BPerOp = int64(v)
		}
		if m[4] != "" {
			e.AllocsPerOp, _ = strconv.ParseInt(m[4], 10, 64)
		}
		entries = append(entries, e)
	}
	return entries
}

// Compare checks current against baseline. It returns a slice of Regression
// entries for any benchmark whose ns/op exceeds baseline by more than
// RegressionThreshold percent. Benchmarks absent from baseline are skipped.
func Compare(baseline, current []BenchEntry) []Regression {
	bmap := make(map[string]BenchEntry, len(baseline))
	for _, b := range baseline {
		bmap[b.Name] = b
	}

	var regressions []Regression
	for _, c := range current {
		b, ok := bmap[c.Name]
		if !ok || b.NsPerOp == 0 {
			continue
		}
		delta := (c.NsPerOp - b.NsPerOp) / b.NsPerOp * 100.0
		if delta > RegressionThreshold {
			regressions = append(regressions, Regression{
				Name:       c.Name,
				BaselineNs: b.NsPerOp,
				CurrentNs:  c.NsPerOp,
				DeltaPct:   delta,
			})
		}
	}
	return regressions
}

// FormatRegressions returns a human-readable block describing regressions.
func FormatRegressions(regressions []Regression, baselinePath string) string {
	if len(regressions) == 0 {
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Benchmark regressions (>%.0f%% slower than %s):\n", RegressionThreshold, baselinePath)
	for _, r := range regressions {
		fmt.Fprintf(&sb, "  %-48s  baseline: %.1f ns/op  current: %.1f ns/op  delta: +%.1f%%\n",
			r.Name, r.BaselineNs, r.CurrentNs, r.DeltaPct)
	}
	return sb.String()
}
