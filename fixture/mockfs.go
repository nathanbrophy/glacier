// SPDX-License-Identifier: Apache-2.0

package fixture

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"
)

// memFS is a read-only in-memory fs.FS satisfying fs.ReadFileFS and fs.ReadDirFS.
type memFS struct {
	// files maps clean slash-separated paths to file contents.
	files map[string][]byte
	// dirs is the set of directory paths (including all implicit parent dirs).
	dirs map[string]struct{}
}

// NewFS constructs a read-only in-memory fs.FS from files. Keys are path
// strings (forward-slash separated); values are file contents. NewFS panics
// at construction if any two paths conflict (a path is simultaneously claimed
// as a file and as a directory).
//
// The returned FS satisfies fs.ReadFileFS and fs.ReadDirFS.
func NewFS(files map[string][]byte) fs.FS {
	m := &memFS{
		files: make(map[string][]byte, len(files)),
		dirs:  make(map[string]struct{}),
	}
	// Always include the root directory.
	m.dirs["."] = struct{}{}

	for rawPath, data := range files {
		// Normalize the path.
		cleaned := path.Clean(strings.TrimPrefix(rawPath, "/"))
		if cleaned == "." {
			panic(fmt.Sprintf("fixture: NewFS: invalid path %q", rawPath))
		}
		if !fs.ValidPath(cleaned) {
			panic(fmt.Sprintf("fixture: NewFS: invalid fs.FS path %q", cleaned))
		}

		// Check for file/dir conflict: is cleaned already a directory?
		if _, isDir := m.dirs[cleaned]; isDir {
			panic(fmt.Sprintf("fixture: NewFS: conflict at path %q: both file and directory", cleaned))
		}
		m.files[cleaned] = data

		// Register all implicit parent directories; check for dir/file conflict.
		dir := path.Dir(cleaned)
		for dir != "." {
			if _, isFile := m.files[dir]; isFile {
				panic(fmt.Sprintf("fixture: NewFS: conflict at path %q: both file and directory", dir))
			}
			m.dirs[dir] = struct{}{}
			dir = path.Dir(dir)
		}
	}
	return m
}

// Open implements fs.FS.
func (m *memFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	// Directory?
	if _, ok := m.dirs[name]; ok {
		return &memDir{fs: m, path: name}, nil
	}
	// File?
	data, ok := m.files[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return &memFile{path: name, data: data}, nil
}

// ReadFile implements fs.ReadFileFS.
func (m *memFS) ReadFile(name string) ([]byte, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrInvalid}
	}
	data, ok := m.files[name]
	if !ok {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrNotExist}
	}
	// Return a copy so callers cannot mutate internal state.
	out := make([]byte, len(data))
	copy(out, data)
	return out, nil
}

// ReadDir implements fs.ReadDirFS.
func (m *memFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}
	if _, ok := m.dirs[name]; !ok {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	return m.listDir(name), nil
}

// listDir returns the immediate children of dir, sorted by name.
func (m *memFS) listDir(dir string) []fs.DirEntry {
	seen := make(map[string]struct{})
	var entries []fs.DirEntry

	prefix := dir + "/"
	if dir == "." {
		prefix = ""
	}

	// Files.
	for p := range m.files {
		if !strings.HasPrefix(p, prefix) {
			continue
		}
		rest := p[len(prefix):]
		if rest == "" || strings.Contains(rest, "/") {
			continue
		}
		if _, ok := seen[rest]; ok {
			continue
		}
		seen[rest] = struct{}{}
		entries = append(entries, &memDirEntry{name: rest, isDir: false, data: m.files[p]})
	}

	// Sub-directories.
	for d := range m.dirs {
		if d == dir || d == "." {
			continue
		}
		if !strings.HasPrefix(d, prefix) {
			continue
		}
		rest := d[len(prefix):]
		if rest == "" || strings.Contains(rest, "/") {
			continue
		}
		if _, ok := seen[rest]; ok {
			continue
		}
		seen[rest] = struct{}{}
		entries = append(entries, &memDirEntry{name: rest, isDir: true})
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	return entries
}

// ── memFile ───────────────────────────────────────────────────────────────────

type memFile struct {
	path   string
	data   []byte
	offset int
}

// Stat implements fs.File.
func (f *memFile) Stat() (fs.FileInfo, error) {
	return &memFileInfo{name: path.Base(f.path), size: int64(len(f.data)), isDir: false}, nil
}

// Read implements fs.File.
func (f *memFile) Read(b []byte) (int, error) {
	if f.offset >= len(f.data) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(b, f.data[f.offset:])
	f.offset += n
	return n, nil
}

// Close implements fs.File and is a no-op for in-memory files.
func (f *memFile) Close() error { return nil }

// ── memDir ────────────────────────────────────────────────────────────────────

type memDir struct {
	fs   *memFS
	path string
}

// Stat implements fs.File.
func (d *memDir) Stat() (fs.FileInfo, error) {
	return &memFileInfo{name: path.Base(d.path), size: 0, isDir: true}, nil
}

// Read implements fs.File and always returns a "is a directory" error.
func (d *memDir) Read(_ []byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.path, Err: fmt.Errorf("is a directory")}
}

// Close implements fs.File and is a no-op for in-memory directories.
func (d *memDir) Close() error { return nil }

// ── memFileInfo ───────────────────────────────────────────────────────────────

type memFileInfo struct {
	name  string
	size  int64
	isDir bool
}

// Name implements fs.FileInfo.
func (i *memFileInfo) Name() string { return i.name }

// Size implements fs.FileInfo.
func (i *memFileInfo) Size() int64 { return i.size }

// Mode implements fs.FileInfo and reports a fixed read-only mode.
func (i *memFileInfo) Mode() fs.FileMode { return 0o444 }

// ModTime implements fs.FileInfo and returns the zero time.
func (i *memFileInfo) ModTime() time.Time { return time.Time{} }

// IsDir implements fs.FileInfo.
func (i *memFileInfo) IsDir() bool { return i.isDir }

// Sys implements fs.FileInfo and returns nil (no underlying system source).
func (i *memFileInfo) Sys() any { return nil }

// ── memDirEntry ───────────────────────────────────────────────────────────────

type memDirEntry struct {
	name  string
	isDir bool
	data  []byte
}

// Name implements fs.DirEntry.
func (e *memDirEntry) Name() string { return e.name }

// IsDir implements fs.DirEntry.
func (e *memDirEntry) IsDir() bool { return e.isDir }

// Type implements fs.DirEntry.
func (e *memDirEntry) Type() fs.FileMode {
	if e.isDir {
		return fs.ModeDir
	}
	return 0
}

// Info implements fs.DirEntry.
func (e *memDirEntry) Info() (fs.FileInfo, error) {
	if e.isDir {
		return &memFileInfo{name: e.name, size: 0, isDir: true}, nil
	}
	return &memFileInfo{name: e.name, size: int64(len(e.data)), isDir: false}, nil
}
