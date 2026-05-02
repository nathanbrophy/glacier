//go:build glacier_debug

// SPDX-License-Identifier: Apache-2.0

package concur

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

const defaultHoldTimeout = 5 * time.Second

// Mutex with hold-timeout diagnostics in glacier_debug builds.
type Mutex struct {
	mu          sync.Mutex
	acquiredAt  time.Time
	callerStack string
	holdTimeout time.Duration
	timer       *time.Timer
	timerMu     sync.Mutex
}

func (m *Mutex) Lock() {
	m.mu.Lock()
	m.recordAcquire()
}

func (m *Mutex) Unlock() {
	m.cancelTimer()
	m.acquiredAt = time.Time{}
	m.callerStack = ""
	m.mu.Unlock()
}

func (m *Mutex) LockCtx(ctx context.Context) error {
	if m.mu.TryLock() {
		m.recordAcquire()
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
		if m.mu.TryLock() {
			m.recordAcquire()
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

func (m *Mutex) recordAcquire() {
	m.acquiredAt = time.Now()
	m.callerStack = captureStack()
	timeout := m.holdTimeout
	if timeout == 0 {
		timeout = defaultHoldTimeout
	}
	m.timerMu.Lock()
	m.timer = time.AfterFunc(timeout, func() {
		slog.Warn("concur: mutex held too long",
			"holder", m.callerStack,
			"elapsed", time.Since(m.acquiredAt),
		)
	})
	m.timerMu.Unlock()
}

func (m *Mutex) cancelTimer() {
	m.timerMu.Lock()
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
	m.timerMu.Unlock()
}

func captureStack() string {
	buf := make([]byte, 2048)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// RWMutex in debug build (simplified — same as non-debug for now since hold-timeout on RWMutex adds complexity).
type RWMutex struct {
	mu sync.RWMutex
}

func (m *RWMutex) Lock()    { m.mu.Lock() }
func (m *RWMutex) Unlock()  { m.mu.Unlock() }
func (m *RWMutex) RLock()   { m.mu.RLock() }
func (m *RWMutex) RUnlock() { m.mu.RUnlock() }

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
