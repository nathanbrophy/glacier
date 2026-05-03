// SPDX-License-Identifier: Apache-2.0

package cache_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/cache"
)

func TestDiskHit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c, err := cache.NewDisk[string](dir)
	require.Nil(t, err)
	c.Set("k", "v")

	v, ok := c.Get("k")
	require.True(t, ok)
	assert.Equal(t, "v", v)

	// File exists at <hash>.json.
	sum := sha256.Sum256([]byte("k"))
	path := filepath.Join(dir, hex.EncodeToString(sum[:])+".json")
	_, statErr := os.Stat(path)
	assert.Nil(t, statErr)
}

func TestDiskMiss(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c, err := cache.NewDisk[string](dir)
	require.Nil(t, err)
	_, ok := c.Get("absent")
	assert.False(t, ok)
}

func TestDiskExpiry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	clock := newFakeClock()
	c, err := cache.NewDisk[int](dir, cache.WithClock(clock.Now))
	require.Nil(t, err)
	c.SetWithTTL("k", 1, 1*time.Minute)

	// Within TTL.
	_, ok := c.Get("k")
	require.True(t, ok)

	// Past TTL.
	clock.Advance(2 * time.Minute)
	_, ok = c.Get("k")
	assert.False(t, ok)
}

func TestDiskPersistsAcrossInstances(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c1, err := cache.NewDisk[string](dir)
	require.Nil(t, err)
	c1.Set("k", "v")

	// Fresh instance pointing at same dir.
	c2, err := cache.NewDisk[string](dir)
	require.Nil(t, err)
	v, ok := c2.Get("k")
	require.True(t, ok)
	assert.Equal(t, "v", v)
}

func TestDiskCorruptFileMisses(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c, err := cache.NewDisk[string](dir)
	require.Nil(t, err)

	// Manually plant a corrupt file at the right hash path.
	sum := sha256.Sum256([]byte("k"))
	path := filepath.Join(dir, hex.EncodeToString(sum[:])+".json")
	require.Nil(t, os.WriteFile(path, []byte("not json"), 0o600))

	_, ok := c.Get("k")
	assert.False(t, ok)

	// File should be deleted as side effect.
	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr))
}

func TestDiskPathTraversalBlocked(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c, err := cache.NewDisk[string](dir)
	require.Nil(t, err)

	// Even a key full of path-traversal bytes lands inside the cache root
	// because the filename is sha256(key), not key itself.
	c.Set("../../../etc/passwd", "v")

	// Walk the cache root: every regular file lives directly under it.
	walkErr := filepath.Walk(dir, func(p string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.Mode().IsRegular() {
			rel, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			// Files should be at the top level of the cache root, not nested.
			assert.False(t, strings.Contains(rel, string(filepath.Separator)),
				"file %q escaped the cache root", rel)
		}
		return nil
	})
	assert.Nil(t, walkErr)
}

func TestDiskDeleteRemovesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c, err := cache.NewDisk[string](dir)
	require.Nil(t, err)
	c.Set("k", "v")
	c.Delete("k")
	_, ok := c.Get("k")
	assert.False(t, ok)
}

func TestNewDiskRejectsEmptyPath(t *testing.T) {
	t.Parallel()
	_, err := cache.NewDisk[string]("")
	assert.NotNil(t, err)
}

func TestNewDiskCreatesDir(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	target := filepath.Join(parent, "newdir")
	_, err := cache.NewDisk[string](target)
	require.Nil(t, err)
	info, statErr := os.Stat(target)
	require.Nil(t, statErr)
	assert.True(t, info.IsDir())
}

func TestDiskRejectsNonDirPath(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	file := filepath.Join(parent, "afile")
	require.Nil(t, os.WriteFile(file, []byte("x"), 0o600))
	_, err := cache.NewDisk[string](file)
	assert.NotNil(t, err)
}

func TestDiskHashFilenameUnique(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c, err := cache.NewDisk[string](dir)
	require.Nil(t, err)
	const N = 1000
	for i := 0; i < N; i++ {
		key := strings.Repeat("a", i+1)
		c.Set(key, "v")
	}
	entries, err := os.ReadDir(dir)
	require.Nil(t, err)
	jsonCount := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			jsonCount++
		}
	}
	assert.Equal(t, N, jsonCount)
}
