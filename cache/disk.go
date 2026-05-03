// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nathanbrophy/glacier/internal/lockfile"
	"github.com/nathanbrophy/glacier/internal/safefile"
)

// diskCache is the disk-backed Cache[V] implementation. Each key is stored
// as a separate JSON file at <root>/<sha256(key)>.json. A sibling .lock file
// arbitrates cross-process access via internal/lockfile.
type diskCache[V any] struct {
	cfg  config
	root string
	sf   singleflight[V]
	// lockTimeout bounds how long any single Get/Set will wait for the file
	// lock. Defaults to 30s; tests override via package-internal accessor.
	lockTimeout time.Duration
}

// diskRecord is the on-disk JSON shape (one record per file).
type diskRecord[V any] struct {
	Value     json.RawMessage `json:"value"`
	StoredAt  time.Time       `json:"stored_at"`
	ExpiresAt time.Time       `json:"expires_at,omitempty"`
}

// NewDisk constructs a disk-backed Cache[V] rooted at path. The directory is
// created with 0o700 if it does not exist. Returns an error only if path is
// not a directory or cannot be created.
func NewDisk[V any](path string, opts ...Option) (Cache[V], error) {
	if path == "" {
		return nil, errors.New("cache: NewDisk: path is required")
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return nil, fmt.Errorf("cache: NewDisk: mkdir %s: %w", path, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cache: NewDisk: stat %s: %w", path, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("cache: NewDisk: %s: not a directory", path)
	}
	return &diskCache[V]{
		cfg:         applyOptions(opts...),
		root:        path,
		lockTimeout: 30 * time.Second,
	}, nil
}

// keyFile returns the absolute path to the data file for key.
// The filename is sha256(key) so untrusted keys cannot escape the cache root.
func (d *diskCache[V]) keyFile(key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(d.root, hex.EncodeToString(sum[:])+".json")
}

// lockPath returns the absolute path to the per-key lock file.
func (d *diskCache[V]) lockPath(key string) string {
	return d.keyFile(key) + ".lock"
}

// Get implements Cache.
func (d *diskCache[V]) Get(key string) (V, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), d.lockTimeout)
	defer cancel()

	unlock, err := lockfile.RLock(ctx, d.lockPath(key))
	if err != nil {
		d.cfg.logger.Warn("cache disk: lock timeout on Get", "key_hash_prefix", hashPrefix(key), "err", err)
		return zero[V](), false
	}
	defer func() { _ = unlock() }()

	data, err := os.ReadFile(d.keyFile(key))
	if err != nil {
		return zero[V](), false
	}

	var rec diskRecord[V]
	if err := json.Unmarshal(data, &rec); err != nil {
		// Corrupt file: delete and miss.
		d.cfg.logger.Warn("cache disk: corrupt entry deleted", "key_hash_prefix", hashPrefix(key), "err", err)
		_ = os.Remove(d.keyFile(key))
		return zero[V](), false
	}

	if !rec.ExpiresAt.IsZero() && !d.cfg.clock().Before(rec.ExpiresAt) {
		// Expired: prune.
		_ = os.Remove(d.keyFile(key))
		return zero[V](), false
	}

	var value V
	if err := json.Unmarshal(rec.Value, &value); err != nil {
		d.cfg.logger.Warn("cache disk: value unmarshal failed", "key_hash_prefix", hashPrefix(key), "err", err)
		_ = os.Remove(d.keyFile(key))
		return zero[V](), false
	}
	return value, true
}

// Set implements Cache.
func (d *diskCache[V]) Set(key string, value V) {
	d.SetWithTTL(key, value, d.cfg.defaultTTL)
}

// SetWithTTL implements Cache.
func (d *diskCache[V]) SetWithTTL(key string, value V, ttl time.Duration) {
	now := d.cfg.clock()
	rec := diskRecord[V]{StoredAt: now}
	if ttl > 0 {
		rec.ExpiresAt = now.Add(ttl)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		d.cfg.logger.Warn("cache disk: value marshal failed", "key_hash_prefix", hashPrefix(key), "err", err)
		return
	}
	rec.Value = encoded

	data, err := json.Marshal(rec)
	if err != nil {
		d.cfg.logger.Warn("cache disk: record marshal failed", "key_hash_prefix", hashPrefix(key), "err", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.lockTimeout)
	defer cancel()
	unlock, err := lockfile.Lock(ctx, d.lockPath(key))
	if err != nil {
		d.cfg.logger.Warn("cache disk: lock timeout on Set", "key_hash_prefix", hashPrefix(key), "err", err)
		return
	}
	defer func() { _ = unlock() }()

	relPath, err := filepath.Rel(d.root, d.keyFile(key))
	if err != nil {
		d.cfg.logger.Warn("cache disk: rel path failed", "err", err)
		return
	}
	if err := safefile.WriteFileAtomic(d.root, filepath.ToSlash(relPath), data, 0o600); err != nil {
		d.cfg.logger.Warn("cache disk: atomic write failed", "key_hash_prefix", hashPrefix(key), "err", err)
		return
	}
}

// Delete implements Cache.
func (d *diskCache[V]) Delete(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), d.lockTimeout)
	defer cancel()
	unlock, err := lockfile.Lock(ctx, d.lockPath(key))
	if err != nil {
		d.cfg.logger.Warn("cache disk: lock timeout on Delete", "key_hash_prefix", hashPrefix(key), "err", err)
		return
	}
	defer func() { _ = unlock() }()
	_ = os.Remove(d.keyFile(key))
}

// GetOrLoad implements Cache.
func (d *diskCache[V]) GetOrLoad(ctx context.Context, key string, loader func(context.Context) (V, error)) (V, error) {
	if v, ok := d.Get(key); ok {
		return v, nil
	}
	return d.sf.do(ctx, key, func(ctx context.Context) (V, error) {
		if v, ok := d.Get(key); ok {
			return v, nil
		}
		v, err := loader(ctx)
		if err != nil {
			return zero[V](), err
		}
		d.Set(key, v)
		return v, nil
	})
}

// hashPrefix returns the first 8 hex characters of sha256(key) for safe logging.
// Logging the full key may leak sensitive data; the prefix is enough to
// correlate operations on the same key in logs without exposing it.
func hashPrefix(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])[:8]
}
