// SPDX-License-Identifier: Apache-2.0

package mock_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestNoUnsafeImports audits the mock and internal/reflectx source files for
// any use of the unsafe package or unsafe reflect functions (§21.10 NF3, Falcon §1.2).
func TestNoUnsafeImports(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}

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
				t.Errorf("failed to read %s: %v", path, err)
				continue
			}
			src := string(content)
			for _, sym := range forbidden {
				if strings.Contains(src, sym) {
					t.Errorf("file %s contains forbidden symbol %q", path, sym)
				}
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
	if !ok {
		t.Fatal("could not determine test file path")
	}
	mockDir := filepath.Dir(thisFile)

	writeAPIs := []string{
		"os.Create(",
		"os.OpenFile(",
		"ioutil.WriteFile(",
		"os.WriteFile(",
	}

	entries, err := os.ReadDir(mockDir)
	if err != nil {
		t.Fatalf("cannot read mock dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") ||
			strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(mockDir, e.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("cannot read %s: %v", path, err)
			continue
		}
		src := string(content)
		for _, api := range writeAPIs {
			if strings.Contains(src, api) {
				t.Errorf("production file %s contains file-write API %q", path, api)
			}
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
	if err != nil {
		t.Errorf("go vet ./mock/... failed:\n%s", out)
	}
}
