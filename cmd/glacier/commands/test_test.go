// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/benchcmp"
	"github.com/nathanbrophy/glacier/fixture"
	"github.com/nathanbrophy/glacier/term"
)

// fakeRunner is a test double for TestRunner. It feeds canned event lines and
// bench-output lines through channels, simulating a `go test -json` subprocess.
type fakeRunner struct {
	eventLines []string
	benchLines []string
	waitErr    error
}

// Run returns pre-loaded channels immediately.
func (f *fakeRunner) Run(_ context.Context, _ ...string) (<-chan string, <-chan string, error) {
	eventsCh := make(chan string, len(f.eventLines)+1)
	benchOutCh := make(chan string, len(f.benchLines)+1)
	for _, l := range f.eventLines {
		eventsCh <- l
	}
	close(eventsCh)
	for _, l := range f.benchLines {
		benchOutCh <- l
	}
	close(benchOutCh)
	return eventsCh, benchOutCh, nil
}

// Wait returns the configured exit error.
func (f *fakeRunner) Wait() error { return f.waitErr }

// mustMarshal serialises v to JSON or panics.
func mustMarshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// makeEvent builds a canned go test -json event JSON string.
func makeEvent(action, pkg, test string, elapsed float64, output string) string {
	type ev struct {
		Action  string  `json:"Action"`
		Package string  `json:"Package,omitempty"`
		Test    string  `json:"Test,omitempty"`
		Output  string  `json:"Output,omitempty"`
		Elapsed float64 `json:"Elapsed,omitempty"`
	}
	return mustMarshal(ev{Action: action, Package: pkg, Test: test, Elapsed: elapsed, Output: output})
}

// --- parseCoverage ---

func TestParseCoverage(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		lines []string
		want  float64
	}{
		{
			name:  "standard line",
			lines: []string{"coverage: 91.4% of statements"},
			want:  0.914,
		},
		{
			name:  "embedded in ok line",
			lines: []string{"ok  github.com/foo/bar  0.123s  coverage: 75.0% of statements"},
			want:  0.75,
		},
		{
			name:  "no coverage",
			lines: []string{"PASS"},
			want:  0,
		},
		{
			name:  "empty",
			lines: nil,
			want:  0,
		},
		{
			name:  "100 percent",
			lines: []string{"coverage: 100.0% of statements"},
			want:  1.0,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := parseCoverage(tc.lines)
			assert.Equal(t, got, tc.want)
		})
	}
}

// --- buildSummary ---

func TestBuildSummary_Counts(t *testing.T) {
	t.Parallel()
	results := []testResult{
		{Package: "pkg", Test: "TestA", Action: "pass", Elapsed: 0.5},
		{Package: "pkg", Test: "TestB", Action: "fail", Elapsed: 0.2, Output: "boom\n"},
		{Package: "pkg", Test: "TestC", Action: "skip", Elapsed: 0.0},
		{Package: "pkg", Test: "TestD", Action: "pass", Elapsed: 1.2},
	}
	sum := buildSummary(results, 0.9, 3.0, 1, 2)
	assert.Equal(t, sum.Pass, 2)
	assert.Equal(t, sum.Fail, 1)
	assert.Equal(t, sum.Skip, 1)
	assert.Equal(t, sum.Packages, 1)
	assert.Equal(t, sum.Coverage, 0.9)
	assert.Equal(t, len(sum.Failures), 1)
	assert.Equal(t, sum.Failures[0].Test, "TestB")
	// Slowest 2: TestD (1.2s) then TestA (0.5s).
	assert.Equal(t, len(sum.Slowest), 2)
	assert.Equal(t, sum.Slowest[0].Test, "TestD")
}

// --- buildJSONAggregate ---

func TestBuildJSONAggregate(t *testing.T) {
	t.Parallel()
	sum := Summary{
		Packages:    3,
		Pass:        10,
		Fail:        1,
		Skip:        2,
		Coverage:    0.914,
		WallSeconds: 24.3,
		Slowest: []testResult{
			{Package: "pkg/a", Test: "TestSlow", Elapsed: 1.84},
		},
		Failures: []testResult{
			{Package: "pkg/b", Test: "TestBad", Output: "expected x, got y\n"},
		},
	}
	agg := buildJSONAggregate(sum)
	assert.Equal(t, agg.Action, "glacier-summary")
	assert.Equal(t, agg.Packages, 3)
	assert.Equal(t, agg.Pass, 10)
	assert.Equal(t, agg.Fail, 1)
	assert.Equal(t, agg.Skip, 2)
	assert.Equal(t, agg.Coverage, 0.914)
	assert.Equal(t, agg.WallSeconds, 24.3)
	assert.Equal(t, len(agg.Slowest), 1)
	assert.Equal(t, agg.Slowest[0].Package, "pkg/a")
	assert.Equal(t, len(agg.Failures), 1)
	assert.Equal(t, agg.Failures[0].Output, "expected x, got y\n")

	// Round-trip through JSON to verify schema shape.
	b, err := json.Marshal(agg)
	require.NoError(t, err)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(b, &raw))
	assert.Equal(t, raw["action"], "glacier-summary")
}

// --- emitJUnit (golden) ---

func TestEmitJUnit_Golden(t *testing.T) {
	t.Parallel()
	sum := Summary{
		Packages:    1,
		Pass:        2,
		Fail:        1,
		Skip:        0,
		WallSeconds: 1.5,
		Slowest: []testResult{
			{Package: "mypkg", Test: "TestPass1", Action: "pass", Elapsed: 0.4},
			{Package: "mypkg", Test: "TestPass2", Action: "pass", Elapsed: 0.3},
		},
		Failures: []testResult{
			{Package: "mypkg", Test: "TestFail", Action: "fail", Elapsed: 0.1, Output: "got nil, want error\n"},
		},
	}
	got, err := emitJUnit(sum)
	require.NoError(t, err)
	fixture.Golden(t, "junit_basic.xml", got)
}

// --- emitSARIF (golden) ---

func TestEmitSARIF_Golden(t *testing.T) {
	t.Parallel()
	sum := Summary{
		Fail: 2,
		Failures: []testResult{
			{Package: "mypkg/sub", Test: "TestBad", Action: "fail"},
			{Package: "mypkg/other", Test: "TestAlsoBad", Action: "fail"},
		},
	}
	got, err := emitSARIF(sum)
	require.NoError(t, err)
	fixture.Golden(t, "sarif_basic.json", got)
}

// --- JSON aggregate golden ---

func TestJSONAggregate_Golden(t *testing.T) {
	t.Parallel()
	sum := Summary{
		Packages:    2,
		Pass:        5,
		Fail:        1,
		Skip:        1,
		Coverage:    0.85,
		WallSeconds: 10.0,
		Slowest: []testResult{
			{Package: "a", Test: "TestSlow", Elapsed: 2.0},
		},
		Failures: []testResult{
			{Package: "b", Test: "TestFail", Output: "nope\n"},
		},
	}
	agg := buildJSONAggregate(sum)
	b, err := json.MarshalIndent(agg, "", "  ")
	require.NoError(t, err)
	fixture.Golden(t, "aggregate_basic.json", b)
}

// --- buildTestArgs ---

func TestBuildTestArgs(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		cmd      TestCmd
		patterns []string
		want     []string
	}{
		{
			name:     "plain",
			cmd:      TestCmd{},
			patterns: []string{"./..."},
			want:     []string{"test", "-json", "./..."},
		},
		{
			name:     "race and cover",
			cmd:      TestCmd{Race: true, Cover: true},
			patterns: []string{"./pkg/..."},
			want:     []string{"test", "-json", "-race", "-cover", "-coverprofile=.glacier/coverage.out", "./pkg/..."},
		},
		{
			name:     "bench",
			cmd:      TestCmd{Bench: "BenchmarkFoo"},
			patterns: []string{"./..."},
			want:     []string{"test", "-json", "-bench=BenchmarkFoo", "-benchmem", "./..."},
		},
		{
			name:     "fuzz",
			cmd:      TestCmd{Fuzz: "FuzzBar"},
			patterns: []string{"./..."},
			want:     []string{"test", "-json", "-fuzz=FuzzBar", "./..."},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := buildTestArgs(&tc.cmd, tc.patterns)
			assert.Equal(t, got, tc.want)
		})
	}
}

// --- addPanelKey ring buffer ---

func TestAddPanelKey_RingBuffer(t *testing.T) {
	t.Parallel()
	// Fill to capacity using a real StatusBar (no I/O in tests, purely in-memory).
	keys := make([]string, 0, statusPanelMaxRows+2)
	sb := term.NewStatusBar()
	for i := range statusPanelMaxRows {
		addPanelKey(sb, &keys, pkgName(i), "content")
	}
	assert.Equal(t, len(keys), statusPanelMaxRows)

	// Adding one more should evict the oldest.
	addPanelKey(sb, &keys, pkgName(statusPanelMaxRows), "extra")
	assert.Equal(t, len(keys), statusPanelMaxRows)
	assert.Equal(t, keys[0], pkgName(1)) // pkg0 evicted
	assert.Equal(t, keys[statusPanelMaxRows-1], pkgName(statusPanelMaxRows))
}

func TestAddPanelKey_UpdateExisting(t *testing.T) {
	t.Parallel()
	keys := make([]string, 0, 4)
	sb := term.NewStatusBar()
	addPanelKey(sb, &keys, "a", "old")
	addPanelKey(sb, &keys, "b", "x")
	addPanelKey(sb, &keys, "a", "new") // update, not insert
	assert.Equal(t, len(keys), 2)
	assert.Equal(t, keys[0], "a")
}

func TestRemovePanelKey(t *testing.T) {
	t.Parallel()
	keys := []string{"a", "b", "c"}
	removePanelKey(&keys, "b")
	assert.Equal(t, keys, []string{"a", "c"})
	removePanelKey(&keys, "z") // no-op
	assert.Equal(t, len(keys), 2)
}

// pkgName returns a deterministic package name for index i.
func pkgName(i int) string { return "pkg" + string(rune('0'+i)) }

// --- Bench baseline ---

func TestBenchBaseline_WriteAndLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	baselinePath := filepath.Join(dir, ".glacier", "bench-baseline.json")

	entries := []benchcmp.BenchEntry{
		{Name: "BenchmarkFoo-8", NsPerOp: 100, AllocsPerOp: 1, BPerOp: 64},
		{Name: "BenchmarkBar-8", NsPerOp: 200, AllocsPerOp: 0, BPerOp: 0},
	}

	require.NoError(t, writeBaseline(baselinePath, entries))

	loaded, err := loadBaseline(baselinePath)
	require.NoError(t, err)
	assert.Equal(t, len(loaded), 2)
	assert.Equal(t, loaded[0].Name, "BenchmarkFoo-8")
	assert.Equal(t, loaded[0].NsPerOp, 100.0)
	assert.Equal(t, loaded[1].NsPerOp, 200.0)
}

func TestBenchBaseline_Shape(t *testing.T) {
	t.Parallel()
	// Confirm the on-disk JSON shape matches the spec (array of objects with
	// name/ns_per_op/allocs_per_op/b_per_op fields).
	dir := t.TempDir()
	baselinePath := filepath.Join(dir, "bench-baseline.json")
	entries := []benchcmp.BenchEntry{
		{Name: "BenchmarkX-8", NsPerOp: 42.5, AllocsPerOp: 2, BPerOp: 128},
	}
	require.NoError(t, writeBaseline(baselinePath, entries))

	raw, err := os.ReadFile(baselinePath)
	require.NoError(t, err)
	var parsed []map[string]any
	require.NoError(t, json.Unmarshal(raw, &parsed))
	assert.Equal(t, len(parsed), 1)
	assert.Equal(t, parsed[0]["name"], "BenchmarkX-8")
	_, hasNs := parsed[0]["ns_per_op"]
	assert.Equal(t, hasNs, true)
	_, hasAllocs := parsed[0]["allocs_per_op"]
	assert.Equal(t, hasAllocs, true)
	_, hasB := parsed[0]["b_per_op"]
	assert.Equal(t, hasB, true)
}

// --- Bench regression detection ---

func TestBenchRegression_NoRegression(t *testing.T) {
	t.Parallel()
	baseline := []benchcmp.BenchEntry{
		{Name: "BenchmarkFoo-8", NsPerOp: 100},
	}
	current := []benchcmp.BenchEntry{
		{Name: "BenchmarkFoo-8", NsPerOp: 104}, // 4% slower :  within 5% threshold
	}
	regressions := benchcmp.Compare(baseline, current)
	assert.Equal(t, len(regressions), 0)
}

func TestBenchRegression_JustAtThreshold(t *testing.T) {
	t.Parallel()
	baseline := []benchcmp.BenchEntry{
		{Name: "BenchmarkFoo-8", NsPerOp: 100},
	}
	current := []benchcmp.BenchEntry{
		{Name: "BenchmarkFoo-8", NsPerOp: 105}, // exactly 5% :  at threshold, not over
	}
	regressions := benchcmp.Compare(baseline, current)
	assert.Equal(t, len(regressions), 0)
}

func TestBenchRegression_OverThreshold(t *testing.T) {
	t.Parallel()
	baseline := []benchcmp.BenchEntry{
		{Name: "BenchmarkFoo-8", NsPerOp: 100},
	}
	current := []benchcmp.BenchEntry{
		{Name: "BenchmarkFoo-8", NsPerOp: 106}, // 6% slower :  regression
	}
	regressions := benchcmp.Compare(baseline, current)
	assert.Equal(t, len(regressions), 1)
	assert.Equal(t, regressions[0].Name, "BenchmarkFoo-8")
	if regressions[0].DeltaPct <= 5.0 {
		t.Errorf("expected DeltaPct > 5.0, got %f", regressions[0].DeltaPct)
	}
}

func TestBenchRegression_NewBenchNotInBaseline(t *testing.T) {
	t.Parallel()
	// Benchmarks absent from baseline are skipped; no regression reported.
	baseline := []benchcmp.BenchEntry{
		{Name: "BenchmarkOld-8", NsPerOp: 100},
	}
	current := []benchcmp.BenchEntry{
		{Name: "BenchmarkNew-8", NsPerOp: 9999},
	}
	regressions := benchcmp.Compare(baseline, current)
	assert.Equal(t, len(regressions), 0)
}

func TestBenchRegression_MultipleRegressions(t *testing.T) {
	t.Parallel()
	baseline := []benchcmp.BenchEntry{
		{Name: "BenchmarkA-8", NsPerOp: 100},
		{Name: "BenchmarkB-8", NsPerOp: 200},
		{Name: "BenchmarkC-8", NsPerOp: 50},
	}
	current := []benchcmp.BenchEntry{
		{Name: "BenchmarkA-8", NsPerOp: 120}, // 20% :  regression
		{Name: "BenchmarkB-8", NsPerOp: 202}, // 1% :  ok
		{Name: "BenchmarkC-8", NsPerOp: 60},  // 20% :  regression
	}
	regressions := benchcmp.Compare(baseline, current)
	assert.Equal(t, len(regressions), 2)
}

func TestBenchRegression_Improvement(t *testing.T) {
	t.Parallel()
	baseline := []benchcmp.BenchEntry{
		{Name: "BenchmarkFoo-8", NsPerOp: 100},
	}
	current := []benchcmp.BenchEntry{
		{Name: "BenchmarkFoo-8", NsPerOp: 50}, // 50% faster :  no regression
	}
	regressions := benchcmp.Compare(baseline, current)
	assert.Equal(t, len(regressions), 0)
}

// --- UpdateBaseline via Run ---

func TestRun_UpdateBaseline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	baselinePath := filepath.Join(dir, "bench-baseline.json")

	benchLine := "BenchmarkFoo-8   1000   150.0 ns/op   64 B/op   2 allocs/op"

	runner := &fakeRunner{
		eventLines: []string{
			makeEvent("run", "mypkg", "TestX", 0, ""),
			makeEvent("pass", "mypkg", "TestX", 0.1, ""),
			makeEvent("pass", "mypkg", "", 0.1, ""),
		},
		benchLines: []string{benchLine},
	}

	cmd := &TestCmd{
		Patterns:       []string{"./..."},
		Bench:          "BenchmarkFoo",
		Baseline:       baselinePath,
		UpdateBaseline: true,
		Format:         "text",
		Slowest:        5,
		NoStatus:       true,
		runner:         runner,
	}

	err := cmd.Run(context.Background())
	require.NoError(t, err)

	loaded, loadErr := loadBaseline(baselinePath)
	require.NoError(t, loadErr)
	assert.Equal(t, len(loaded), 1)
	assert.Equal(t, loaded[0].Name, "BenchmarkFoo-8")
	assert.Equal(t, loaded[0].NsPerOp, 150.0)
}

// --- Run with format=json emits glacier-summary ---

// Not t.Parallel: the test rebinds the process-wide os.Stdout to a pipe
// while TestCmd.Run runs. Other tests in this package that swap os.Stdout
// (lint_test.TestExitCodeStability, cross_cutting.TestBannerSuppressedOnSubcommands)
// would race here if any ran in parallel.
func TestRun_FormatJSON_AggregateEmitted(t *testing.T) {
	runner := &fakeRunner{
		eventLines: []string{
			makeEvent("run", "mypkg", "TestA", 0, ""),
			makeEvent("pass", "mypkg", "TestA", 0.2, ""),
			makeEvent("pass", "mypkg", "", 0.2, ""),
		},
	}

	// Capture stdout.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &TestCmd{
		Patterns: []string{"./..."},
		Format:   "json",
		Slowest:  5,
		NoStatus: true,
		runner:   runner,
	}
	err := cmd.Run(context.Background())

	w.Close()
	os.Stdout = oldStdout

	require.NoError(t, err)

	var buf strings.Builder
	tmp := make([]byte, 4096)
	for {
		n, readErr := r.Read(tmp)
		if n > 0 {
			buf.Write(tmp[:n])
		}
		if readErr != nil {
			break
		}
	}
	r.Close()

	output := buf.String()
	// Last non-empty line should be the glacier-summary aggregate.
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	lastLine := lines[len(lines)-1]
	var agg map[string]any
	require.NoError(t, json.Unmarshal([]byte(lastLine), &agg))
	assert.Equal(t, agg["action"], "glacier-summary")
	_, hasPackages := agg["packages"]
	assert.Equal(t, hasPackages, true)
}

// --- Run with failing tests returns exit 66 error ---

func TestRun_FailureReturnsExitCode66(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{
		eventLines: []string{
			makeEvent("run", "mypkg", "TestBad", 0, ""),
			makeEvent("fail", "mypkg", "TestBad", 0.1, ""),
			makeEvent("fail", "mypkg", "", 0.1, ""),
		},
	}

	cmd := &TestCmd{
		Patterns: []string{"./..."},
		Format:   "text",
		Slowest:  5,
		NoStatus: true,
		runner:   runner,
	}

	err := cmd.Run(context.Background())
	assert.Error(t, err)
	var ec *exitCodeError
	assert.ErrorAs(t, err, &ec)
	assert.Equal(t, ec.code, exitTestsFailed)
}

// --- Run with bench regression returns exit 66 ---

func TestRun_BenchRegression_ExitCode66(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	baselinePath := filepath.Join(dir, "bench-baseline.json")

	// Write a baseline with 100 ns/op.
	require.NoError(t, writeBaseline(baselinePath, []benchcmp.BenchEntry{
		{Name: "BenchmarkFoo-8", NsPerOp: 100},
	}))

	// Fake runner reports 200 ns/op (100% regression).
	runner := &fakeRunner{
		eventLines: []string{
			makeEvent("pass", "mypkg", "", 0.1, ""),
		},
		benchLines: []string{
			"BenchmarkFoo-8   1000   200.0 ns/op   0 B/op   0 allocs/op",
		},
	}

	cmd := &TestCmd{
		Patterns: []string{"./..."},
		Bench:    "BenchmarkFoo",
		Baseline: baselinePath,
		Format:   "text",
		Slowest:  5,
		NoStatus: true,
		runner:   runner,
	}

	err := cmd.Run(context.Background())
	assert.Error(t, err)
	var ec *exitCodeError
	assert.ErrorAs(t, err, &ec)
	assert.Equal(t, ec.code, exitTestsFailed)
}

// --- Unknown format returns exit 2 ---

func TestRun_UnknownFormat(t *testing.T) {
	t.Parallel()

	cmd := &TestCmd{
		Format:   "csv",
		Slowest:  5,
		NoStatus: true,
		runner:   &fakeRunner{},
	}

	err := cmd.Run(context.Background())
	assert.Error(t, err)
	var ec *exitCodeError
	assert.ErrorAs(t, err, &ec)
	assert.Equal(t, ec.code, exitUsage)
}

// --- benchcmp.Parse ---

func TestBenchcmpParse(t *testing.T) {
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
}

// --- packageToURI ---

func TestPackageToURI(t *testing.T) {
	t.Parallel()
	cases := []struct {
		pkg  string
		want string
	}{
		{"github.com/foo/bar/pkg", "pkg/"},
		{"mypkg", "mypkg/"},
		{"a/b/c/d", "d/"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.pkg, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, packageToURI(tc.pkg), tc.want)
		})
	}
}
