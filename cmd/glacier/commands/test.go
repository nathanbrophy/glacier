// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nathanbrophy/glacier/cmd/glacier/internal/benchcmp"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/report"
	"github.com/nathanbrophy/glacier/internal/safefile"
	"github.com/nathanbrophy/glacier/internal/safejson"
	"github.com/nathanbrophy/glacier/term"
)

// TestCmd runs the test suite.
//
// +glacier:command name=test parent=glacier
type TestCmd struct {
	// Patterns are the packages to test (default ./...).
	//
	// +glacier:positional
	Patterns []string

	// Race enables the race detector.
	//
	// +glacier:default false
	Race bool

	// Cover enables coverage reporting.
	//
	// +glacier:default false
	Cover bool

	// Fuzz is a regexp selecting fuzz targets to run.
	Fuzz string

	// Bench is a regexp selecting benchmarks to run.
	Bench string

	// Baseline is the path to the benchmark baseline file.
	//
	// +glacier:default ".glacier/bench-baseline.json"
	Baseline string

	// UpdateBaseline writes a new benchmark baseline.
	//
	// +glacier:default false
	UpdateBaseline bool

	// Format is the output format: text, junit, sarif, or json.
	//
	// +glacier:choices text|junit|sarif|json
	// +glacier:default text
	Format string

	// Slowest prints the N slowest tests.
	//
	// +glacier:default 5
	Slowest int

	// NoStatus disables the status panel animation.
	//
	// +glacier:default false
	NoStatus bool

	// runner is used for dependency injection in tests; nil means the real subprocess.
	runner TestRunner
}

// TestRunner runs `go test -json` and returns channels of event lines and bench
// output. The interface exists for testability; inject a fake via TestCmd.runner.
//
// +glacier:mock
type TestRunner interface {
	// Run starts the test subprocess with the given arguments and returns a
	// channel of raw `go test -json` output lines (one JSON object per line),
	// a channel of non-JSON output lines (benchmark output), and any start
	// error.
	Run(ctx context.Context, args ...string) (events <-chan string, benchOut <-chan string, err error)
	// Wait blocks until the subprocess exits and returns its exit error.
	Wait() error
}

// testEvent is a single event from go test -json output.
type testEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

// testResult summarizes a completed test.
type testResult struct {
	Package string
	Test    string
	Action  string // "pass", "fail", "skip"
	Elapsed float64
	Output  string // accumulated output for failure diagnostics
}

// Summary holds aggregate test-run statistics. It is passed to format emitters.
type Summary struct {
	// Packages is the number of tested packages.
	Packages int
	// Pass is the number of passing test functions.
	Pass int
	// Fail is the number of failing test functions.
	Fail int
	// Skip is the number of skipped test functions.
	Skip int
	// Coverage is the measured fraction [0,1]; zero means not measured.
	Coverage float64
	// WallSeconds is total elapsed wall-clock time for the run.
	WallSeconds float64
	// Slowest is the N slowest test results by elapsed time.
	Slowest []testResult
	// Failures holds all failing test results, for display in the summary.
	Failures []testResult
}

// statusPanelMaxRows is the cap on simultaneously displayed in-flight packages.
const statusPanelMaxRows = 10

// Run executes `go test -json` and streams results.
func (c *TestCmd) Run(ctx context.Context) error {
	report.Status(report.Calm, "glacier test")

	switch c.Format {
	case "", "text", "junit", "sarif", "json":
		// valid format
	default:
		return &exitCodeError{
			code:  exitUsage,
			cause: fmt.Errorf("unknown format %q", c.Format),
		}
	}

	patterns := c.Patterns
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	args := buildTestArgs(c, patterns)

	runner := c.runner
	if runner == nil {
		runner = &goTestRunnerImpl{}
	}

	events, benchOut, err := runner.Run(ctx, args...)
	if err != nil {
		return &exitCodeError{
			code:  exitSubprocess,
			cause: fmt.Errorf("cannot start `go test`: %w", err),
		}
	}

	// Rolling status panel (TTY + text format only).
	var sb *term.StatusBar
	var animator *term.Animator
	caps := term.Capability(os.Stderr)
	usePanel := !c.NoStatus && caps.IsTTY && (c.Format == "text" || c.Format == "")
	if usePanel {
		sb = term.NewStatusBar()
		animator = term.NewAnimator(slog.Default())
		animator.Add(sb)
		animCtx, animCancel := context.WithCancel(ctx)
		defer animCancel()
		go func() {
			_ = animator.Run(animCtx)
		}()
	}

	start := time.Now()

	var results []testResult
	// testOutputs accumulates per-test output keyed by "package/test".
	testOutputs := make(map[string]*strings.Builder)
	// packageSet tracks distinct tested packages.
	packageSet := make(map[string]struct{})
	failed := 0

	// panelKeys is the ring-buffer of currently displayed packages.
	panelKeys := make([]string, 0, statusPanelMaxRows+1)

	// coverageLines accumulates lines that may carry coverage percentages.
	var coverageLines []string

	// Drain bench output concurrently.
	var benchLines []string
	doneBench := make(chan struct{})
	go func() {
		defer close(doneBench)
		for line := range benchOut {
			benchLines = append(benchLines, line)
		}
	}()

	// Process JSON event stream.
	for line := range events {
		if c.Format == "json" {
			// Forward verbatim before parsing; tools expecting raw go test -json
			// output see every event unmodified.
			fmt.Fprintln(os.Stdout, line)
		}

		var ev testEvent
		if jsonErr := json.Unmarshal([]byte(line), &ev); jsonErr != nil {
			continue
		}

		if ev.Package != "" {
			packageSet[ev.Package] = struct{}{}
		}

		switch ev.Action {
		case "run":
			if sb != nil && ev.Test != "" {
				addPanelKey(sb, &panelKeys, ev.Package, "ʕ•_•ʔ "+ev.Package+"/"+ev.Test)
			}

		case "output":
			if ev.Test != "" {
				k := ev.Package + "/" + ev.Test
				if testOutputs[k] == nil {
					testOutputs[k] = &strings.Builder{}
				}
				testOutputs[k].WriteString(ev.Output)
			}
			// Watch for coverage summary lines regardless of test association.
			if strings.Contains(ev.Output, "coverage:") {
				coverageLines = append(coverageLines, ev.Output)
			}

		case "pass", "fail", "skip":
			if ev.Test != "" {
				out := ""
				if b := testOutputs[ev.Package+"/"+ev.Test]; b != nil {
					out = b.String()
				}
				results = append(results, testResult{
					Package: ev.Package,
					Test:    ev.Test,
					Action:  ev.Action,
					Elapsed: ev.Elapsed,
					Output:  out,
				})
				if ev.Action == "fail" {
					failed++
				}
			} else {
				// Package-level result; update the status panel.
				if sb != nil {
					sb.Remove(ev.Package)
					removePanelKey(&panelKeys, ev.Package)
					emitPackageResultLine(os.Stderr, ev.Action, ev.Package, ev.Elapsed)
				}
			}
		}
	}

	<-doneBench

	if waitErr := runner.Wait(); waitErr != nil && failed == 0 {
		failed++
	}

	wallSeconds := time.Since(start).Seconds()

	if animator != nil {
		animator.Close()
	}

	coverage := parseCoverage(coverageLines)

	sum := buildSummary(results, coverage, wallSeconds, len(packageSet), c.Slowest)

	// Harvest benchmark result lines from per-test output. With go test -json,
	// the bench result line "BenchmarkFoo-N  iter ns/op B/op allocs/op" is
	// frequently split across multiple "output" events; accumulating the
	// full per-test buffer and parsing it whole avoids partial lines.
	for k, b := range testOutputs {
		// Only consider tests whose name starts with "Benchmark" to keep
		// non-bench test output out of the bench parser.
		idx := strings.LastIndex(k, "/")
		if idx < 0 || !strings.HasPrefix(k[idx+1:], "Benchmark") {
			continue
		}
		benchLines = append(benchLines, b.String())
	}

	// Benchmark baseline compare / update.
	benchEntries := benchcmp.Parse(strings.Join(benchLines, "\n"))

	if c.UpdateBaseline && len(benchEntries) > 0 {
		if wErr := writeBaseline(c.Baseline, benchEntries); wErr != nil {
			return fmt.Errorf("test: write baseline: %w", wErr)
		}
	}

	if c.Bench != "" && len(benchEntries) > 0 {
		// Always print a human-readable bench summary when benchmarks ran.
		// Without this, only the bare go-test output and a regression block
		// on failure were visible; passing benchmarks were silent.
		baseline, _ := loadBaseline(c.Baseline)
		printBenchSummary(os.Stderr, benchEntries, baseline)
	}

	if c.Bench != "" && !c.UpdateBaseline && len(benchEntries) > 0 {
		baseline, loadErr := loadBaseline(c.Baseline)
		if loadErr == nil && len(baseline) > 0 {
			regressions := benchcmp.Compare(baseline, benchEntries)
			if len(regressions) > 0 {
				block := benchcmp.FormatRegressions(regressions, c.Baseline)
				box := term.Box(block,
					term.WithTitle("bench regressions"),
					term.WithRoundedCorners(),
					term.WithPadding(1, 2, 1, 2),
				)
				fmt.Fprintln(os.Stderr, box)
				return &exitCodeError{
					code:  exitTestsFailed,
					cause: fmt.Errorf("test: bench regression(s) detected"),
				}
			}
		}
	}

	// Emit in the requested format.
	switch c.Format {
	case "junit":
		out, emitErr := emitJUnit(sum)
		if emitErr != nil {
			return fmt.Errorf("test: junit: %w", emitErr)
		}
		_, _ = os.Stdout.Write(out)
	case "sarif":
		out, emitErr := emitSARIF(sum)
		if emitErr != nil {
			return fmt.Errorf("test: sarif: %w", emitErr)
		}
		_, _ = os.Stdout.Write(out)
	case "json":
		agg := buildJSONAggregate(sum)
		line, marshalErr := json.Marshal(agg)
		if marshalErr != nil {
			return fmt.Errorf("test: json aggregate: %w", marshalErr)
		}
		fmt.Fprintln(os.Stdout, string(line))
	default:
		printTestSummary(sum)
	}

	if failed > 0 {
		report.Status(report.Err, fmt.Sprintf("%d test(s) failed", failed))
		return &exitCodeError{
			code:  exitTestsFailed,
			cause: fmt.Errorf("test: %d test(s) failed", failed),
		}
	}

	report.Status(report.Confident, "that went well.")
	return nil
}

// addPanelKey adds key to the status bar, evicting the oldest entry when the
// ring buffer is at capacity. Updates content if already present.
func addPanelKey(sb *term.StatusBar, keys *[]string, key, content string) {
	for _, k := range *keys {
		if k == key {
			sb.SetSection(key, content)
			return
		}
	}
	if len(*keys) >= statusPanelMaxRows {
		oldest := (*keys)[0]
		*keys = (*keys)[1:]
		sb.Remove(oldest)
	}
	*keys = append(*keys, key)
	sb.SetSection(key, content)
}

// removePanelKey removes key from the ring-buffer slice.
func removePanelKey(keys *[]string, key string) {
	for i, k := range *keys {
		if k == key {
			*keys = append((*keys)[:i], (*keys)[i+1:]...)
			return
		}
	}
}

// buildTestArgs assembles the go test argument list.
func buildTestArgs(c *TestCmd, patterns []string) []string {
	args := []string{"test", "-json"}
	if c.Race {
		args = append(args, "-race")
	}
	if c.Cover {
		args = append(args, "-cover", "-coverprofile=.glacier/coverage.out")
	}
	if c.Fuzz != "" {
		args = append(args, "-fuzz="+c.Fuzz)
	}
	if c.Bench != "" {
		args = append(args, "-bench="+c.Bench, "-benchmem")
	}
	args = append(args, patterns...)
	return args
}

// buildSummary constructs a Summary from raw test results.
func buildSummary(results []testResult, coverage, wallSeconds float64, packages, slowestN int) Summary {
	var sum Summary
	sum.Packages = packages
	sum.Coverage = coverage
	sum.WallSeconds = wallSeconds

	for _, r := range results {
		switch r.Action {
		case "pass":
			sum.Pass++
		case "fail":
			sum.Fail++
			sum.Failures = append(sum.Failures, r)
		case "skip":
			sum.Skip++
		}
	}

	sorted := make([]testResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Elapsed > sorted[j].Elapsed
	})
	cap := slowestN
	if cap > len(sorted) {
		cap = len(sorted)
	}
	sum.Slowest = sorted[:cap]
	return sum
}

// parseCoverage extracts the coverage fraction from lines like
// "coverage: 91.4% of statements".
func parseCoverage(lines []string) float64 {
	for _, line := range lines {
		idx := strings.Index(line, "coverage:")
		if idx < 0 {
			continue
		}
		rest := strings.TrimSpace(line[idx+len("coverage:"):])
		var pct float64
		if _, err := fmt.Sscanf(rest, "%f%%", &pct); err == nil {
			return pct / 100.0
		}
	}
	return 0
}

// printBenchSummary renders a human-readable benchmark summary box to w.
// Each entry shows ns/op, B/op, and allocs/op aligned in columns. When a
// baseline is supplied, an extra column shows the delta vs baseline as a
// percentage with green / yellow / red coloring (faster / similar / slower).
func printBenchSummary(w io.Writer, current, baseline []benchcmp.BenchEntry) {
	useColor := term.ShouldColor(w)
	const (
		reset       = "\x1b[0m"
		bold        = "\x1b[1m"
		dim         = "\x1b[2m"
		labelDim    = "\x1b[2;38;5;87m"
		nameMagenta = "\x1b[38;5;213m"
		green       = "\x1b[38;5;84m"
		yellow      = "\x1b[38;5;228m"
		red         = "\x1b[38;5;203m"
		cyan        = "\x1b[38;5;87m"
	)
	wrap := func(prefix, s string) string {
		if !useColor {
			return s
		}
		return prefix + s + reset
	}
	// padLeft pads s to width n by prepending spaces. Always operates on
	// visible content (no ANSI), so callers should pad BEFORE wrapping color.
	padLeft := func(s string, n int) string {
		if len(s) >= n {
			return s
		}
		return strings.Repeat(" ", n-len(s)) + s
	}
	// padRight pads s to width n by appending spaces.
	padRight := func(s string, n int) string {
		if len(s) >= n {
			return s
		}
		return s + strings.Repeat(" ", n-len(s))
	}

	// Index baseline by name for quick lookup.
	base := make(map[string]benchcmp.BenchEntry, len(baseline))
	for _, b := range baseline {
		base[b.Name] = b
	}

	// Compute column widths for alignment from VISIBLE strings.
	type row struct {
		name, ns, b, alloc, delta string
		deltaColor                string
	}
	rows := make([]row, 0, len(current))
	const sep = "   "
	colW := struct {
		name, ns, b, alloc, delta int
	}{
		name: len("name"), ns: len("ns/op"), b: len("B/op"),
		alloc: len("allocs/op"), delta: len("vs base"),
	}
	hasBaseline := len(baseline) > 0
	for _, e := range current {
		r := row{
			name:  e.Name,
			ns:    formatNs(e.NsPerOp),
			b:     fmt.Sprintf("%d", e.BPerOp),
			alloc: fmt.Sprintf("%d", e.AllocsPerOp),
		}
		if hasBaseline {
			if be, ok := base[e.Name]; ok && be.NsPerOp > 0 {
				delta := (e.NsPerOp - be.NsPerOp) / be.NsPerOp * 100.0
				r.delta = fmt.Sprintf("%+6.1f%%", delta)
				switch {
				case delta < -1.0:
					r.deltaColor = green
				case delta > 5.0:
					r.deltaColor = red
				case delta > 1.0:
					r.deltaColor = yellow
				default:
					r.deltaColor = labelDim
				}
			} else {
				r.delta = "(new)"
				r.deltaColor = labelDim
			}
		}
		if len(r.name) > colW.name {
			colW.name = len(r.name)
		}
		if len(r.ns) > colW.ns {
			colW.ns = len(r.ns)
		}
		if len(r.b) > colW.b {
			colW.b = len(r.b)
		}
		if len(r.alloc) > colW.alloc {
			colW.alloc = len(r.alloc)
		}
		if len(r.delta) > colW.delta {
			colW.delta = len(r.delta)
		}
		rows = append(rows, r)
	}

	var sb strings.Builder

	// Header row: pad first, then color, so column widths match.
	hdrName := padRight("name", colW.name)
	hdrNs := padLeft("ns/op", colW.ns)
	hdrB := padLeft("B/op", colW.b)
	hdrAlloc := padLeft("allocs/op", colW.alloc)
	headerLine := wrap(bold+cyan, hdrName) + sep +
		wrap(bold+cyan, hdrNs) + sep +
		wrap(bold+cyan, hdrB) + sep +
		wrap(bold+cyan, hdrAlloc)
	if hasBaseline {
		headerLine += sep + wrap(bold+cyan, padLeft("vs base", colW.delta))
	}
	sb.WriteString(headerLine)
	sb.WriteString("\n")

	// Underline.
	totalVis := colW.name + colW.ns + colW.b + colW.alloc + 3*len(sep)
	if hasBaseline {
		totalVis += colW.delta + len(sep)
	}
	sb.WriteString(wrap(dim+cyan, strings.Repeat("─", totalVis)))
	sb.WriteString("\n")

	for _, r := range rows {
		line := wrap(nameMagenta, padRight(r.name, colW.name)) + sep +
			wrap(cyan, padLeft(r.ns, colW.ns)) + sep +
			wrap(cyan, padLeft(r.b, colW.b)) + sep +
			wrap(cyan, padLeft(r.alloc, colW.alloc))
		if hasBaseline {
			line += sep + wrap(r.deltaColor, padLeft(r.delta, colW.delta))
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	if hasBaseline {
		sb.WriteString(wrap(labelDim, "baseline: "+relativeBaselinePath()))
	} else {
		sb.WriteString(wrap(labelDim, "no baseline yet; run with --update-baseline to record"))
	}
	sb.WriteString("\n")

	box := term.Box(
		sb.String(),
		term.WithTitle(fmt.Sprintf("bench results  %d benchmark(s)", len(current))),
		term.WithRoundedCorners(),
		term.WithPadding(1, 2, 1, 2),
	)
	fmt.Fprintln(w, box)
}

// formatNs formats a ns/op value at human-readable scale (ns / µs / ms / s).
func formatNs(ns float64) string {
	switch {
	case ns < 1e3:
		return fmt.Sprintf("%.1f ns", ns)
	case ns < 1e6:
		return fmt.Sprintf("%.2f µs", ns/1e3)
	case ns < 1e9:
		return fmt.Sprintf("%.2f ms", ns/1e6)
	default:
		return fmt.Sprintf("%.2f s", ns/1e9)
	}
}

// relativeBaselinePath returns the bench baseline file path relative to the
// current working directory if possible, falling back to the absolute path.
func relativeBaselinePath() string {
	const def = ".glacier/bench-baseline.json"
	if _, err := os.Stat(def); err == nil {
		return def
	}
	return def
}

// lenVisible returns the visible character count of s after stripping ANSI
// escapes. Local helper because the term package's visibleWidth is unexported
// at the package boundary; this duplicates the small CSI-strip logic.
func lenVisible(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			// Skip until final byte 0x40..0x7E
			i += 2
			for i < len(s) && (s[i] < 0x40 || s[i] > 0x7E) {
				i++
			}
			continue
		}
		n++
	}
	return n
}

// emitPackageResultLine writes one streaming "PKG (Xs)" line per completed
// package, colored by pass/fail/skip semantics.
func emitPackageResultLine(w io.Writer, action, pkg string, elapsed float64) {
	useColor := term.ShouldColor(w)
	const reset = "\x1b[0m"
	var glyph, glyphColor, pkgColor, elapsedColor string
	switch action {
	case "fail":
		glyph = "ʕ× ×ʔ"
		glyphColor = "\x1b[1;38;5;203m" // bold red
		pkgColor = "\x1b[38;5;203m"
		elapsedColor = "\x1b[2;38;5;203m"
	case "skip":
		glyph = "ʕ•_•ʔ"
		glyphColor = "\x1b[1;38;5;228m" // bold yellow
		pkgColor = "\x1b[38;5;228m"
		elapsedColor = "\x1b[2;38;5;228m"
	default: // pass
		glyph = "ʕ⌐■-■ʔ"
		glyphColor = "\x1b[1;38;5;84m" // bold green
		pkgColor = "\x1b[38;5;75m"
		elapsedColor = "\x1b[2;38;5;87m"
	}
	if !useColor {
		fmt.Fprintf(w, "%s %s (%.3fs)\n", glyph, pkg, elapsed)
		return
	}
	fmt.Fprintf(w, "%s%s%s %s%s%s %s(%.3fs)%s\n",
		glyphColor, glyph, reset,
		pkgColor, pkg, reset,
		elapsedColor, elapsed, reset)
}

// printTestSummary renders the summary box to stderr with rich color
// (numerical colors per metric, fail/slow/output highlighting).
func printTestSummary(sum Summary) {
	useColor := term.ShouldColor(os.Stderr)
	c := newSummaryColors(useColor)

	var sb strings.Builder
	total := sum.Pass + sum.Fail + sum.Skip

	// Top metrics line: each count tinted by semantic.
	sb.WriteString(c.label("packages: "))
	sb.WriteString(c.numNeutral(fmt.Sprintf("%d", sum.Packages)))
	sb.WriteString("  ")
	sb.WriteString(c.label("tests: "))
	sb.WriteString(c.numNeutral(fmt.Sprintf("%d", total)))
	sb.WriteString("  ")
	sb.WriteString(c.label("pass: "))
	sb.WriteString(c.pass(fmt.Sprintf("%d", sum.Pass)))
	sb.WriteString("  ")
	sb.WriteString(c.label("fail: "))
	if sum.Fail > 0 {
		sb.WriteString(c.fail(fmt.Sprintf("%d", sum.Fail)))
	} else {
		sb.WriteString(c.dim("0"))
	}
	sb.WriteString("  ")
	sb.WriteString(c.label("skip: "))
	if sum.Skip > 0 {
		sb.WriteString(c.skip(fmt.Sprintf("%d", sum.Skip)))
	} else {
		sb.WriteString(c.dim("0"))
	}
	sb.WriteString("\n")

	if sum.Coverage > 0 {
		covStr := fmt.Sprintf("%.1f%%", sum.Coverage*100)
		sb.WriteString(c.label("coverage: "))
		switch {
		case sum.Coverage >= 0.9:
			sb.WriteString(c.pass(covStr))
		case sum.Coverage >= 0.75:
			sb.WriteString(c.numNeutral(covStr))
		default:
			sb.WriteString(c.warn(covStr))
		}
		sb.WriteString("\n")
	}
	sb.WriteString(c.label("wall: "))
	sb.WriteString(c.numNeutral(fmt.Sprintf("%.1fs", sum.WallSeconds)))
	sb.WriteString("\n")

	if len(sum.Slowest) > 0 {
		sb.WriteString("\n")
		sb.WriteString(c.heading(fmt.Sprintf("Slowest %d tests:", len(sum.Slowest))))
		sb.WriteString("\n")
		for i, r := range sum.Slowest {
			fmt.Fprintf(&sb, "  %s. %s%s  %s\n",
				c.dim(fmt.Sprintf("%d", i+1)),
				c.pkg(r.Package+"/"),
				c.testName(r.Test),
				c.numNeutral(fmt.Sprintf("%.3fs", r.Elapsed)))
		}
	}

	if len(sum.Failures) > 0 {
		sb.WriteString("\n")
		sb.WriteString(c.fail("ʕ× ×ʔ Failures:"))
		sb.WriteString("\n")
		for _, r := range sum.Failures {
			fmt.Fprintf(&sb, "  %s%s\n",
				c.pkg(r.Package+"/"),
				c.testName(r.Test))
			if r.Output != "" {
				for _, outLine := range strings.Split(strings.TrimRight(r.Output, "\n"), "\n") {
					sb.WriteString("    " + c.dim(outLine) + "\n")
				}
			}
		}
		sb.WriteString("\n  Try running the failing test in isolation:\n")
		if len(sum.Failures) == 1 {
			f := sum.Failures[0]
			fmt.Fprintf(&sb, "    %s\n", c.cmd(fmt.Sprintf("glacier test ./%s/ -run %s -v", f.Package, f.Test)))
		}
		fmt.Fprintf(&sb, "\n  Run %s for exit-code details.\n", c.cmd("`glacier explain exit:66`"))
	}

	title := "glacier test summary  " + time.Now().Format("15:04:05")
	box := term.Box(
		sb.String(),
		term.WithTitle(title),
		term.WithRoundedCorners(),
		term.WithPadding(1, 2, 1, 2),
	)
	fmt.Fprintln(os.Stderr, box)
}

// summaryColors holds the styled-string functions used by printTestSummary.
// Centralized so the same palette is consistent across pass/fail/skip lines.
type summaryColors struct {
	pass, fail, skip, warn, dim               func(string) string
	pkg, testName, label, numNeutral, heading func(string) string
	cmd                                       func(string) string
}

func newSummaryColors(useColor bool) summaryColors {
	if !useColor {
		id := func(s string) string { return s }
		return summaryColors{pass: id, fail: id, skip: id, warn: id, dim: id, pkg: id, testName: id, label: id, numNeutral: id, heading: id, cmd: id}
	}
	const (
		green   = "\x1b[38;5;84m"
		red     = "\x1b[38;5;203m"
		yellow  = "\x1b[38;5;228m"
		orange  = "\x1b[38;5;215m"
		cyan    = "\x1b[38;5;87m"
		blue    = "\x1b[38;5;75m"
		magenta = "\x1b[38;5;213m"
		bold    = "\x1b[1m"
		dim     = "\x1b[2m"
		under   = "\x1b[4m"
		reset   = "\x1b[0m"
	)
	wrap := func(prefix string) func(string) string {
		return func(s string) string { return prefix + s + reset }
	}
	return summaryColors{
		pass:       wrap(bold + green),
		fail:       wrap(bold + red),
		skip:       wrap(yellow),
		warn:       wrap(orange),
		dim:        wrap(dim),
		pkg:        wrap(blue),
		testName:   wrap(magenta),
		label:      wrap(dim + cyan),
		numNeutral: wrap(cyan),
		heading:    wrap(bold + cyan + under),
		cmd:        wrap(bold + cyan),
	}
}

// --- JSON aggregate ---

// jsonAggregate is the glacier-summary event emitted at the end of --format=json.
type jsonAggregate struct {
	Action      string         `json:"action"`
	Packages    int            `json:"packages"`
	Pass        int            `json:"pass"`
	Fail        int            `json:"fail"`
	Skip        int            `json:"skip"`
	Coverage    float64        `json:"coverage"`
	WallSeconds float64        `json:"wall_seconds"`
	Slowest     []jsonSlowTest `json:"slowest"`
	Failures    []jsonFailure  `json:"failures"`
}

// jsonSlowTest is one entry in the slowest list within jsonAggregate.
type jsonSlowTest struct {
	Package string  `json:"package"`
	Test    string  `json:"test"`
	Elapsed float64 `json:"elapsed"`
}

// jsonFailure is one entry in the failures list within jsonAggregate.
type jsonFailure struct {
	Package string `json:"package"`
	Test    string `json:"test"`
	Output  string `json:"output,omitempty"`
}

// buildJSONAggregate constructs the glacier-summary aggregate event from a Summary.
func buildJSONAggregate(sum Summary) jsonAggregate {
	agg := jsonAggregate{
		Action:      "glacier-summary",
		Packages:    sum.Packages,
		Pass:        sum.Pass,
		Fail:        sum.Fail,
		Skip:        sum.Skip,
		Coverage:    sum.Coverage,
		WallSeconds: sum.WallSeconds,
	}
	for _, s := range sum.Slowest {
		agg.Slowest = append(agg.Slowest, jsonSlowTest{
			Package: s.Package,
			Test:    s.Test,
			Elapsed: s.Elapsed,
		})
	}
	for _, f := range sum.Failures {
		agg.Failures = append(agg.Failures, jsonFailure{
			Package: f.Package,
			Test:    f.Test,
			Output:  f.Output,
		})
	}
	return agg
}

// --- JUnit emitter ---

// junitTestSuites is the root XML element for JUnit output.
type junitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	TestSuites []junitTestSuite `xml:"testsuite"`
}

// junitTestSuite represents one tested Go package.
type junitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Skipped   int             `xml:"skipped,attr"`
	Time      string          `xml:"time,attr"`
	TestCases []junitTestCase `xml:"testcase"`
}

// junitTestCase represents one test function.
type junitTestCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	Classname string        `xml:"classname,attr"`
	Name      string        `xml:"name,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
	Skipped   *junitSkipped `xml:"skipped,omitempty"`
}

// junitFailure carries failure details.
type junitFailure struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

// junitSkipped marks a skipped test case.
type junitSkipped struct{}

// emitJUnit produces standard JUnit XML from a Summary.
// It is a pure function over Summary, with no side effects.
func emitJUnit(sum Summary) ([]byte, error) {
	type pkgAccum struct {
		cases    []junitTestCase
		failures int
		skipped  int
	}
	groups := make(map[string]*pkgAccum)
	var order []string

	// Combine slowest + failures into a unified result set.
	seen := make(map[string]struct{})
	var all []testResult
	for _, r := range sum.Slowest {
		k := r.Package + "/" + r.Test
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			all = append(all, r)
		}
	}
	for _, r := range sum.Failures {
		k := r.Package + "/" + r.Test
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			all = append(all, r)
		}
	}

	for _, r := range all {
		if _, ok := groups[r.Package]; !ok {
			groups[r.Package] = &pkgAccum{}
			order = append(order, r.Package)
		}
		g := groups[r.Package]
		tc := junitTestCase{
			Classname: r.Package,
			Name:      r.Test,
			Time:      fmt.Sprintf("%.3f", r.Elapsed),
		}
		switch r.Action {
		case "fail":
			tc.Failure = &junitFailure{
				Message: fmt.Sprintf("%s/%s failed", r.Package, r.Test),
				Text:    r.Output,
			}
			g.failures++
		case "skip":
			tc.Skipped = &junitSkipped{}
			g.skipped++
		}
		g.cases = append(g.cases, tc)
	}

	suites := junitTestSuites{
		Tests:    sum.Pass + sum.Fail + sum.Skip,
		Failures: sum.Fail,
	}
	for _, pkg := range order {
		g := groups[pkg]
		suites.TestSuites = append(suites.TestSuites, junitTestSuite{
			Name:      pkg,
			Tests:     len(g.cases),
			Failures:  g.failures,
			Skipped:   g.skipped,
			Time:      fmt.Sprintf("%.3f", sum.WallSeconds),
			TestCases: g.cases,
		})
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(suites); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// --- SARIF emitter ---

// sarifRoot is the top-level SARIF 2.1.0 document.
type sarifRoot struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

// sarifRun is one analysis run.
type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

// sarifTool describes the analysis tool.
type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

// sarifDriver identifies the tool and its rules.
type sarifDriver struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Rules   []sarifRule `json:"rules"`
}

// sarifRule defines a rule referenced by results.
type sarifRule struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	ShortDescription sarifText `json:"shortDescription"`
}

// sarifText is a plain-text message wrapper.
type sarifText struct {
	Text string `json:"text"`
}

// sarifResult is one finding (one failing test).
type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Message   sarifText       `json:"message"`
	Locations []sarifLocation `json:"locations,omitempty"`
}

// sarifLocation points to a source artifact.
type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

// sarifPhysicalLocation holds an artifact URI.
type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
}

// sarifArtifactLocation is the relative URI of the source file.
type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

// emitSARIF produces a minimal SARIF 2.1.0 document from a Summary.
// It is a pure function over Summary, with no side effects.
func emitSARIF(sum Summary) ([]byte, error) {
	results := make([]sarifResult, 0, len(sum.Failures))
	for _, f := range sum.Failures {
		results = append(results, sarifResult{
			RuleID:  "test-failure",
			Message: sarifText{Text: fmt.Sprintf("%s/%s failed", f.Package, f.Test)},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{URI: packageToURI(f.Package)},
					},
				},
			},
		})
	}

	doc := sarifRoot{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:    "glacier-test",
						Version: "0.0.0",
						Rules: []sarifRule{
							{
								ID:               "test-failure",
								Name:             "TestFailure",
								ShortDescription: sarifText{Text: "A Go test function failed."},
							},
						},
					},
				},
				Results: results,
			},
		},
	}
	return json.MarshalIndent(doc, "", "  ")
}

// packageToURI converts a Go package path to a relative URI for SARIF.
func packageToURI(pkg string) string {
	parts := strings.Split(pkg, "/")
	return parts[len(parts)-1] + "/"
}

// --- Baseline I/O ---

// writeBaseline serialises bench entries to the baseline file atomically.
func writeBaseline(baselinePath string, entries []benchcmp.BenchEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal baseline: %w", err)
	}
	dir := filepath.Dir(baselinePath)
	name := filepath.Base(baselinePath)
	return safefile.WriteFileAtomic(dir, name, data, 0o644)
}

// loadBaseline reads and decodes the baseline file.
func loadBaseline(baselinePath string) ([]benchcmp.BenchEntry, error) {
	f, err := os.Open(baselinePath) //nolint:gosec // caller-controlled path from CLI flag
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []benchcmp.BenchEntry
	if err := safejson.Decode(f, &entries); err != nil {
		return nil, fmt.Errorf("load baseline: %w", err)
	}
	return entries, nil
}

// --- Real subprocess runner ---

// goTestRunnerImpl is the production TestRunner that shells out to `go test`.
type goTestRunnerImpl struct {
	proc *goRunnerProc
}

// goRunnerProc holds live subprocess state for goTestRunnerImpl.
type goRunnerProc struct {
	eventsCh   chan string
	benchOutCh chan string
	doneCh     chan struct{}
	waitErr    error
}

// Run starts `go test` and fans out JSON lines and non-JSON (bench) lines.
func (r *goTestRunnerImpl) Run(ctx context.Context, args ...string) (<-chan string, <-chan string, error) {
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Env = os.Environ()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("pipe: %w", err)
	}
	cmd.Stderr = os.Stderr

	if startErr := cmd.Start(); startErr != nil {
		return nil, nil, fmt.Errorf("start: %w", startErr)
	}

	proc := &goRunnerProc{
		eventsCh:   make(chan string, 256),
		benchOutCh: make(chan string, 64),
		doneCh:     make(chan struct{}),
	}
	r.proc = proc

	go func() {
		defer close(proc.eventsCh)
		defer close(proc.benchOutCh)
		defer close(proc.doneCh)

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			// go test -json lines always start with '{'.
			if len(line) > 0 && line[0] == '{' {
				proc.eventsCh <- line
			} else if line != "" {
				proc.benchOutCh <- line
			}
		}
		proc.waitErr = cmd.Wait()
	}()

	return proc.eventsCh, proc.benchOutCh, nil
}

// Wait blocks until the subprocess finishes and returns its exit error.
func (r *goTestRunnerImpl) Wait() error {
	if r.proc == nil {
		return nil
	}
	<-r.proc.doneCh
	return r.proc.waitErr
}
