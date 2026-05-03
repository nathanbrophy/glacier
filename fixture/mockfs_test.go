// SPDX-License-Identifier: Apache-2.0

package fixture_test

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/nathanbrophy/glacier/fixture"
)

// TestMockFSRead: NewFS({"a":bytes}); fs.ReadFile("a") returns bytes. (#38)
func TestMockFSRead(t *testing.T) {
	want := []byte("hello from memory")
	fsys := fixture.NewFS(map[string][]byte{"a.txt": want})
	got, err := fs.ReadFile(fsys, "a.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("ReadFile got %q; want %q", got, want)
	}
}

// TestMockFSReadDir: ReadDir lists entries. (#39)
func TestMockFSReadDir(t *testing.T) {
	fsys := fixture.NewFS(map[string][]byte{
		"a.txt": []byte("a"),
		"b.txt": []byte("b"),
		"c.txt": []byte("c"),
	})
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("ReadDir returned %d entries; want 3", len(entries))
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	// Entries should be sorted.
	if names[0] != "a.txt" || names[1] != "b.txt" || names[2] != "c.txt" {
		t.Fatalf("ReadDir names %v; want [a.txt b.txt c.txt]", names)
	}
}

// TestMockFSReadFileInterface: Returned FS satisfies fs.ReadFileFS and fs.ReadDirFS.
// (#40)
func TestMockFSReadFileInterface(t *testing.T) {
	fsys := fixture.NewFS(map[string][]byte{"x": []byte("x")})
	if _, ok := fsys.(fs.ReadFileFS); !ok {
		t.Fatal("NewFS result does not implement fs.ReadFileFS")
	}
	if _, ok := fsys.(fs.ReadDirFS); !ok {
		t.Fatal("NewFS result does not implement fs.ReadDirFS")
	}
}

// TestMockFSConflictPanics: NewFS with file-and-dir conflict panics. (#41)
func TestMockFSConflictPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("NewFS did not panic on file/dir conflict")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("panic value is not a string: %T", r)
		}
		// Check for the expected message pattern.
		if !contains(msg, "conflict") {
			t.Fatalf("panic message %q does not contain 'conflict'", msg)
		}
	}()
	// "foo" as a file and "foo/bar" implies "foo" as a directory :  conflict.
	fixture.NewFS(map[string][]byte{
		"foo":     []byte("file"),
		"foo/bar": []byte("other"),
	})
}

// TestMockFSNestedPaths: Path "a/b/c.txt" accessible. (#42)
func TestMockFSNestedPaths(t *testing.T) {
	content := []byte("deep content")
	fsys := fixture.NewFS(map[string][]byte{
		"a/b/c.txt": content,
	})
	got, err := fs.ReadFile(fsys, "a/b/c.txt")
	if err != nil {
		t.Fatalf("ReadFile nested: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("nested read got %q; want %q", got, content)
	}
}

// TestMockFSReadOnly: No write methods on returned FS. (#43)
func TestMockFSReadOnly(t *testing.T) {
	fsys := fixture.NewFS(map[string][]byte{"r.txt": []byte("readonly")})
	// fs.FS interface only has Open; write interfaces do not exist in stdlib.
	// Verify the FS does not implement any writable interface.
	type creator interface{ Create(string) (fs.File, error) }
	type writer interface {
		WriteFile(string, []byte, fs.FileMode) error
	}
	if _, ok := fsys.(creator); ok {
		t.Fatal("NewFS result unexpectedly implements Create")
	}
	if _, ok := fsys.(writer); ok {
		t.Fatal("NewFS result unexpectedly implements WriteFile")
	}
}

// TestMockFSMissingFile: ReadFile on missing path returns fs.ErrNotExist. (#44)
func TestMockFSMissingFile(t *testing.T) {
	fsys := fixture.NewFS(map[string][]byte{})
	_, err := fs.ReadFile(fsys, "missing.txt")
	if !isNotExist(err) {
		t.Fatalf("ReadFile(missing) = %v; want fs.ErrNotExist", err)
	}
}

// TestMockFSStat: Stat on a directory entry reports IsDir() == true. (#45, EX4)
func TestMockFSStat(t *testing.T) {
	fsys := fixture.NewFS(map[string][]byte{
		"dir/file.txt": []byte("content"),
	})
	f, err := fsys.Open("dir")
	if err != nil {
		t.Fatalf("Open(dir): %v", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat(dir): %v", err)
	}
	if !info.IsDir() {
		t.Fatal("Stat(dir).IsDir() = false; want true")
	}
}

// TestMockFSEmptyFS: Empty FS works without panic.
func TestMockFSEmptyFS(t *testing.T) {
	fsys := fixture.NewFS(nil)
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		t.Fatalf("ReadDir(.) on empty FS: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("ReadDir(.) on empty FS: got %d entries, want 0", len(entries))
	}
}

// Table-driven FS tests.
func TestMockFSTableDriven(t *testing.T) {
	cases := []struct {
		name    string
		files   map[string][]byte
		readKey string
		wantOK  bool
	}{
		{"single_file", map[string][]byte{"x.txt": []byte("x")}, "x.txt", true},
		{"missing", map[string][]byte{"x.txt": []byte("x")}, "y.txt", false},
		{"nested", map[string][]byte{"a/b.txt": []byte("nested")}, "a/b.txt", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fsys := fixture.NewFS(tc.files)
			_, err := fs.ReadFile(fsys, tc.readKey)
			if (err == nil) != tc.wantOK {
				t.Fatalf("ReadFile(%q) error=%v; wantOK=%v", tc.readKey, err, tc.wantOK)
			}
		})
	}
}

// helpers used across tests.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func isNotExist(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, fs.ErrNotExist)
}
