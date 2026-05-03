// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"sync"
)

// singleflight collapses concurrent identical loader calls onto one execution.
// Generic over the value type V. The zero value is ready to use.
type singleflight[V any] struct {
	mu      sync.Mutex
	pending map[string]*sfCall[V]
}

// sfCall is one in-flight call shared by multiple waiters.
type sfCall[V any] struct {
	done  chan struct{}
	value V
	err   error
}

// do calls fn for the given key, ensuring only one fn invocation per key is
// in flight at a time. Other concurrent callers for the same key wait for the
// first call to finish and observe the same result.
func (s *singleflight[V]) do(ctx context.Context, key string, fn func(context.Context) (V, error)) (V, error) {
	s.mu.Lock()
	if s.pending == nil {
		s.pending = make(map[string]*sfCall[V])
	}
	if call, ok := s.pending[key]; ok {
		s.mu.Unlock()
		select {
		case <-call.done:
			return call.value, call.err
		case <-ctx.Done():
			var zero V
			return zero, ctx.Err()
		}
	}
	call := &sfCall[V]{done: make(chan struct{})}
	s.pending[key] = call
	s.mu.Unlock()

	call.value, call.err = fn(ctx)
	close(call.done)

	s.mu.Lock()
	delete(s.pending, key)
	s.mu.Unlock()

	return call.value, call.err
}
