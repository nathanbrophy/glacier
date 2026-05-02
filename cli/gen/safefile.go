// SPDX-License-Identifier: Apache-2.0

package gen

import (
	"fmt"
	"path/filepath"
	"strings"
)

// canonicalizePath validates and canonicalizes the output file path for codegen.
// root is the absolute module root. rel is the relative path to canonicalize.
//
// Rules enforced (§23.8):
//   - filepath.Clean applied.
//   - No ".." components.
//   - No absolute paths.
//   - No UNC paths (\\).
//   - Resolved path must sit under root.
func canonicalizePath(root, rel string) (string, error) {
	// Reject UNC paths.
	if strings.HasPrefix(rel, `\\`) || strings.HasPrefix(rel, `//`) {
		return "", fmt.Errorf("gen: output path %q is a UNC path", rel)
	}

	// Reject absolute paths.
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("gen: output path %q is absolute", rel)
	}
	// Windows drive letter check.
	if len(rel) >= 2 && rel[1] == ':' {
		return "", fmt.Errorf("gen: output path %q is absolute (drive letter)", rel)
	}

	clean := filepath.Clean(rel)

	// Reject ".." components.
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("gen: output path %q contains path traversal", rel)
	}
	parts := strings.Split(clean, string(filepath.Separator))
	for _, p := range parts {
		if p == ".." {
			return "", fmt.Errorf("gen: output path %q contains path traversal", rel)
		}
	}

	abs := filepath.Join(root, clean)

	// Confirm it sits under root.
	rootClean := filepath.Clean(root) + string(filepath.Separator)
	if !strings.HasPrefix(abs+string(filepath.Separator), rootClean) {
		return "", fmt.Errorf("gen: output path %q escapes module root", rel)
	}

	return abs, nil
}
