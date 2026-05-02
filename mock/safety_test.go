// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
)

// TestNoUnsafeImports audits the mock and internal/reflectx source files for
// any use of the unsafe package or unsafe reflect functions (§21.10 NF3, Falcon §1.2).
func TestNoUnsafeImports(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "could not determine test file path")

	// Find the module root (directory containing go.mod).
	dir := filepath.Dir(thisFile) // mock/
	root := filepath.Dir(dir)     // module root

	// Directories to audit.
	toAudit := []string{
		filepath.Join(root, "mock"),
		filepath.Join(root, "internal", "reflectx"),
	}

	forbidden := []string{
		"unsafe.Pointer",
		"UnsafeAddr",
		"UnsafePointer",
		"reflect.NewAt",
	}

	for _, dir := range toAudit {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Logf("skipping %s: %v", dir, err)
			continue
		}
		for _, e := range entries {
			name := e.Name()
			// Skip directories, test files, and this safety test file itself.
			if e.IsDir() || !strings.HasSuffix(name, ".go") ||
				strings.HasSuffix(name, "_test.go") {
				continue
			}
			path := filepath.Join(dir, name)
			content, err := os.ReadFile(path)
			if err != nil {
				assert.True(t, false, "failed to read "+path+": "+err.Error())
				continue
			}
			src := string(content)
			for _, sym := range forbidden {
				assert.False(t, strings.Contains(src, sym),
					"file "+path+" contains forbidden symbol "+sym)
			}
		}
	}
}

// TestNoOnDiskEmissionAtRuntime verifies that the mock package does not open
// any files for writing at test runtime (§21.10 NF4).
// This is a best-effort audit: we verify that the mock source files contain no
// os.Create, os.OpenFile with os.O_WRONLY/O_CREATE, or ioutil.WriteFile calls.
func TestNoOnDiskEmissionAtRuntime(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "could not determine test file path")
	mockDir := filepath.Dir(thisFile)

	writeAPIs := []string{
		"os.Create(",
		"os.OpenFile(",
		"ioutil.WriteFile(",
		"os.WriteFile(",
	}

	entries, err := os.ReadDir(mockDir)
	require.NoError(t, err, "cannot read mock dir")
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") ||
			strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(mockDir, e.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			assert.True(t, false, "cannot read "+path+": "+err.Error())
			continue
		}
		src := string(content)
		for _, api := range writeAPIs {
			assert.False(t, strings.Contains(src, api),
				"production file "+path+" contains file-write API "+api)
		}
	}
}

// TestGoVetMock runs `go vet ./mock/...` and reports any issues.
func TestGoVetMock(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Skip("cannot determine test file path")
	}
	root := filepath.Dir(filepath.Dir(thisFile))
	cmd := exec.Command("go", "vet", "./mock/...")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	assert.True(t, err == nil, "go vet ./mock/... failed:\n"+string(out))
}
