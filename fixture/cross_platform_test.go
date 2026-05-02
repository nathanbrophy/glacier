// SPDX-License-Identifier: Apache-2.0

package fixture_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nathanbrophy/glacier/fixture"
)

// TestGoldenCrossPlatformLineEndings verifies that golden file comparison
// works correctly on all platforms regardless of OS newline conventions.
func TestGoldenCrossPlatformLineEndings(t *testing.T) {
	dir := t.TempDir()
	// Write a golden file with LF endings.
	content := []byte("line1\nline2\nline3\n")
	if err := os.WriteFile(filepath.Join(dir, "lf.txt"), content, 0o644); err != nil {
		t.Fatal(err)
	}
	ok := fixture.Golden(t, "lf.txt", content, fixture.WithRoot(dir))
	if !ok {
		t.Fatal("Golden failed on LF-terminated file")
	}
}

// TestLoadCrossPlatform verifies that Load works on all platforms.
func TestLoadCrossPlatform(t *testing.T) {
	dir := t.TempDir()
	want := []byte("cross platform data")
	if err := os.WriteFile(filepath.Join(dir, "cp.txt"), want, 0o644); err != nil {
		t.Fatal(err)
	}
	got := fixture.Load(t, "cp.txt", fixture.WithRoot(dir))
	if string(got) != string(want) {
		t.Fatalf("Load cross-platform: got %q, want %q", got, want)
	}
}

// TestSnapshotCrossPlatform verifies snapshot creation and match on all platforms.
func TestSnapshotCrossPlatform(t *testing.T) {
	t.Setenv("GLACIER_GOLDEN_UPDATE", "1")

	type Pair struct {
		Key   string
		Value int
	}
	v := Pair{Key: "cross", Value: 42}
	ok1 := fixture.Snapshot(t, "cross_platform_snap", v)
	if !ok1 {
		t.Fatal("Snapshot create failed")
	}

	t.Setenv("GLACIER_GOLDEN_UPDATE", "0")
	ok2 := fixture.Snapshot(t, "cross_platform_snap", v)
	if !ok2 {
		t.Fatal("Snapshot match failed on second run")
	}
}

// TestNewFSCrossPlatform verifies that NewFS path separators work on all platforms.
func TestNewFSCrossPlatform(t *testing.T) {
	// fs.FS always uses forward slash separators.
	content := []byte("platform neutral")
	fsys := fixture.NewFS(map[string][]byte{
		"a/b/c.txt": content,
	})
	// fs.ReadFile always uses forward slashes per fs.FS contract.
	got, err := fsys.(interface {
		ReadFile(string) ([]byte, error)
	}).ReadFile("a/b/c.txt")
	if err != nil {
		t.Fatalf("ReadFile(a/b/c.txt): %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("cross-platform FS: got %q, want %q", got, content)
	}
}
