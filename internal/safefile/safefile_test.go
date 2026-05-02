// SPDX-License-Identifier: Apache-2.0

package safefile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
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
			require.NoError(t, err, "Clean("+name+")")
			assert.True(t, clean != "", "Clean("+name+") returned empty string")
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
			assert.ErrorIs(t, err, safefile.ErrTraversal, "Clean("+name+") should return ErrTraversal")
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
			assert.ErrorIs(t, err, safefile.ErrAbsolute, "Clean("+name+") should return ErrAbsolute")
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
			assert.ErrorIs(t, err, safefile.ErrUNC, "Clean("+name+") should return ErrUNC")
		})
	}
}

func TestReadFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	content := []byte("safefile test content")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.txt"), content, 0o644))
	got, err := safefile.ReadFile(dir, "test.txt")
	require.NoError(t, err, "ReadFile")
	assert.Equal(t, string(content), string(got))
}

func TestWriteFileAtomicRoundTrip(t *testing.T) {
	dir := t.TempDir()
	content := []byte("atomic write content")
	require.NoError(t, safefile.WriteFileAtomic(dir, "atomic.txt", content, 0o644), "WriteFileAtomic")
	got, err := os.ReadFile(filepath.Join(dir, "atomic.txt"))
	require.NoError(t, err, "ReadFile after atomic write")
	assert.Equal(t, string(content), string(got))
}

func TestJoinRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	_, err := safefile.Join(dir, "../oops")
	assert.ErrorIs(t, err, safefile.ErrTraversal, "Join with traversal should return ErrTraversal")
}
