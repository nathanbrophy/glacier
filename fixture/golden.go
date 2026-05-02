// SPDX-License-Identifier: Apache-2.0

package fixture

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/internal/safefile"
)

// goldenUpdateEnv is the environment variable that activates auto-write mode.
// Set to "1" to create or update golden and snapshot files on mismatch.
const goldenUpdateEnv = "GLACIER_GOLDEN_UPDATE"

// goldenConfig holds the resolved options for a golden/load operation.
type goldenConfig struct {
	// root is the base directory for file lookups.
	// Zero value ("") means "testdata/" adjacent to the calling test file.
	root string
}

// GoldenOption configures Golden, Snapshot, Load, and LoadJSON behaviour.
type GoldenOption interface{ applyGolden(*goldenConfig) error }

type goldenOptFunc func(*goldenConfig) error

func (f goldenOptFunc) applyGolden(c *goldenConfig) error { return f(c) }

// WithRoot redirects file operations to path instead of the default testdata/
// directory. path may be relative or absolute. Relative paths are resolved
// from the current working directory. Paths that resolve to or start with ".."
// after Clean are rejected to prevent traversal attacks. Windows UNC paths are
// rejected.
func WithRoot(path string) GoldenOption {
	return goldenOptFunc(func(c *goldenConfig) error {
		if path == "" {
			return fmt.Errorf("fixture: WithRoot: path must not be empty")
		}
		// Reject UNC paths.
		if len(path) >= 2 && path[0] == '\\' && path[1] == '\\' {
			return fmt.Errorf("fixture: WithRoot: %w", safefile.ErrUNC)
		}
		if len(path) >= 2 && path[0] == '/' && path[1] == '/' {
			return fmt.Errorf("fixture: WithRoot: %w", safefile.ErrUNC)
		}
		// Reject traversal: clean the path and check for ".." components.
		cleaned := filepath.Clean(path)
		if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
			return fmt.Errorf("fixture: WithRoot: %w", safefile.ErrTraversal)
		}
		c.root = cleaned
		return nil
	})
}

// applyGoldenOptions applies opts to a zero goldenConfig and returns the result.
// Any option error is returned immediately.
func applyGoldenOptions(opts []GoldenOption) (goldenConfig, error) {
	var c goldenConfig
	for _, o := range opts {
		if o != nil {
			if err := o.applyGolden(&c); err != nil {
				return goldenConfig{}, err
			}
		}
	}
	return c, nil
}

// resolveRoot returns the absolute testdata root for the given config.
// If c.root is empty, it is resolved relative to the calling test file's
// directory using runtime.Caller. skip is the number of stack frames to
// skip above resolveRoot itself.
func resolveRoot(c goldenConfig, skip int) (string, error) {
	if c.root != "" {
		if filepath.IsAbs(c.root) {
			return c.root, nil
		}
		// Relative root: resolve from cwd.
		abs, err := filepath.Abs(c.root)
		if err != nil {
			return "", fmt.Errorf("fixture: resolve root: %w", err)
		}
		return abs, nil
	}
	// Locate the calling test file's directory.
	_, file, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return "", fmt.Errorf("fixture: could not determine caller file path")
	}
	return filepath.Join(filepath.Dir(file), "testdata"), nil
}

// Golden compares got against the golden file at testdata/<name> (or the root
// set by WithRoot). If the file is missing and GLACIER_GOLDEN_UPDATE=1, the
// file is created and true is returned. If the file is missing without the env
// var, t.Errorf is called with a "re-run with GLACIER_GOLDEN_UPDATE=1" hint
// and false is returned. On content mismatch, t.Errorf is called with a
// line-by-line diff (text) or hex header (binary) and false is returned.
// On match, true is returned with no output.
//
// name must be a relative path; ".." components are rejected by safefile.
// Golden calls t.Helper before any failure message.
func Golden(t assert.TB, name string, got []byte, opts ...GoldenOption) bool {
	t.Helper()
	cfg, err := applyGoldenOptions(opts)
	if err != nil {
		t.Errorf("fixture: Golden: %v", err)
		return false
	}
	root, err := resolveRoot(cfg, 1)
	if err != nil {
		t.Errorf("fixture: Golden: %v", err)
		return false
	}
	return goldenCompare(t, root, name, got, "testdata/"+name)
}

// goldenCompare is the shared implementation used by Golden and Snapshot.
// root is the absolute path to the testdata directory. display is the
// human-readable path shown in error messages.
func goldenCompare(t assert.TB, root, name string, got []byte, display string) bool {
	t.Helper()

	// Validate the name for path safety.
	cleanName, err := safefile.Clean(name)
	if err != nil {
		t.Errorf("fixture: golden: path rejected for %q: %v", name, err)
		return false
	}

	path := filepath.Join(root, cleanName)

	want, readErr := os.ReadFile(path)
	if readErr != nil {
		if !os.IsNotExist(readErr) {
			t.Errorf("fixture: golden: read %s: %v", display, readErr)
			return false
		}
		// File does not exist.
		if os.Getenv(goldenUpdateEnv) == "1" {
			if err := safefile.WriteFileAtomic(root, cleanName, got, 0o644); err != nil {
				t.Errorf("fixture: golden: write %s: %v", display, err)
				return false
			}
			return true
		}
		t.Errorf("fixture: golden: file %s does not exist; re-run with %s=1 to create it", display, goldenUpdateEnv)
		return false
	}

	if bytes.Equal(want, got) {
		return true
	}

	// Content mismatch.
	if os.Getenv(goldenUpdateEnv) == "1" {
		if err := safefile.WriteFileAtomic(root, cleanName, got, 0o644); err != nil {
			t.Errorf("fixture: golden: update %s: %v", display, err)
			return false
		}
		return true
	}

	t.Errorf("fixture: golden: mismatch for %s:\n%s", display, diffBytes(got, want))
	return false
}

// diffBytes produces a human-readable diff between got and want.
// For valid UTF-8 content, it produces a line-by-line +/- diff.
// For binary content, it shows hex headers of both.
func diffBytes(got, want []byte) string {
	if utf8.Valid(got) && utf8.Valid(want) {
		return textDiff(string(got), string(want))
	}
	return binaryDiff(got, want)
}

// textDiff produces a simple +/- line diff between two text strings.
func textDiff(got, want string) string {
	gotLines := strings.Split(got, "\n")
	wantLines := strings.Split(want, "\n")
	var b strings.Builder
	b.WriteString("--- want\n+++ got\n")
	maxLines := len(gotLines)
	if len(wantLines) > maxLines {
		maxLines = len(wantLines)
	}
	for i := range maxLines {
		var g, w string
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if g != w {
			if i < len(wantLines) {
				fmt.Fprintf(&b, "-%s\n", w)
			}
			if i < len(gotLines) {
				fmt.Fprintf(&b, "+%s\n", g)
			}
		}
	}
	return b.String()
}

// binaryDiff shows hex dumps of the first 64 bytes of each.
func binaryDiff(got, want []byte) string {
	const head = 64
	g := got
	if len(g) > head {
		g = g[:head]
	}
	w := want
	if len(w) > head {
		w = w[:head]
	}
	return fmt.Sprintf("binary mismatch\ngot  (%d bytes): %s\nwant (%d bytes): %s\n",
		len(got), hex.EncodeToString(g),
		len(want), hex.EncodeToString(w))
}
