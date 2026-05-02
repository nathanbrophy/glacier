// SPDX-License-Identifier: Apache-2.0

package concur

import "sync"

// Once[T] memoizes the (T, error) result of the first successful call to Do.
// A panicking Do does not count as "done" — the next call re-attempts.
type Once[T any] struct {
	mu   sync.Mutex
	done bool
	val  T
	err  error
}

// Do calls fn exactly once (on first successful call). Subsequent calls return
// the memoized (T, error). If fn panics, the panic propagates and the next
// call to Do re-attempts.
func (o *Once[T]) Do(fn func() (T, error)) (T, error) {
	o.mu.Lock()
	if o.done {
		v, e := o.val, o.err
		o.mu.Unlock()
		return v, e
	}
	// panicked tracks whether fn panicked so that defer can skip memoization.
	panicked := true
	defer func() {
		if panicked {
			// fn panicked: do not memoize; release lock so the next call can retry.
			o.mu.Unlock()
		}
	}()
	v, e := fn()
	panicked = false
	o.done = true
	o.val = v
	o.err = e
	o.mu.Unlock()
	return v, e
}
