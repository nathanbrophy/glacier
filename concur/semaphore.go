// SPDX-License-Identifier: Apache-2.0

package concur

import (
	"context"
	"sync"
	"sync/atomic"
)

// Semaphore is a counted semaphore with atomic-counter fast path.
type Semaphore struct {
	capacity int64
	count    atomic.Int64
	mu       sync.Mutex
	cond     *sync.Cond
}

// NewSemaphore creates a Semaphore with the given capacity.
// Returns ErrInvalidPermits if capacity <= 0.
func NewSemaphore(capacity int64) (*Semaphore, error) {
	if capacity <= 0 {
		return nil, ErrInvalidPermits
	}
	s := &Semaphore{capacity: capacity}
	s.cond = sync.NewCond(&s.mu)
	return s, nil
}

// Acquire acquires n permits, blocking until available or ctx is cancelled.
// Returns ErrInvalidPermits if n <= 0 or n > capacity.
// Returns ErrCancelled if ctx is cancelled.
func (s *Semaphore) Acquire(ctx context.Context, n int64) error {
	if n <= 0 || n > s.capacity {
		return ErrInvalidPermits
	}
	// Fast path: atomic CAS.
	for {
		cur := s.count.Load()
		if cur+n <= s.capacity {
			if s.count.CompareAndSwap(cur, cur+n) {
				return nil
			}
			continue // CAS raced; retry fast path
		}
		break // need slow path
	}
	// Slow path: wait on cond.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Cancel watcher goroutine: signals cond on ctx cancel.
	done := make(chan struct{})
	go func() {
		defer close(done)
		select {
		case <-ctx.Done():
			s.cond.Broadcast()
		case <-done:
		}
	}()

	s.mu.Lock()
	for {
		if ctx.Err() != nil {
			s.mu.Unlock()
			<-done
			return ErrCancelled
		}
		cur := s.count.Load()
		if cur+n <= s.capacity {
			if s.count.CompareAndSwap(cur, cur+n) {
				s.mu.Unlock()
				cancel() // stop watcher
				<-done
				return nil
			}
		}
		s.cond.Wait()
	}
}

// TryAcquire attempts to acquire n permits without blocking.
// Returns true on success, false if insufficient permits are available.
// Returns ErrInvalidPermits if n <= 0 or n > capacity.
func (s *Semaphore) TryAcquire(n int64) (bool, error) {
	if n <= 0 || n > s.capacity {
		return false, ErrInvalidPermits
	}
	for {
		cur := s.count.Load()
		if cur+n > s.capacity {
			return false, nil
		}
		if s.count.CompareAndSwap(cur, cur+n) {
			return true, nil
		}
	}
}

// Release releases n permits.
// Returns ErrInvalidPermits if n <= 0 or n > capacity.
func (s *Semaphore) Release(n int64) error {
	if n <= 0 || n > s.capacity {
		return ErrInvalidPermits
	}
	s.count.Add(-n)
	s.cond.Broadcast()
	return nil
}
