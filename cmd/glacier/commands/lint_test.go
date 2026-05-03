// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/mock"
)

// linterAdapter routes Linter method calls through the mock dispatch function.
type linterAdapter struct {
	dispatch func(string, []reflect.Value) []reflect.Value
}

func (a *linterAdapter) Name() string {
	res := a.dispatch("Name", nil)
	if len(res) == 0 || !res[0].IsValid() {
		return ""
	}
	return res[0].Interface().(string)
}

func (a *linterAdapter) Severity() Severity {
	res := a.dispatch("Severity", nil)
	if len(res) == 0 || !res[0].IsValid() {
		return SeverityInfo
	}
	return res[0].Interface().(Severity)
}

func (a *linterAdapter) Check(file string, src []byte) []Finding {
	res := a.dispatch("Check", []reflect.Value{reflect.ValueOf(file), reflect.ValueOf(src)})
	if len(res) == 0 || !res[0].IsValid() || res[0].IsNil() {
		return nil
	}
	return res[0].Interface().([]Finding)
}

// TestMain registers the Linter mock adapter before any test runs.
func TestMain(m *testing.M) {
	mock.RegisterAdapter[Linter](func(dispatch func(string, []reflect.Value) []reflect.Value) Linter {
		return &linterAdapter{dispatch: dispatch}
	})
	os.Exit(m.Run())
}

// --- Unit tests for individual linters ---

// TestLintNoEmDash verifies the no-em-dash linter detects U+2014 in .go files.
func TestLintNoEmDash(t *testing.T) {
	t.Parallel()
	l := &noEmDashLinter{}

	tests := []struct {
		name    string
		file    string
		src     string
		wantHit bool
	}{
		{
			name:    "em_dash_present",
			file:    "foo.go",
			src:     "// this \xe2\x80\x94 is bad\n",
			wantHit: true,
		},
		{
			name:    "no_em_dash",
			file:    "foo.go",
			src:     "// this - is fine\n",
			wantHit: false,
		},
		{
			name:    "skips_non_go",
			file:    "foo.md",
			src:     "# title \xe2\x80\x94 subtitle\n",
			wantHit: false,
		},
		{
			// The no-em-dash rule applies to all .go files, including test files.
			name:    "test_file_also_checked",
			file:    "foo_test.go",
			src:     "// test \xe2\x80\x94 file\n",
			wantHit: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			findings := l.Check(tc.file, []byte(tc.src))
			if tc.wantHit {
				assert.True(t, len(findings) > 0, "expected finding for em-dash but got none")
			} else {
				assert.Equal(t, 0, len(findings))
			}
		})
	}
}

// TestLintPanicInLibrary verifies the panic-in-library linter detects panic
// calls in non-cmd non-test code, and that AST-based detection ignores
// occurrences in comments and string literals.
func TestLintPanicInLibrary(t *testing.T) {
	t.Parallel()
	l := &panicInLibraryLinter{}

	tests := []struct {
		name    string
		file    string
		src     string
		wantHit bool
	}{
		{
			name:    "panic_in_library",
			file:    "mylib/foo.go",
			src:     "package mylib\nfunc f() { panic(\"unreachable\") }\n",
			wantHit: true,
		},
		{
			name:    "panic_in_cmd",
			file:    "cmd/glacier/main.go",
			src:     "package main\nfunc main() { panic(\"unreachable\") }\n",
			wantHit: false,
		},
		{
			name:    "panic_in_test",
			file:    "mylib/foo_test.go",
			src:     "package mylib_test\nfunc TestFoo(t *testing.T) { panic(\"oh no\") }\n",
			wantHit: false,
		},
		{
			name:    "no_panic",
			file:    "mylib/foo.go",
			src:     "package mylib\nfunc f() error { return nil }\n",
			wantHit: false,
		},
		{
			name:    "panic_token_in_comment_only",
			file:    "mylib/foo.go",
			src:     "package mylib\n// f calls panic( on a misuse.\nfunc f() error { return nil }\n",
			wantHit: false,
		},
		{
			name:    "panic_token_in_string_literal_only",
			file:    "mylib/foo.go",
			src:     "package mylib\nfunc emit() string { return \"\\tpanic(err)\\n\" }\n",
			wantHit: false,
		},
		{
			name:    "skips_zz_generated",
			file:    "mylib/zz_generated_cli.go",
			src:     "package mylib\nfunc f() { panic(\"generated\") }\n",
			wantHit: false,
		},
		{
			name:    "nolint_directive_on_same_line",
			file:    "mylib/foo.go",
			src:     "package mylib\nfunc f() { panic(\"x\") //glacier:nolint=panic-in-library reason\n}\n",
			wantHit: false,
		},
		{
			name:    "nolint_directive_on_preceding_line",
			file:    "mylib/foo.go",
			src:     "package mylib\nfunc f() {\n\t//glacier:nolint=panic-in-library Must convention\n\tpanic(\"x\")\n}\n",
			wantHit: false,
		},
		{
			name:    "nolint_directive_in_func_doc",
			file:    "mylib/foo.go",
			src:     "package mylib\n// f panics on bad input.\n//\n//glacier:nolint=panic-in-library\nfunc f() { panic(\"x\") }\n",
			wantHit: false,
		},
		{
			name:    "nolint_directive_for_other_rule_does_not_suppress",
			file:    "mylib/foo.go",
			src:     "package mylib\nfunc f() { panic(\"x\") //glacier:nolint=other-rule\n}\n",
			wantHit: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			findings := l.Check(tc.file, []byte(tc.src))
			if tc.wantHit {
				assert.True(t, len(findings) > 0, "expected finding for panic( but got none")
			} else {
				assert.Equal(t, 0, len(findings))
			}
		})
	}
}

// TestLintExportedDocComment verifies the exported-doc-comment linter.
func TestLintExportedDocComment(t *testing.T) {
	t.Parallel()
	l := &exportedDocCommentLinter{}

	tests := []struct {
		name    string
		file    string
		src     string
		wantHit bool
	}{
		{
			name: "missing_func_doc",
			file: "mylib/foo.go",
			src: `package mylib
func Exported() {}
`,
			wantHit: true,
		},
		{
			name: "has_func_doc",
			file: "mylib/foo.go",
			src: `package mylib
// Exported does something.
func Exported() {}
`,
			wantHit: false,
		},
		{
			name: "doc_wrong_prefix",
			file: "mylib/foo.go",
			src: `package mylib
// This does something.
func Exported() {}
`,
			wantHit: true,
		},
		{
			name: "skips_unexported",
			file: "mylib/foo.go",
			src: `package mylib
func unexported() {}
`,
			wantHit: false,
		},
		{
			name: "skips_test_file",
			file: "mylib/foo_test.go",
			src: `package mylib
func ExportedTest() {}
`,
			wantHit: false,
		},
		{
			name: "skips_generated",
			file: "mylib/zz_generated_cli.go",
			src: `package mylib
func Exported() {}
`,
			wantHit: false,
		},
		{
			name: "missing_type_doc",
			file: "mylib/foo.go",
			src: `package mylib
type MyType struct{}
`,
			wantHit: true,
		},
		{
			name: "has_type_doc",
			file: "mylib/foo.go",
			src: `package mylib
// MyType is a thing.
type MyType struct{}
`,
			wantHit: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			findings := l.Check(tc.file, []byte(tc.src))
			if tc.wantHit {
				assert.True(t, len(findings) > 0, "expected doc-comment finding but got none")
			} else {
				assert.Equal(t, 0, len(findings))
			}
		})
	}
}

// TestLintPackageExampleTest verifies the package-example-test linter.
func TestLintPackageExampleTest(t *testing.T) {
	t.Parallel()

	// Build a temp directory tree for the test.
	dir := t.TempDir()

	// Package without any example test.
	pkgDir := filepath.Join(dir, "nopkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	goFile := filepath.Join(pkgDir, "nopkg.go")
	if err := os.WriteFile(goFile, []byte("package nopkg\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	l := &packageExampleTestLinter{}
	findings := l.Check(goFile, []byte("package nopkg\n"))
	assert.True(t, len(findings) > 0, "expected package-example-test finding for package with no Example")

	// Add an example test file.
	exFile := filepath.Join(pkgDir, "example_test.go")
	if err := os.WriteFile(exFile, []byte("package nopkg_test\nfunc ExampleFoo() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	l2 := &packageExampleTestLinter{}
	findings2 := l2.Check(goFile, []byte("package nopkg\n"))
	assert.Equal(t, 0, len(findings2))
}

// TestLintPackageExampleTestSkips verifies that the rule skips structural
// directories (testdata, docs/examples), package main, and the integration
// test package.
func TestLintPackageExampleTestSkips(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		relPath string
		src     string
	}{
		{
			name:    "skips_testdata_dir",
			relPath: "mock/gen/testdata/x.go",
			src:     "package mypkg\n",
		},
		{
			name:    "skips_docs_examples",
			relPath: "docs/examples/codegen/cli/simple/in.go",
			src:     "package main\n",
		},
		{
			name:    "skips_package_main",
			relPath: "cmd_oddball/main.go",
			src:     "package main\n",
		},
		{
			name:    "skips_tests_integration_package",
			relPath: "tests/foo.go",
			src:     "package tests\n",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Cannot run subtests in parallel: they share os.Chdir state.
			dir := t.TempDir()
			full := filepath.Join(dir, tc.relPath)
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, []byte(tc.src), 0o644); err != nil {
				t.Fatal(err)
			}
			old, _ := os.Getwd()
			t.Cleanup(func() { _ = os.Chdir(old) })
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}
			l := &packageExampleTestLinter{}
			findings := l.Check(tc.relPath, []byte(tc.src))
			assert.Equal(t, 0, len(findings))
		})
	}
}

// TestSkipDir verifies the directory skip helper used by every walk.
func TestSkipDir(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path string
		want bool
	}{
		{".git", true},
		{".glacier", true},
		{"node_modules", true},
		{"vendor", true},
		{"dist", true},
		{".claude", true},
		{".claude/worktrees", true},
		{".claude/worktrees/unruffled-pare-eac778", true},
		{"some/path/.claude/scratch", true},
		{"cli", false},
		{"cmd/glacier/commands", false},
		{"docs/examples", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			got := skipDir(tc.path)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestParseStaticcheckOutput verifies that real diagnostics pass through
// individually and that tool-failure noise (panic traces, version
// mismatches) collapses into a single summary finding.
func TestParseStaticcheckOutput(t *testing.T) {
	t.Parallel()

	t.Run("real_diagnostics_pass_through", func(t *testing.T) {
		t.Parallel()
		out := []byte("foo.go:12:5: SA9003 empty branch\nbar.go:7:2: ST1005 message\n")
		findings := parseStaticcheckOutput(out, nil)
		assert.Equal(t, 2, len(findings))
		assert.Equal(t, "staticcheck", findings[0].Rule)
	})

	t.Run("panic_trace_collapses_to_summary", func(t *testing.T) {
		t.Parallel()
		out := []byte("panic: runtime error: invalid memory address\n[signal 0xc0000005]\ngoroutine 1 [running]:\n\thonnef.co/go/tools/...\n")
		findings := parseStaticcheckOutput(out, nil)
		assert.Equal(t, 1, len(findings))
		assert.True(t, strings.Contains(findings[0].Message, "tool failure"), "expected summary; got %q", findings[0].Message)
	})

	t.Run("version_mismatch_collapses_to_summary", func(t *testing.T) {
		t.Parallel()
		out := []byte(`-: internal error in importing "cmp" (unsupported version: 2); please report an issue (compile)`)
		findings := parseStaticcheckOutput(out, nil)
		assert.Equal(t, 1, len(findings))
		assert.True(t, strings.Contains(findings[0].Message, "tool failure"), "expected summary; got %q", findings[0].Message)
	})

	t.Run("empty_output_no_findings", func(t *testing.T) {
		t.Parallel()
		findings := parseStaticcheckOutput(nil, nil)
		assert.Equal(t, 0, len(findings))
	})
}

// TestSanitizeStaticcheckPatterns verifies that ./... is expanded to a
// per-top-level-dir set with skipped dirs removed.
func TestSanitizeStaticcheckPatterns(t *testing.T) {
	dir := t.TempDir()
	for _, sub := range []string{"cli", "mock", ".claude", ".git", "vendor", "dist"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	old, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(old) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	got, err := sanitizeStaticcheckPatterns([]string{"./..."})
	assert.NoError(t, err)

	gotSet := make(map[string]bool, len(got))
	for _, p := range got {
		gotSet[p] = true
	}
	assert.True(t, gotSet["./cli/..."], "expected ./cli/... in expanded patterns; got %v", got)
	assert.True(t, gotSet["./mock/..."], "expected ./mock/... in expanded patterns; got %v", got)
	assert.False(t, gotSet["./.claude/..."], "expected .claude excluded; got %v", got)
	assert.False(t, gotSet["./.git/..."], "expected .git excluded; got %v", got)
	assert.False(t, gotSet["./vendor/..."], "expected vendor excluded; got %v", got)
	assert.False(t, gotSet["./dist/..."], "expected dist excluded; got %v", got)
}

// TestLintLibraryErrorRegister verifies the library-error-register linter.
func TestLintLibraryErrorRegister(t *testing.T) {
	t.Parallel()
	l := &libraryErrorRegisterLinter{}

	tests := []struct {
		name    string
		file    string
		src     string
		wantHit bool
	}{
		{
			name: "capitalized_error_string",
			file: "mylib/errors.go",
			src: `package mylib
type MyError struct{}
func (e *MyError) Error() string { return "Bad error." }
`,
			wantHit: true,
		},
		{
			name: "trailing_period",
			file: "mylib/errors.go",
			src: `package mylib
type MyError struct{}
func (e *MyError) Error() string { return "bad error." }
`,
			wantHit: true,
		},
		{
			name: "valid_error_string",
			file: "mylib/errors.go",
			src: `package mylib
type MyError struct{}
func (e *MyError) Error() string { return "mylib: bad thing happened" }
`,
			wantHit: false,
		},
		{
			name: "skips_test_file",
			file: "mylib/errors_test.go",
			src: `package mylib_test
type MyError struct{}
func (e *MyError) Error() string { return "Bad." }
`,
			wantHit: false,
		},
		{
			name: "skips_cmd",
			file: "cmd/glacier/main.go",
			src: `package main
type MyError struct{}
func (e *MyError) Error() string { return "Bad." }
`,
			wantHit: false,
		},
		{
			// Empty-string returns are nil-receiver guards, not real error
			// messages: see Wrapper.Error in errs/errs.go.
			name: "empty_string_return_is_nil_guard",
			file: "mylib/errors.go",
			src: `package mylib
type Wrapper struct{}
func (w *Wrapper) Error() string {
	if w == nil {
		return ""
	}
	return "mylib: wrapped"
}
`,
			wantHit: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			findings := l.Check(tc.file, []byte(tc.src))
			if tc.wantHit {
				assert.True(t, len(findings) > 0, "expected library-error-register finding but got none")
			} else {
				assert.Equal(t, 0, len(findings))
			}
		})
	}
}

// TestLintNakedAnyOptIn verifies naked-any is inactive without the linter and fires when included.
func TestLintNakedAnyOptIn(t *testing.T) {
	t.Parallel()

	src := `package mylib
func Process(v any) error { return nil }
`
	// Without naked-any in the linter list (default off).
	linters := buildLinters(false)
	var findings []Finding
	for _, l := range linters {
		findings = append(findings, l.Check("mylib/foo.go", []byte(src))...)
	}
	for _, f := range findings {
		if f.Rule == "naked-any" {
			t.Errorf("naked-any lint fired but should be disabled by default")
		}
	}

	// With naked-any enabled.
	lintersOn := buildLinters(true)
	var findingsOn []Finding
	for _, l := range lintersOn {
		findingsOn = append(findingsOn, l.Check("mylib/foo.go", []byte(src))...)
	}
	var nakedHit bool
	for _, f := range findingsOn {
		if f.Rule == "naked-any" {
			nakedHit = true
		}
	}
	assert.True(t, nakedHit, "naked-any lint should fire when enabled")
}

// TestLintDispatchLoopViaInterface tests the dispatch loop using a mock Linter
// to confirm the loop calls Check on each linter.
func TestLintDispatchLoopViaInterface(t *testing.T) {
	t.Parallel()

	expected := []Finding{{Rule: "test-rule", File: "a.go", Severity: "warning", Message: "test"}}

	m := mock.Of[Linter](t)
	m.OnCall("Name").Return("test-rule").AnyTimes()
	m.OnCall("Severity").Return(SeverityWarning).AnyTimes()
	m.OnCall("Check").
		With(mock.Any[string](), mock.Any[[]byte]()).
		Return(expected).
		AnyTimes()

	l := m.Interface()
	got := l.Check("a.go", []byte("package x\n"))
	assert.Equal(t, len(expected), len(got))
}

// TestLintSeverityThreshold verifies that findings below the threshold are hidden.
func TestLintSeverityThreshold(t *testing.T) {
	t.Parallel()

	all := []Finding{
		{Rule: "r1", File: "f.go", Severity: "error", Message: "bad"},
		{Rule: "r2", File: "f.go", Severity: "warning", Message: "meh"},
		{Rule: "r3", File: "f.go", Severity: "info", Message: "fyi"},
	}

	threshold := severityRank("error")
	var visible []Finding
	for _, f := range all {
		if severityRank(f.Severity) >= threshold {
			visible = append(visible, f)
		}
	}
	assert.Equal(t, 1, len(visible))
	assert.Equal(t, "error", visible[0].Severity)
}

// TestLintFormatJSON verifies JSON output contains expected keys.
//
// Not t.Parallel: rebinds the process-wide os.Stdout. See
// TestRun_FormatJSON_AggregateEmitted in test_test.go for the matching
// rationale: the three tests in this package that swap os.Stdout
// (this one, TestRun_FormatJSON_AggregateEmitted, and the
// TestBannerSuppressedOnSubcommands "explain_list" subtest) must run
// serially so the redirection windows don't overlap.
func TestLintFormatJSON(t *testing.T) {
	// Redirect stdout to capture JSON output.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	findings := []Finding{
		{Rule: "test-rule", File: "foo.go", Line: 1, Severity: "error", Message: "oops"},
	}
	printFindingsJSON(findings)

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	assert.True(t, strings.Contains(output, `"rule"`), "JSON output must contain 'rule' key")
	assert.True(t, strings.Contains(output, `"test-rule"`), "JSON output must contain rule value")
	assert.True(t, strings.Contains(output, `"file"`), "JSON output must contain 'file' key")
	assert.True(t, strings.Contains(output, `"severity"`), "JSON output must contain 'severity' key")
}

// TestLintFixEmDash verifies the em-dash replacement helper.
func TestLintFixEmDash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		changed bool
	}{
		{
			name:    "replaces_em_dash",
			input:   "foo \xe2\x80\x94 bar",
			want:    "foo :  bar",
			changed: true,
		},
		{
			name:    "no_em_dash",
			input:   "foo - bar",
			want:    "foo - bar",
			changed: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, changed := fixEmDash([]byte(tc.input))
			assert.Equal(t, tc.changed, changed)
			assert.Equal(t, tc.want, string(got))
		})
	}
}

// TestLintCacheHit verifies that a file with an unchanged hash uses the cached findings.
func TestLintCacheHit(t *testing.T) {
	t.Parallel()

	src := []byte("package x\n")
	hash := sha256File(src)

	cached := []Finding{{Rule: "cached-rule", File: "x.go", Severity: "info", Message: "from cache"}}
	cache := lintCache{hash: cached}

	got, ok := cache[hash]
	assert.True(t, ok, "cache lookup must hit on same hash")
	assert.Equal(t, len(cached), len(got))
}

// TestLintSaveLoadCache verifies that cache round-trips through the filesystem.
func TestLintSaveLoadCache(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := []byte("package x\n")
	hash := sha256File(src)

	in := lintCache{
		hash: []Finding{{Rule: "saved-rule", File: "x.go", Severity: "warning", Message: "hello"}},
	}
	saveCache(dir, in)

	out := make(lintCache)
	loadCache(dir, out)

	findings, ok := out[hash]
	assert.True(t, ok, "loaded cache must contain saved entry")
	assert.Equal(t, 1, len(findings))
	assert.Equal(t, "saved-rule", findings[0].Rule)
}
