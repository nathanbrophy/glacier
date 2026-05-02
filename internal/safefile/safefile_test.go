// SPDX-License-Identifier: Apache-2.0

package safefile_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/nathanbrophy/glacier/internal/safefile"
)

func TestCleanAcceptsRelative(t *testing.T) {
	cases := []string{
		"foo.txt",
		"a/b/c.txt",
		"testdata/golden.txt",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			clean, err := safefile.Clean(name)
			if err != nil {
				t.Fatalf("Clean(%q) = %v; want nil", name, err)
			}
			if clean == "" {
				t.Fatalf("Clean(%q) returned empty string", name)
			}
		})
	}
}

func TestCleanRejectsTraversal(t *testing.T) {
	cases := []string{
		"../foo",
		"a/../../b",
		"..",
		"a/../..",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := safefile.Clean(name)
			if !errors.Is(err, safefile.ErrTraversal) {
				t.Fatalf("Clean(%q) = %v; want ErrTraversal", name, err)
			}
		})
	}
}

func TestCleanRejectsAbsolute(t *testing.T) {
	cases := []string{
		"/etc/passwd",
		"/foo/bar",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := safefile.Clean(name)
			if !errors.Is(err, safefile.ErrAbsolute) {
				t.Fatalf("Clean(%q) = %v; want ErrAbsolute", name, err)
			}
		})
	}
}

func TestCleanRejectsUNC(t *testing.T) {
	cases := []string{
		`\\server\share`,
		`\\?\C:\foo`,
		"//server/share",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := safefile.Clean(name)
			if !errors.Is(err, safefile.ErrUNC) {
				t.Fatalf("Clean(%q) = %v; want ErrUNC", name, err)
			}
		})
	}
}

func TestReadFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	content := []byte("safefile test content")
	if err := os.WriteFile(filepath.Join(dir, "test.txt"), content, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := safefile.ReadFile(dir, "test.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("ReadFile got %q; want %q", got, content)
	}
}

func TestWriteFileAtomicRoundTrip(t *testing.T) {
	dir := t.TempDir()
	content := []byte("atomic write content")
	if err := safefile.WriteFileAtomic(dir, "atomic.txt", content, 0o644); err != nil {
		t.Fatalf("WriteFileAtomic: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "atomic.txt"))
	if err != nil {
		t.Fatalf("ReadFile after atomic write: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("atomic write content mismatch: got %q; want %q", got, content)
	}
}

func TestJoinRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	_, err := safefile.Join(dir, "../oops")
	if !errors.Is(err, safefile.ErrTraversal) {
		t.Fatalf("Join with traversal = %v; want ErrTraversal", err)
	}
}
