// SPDX-License-Identifier: Apache-2.0

package concur

import (
	"context"
	"runtime"
	"sync"

	"github.com/nathanbrophy/glacier/errs"
	"github.com/nathanbrophy/glacier/option"
)

const unlimited = -1

type groupConfig struct {
	limit int // invariant: >= 1 or unlimited (-1)
}

// Group collects errors from all goroutines and offers a concurrency cap.
type Group struct {
	sem     *Semaphore
	wg      sync.WaitGroup
	mu      sync.Mutex
	errList []error
	closed  bool
}

// NewGroup creates a Group with optional configuration.
// Default concurrency limit: runtime.NumCPU() * 64.
func NewGroup(opts ...option.Option[groupConfig]) *Group {
	cfg, _ := option.Apply(opts)
	if cfg.limit == 0 {
		cfg.limit = runtime.NumCPU() * 64
	}
	g := &Group{}
	if cfg.limit != unlimited {
		s, _ := NewSemaphore(int64(cfg.limit))
		g.sem = s
	}
	return g
}

// WithLimit sets the maximum number of concurrent goroutines. Must be >= 1.
func WithLimit(n int) option.Option[groupConfig] {
	return option.OptionFunc[groupConfig](func(c *groupConfig) error {
		if n < 1 {
			return ErrInvalidPermits
		}
		c.limit = n
		return nil
	})
}

// WithUnlimited removes the concurrency cap.
func WithUnlimited() option.Option[groupConfig] {
	return option.OptionFunc[groupConfig](func(c *groupConfig) error {
		c.limit = unlimited
		return nil
	})
}

// Go schedules fn in a new goroutine. If at concurrency cap, blocks until a
// slot is available or ctx is cancelled. Panics if called after WaitDone.
func (g *Group) Go(ctx context.Context, fn func() error) {
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		//glacier:nolint=panic-in-library programmer error: Go after WaitDone is documented as panic in func doc.
		panic("concur: Group.Go called after WaitDone")
	}
	g.mu.Unlock()

	if g.sem != nil {
		if err := g.sem.Acquire(ctx, 1); err != nil {
			g.mu.Lock()
			g.errList = append(g.errList, err)
			g.mu.Unlock()
			return
		}
	}
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if g.sem != nil {
			defer func() { _ = g.sem.Release(1) }()
		}
		defer func() {
			if r := recover(); r != nil {
				g.mu.Lock()
				g.errList = append(g.errList, &PanicError{Value: r})
				g.mu.Unlock()
			}
		}()
		if err := fn(); err != nil {
			g.mu.Lock()
			g.errList = append(g.errList, err)
			g.mu.Unlock()
		}
	}()
}

// TryGo schedules fn if a slot is immediately available. Returns true if
// scheduled, false if at capacity. Panics if called after WaitDone.
func (g *Group) TryGo(fn func() error) bool {
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		//glacier:nolint=panic-in-library programmer error: TryGo after WaitDone is documented as panic in func doc.
		panic("concur: Group.TryGo called after WaitDone")
	}
	g.mu.Unlock()

	if g.sem != nil {
		ok, err := g.sem.TryAcquire(1)
		if err != nil || !ok {
			return false
		}
	}
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if g.sem != nil {
			defer func() { _ = g.sem.Release(1) }()
		}
		defer func() {
			if r := recover(); r != nil {
				g.mu.Lock()
				g.errList = append(g.errList, &PanicError{Value: r})
				g.mu.Unlock()
			}
		}()
		if err := fn(); err != nil {
			g.mu.Lock()
			g.errList = append(g.errList, err)
			g.mu.Unlock()
		}
	}()
	return true
}

// WaitDone waits for all scheduled goroutines to complete (or ctx to cancel).
// Returns errs.Join over all collected errors (including PanicErrors).
// After WaitDone returns, the Group is terminal :  further Go calls panic.
func (g *Group) WaitDone(ctx context.Context) error {
	waitDone := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(waitDone)
	}()
	select {
	case <-waitDone:
	case <-ctx.Done():
		// Context cancelled; we still mark closed to prevent further scheduling.
	}
	g.mu.Lock()
	g.closed = true
	collected := g.errList
	g.mu.Unlock()
	return errs.Join(collected...)
}
