// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/nathanbrophy/glacier/option"
)

// Cache is the generic key-value cache contract. Every implementation in this
// package satisfies Cache[V] so callers can compose them and tests can mock
// them. The zero value of V is returned alongside ok=false on any miss.
//
// All methods are goroutine-safe. The disk-backed implementation is also
// process-safe via advisory file locking.
//
// +glacier:mock
type Cache[V any] interface {
	// Get returns the value for key and whether it was found and unexpired.
	// On miss (absent or expired), the zero value of V is returned with ok=false.
	Get(key string) (value V, ok bool)

	// Set stores value under key with the cache's default TTL. If no default
	// TTL is configured (TTL == 0), the entry is stored without expiry.
	Set(key string, value V)

	// SetWithTTL stores value under key with an explicit TTL. ttl <= 0 stores
	// the entry without expiry.
	SetWithTTL(key string, value V, ttl time.Duration)

	// Delete removes the entry for key. No-op if absent.
	Delete(key string)

	// GetOrLoad returns the value for key. On miss, loader is called and the
	// result is stored before being returned. Concurrent misses on the same
	// key share a single loader call (singleflight). The loader's context is
	// ctx; loader errors are not cached.
	GetOrLoad(ctx context.Context, key string, loader func(context.Context) (V, error)) (V, error)
}

// Option configures a cache implementation at construction time.
// Cache options dogfood the framework's option.Option pattern (D-C20):
// they compose with option.Apply so callers can reason about ordering
// and validation uniformly across Glacier packages.
type Option = option.Option[config]

// config aggregates Option values inside the cache constructors.
// invariant: defaultTTL >= 0
// invariant: clock != nil after applyOptions
// invariant: logger != nil after applyOptions
type config struct {
	defaultTTL time.Duration
	clock      func() time.Time
	logger     *slog.Logger
}

// applyOptions starts from sensible defaults and folds in the given Options
// via option.Apply. Errors from option.Apply panic in v0; cache options never
// return errors today, so this is a no-op in practice.
func applyOptions(opts ...Option) config {
	c, err := option.Apply(opts)
	if err != nil {
		// option.Apply only errors when an Option's Apply returns one.
		// Cache options never return errors at v0, so this is unreachable.
		//glacier:nolint=panic-in-library unreachable: cache options never return errors at v0.
		panic("cache: option.Apply returned unexpected error: " + err.Error())
	}
	if c.clock == nil {
		c.clock = time.Now
	}
	if c.logger == nil {
		c.logger = slog.Default()
	}
	return c
}

// WithDefaultTTL sets the default TTL applied by Set. ttl <= 0 means no expiry.
// The default is 0 (no expiry).
func WithDefaultTTL(ttl time.Duration) Option {
	return option.OptionFunc[config](func(c *config) error {
		if ttl < 0 {
			ttl = 0
		}
		c.defaultTTL = ttl
		return nil
	})
}

// WithLogger sets the slog.Logger used for non-fatal cache messages such as
// "corrupt entry deleted". The default is slog.Default().
func WithLogger(l *slog.Logger) Option {
	return option.OptionFunc[config](func(c *config) error {
		if l != nil {
			c.logger = l
		}
		return nil
	})
}

// WithClock injects a clock function. Tests pass a deterministic clock to
// exercise expiry without sleeping. The default is time.Now.
func WithClock(now func() time.Time) Option {
	return option.OptionFunc[config](func(c *config) error {
		if now != nil {
			c.clock = now
		}
		return nil
	})
}
