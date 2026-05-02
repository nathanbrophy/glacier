// SPDX-License-Identifier: Apache-2.0

// Package safefile provides path-canonicalization and safe file-open helpers
// for Glacier's test-data operations. All callers must route file paths
// through this package to prevent path-traversal attacks.
//
// Rules enforced on every path:
//   - filepath.Clean is applied before any filesystem call.
//   - Any post-Clean path component equal to ".." is rejected.
//   - Absolute paths are rejected (prefix "/" on Unix, drive-letter or UNC on Windows).
//   - Windows UNC paths (\\server\share and \\?\) are always rejected.
package safefile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ErrTraversal is returned when a path contains ".." after canonicalization.
var ErrTraversal = fmt.Errorf("safefile: path traversal rejected")

// ErrAbsolute is returned when a path is absolute.
var ErrAbsolute = fmt.Errorf("safefile: absolute path rejected")

// ErrUNC is returned when a Windows UNC path is detected.
var ErrUNC = fmt.Errorf("safefile: UNC path rejected")

// Clean canonicalizes name and returns the cleaned path plus any rejection
// error. A nil error means the path is safe; callers must not proceed on
// a non-nil error.
//
// name must be a relative path: no ".." after Clean, no leading "/",
// no Windows drive prefix, no UNC.
func Clean(name string) (string, error) {
	// Reject UNC paths immediately — before Clean can normalize them.
	// UNC: \\server\share or \\?\ on Windows.
	if len(name) >= 2 && name[0] == '\\' && name[1] == '\\' {
		return "", ErrUNC
	}
	if len(name) >= 2 && name[0] == '/' && name[1] == '/' {
		// Double-slash prefix — treat as UNC-like.
		return "", ErrUNC
	}

	// Reject paths starting with / or \ (Unix absolute or Windows root-relative).
	if len(name) > 0 && (name[0] == '/' || name[0] == '\\') {
		return "", ErrAbsolute
	}

	clean := filepath.Clean(name)

	// Reject absolute paths (after Clean, so we handle "/" → "/" cleanly).
	if filepath.IsAbs(clean) {
		return "", ErrAbsolute
	}
	// Windows drive letter: e.g. "C:" or "C:foo".
	if len(clean) >= 2 && clean[1] == ':' {
		return "", ErrAbsolute
	}
	// Reject paths that start with a separator after Clean (root-relative on Windows).
	if len(clean) > 0 && (clean[0] == '/' || clean[0] == '\\') {
		return "", ErrAbsolute
	}

	// Reject any path component equal to "..".
	// filepath.Clean collapses ".." elements, but the result may still
	// start with ".." if the name itself starts with "..".
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", ErrTraversal
	}
	// Walk every component.
	parts := strings.Split(clean, string(filepath.Separator))
	for _, p := range parts {
		if p == ".." {
			return "", ErrTraversal
		}
	}

	return clean, nil
}

// Join calls Clean on name and then joins root and the clean name.
// root must already be an absolute, canonical path. An error is returned
// if name fails the safety checks.
func Join(root, name string) (string, error) {
	clean, err := Clean(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, clean), nil
}

// Open opens the file at root/name after safety-checking name.
// root is a trusted absolute path (e.g. t.TempDir(), testdata/).
func Open(root, name string) (*os.File, error) {
	path, err := Join(root, name)
	if err != nil {
		return nil, err
	}
	return os.Open(path) //nolint:gosec // path has been canonicalized and validated above
}

// ReadFile reads root/name after safety-checking name.
func ReadFile(root, name string) ([]byte, error) {
	f, err := Open(root, name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// WriteFileAtomic writes data to root/name atomically using a temp file +
// rename strategy. The target directory must already exist.
func WriteFileAtomic(root, name string, data []byte, perm os.FileMode) error {
	path, err := Join(root, name)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("safefile: mkdir %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".safefile-*")
	if err != nil {
		return fmt.Errorf("safefile: create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("safefile: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("safefile: close temp: %w", err)
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("safefile: chmod temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("safefile: rename: %w", err)
	}
	return nil
}
