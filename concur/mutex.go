//go:build !glacier_debug

// SPDX-License-Identifier: Apache-2.0

package concur

import (
	"context"
	"sync"
	"time"
)

// Mutex is a mutual-exclusion lock. In default builds, byte-equivalent to sync.Mutex.
// LockCtx is ctx-aware: it polls with a short backoff until the lock is acquired or ctx is cancelled.
type Mutex struct {
	mu sync.Mutex
}

func (m *Mutex) Lock()   { m.mu.Lock() }
func (m *Mutex) Unlock() { m.mu.Unlock() }

// LockCtx attempts to acquire the lock, returning ErrCancelled if ctx is
// cancelled before the lock is obtained. Uses a try-lock-with-backoff approach
// with 1ms initial backoff doubling to 16ms max.
func (m *Mutex) LockCtx(ctx context.Context) error {
	// Fast path: try once with no blocking.
	if m.mu.TryLock() {
		return nil
	}
	// Slow path: poll with exponential backoff.
	backoff := time.Millisecond
	const maxBackoff = 16 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return ErrCancelled
		default:
		}
		if m.mu.TryLock() {
			return nil
		}
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ErrCancelled
		case <-timer.C:
		}
		if backoff < maxBackoff {
			backoff *= 2
		}
	}
}

// RWMutex is a reader/writer mutual-exclusion lock. Byte-equivalent to sync.RWMutex.
type RWMutex struct {
	mu sync.RWMutex
}

func (m *RWMutex) Lock()    { m.mu.Lock() }
func (m *RWMutex) Unlock()  { m.mu.Unlock() }
func (m *RWMutex) RLock()   { m.mu.RLock() }
func (m *RWMutex) RUnlock() { m.mu.RUnlock() }

// RLockCtx is the ctx-aware read-lock. Same backoff approach as LockCtx.
func (m *RWMutex) RLockCtx(ctx context.Context) error {
	if m.mu.TryRLock() {
		return nil
	}
	backoff := time.Millisecond
	const maxBackoff = 16 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return ErrCancelled
		default:
		}
		if m.mu.TryRLock() {
			return nil
		}
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ErrCancelled
		case <-timer.C:
		}
		if backoff < maxBackoff {
			backoff *= 2
		}
	}
}
