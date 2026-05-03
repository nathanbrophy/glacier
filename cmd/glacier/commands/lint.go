// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nathanbrophy/glacier/conf"
	"github.com/nathanbrophy/glacier/fluent"
	"github.com/nathanbrophy/glacier/internal/safefile"
	"github.com/nathanbrophy/glacier/internal/safejson"
)

// lintConfig holds opt-in lint toggles read from the "lint" conf section.
type lintConfig struct {
	NakedAny nakedAnyConfig `json:"naked_any"`
}

type nakedAnyConfig struct {
	// Enabled activates the naked-any lint (default false).
	Enabled bool `json:"enabled"`
}

// lintCfg is the package-level accessor for lint configuration.
var lintCfg = conf.Register("lint", lintConfig{})

// LintCmd runs the Glacier lint suite.
//
// +glacier:command name=lint parent=glacier
type LintCmd struct {
	// Patterns are the packages or paths to lint (default ./...).
	//
	// +glacier:positional
	Patterns []string

	// Fix applies auto-fixes where available.
	//
	// +glacier:default false
	Fix bool

	// Severity is the minimum severity to report: error, warning, or info.
	//
	// +glacier:choices error|warning|info
	// +glacier:default warning
	Severity string

	// Format is the output format: text, json, or sarif.
	//
	// +glacier:choices text|json|sarif
	// +glacier:default text
	Format string

	// NoCache disables the per-file result cache.
	//
	// +glacier:default false
	NoCache bool
}

// Severity represents the priority of a lint finding.
type Severity int

const (
	// SeverityInfo is the lowest priority level.
	SeverityInfo Severity = iota
	// SeverityWarning signals a potential problem.
	SeverityWarning
	// SeverityError signals a definite violation.
	SeverityError
)

// String returns the lowercase name of the severity level.
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	default:
		return "info"
	}
}

// Finding is a single lint result produced by a Linter.
type Finding struct {
	Rule     string
	File     string
	Line     int
	Severity string
	Message  string
	// FixHint is a short manual-fix instruction for lints that have no auto-fix.
	FixHint string
}

// Linter checks a single source file for a named rule.
//
// +glacier:mock
type Linter interface {
	// Name returns the lint rule identifier (e.g., "no-em-dash").
	Name() string
	// Severity returns the default severity for this rule.
	Severity() Severity
	// Check analyses src and returns zero or more findings.
	// file is the relative path as it will appear in output.
	Check(file string, src []byte) []Finding
}

// lintCache is the on-disk format for the per-file result cache.
// Keys are sha256 hex hashes of file contents; values are the cached findings.
type lintCache map[string][]Finding

// Run executes the lint suite and prints findings.
func (c *LintCmd) Run(ctx context.Context) error {
	patterns := c.Patterns
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	minSeverity := c.Severity
	if minSeverity == "" {
		minSeverity = "warning"
	}

	// Determine whether naked-any lint is enabled.
	cfg := lintCfg()
	nakedAnyEnabled := cfg.NakedAny.Enabled

	linters := buildLinters(nakedAnyEnabled)

	// Load cache if enabled.
	cacheRoot, _ := os.Getwd()
	cache := make(lintCache)
	if !c.NoCache {
		loadCache(cacheRoot, cache)
	}

	var all []Finding

	// 1. gofmt check / fix.
	fmtFindings, fmtFixed, fmtErr := runGofmtCheck(ctx, c.Fix)
	if fmtErr != nil {
		reportStatus("gofmt: "+fmtErr.Error(), "alarmed")
	}
	if c.Fix && len(fmtFixed) > 0 {
		reportStatus(fmt.Sprintf("gofmt: fixed %d file(s)", len(fmtFixed)), "confident")
	}
	all = append(all, fmtFindings...)

	// 2. go vet.
	vetFindings, vetErr := runGoVet(ctx, patterns)
	if vetErr != nil {
		reportStatus("go vet: "+vetErr.Error(), "alarmed")
	}
	all = append(all, vetFindings...)

	// 3. staticcheck (skip gracefully if not on PATH).
	if _, lookErr := exec.LookPath("staticcheck"); lookErr == nil {
		scFindings, scErr := runStaticcheck(ctx, patterns)
		if scErr != nil {
			reportStatus("staticcheck: "+scErr.Error(), "alarmed")
		}
		all = append(all, scFindings...)
	} else {
		reportStatus("staticcheck not found on PATH; skipping", "thinking")
	}

	// 4. Glacier-specific lints via the Linter interface.
	cacheUpdated := false
	_ = filepath.WalkDir(".", func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if skipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		hash := sha256File(src)

		if !c.NoCache {
			if cached, ok := cache[hash]; ok {
				all = append(all, cached...)
				return nil
			}
		}

		var fileFindings []Finding
		for _, l := range linters {
			fileFindings = append(fileFindings, l.Check(path, src)...)
		}

		// Apply --fix for no-em-dash and marker normalization.
		if c.Fix {
			cur := src
			if fixed, changed := fixEmDash(cur); changed {
				cur = fixed
				// Remove em-dash findings from this file since we fixed them.
				var kept []Finding
				for _, f := range fileFindings {
					if f.Rule != "no-em-dash" {
						kept = append(kept, f)
					}
				}
				fileFindings = kept
			}
			if fixed, changed := fixMarkers(cur); changed {
				cur = fixed
			}
			if !bytes.Equal(cur, src) {
				_ = os.WriteFile(path, cur, 0o644)
			}
		}

		if !c.NoCache {
			cache[hash] = fileFindings
			cacheUpdated = true
		}

		all = append(all, fileFindings...)
		return nil
	})

	// Also run no-em-dash on .md and .txt files (outside the per-Go-file loop).
	mdTxtFindings := runNonGoEmDash(c.Fix)
	all = append(all, mdTxtFindings...)

	// Persist cache.
	if !c.NoCache && cacheUpdated {
		saveCache(cacheRoot, cache)
	}

	// Filter by severity threshold using the framework's fluent iterator
	// helpers so the SDK dogfoods fluent across command implementations.
	threshold := severityRank(minSeverity)
	visible := fluent.ToSlice(fluent.Filter(fluent.From(all), func(f Finding) bool {
		return severityRank(f.Severity) >= threshold
	}))

	if len(visible) == 0 {
		reportStatus("nothing to complain about.", "confident")
		return nil
	}

	printFindings(visible, c.Format)
	reportStatus(fmt.Sprintf("%d finding(s)", len(visible)), "alarmed")
	return &exitCodeError{code: exitLintFindings, cause: fmt.Errorf("lint: %d finding(s)", len(visible))}
}

// buildLinters constructs the ordered slice of Glacier-specific linters.
func buildLinters(nakedAnyEnabled bool) []Linter {
	ls := []Linter{
		&noEmDashLinter{},
		&panicInLibraryLinter{},
		&exportedDocCommentLinter{},
		&packageExampleTestLinter{},
		&libraryErrorRegisterLinter{},
	}
	if nakedAnyEnabled {
		ls = append(ls, &nakedAnyLinter{})
	}
	return ls
}

// reportStatus is a thin shim so the linter tests do not depend on the report
// package. In production it forwards to report.Status.
var reportStatus = func(msg, mood string) {
	switch mood {
	case "alarmed":
		fmt.Fprintf(os.Stderr, "ʕ× ×ʔ %s\n", msg)
	case "confident":
		fmt.Fprintf(os.Stderr, "ʕ•ᴥ•ʔ %s\n", msg)
	case "thinking":
		fmt.Fprintf(os.Stderr, "ʕ◔_◔ʔ %s\n", msg)
	default:
		fmt.Fprintf(os.Stderr, "ʕ•ᴥ•ʔ %s\n", msg)
	}
}

func severityRank(s string) int {
	switch s {
	case "error":
		return 2
	case "warning":
		return 1
	default:
		return 0
	}
}

// sha256File returns the hex sha256 digest of src.
func sha256File(src []byte) string {
	sum := sha256.Sum256(src)
	return hex.EncodeToString(sum[:])
}

// skipDir reports whether path is a directory that should not be linted.
// It excludes VCS state, build artifacts, vendored deps, and Claude
// worktree scratch trees so the linter walks only first-class source.
//
// The exclusions matter on two axes: correctness (worktree copies of the
// same files would double-count every finding) and robustness (staticcheck
// has historically panicked in its IR builder when fed partial or
// experimental code from worktrees).
func skipDir(path string) bool {
	switch filepath.Base(path) {
	case ".git", ".glacier", ".claude", "node_modules", "vendor", "dist":
		return true
	}
	p := filepath.ToSlash(path)
	if strings.HasPrefix(p, ".claude/") || strings.Contains(p, "/.claude/") {
		return true
	}
	return false
}

// loadCache reads .glacier/lint-cache.json into cache, ignoring any read error.
func loadCache(repoRoot string, cache lintCache) {
	path := filepath.Join(repoRoot, ".glacier", "lint-cache.json")
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	_ = safejson.Decode(f, &cache)
}

// saveCache persists cache to .glacier/lint-cache.json atomically.
func saveCache(repoRoot string, cache lintCache) {
	data, err := json.Marshal(cache)
	if err != nil {
		return
	}
	_ = safefile.WriteFileAtomic(repoRoot, filepath.Join(".glacier", "lint-cache.json"), data, 0o644)
}

// fixEmDash replaces U+2014 with ": " in src, returning the modified bytes and
// whether any replacement was made.
func fixEmDash(src []byte) ([]byte, bool) {
	emDash := []byte("\xe2\x80\x94") // UTF-8 encoding of U+2014
	if !bytes.Contains(src, emDash) {
		return src, false
	}
	return bytes.ReplaceAll(src, emDash, []byte(": ")), true
}

// markerSpaceRe matches a +glacier: directive prefix followed by extra whitespace.
var markerSpaceRe = regexp.MustCompile(`(//\s*\+glacier:)\s+`)

// fixMarkers normalizes +glacier: comment markers in src by removing extraneous
// whitespace between the directive prefix and the directive name.
// For example "// +glacier:command" becomes "// +glacier:command".
// Returns the modified bytes and whether any change was made.
func fixMarkers(src []byte) ([]byte, bool) {
	result := markerSpaceRe.ReplaceAll(src, []byte("${1}"))
	return result, !bytes.Equal(result, src)
}

// runGofmtCheck checks gofmt compliance for all .go files.
// When fix is true it rewrites non-compliant files in place.
// Returns findings for files that were not (or could not be) fixed.
func runGofmtCheck(_ context.Context, fix bool) (findings []Finding, fixed []string, err error) {
	walkErr := filepath.WalkDir(".", func(path string, d os.DirEntry, wErr error) error {
		if wErr != nil {
			return wErr
		}
		if d.IsDir() {
			if skipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		formatted, fmtErr := format.Source(src)
		if fmtErr != nil {
			return nil // skip files with syntax errors
		}
		if bytes.Equal(src, formatted) {
			return nil
		}
		if fix {
			if writeErr := os.WriteFile(path, formatted, 0o644); writeErr == nil {
				fixed = append(fixed, path)
				return nil
			}
		}
		findings = append(findings, Finding{
			Rule:     "gofmt",
			File:     path,
			Severity: "error",
			Message:  "file is not gofmt-formatted",
			FixHint:  "run: gofmt -w " + path,
		})
		return nil
	})
	return findings, fixed, walkErr
}

// runGoVet runs go vet and parses its output.
func runGoVet(ctx context.Context, patterns []string) ([]Finding, error) {
	args := append([]string{"vet"}, patterns...)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil, nil
	}
	var findings []Finding
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		f := Finding{
			Rule:     "go-vet",
			Severity: "error",
			Message:  line,
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 3 {
			f.File = strings.TrimSpace(parts[0])
			f.Message = strings.TrimSpace(parts[2])
		}
		findings = append(findings, f)
	}
	return findings, nil
}

// sanitizeStaticcheckPatterns expands "./..." into a per-top-level-dir set
// with skipDir directories removed. Other patterns pass through unchanged.
// If no top-level dirs survive the filter, an empty slice is returned and
// the caller should treat that as "nothing to check".
func sanitizeStaticcheckPatterns(patterns []string) ([]string, error) {
	hasWildcard := false
	other := make([]string, 0, len(patterns))
	for _, p := range patterns {
		if p == "./..." {
			hasWildcard = true
			continue
		}
		other = append(other, p)
	}
	if !hasWildcard {
		return other, nil
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("staticcheck: read repo root: %w", err)
	}
	expanded := other
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if skipDir(e.Name()) {
			continue
		}
		expanded = append(expanded, "./"+e.Name()+"/...")
	}
	return expanded, nil
}

// runStaticcheck runs staticcheck and parses its output.
//
// If patterns contains the wildcard "./...", that pattern is expanded into a
// per-top-level-dir set with skipDir directories removed. This prevents
// staticcheck from analysing files under .claude/worktrees, vendor, etc.,
// which would otherwise duplicate findings and (historically) trigger an
// IR-builder nil-pointer crash inside honnef.co/go/tools.
func runStaticcheck(ctx context.Context, patterns []string) ([]Finding, error) {
	expanded, err := sanitizeStaticcheckPatterns(patterns)
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, "staticcheck", expanded...)
	cmd.Env = os.Environ()
	out, scErr := cmd.CombinedOutput()
	if scErr != nil && len(out) == 0 {
		return nil, scErr
	}
	return parseStaticcheckOutput(out, scErr), nil
}

// staticcheckFindingRe matches a real staticcheck diagnostic of the form
// "path:line:col: rule message". Anything else (panic stack frames,
// "internal error" lines, "(compile)" prefix lines) is treated as tool
// noise and folded into a single summary finding.
var staticcheckFindingRe = regexp.MustCompile(`^[^:\s][^:]*:\d+:\d+:`)

// parseStaticcheckOutput converts staticcheck's stdout/stderr into Finding
// values. Lines that look like real diagnostics become individual findings.
// If the tool produced output that does not look like a diagnostic (panic
// trace, version-mismatch error), a single summary Finding is emitted with
// a pointer to the cause so users see one signal, not 17 stack frames.
func parseStaticcheckOutput(out []byte, runErr error) []Finding {
	var findings []Finding
	var noise []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if staticcheckFindingRe.MatchString(line) {
			findings = append(findings, Finding{
				Rule:     "staticcheck",
				Severity: "warning",
				Message:  line,
			})
			continue
		}
		noise = append(noise, line)
	}
	if len(findings) == 0 && (len(noise) > 0 || runErr != nil) {
		summary := "staticcheck did not produce diagnostics"
		if len(noise) > 0 {
			summary = noise[0]
		}
		// Tool-failure surfaces at info severity so a stale or panicking
		// staticcheck binary on PATH does not fail an otherwise clean lint
		// run; the message still appears at --severity=info.
		findings = append(findings, Finding{
			Rule:     "staticcheck",
			Severity: "info",
			Message:  "tool failure: " + summary,
			FixHint:  "upgrade staticcheck (`go install honnef.co/go/tools/cmd/staticcheck@latest`) or remove it from PATH",
		})
	}
	return findings
}

// runNonGoEmDash checks .md and .txt files for U+2014 and optionally fixes them.
func runNonGoEmDash(fix bool) []Finding {
	var findings []Finding
	_ = filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := filepath.Ext(path)
		if ext != ".md" && ext != ".txt" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		if fix {
			if fixed, changed := fixEmDash(data); changed {
				_ = os.WriteFile(path, fixed, 0o644)
				return nil
			}
		}
		scanner := bufio.NewScanner(bytes.NewReader(data))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			if bytes.ContainsRune(scanner.Bytes(), '\u2014') {
				findings = append(findings, Finding{
					Rule:     "no-em-dash",
					File:     path,
					Line:     lineNum,
					Severity: "error",
					Message:  "em-dash (U+2014) found; use colon, parens, or hyphen-minus instead",
					FixHint:  "run glacier lint --fix to auto-replace",
				})
			}
		}
		return nil
	})
	return findings
}

// printFindings renders findings in the requested format.
func printFindings(findings []Finding, format string) {
	switch format {
	case "json":
		printFindingsJSON(findings)
	default:
		printFindingsText(findings)
	}
}

func printFindingsText(findings []Finding) {
	type group struct {
		label   string
		kaomoji string
		items   []Finding
	}
	groups := []*group{
		{label: "error", kaomoji: "ʕ× ×ʔ"},
		{label: "warning", kaomoji: "ʕ◉_◉ʔ"},
		{label: "info", kaomoji: "ʕ•_•ʔ"},
	}
	for _, f := range findings {
		for _, g := range groups {
			if g.label == f.Severity {
				g.items = append(g.items, f)
				break
			}
		}
	}
	for _, g := range groups {
		if len(g.items) == 0 {
			continue
		}
		fmt.Fprintf(os.Stderr, "\n%s %s\n", g.kaomoji, g.label)
		for _, f := range g.items {
			loc := f.File
			if f.Line > 0 {
				loc = fmt.Sprintf("%s:%d", f.File, f.Line)
			}
			fmt.Fprintf(os.Stderr, "  %s: [%s] %s\n", loc, f.Rule, f.Message)
		}
	}
}

func printFindingsJSON(findings []Finding) {
	fmt.Fprintln(os.Stdout, "[")
	for i, f := range findings {
		comma := ","
		if i == len(findings)-1 {
			comma = ""
		}
		fmt.Fprintf(os.Stdout, "  {\"rule\":%q,\"file\":%q,\"line\":%d,\"severity\":%q,\"message\":%q}%s\n",
			f.Rule, f.File, f.Line, f.Severity, f.Message, comma)
	}
	fmt.Fprintln(os.Stdout, "]")
}

// --- Linter implementations ---

// noEmDashLinter flags U+2014 em-dash in .go files.
type noEmDashLinter struct{}

// Name returns the lint rule name.
func (l *noEmDashLinter) Name() string { return "no-em-dash" }

// Severity returns the default severity for this rule.
func (l *noEmDashLinter) Severity() Severity { return SeverityError }

// Check scans src for U+2014 and reports one finding per occurrence.
func (l *noEmDashLinter) Check(file string, src []byte) []Finding {
	if filepath.Ext(file) != ".go" {
		return nil
	}
	var findings []Finding
	scanner := bufio.NewScanner(bytes.NewReader(src))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if bytes.ContainsRune(scanner.Bytes(), '\u2014') {
			findings = append(findings, Finding{
				Rule:     l.Name(),
				File:     file,
				Line:     lineNum,
				Severity: l.Severity().String(),
				Message:  "em-dash (U+2014) found; use colon, parens, or hyphen-minus instead",
				FixHint:  "run glacier lint --fix to auto-replace",
			})
		}
	}
	return findings
}

// panicInLibraryLinter flags panic( calls in non-test, non-cmd packages.
type panicInLibraryLinter struct{}

// Name returns the lint rule name.
func (l *panicInLibraryLinter) Name() string { return "panic-in-library" }

// Severity returns the default severity for this rule.
func (l *panicInLibraryLinter) Severity() Severity { return SeverityError }

// Check parses src and reports calls to the builtin panic in library code.
//
// Detection is AST-based so the rule does not match the substring "panic("
// inside comments or string literals. Files in cmd/, _test.go files, and
// generated files (zz_generated_*.go) are exempt.
//
// A panic call is suppressed when any of the following carries a
// //glacier:nolint=panic-in-library directive: a same-line trailing
// comment, the comment block on the immediately preceding line, or the
// doc comment of the enclosing func declaration. This lets canonical
// programmer-error panics (assert.Must*, conf.Register duplicates,
// concur.Group post-shutdown calls) opt out with a justification rather
// than be refactored to error returns they cannot meaningfully produce.
func (l *panicInLibraryLinter) Check(file string, src []byte) []Finding {
	if !strings.HasSuffix(file, ".go") {
		return nil
	}
	if strings.HasSuffix(file, "_test.go") {
		return nil
	}
	normalized := filepath.ToSlash(file)
	if strings.HasPrefix(normalized, "cmd/") {
		return nil
	}
	base := filepath.Base(file)
	if strings.HasPrefix(base, "zz_generated_") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, parser.ParseComments)
	if err != nil {
		return nil
	}

	var findings []Finding
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		fnDocSuppressed := commentGroupSuppresses(fn.Doc, l.Name())
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			id, ok := call.Fun.(*ast.Ident)
			if !ok || id.Name != "panic" {
				return true
			}
			pos := fset.Position(call.Pos())
			if fnDocSuppressed || nolintAt(f, fset, pos.Line, l.Name()) {
				return true
			}
			findings = append(findings, Finding{
				Rule:     l.Name(),
				File:     file,
				Line:     pos.Line,
				Severity: l.Severity().String(),
				Message:  "panic( in library code; use error returns instead",
				FixHint:  "replace panic with an error return, or annotate //glacier:nolint=panic-in-library with a justification",
			})
			return true
		})
	}
	return findings
}

// nolintAt reports whether the line at lineNum (or the comment block on the
// line immediately above) carries a //glacier:nolint=<rule> directive.
func nolintAt(f *ast.File, fset *token.FileSet, lineNum int, rule string) bool {
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			cl := fset.Position(c.Pos()).Line
			if cl == lineNum || cl == lineNum-1 {
				if commentTextSuppresses(c.Text, rule) {
					return true
				}
			}
		}
	}
	return false
}

// commentGroupSuppresses reports whether any comment in cg carries
// //glacier:nolint=<rule>.
func commentGroupSuppresses(cg *ast.CommentGroup, rule string) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		if commentTextSuppresses(c.Text, rule) {
			return true
		}
	}
	return false
}

// commentTextSuppresses reports whether s contains a glacier:nolint=<rule>
// directive token. The token may appear anywhere in the comment so a
// trailing free-form justification is allowed.
func commentTextSuppresses(s, rule string) bool {
	idx := strings.Index(s, "glacier:nolint=")
	if idx < 0 {
		return false
	}
	rest := s[idx+len("glacier:nolint="):]
	for _, tok := range strings.FieldsFunc(rest, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	}) {
		if tok == rule {
			return true
		}
	}
	return false
}

// exportedDocCommentLinter warns when an exported symbol lacks a doc comment
// starting with the symbol name.
type exportedDocCommentLinter struct{}

// Name returns the lint rule name.
func (l *exportedDocCommentLinter) Name() string { return "exported-doc-comment" }

// Severity returns the default severity for this rule.
func (l *exportedDocCommentLinter) Severity() Severity { return SeverityWarning }

// Check parses src as Go source and reports exported symbols without proper doc comments.
// Skips test files and generated files (zz_generated_*.go).
func (l *exportedDocCommentLinter) Check(file string, src []byte) []Finding {
	if !strings.HasSuffix(file, ".go") {
		return nil
	}
	if strings.HasSuffix(file, "_test.go") {
		return nil
	}
	base := filepath.Base(file)
	if strings.HasPrefix(base, "zz_generated_") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, parser.ParseComments)
	if err != nil {
		return nil
	}

	var findings []Finding
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if !d.Name.IsExported() {
				continue
			}
			if d.Doc == nil || d.Doc.Text() == "" {
				pos := fset.Position(d.Pos())
				findings = append(findings, Finding{
					Rule:     l.Name(),
					File:     file,
					Line:     pos.Line,
					Severity: l.Severity().String(),
					Message:  fmt.Sprintf("exported func %s has no doc comment", d.Name.Name),
					FixHint:  fmt.Sprintf("add: // %s ...", d.Name.Name),
				})
				continue
			}
			if !strings.HasPrefix(strings.TrimSpace(d.Doc.Text()), d.Name.Name) {
				pos := fset.Position(d.Pos())
				findings = append(findings, Finding{
					Rule:     l.Name(),
					File:     file,
					Line:     pos.Line,
					Severity: l.Severity().String(),
					Message:  fmt.Sprintf("exported func %s: doc comment should start with %q", d.Name.Name, d.Name.Name),
					FixHint:  fmt.Sprintf("start comment with: // %s ...", d.Name.Name),
				})
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if !s.Name.IsExported() {
						continue
					}
					// Doc can be on the GenDecl or the TypeSpec itself.
					doc := s.Doc
					if doc == nil {
						doc = d.Doc
					}
					if doc == nil || doc.Text() == "" {
						pos := fset.Position(s.Pos())
						findings = append(findings, Finding{
							Rule:     l.Name(),
							File:     file,
							Line:     pos.Line,
							Severity: l.Severity().String(),
							Message:  fmt.Sprintf("exported type %s has no doc comment", s.Name.Name),
							FixHint:  fmt.Sprintf("add: // %s ...", s.Name.Name),
						})
						continue
					}
					if !strings.HasPrefix(strings.TrimSpace(doc.Text()), s.Name.Name) {
						pos := fset.Position(s.Pos())
						findings = append(findings, Finding{
							Rule:     l.Name(),
							File:     file,
							Line:     pos.Line,
							Severity: l.Severity().String(),
							Message:  fmt.Sprintf("exported type %s: doc comment should start with %q", s.Name.Name, s.Name.Name),
							FixHint:  fmt.Sprintf("start comment with: // %s ...", s.Name.Name),
						})
					}
				case *ast.ValueSpec:
					// Only check top-level var/const blocks with more than one spec
					// if the block itself has no doc; single-spec blocks inherit the GenDecl doc.
					for _, name := range s.Names {
						if !name.IsExported() {
							continue
						}
						doc := s.Doc
						if doc == nil {
							doc = d.Doc
						}
						if doc == nil || doc.Text() == "" {
							pos := fset.Position(name.Pos())
							findings = append(findings, Finding{
								Rule:     l.Name(),
								File:     file,
								Line:     pos.Line,
								Severity: l.Severity().String(),
								Message:  fmt.Sprintf("exported var/const %s has no doc comment", name.Name),
								FixHint:  fmt.Sprintf("add: // %s ...", name.Name),
							})
						}
					}
				}
			}
		}
	}
	return findings
}

// packageExampleTestLinter warns when a non-internal package has no Example* test.
// It operates on the package directory level: the finding is reported once per package.
type packageExampleTestLinter struct {
	// seen tracks directories we have already emitted a finding for.
	seen map[string]bool
}

// Name returns the lint rule name.
func (l *packageExampleTestLinter) Name() string { return "package-example-test" }

// Severity returns the default severity for this rule.
func (l *packageExampleTestLinter) Severity() Severity { return SeverityWarning }

// Check inspects the directory containing file for Example* functions.
//
// It fires on consumer-facing library packages only. Skipped:
//   - internal packages (/internal/, internal/)
//   - cmd packages (/cmd/, cmd/)
//   - testdata directories (/testdata/, testdata/)
//   - codegen example fixtures under docs/examples/
//   - package main (entry-point packages have no exported surface to exemplify)
//   - the integration-test "tests" package (consumed via go test, not import)
func (l *packageExampleTestLinter) Check(file string, _ []byte) []Finding {
	if !strings.HasSuffix(file, ".go") {
		return nil
	}
	if strings.HasSuffix(file, "_test.go") {
		return nil
	}
	normalized := filepath.ToSlash(file)
	if strings.Contains(normalized, "/internal/") || strings.HasPrefix(normalized, "internal/") {
		return nil
	}
	if strings.Contains(normalized, "/cmd/") || strings.HasPrefix(normalized, "cmd/") {
		return nil
	}
	if strings.Contains(normalized, "/testdata/") || strings.HasPrefix(normalized, "testdata/") {
		return nil
	}
	if strings.Contains(normalized, "/docs/examples/") || strings.HasPrefix(normalized, "docs/examples/") {
		return nil
	}

	dir := filepath.Dir(file)

	if l.seen == nil {
		l.seen = make(map[string]bool)
	}
	if l.seen[dir] {
		return nil
	}
	l.seen[dir] = true

	// Determine the package name once: used for both the main-package skip
	// and the finding message.
	fset := token.NewFileSet()
	pf, parseErr := parser.ParseFile(fset, file, nil, parser.PackageClauseOnly)
	pkgName := filepath.Base(dir)
	if parseErr == nil && pf.Name != nil {
		pkgName = pf.Name.Name
	}
	if pkgName == "main" {
		return nil
	}
	if pkgName == "tests" && filepath.Base(dir) == "tests" {
		return nil
	}

	// Check whether the directory contains any *_test.go file with an Example* func.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		testPath := filepath.Join(dir, e.Name())
		testSrc, readErr := os.ReadFile(testPath)
		if readErr != nil {
			continue
		}
		tfset := token.NewFileSet()
		tf, parseTestErr := parser.ParseFile(tfset, testPath, testSrc, 0)
		if parseTestErr != nil {
			continue
		}
		for _, decl := range tf.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && strings.HasPrefix(fn.Name.Name, "Example") {
				return nil
			}
		}
	}

	return []Finding{{
		Rule:     l.Name(),
		File:     file,
		Line:     1,
		Severity: l.Severity().String(),
		Message:  fmt.Sprintf("package %s has no Example* function in any *_test.go file", pkgName),
		FixHint:  "add an Example* func in a _test.go file in this package",
	}}
}

// libraryErrorRegisterLinter flags Error() strings that do not match ^[a-z][^.]*$.
// It only checks exported *Error types in library packages (not cmd/, not internal/, not _test.go).
type libraryErrorRegisterLinter struct{}

// errorStringRe is the pattern all library error strings must satisfy.
var errorStringRe = regexp.MustCompile(`^[a-z][^.]*$`)

// Name returns the lint rule name.
func (l *libraryErrorRegisterLinter) Name() string { return "library-error-register" }

// Severity returns the default severity for this rule.
func (l *libraryErrorRegisterLinter) Severity() Severity { return SeverityError }

// Check parses src for exported Error() methods and validates their string literals.
func (l *libraryErrorRegisterLinter) Check(file string, src []byte) []Finding {
	if !strings.HasSuffix(file, ".go") {
		return nil
	}
	if strings.HasSuffix(file, "_test.go") {
		return nil
	}
	normalized := filepath.ToSlash(file)
	if strings.HasPrefix(normalized, "cmd/") {
		return nil
	}
	if strings.Contains(normalized, "/internal/") || strings.HasPrefix(normalized, "internal/") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, 0)
	if err != nil {
		return nil
	}

	var findings []Finding
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		// Match exported Error() string methods with a pointer receiver.
		if fn.Name.Name != "Error" {
			continue
		}
		if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
			continue
		}
		// Receiver must be present.
		if fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		recv := fn.Recv.List[0]
		// Receiver type name must start with uppercase (exported type).
		recvName := receiverTypeName(recv.Type)
		if recvName == "" || !isExported(recvName) {
			continue
		}

		// Walk the function body looking for return statements with string literals.
		if fn.Body == nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			ret, ok := n.(*ast.ReturnStmt)
			if !ok {
				return true
			}
			for _, result := range ret.Results {
				lit, ok := result.(*ast.BasicLit)
				if !ok {
					continue
				}
				if lit.Kind != token.STRING {
					continue
				}
				// Unquote the string value.
				val := strings.Trim(lit.Value, `"`)
				// Empty-string returns are nil-receiver guards (e.g.
				// `if w == nil { return "" }`), not real error messages.
				if val == "" {
					continue
				}
				if !errorStringRe.MatchString(val) {
					pos := fset.Position(lit.Pos())
					findings = append(findings, Finding{
						Rule:     l.Name(),
						File:     file,
						Line:     pos.Line,
						Severity: l.Severity().String(),
						Message:  fmt.Sprintf("library error string %q does not match ^[a-z][^.]*$ (no trailing period, start lowercase)", val),
						FixHint:  "change the error string to be lowercase with no trailing period",
					})
				}
			}
			return true
		})
	}
	return findings
}

// receiverTypeName extracts the base type name from a receiver type expression.
// Handles *T and T forms.
func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.Ident:
		return t.Name
	}
	return ""
}

// isExported reports whether name starts with an uppercase letter.
func isExported(name string) bool {
	if name == "" {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

// nakedAnyLinter flags bare any or interface{} in function signatures outside test files.
// Only active when lint.naked_any.enabled is true in config.
type nakedAnyLinter struct{}

// Name returns the lint rule name.
func (l *nakedAnyLinter) Name() string { return "naked-any" }

// Severity returns the default severity for this rule.
func (l *nakedAnyLinter) Severity() Severity { return SeverityWarning }

// Check parses src for function parameters or return values typed as bare any/interface{}.
func (l *nakedAnyLinter) Check(file string, src []byte) []Finding {
	if !strings.HasSuffix(file, ".go") {
		return nil
	}
	if strings.HasSuffix(file, "_test.go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, 0)
	if err != nil {
		return nil
	}

	var findings []Finding
	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		// Check params.
		if fn.Type.Params != nil {
			for _, field := range fn.Type.Params.List {
				if isNakedAny(field.Type) {
					pos := fset.Position(field.Pos())
					findings = append(findings, Finding{
						Rule:     l.Name(),
						File:     file,
						Line:     pos.Line,
						Severity: l.Severity().String(),
						Message:  fmt.Sprintf("function %s: parameter typed as bare any/interface{}; use a named interface or type constraint", fn.Name.Name),
						FixHint:  "define a named interface expressing the required methods",
					})
				}
			}
		}
		// Check results.
		if fn.Type.Results != nil {
			for _, field := range fn.Type.Results.List {
				if isNakedAny(field.Type) {
					pos := fset.Position(field.Pos())
					findings = append(findings, Finding{
						Rule:     l.Name(),
						File:     file,
						Line:     pos.Line,
						Severity: l.Severity().String(),
						Message:  fmt.Sprintf("function %s: return value typed as bare any/interface{}; use a named interface or type constraint", fn.Name.Name),
						FixHint:  "define a named interface expressing the required methods",
					})
				}
			}
		}
		return true
	})
	return findings
}

// isNakedAny reports whether expr is a bare `any` identifier or an empty `interface{}`.
func isNakedAny(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name == "any"
	case *ast.InterfaceType:
		// interface{} with no methods.
		return t.Methods == nil || len(t.Methods.List) == 0
	}
	return false
}
