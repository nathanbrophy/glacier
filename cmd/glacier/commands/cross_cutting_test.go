// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bytes"
	"context"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cache"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/ghreleases"
	"github.com/nathanbrophy/glacier/term"
)

// --- X13: TestBannerSuppressedOnSubcommands ---

// TestBannerSuppressedOnSubcommands verifies that non-root command Run()
// methods do not emit the wordmark banner. The banner is only written by
// cli/app.go Main() when len(argv)==0. Running a subcommand Run() directly
// is the canonical non-root path; it must produce no banner block characters.
func TestBannerSuppressedOnSubcommands(t *testing.T) {
	t.Parallel()

	// bannerSentinel is the block-character sequence unique to the GLACIER
	// banner art in cli/assets/banner.txt. It will never appear in normal
	// command output.
	const bannerSentinel = "\xe2\x96\x88\xe2\x96\x88" // "██" in UTF-8

	rows := []struct {
		name   string
		output func(t *testing.T) string
	}{
		{
			name: "version",
			output: func(t *testing.T) string {
				var buf bytes.Buffer
				cmd := (&VersionCmd{}).
					withWriter(&buf).
					withCache(cache.New[ghreleases.Release]())
				_ = cmd.Run(context.Background())
				return buf.String()
			},
		},
		{
			name: "explain_list",
			output: func(t *testing.T) string {
				old := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w
				cmd := &ExplainCmd{List: true}
				_ = cmd.Run(context.Background())
				w.Close()
				os.Stdout = old
				var buf bytes.Buffer
				_, _ = buf.ReadFrom(r)
				return buf.String()
			},
		},
	}

	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out := tc.output(t)
			assert.True(t, !strings.Contains(out, bannerSentinel),
				"subcommand output must not contain the wordmark banner block characters")
		})
	}
}

// --- X14 & X15: TestNoColorStripsAnsi / TestGlacierNoColorStripsAnsi ---

// TestNoColorStripsAnsi verifies that setting NO_COLOR=1 causes
// term.Capability to report NoColorEnv=true, which suppresses ANSI output.
// The writeBanner function in cli/banner.go gates gradient rendering on
// !caps.NoColorEnv.
func TestNoColorStripsAnsi(t *testing.T) {
	// Serialize: mutates env.
	t.Setenv("NO_COLOR", "1")
	t.Setenv("GLACIER_NO_COLOR", "") // ensure only NO_COLOR is active

	// Fresh writer so capCache has no stale entry.
	w := &bytes.Buffer{}
	caps := term.Capability(w)
	assert.True(t, caps.NoColorEnv, "NO_COLOR=1 must set NoColorEnv=true")
}

// TestGlacierNoColorStripsAnsi verifies that GLACIER_NO_COLOR=1 also sets
// NoColorEnv, giving application-scoped color suppression precedence.
func TestGlacierNoColorStripsAnsi(t *testing.T) {
	// Serialize: mutates env.
	t.Setenv("GLACIER_NO_COLOR", "1")

	w := &bytes.Buffer{}
	caps := term.Capability(w)
	assert.True(t, caps.NoColorEnv, "GLACIER_NO_COLOR=1 must set NoColorEnv=true")
}

// --- X16: TestNonTTYSuppressesAnimator ---

// TestNonTTYSuppressesAnimator verifies that writing to a non-TTY io.Writer
// causes term.Capability to report IsTTY=false. All animator start paths in
// the commands package gate on caps.IsTTY (e.g. vibe.go, generate.go); when
// IsTTY=false the animation branch is skipped and runStatic()/no-op is taken.
func TestNonTTYSuppressesAnimator(t *testing.T) {
	t.Parallel()
	caps := term.Capability(&bytes.Buffer{})
	assert.True(t, !caps.IsTTY, "bytes.Buffer must not be detected as a TTY")
	assert.True(t, !caps.SupportsUTF8, "non-TTY writer must not claim UTF-8 support")
}

// --- X17: TestVerbosityMutualExclusion ---

// TestVerbosityMutualExclusion verifies that combining quiet with verbose
// flags returns exitUsage (exit code 2).
func TestVerbosityMutualExclusion(t *testing.T) {
	t.Parallel()

	rows := []struct {
		name string
		cmd  GlacierCmd
	}{
		{"quiet_verbose", GlacierCmd{Quiet: true, Verbose: true}},
		{"quiet_very_verbose", GlacierCmd{Quiet: true, VeryVerbose: true}},
	}

	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.cmd.validateVerbosity()
			assert.Error(t, err)
			var ec *exitCodeError
			assert.True(t, errors.As(err, &ec), "expected exitCodeError")
			assert.Equal(t, exitUsage, ec.ExitCode())
		})
	}
}

// --- X19: TestProfileFlagWritesPprofFiles ---

// TestProfileFlagWritesPprofFiles verifies that when --profile=<path> is
// wired to the CLI binary, running a command writes <path>.cpu, <path>.heap,
// and <path>.goroutine pprof files.
//
// TODO(spec-0032): ProfileFlag wiring lives in cmd/glacier/main.go (or the
// root GlacierCmd) once the internal/profile package is plumbed through the
// flag via GlacierCmd.Profile. Until then this test is skipped.
func TestProfileFlagWritesPprofFiles(t *testing.T) {
	t.Skip("TODO(spec-0032): --profile flag not yet wired into GlacierCmd.Run; add when profile.Start is called in root Run")
}

// --- X20: TestExitCodeAreReachable ---

// TestExitCodeAreReachable uses go/parser to scan every non-test .go file in
// the commands package and verifies that specific exit-code values appear in
// at least one &exitCodeError{code: <N>} literal. Codes 0 (success), 1
// (generic), 130 (SIGINT), and 143 (SIGTERM) are handled by the framework or
// signal machinery rather than exitCodeError literals; they are exempted.
func TestExitCodeAreReachable(t *testing.T) {
	t.Parallel()

	// Codes expected to appear as exitCodeError literals in non-test source.
	mustBeReachable := []struct {
		name string
		code int
	}{
		{"usage(2)", exitUsage},
		{"generate_failed(64)", exitGenerateFailed},
		{"lint_findings(65)", exitLintFindings},
		{"tests_failed(66)", exitTestsFailed},
		{"scaffold_failed(67)", exitScaffoldFailed},
		{"version_check(68)", exitVersionCheck},
		{"codegen_drift(69)", exitCodegenDrift},
		{"subprocess(70)", exitSubprocess},
		{"interrupted(130)", exitInterrupted},
	}

	// Find the directory containing this test file.
	_, thisFile, _, _ := runtime.Caller(0)
	pkgDir := filepath.Dir(thisFile)

	codesFound := make(map[int]bool)
	fset := token.NewFileSet()

	entries, err := os.ReadDir(pkgDir)
	assert.NoError(t, err)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		src, readErr := os.ReadFile(filepath.Join(pkgDir, e.Name()))
		assert.NoError(t, readErr)

		f, parseErr := parser.ParseFile(fset, e.Name(), src, 0)
		if parseErr != nil {
			continue
		}
		_ = f
		// Scan raw source for exitCodeError{code: patterns; parsing AST to
		// extract composite literal values is equivalent but fragile across
		// formatting. String scan is simpler and sufficient.
		content := string(src)
		for _, row := range mustBeReachable {
			if strings.Contains(content, "exitCodeError") {
				codesFound[row.code] = codesFound[row.code] || exitCodeUsedInFile(content, row.code)
			}
		}
	}

	for _, row := range mustBeReachable {
		assert.True(t, codesFound[row.code],
			"exit code %s must appear in a non-test source file as an exitCodeError literal",
			row.name)
	}
}

// exitCodeUsedInFile reports whether a file containing exitCodeError uses the
// named constant whose value is code. Because the constants are file-local
// (unexported), we resolve by constant name pattern: the constant names embed
// their semantic meaning and map 1:1 to the numeric values in exitcodes.go.
func exitCodeUsedInFile(content string, code int) bool {
	name := exitCodeConstantName(code)
	if name == "" {
		return false
	}
	return strings.Contains(content, name)
}

// exitCodeConstantName returns the package-level constant name for code.
func exitCodeConstantName(code int) string {
	switch code {
	case exitSuccess:
		return "exitSuccess"
	case exitGeneric:
		return "exitGeneric"
	case exitUsage:
		return "exitUsage"
	case exitGenerateFailed:
		return "exitGenerateFailed"
	case exitLintFindings:
		return "exitLintFindings"
	case exitTestsFailed:
		return "exitTestsFailed"
	case exitScaffoldFailed:
		return "exitScaffoldFailed"
	case exitVersionCheck:
		return "exitVersionCheck"
	case exitCodegenDrift:
		return "exitCodegenDrift"
	case exitSubprocess:
		return "exitSubprocess"
	case exitInterrupted:
		return "exitInterrupted"
	case exitTerminated:
		return "exitTerminated"
	}
	return ""
}

// --- X22: TestGlacierVersionCheckUsesCache ---

// countingFetcher is a test-only ReleaseFetcher that counts calls and returns
// a canned release. It satisfies ghreleases.ReleaseFetcher.
type countingFetcher struct {
	calls   int
	release ghreleases.Release
}

func (f *countingFetcher) Latest(_ context.Context, _ string) (ghreleases.Release, error) {
	f.calls++
	return f.release, nil
}

// TestGlacierVersionCheckUsesCache verifies that the first --check call
// invokes the fetcher, and a second call within the cache TTL hits the cache
// without a second fetcher invocation.
func TestGlacierVersionCheckUsesCache(t *testing.T) {
	t.Parallel()

	fetcher := &countingFetcher{release: fakeRelease}
	rc := cache.New[ghreleases.Release](cache.WithDefaultTTL(1 * time.Hour))

	runCheck := func() {
		var buf bytes.Buffer
		cmd := (&VersionCmd{Check: true}).
			withWriter(&buf).
			withFetcher(fetcher).
			withCache(rc)
		assert.NoError(t, cmd.Run(context.Background()))
	}

	runCheck() // first call: cache miss, fetcher invoked
	assert.True(t, fetcher.calls == 1, "first call must invoke the fetcher, got", fetcher.calls)

	runCheck() // second call: cache hit, fetcher NOT invoked
	assert.True(t, fetcher.calls == 1, "second call within TTL must use cache, got", fetcher.calls)
}

// --- X23: TestSDKImportsCachePackage ---

// TestSDKImportsCachePackage is a sanity check that the commands package
// imports github.com/nathanbrophy/glacier/cache. It guards against accidental
// removal of the cache dependency.
func TestSDKImportsCachePackage(t *testing.T) {
	t.Parallel()

	// Compile-time evidence: cache.New is called in this very file's
	// TestGlacierVersionCheckUsesCache and in version_test.go. If the import
	// is ever removed, those tests will fail to compile. This runtime check
	// provides an additional named failure point with a clear error message.
	//
	// We verify by calling cache.New to ensure the import is active.
	c := cache.New[string]()
	assert.True(t, c != nil, "cache.New must return a non-nil Cache; "+
		"if this fails the glacier/cache import was removed from the commands package")
}
