// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"sync"
	"time"
)

// memEntry holds one in-memory cache entry.
// expiresAt is the zero time when the entry has no expiry.
type memEntry[V any] struct {
	value     V
	expiresAt time.Time
}

// memCache is the in-memory Cache[V] implementation.
type memCache[V any] struct {
	cfg     config
	mu      sync.RWMutex
	entries map[string]memEntry[V]
	sf      singleflight[V]
}

// New constructs an in-memory Cache[V]. The returned cache is ready to use
// and goroutine-safe.
func New[V any](opts ...Option) Cache[V] {
	return &memCache[V]{
		cfg:     applyOptions(opts...),
		entries: make(map[string]memEntry[V]),
	}
}

// Get implements Cache.
func (m *memCache[V]) Get(key string) (V, bool) {
	m.mu.RLock()
	e, ok := m.entries[key]
	m.mu.RUnlock()
	if !ok {
		return zero[V](), false
	}
	if !e.expiresAt.IsZero() && !m.cfg.clock().Before(e.expiresAt) {
		// Expired; prune lazily.
		m.mu.Lock()
		// Recheck under write lock to avoid racing with a concurrent Set.
		if cur, still := m.entries[key]; still && cur.expiresAt == e.expiresAt {
			delete(m.entries, key)
		}
		m.mu.Unlock()
		return zero[V](), false
	}
	return e.value, true
}

// Set implements Cache.
func (m *memCache[V]) Set(key string, value V) {
	m.SetWithTTL(key, value, m.cfg.defaultTTL)
}

// SetWithTTL implements Cache.
func (m *memCache[V]) SetWithTTL(key string, value V, ttl time.Duration) {
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = m.cfg.clock().Add(ttl)
	}
	m.mu.Lock()
	m.entries[key] = memEntry[V]{value: value, expiresAt: expiresAt}
	m.mu.Unlock()
}

// Delete implements Cache.
func (m *memCache[V]) Delete(key string) {
	m.mu.Lock()
	delete(m.entries, key)
	m.mu.Unlock()
}

// GetOrLoad implements Cache.
func (m *memCache[V]) GetOrLoad(ctx context.Context, key string, loader func(context.Context) (V, error)) (V, error) {
	if v, ok := m.Get(key); ok {
		return v, nil
	}
	return m.sf.do(ctx, key, func(ctx context.Context) (V, error) {
		// Re-check inside singleflight to handle the case where another
		// caller populated the entry while we waited.
		if v, ok := m.Get(key); ok {
			return v, nil
		}
		v, err := loader(ctx)
		if err != nil {
			return zero[V](), err
		}
		m.Set(key, v)
		return v, nil
	})
}

// zero returns the zero value of V. Defined as a helper for readability.
func zero[V any]() V {
	var z V
	return z
}
