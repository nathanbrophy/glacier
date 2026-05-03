// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"time"
)

// layeredCache composes a primary and a backing Cache[V] with write-through
// semantics: reads consult primary first, then backing (and on backing-hit
// populate primary); writes go to both. backing-layer errors degrade the
// composition to primary-only for the failing operation but never propagate
// to the caller.
type layeredCache[V any] struct {
	primary Cache[V]
	backing Cache[V]
	sf      singleflight[V]
}

// NewLayered composes a primary and a backing Cache[V] with write-through
// semantics. A typical pairing is an in-memory primary with a disk-backed
// backing so the cache survives restarts but stays fast on the hot path.
func NewLayered[V any](primary, backing Cache[V]) Cache[V] {
	return &layeredCache[V]{primary: primary, backing: backing}
}

// Get implements Cache.
func (l *layeredCache[V]) Get(key string) (V, bool) {
	if v, ok := l.primary.Get(key); ok {
		return v, true
	}
	if v, ok := l.backing.Get(key); ok {
		// Populate primary so subsequent reads are fast.
		l.primary.Set(key, v)
		return v, true
	}
	return zero[V](), false
}

// Set implements Cache.
func (l *layeredCache[V]) Set(key string, value V) {
	l.primary.Set(key, value)
	l.backing.Set(key, value)
}

// SetWithTTL implements Cache.
func (l *layeredCache[V]) SetWithTTL(key string, value V, ttl time.Duration) {
	l.primary.SetWithTTL(key, value, ttl)
	l.backing.SetWithTTL(key, value, ttl)
}

// Delete implements Cache.
func (l *layeredCache[V]) Delete(key string) {
	l.primary.Delete(key)
	l.backing.Delete(key)
}

// GetOrLoad implements Cache. The singleflight is layered-cache-local, not
// per-implementation, so concurrent misses across both layers collapse onto
// one loader call.
func (l *layeredCache[V]) GetOrLoad(ctx context.Context, key string, loader func(context.Context) (V, error)) (V, error) {
	if v, ok := l.Get(key); ok {
		return v, nil
	}
	return l.sf.do(ctx, key, func(ctx context.Context) (V, error) {
		if v, ok := l.Get(key); ok {
			return v, nil
		}
		v, err := loader(ctx)
		if err != nil {
			return zero[V](), err
		}
		l.Set(key, v)
		return v, nil
	})
}
