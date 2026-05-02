// SPDX-License-Identifier: Apache-2.0

package concur

import "sync"

// Pool[T] is a type-safe wrapper over sync.Pool. Get returns T; Put accepts T.
// No any conversion at the call site.
type Pool[T any] struct {
	p sync.Pool
}

// NewPool creates a Pool[T] with the given factory function.
// factory is called when Get is called on an empty pool.
// If factory is nil, Get returns the zero value of T.
func NewPool[T any](factory func() T) *Pool[T] {
	pl := &Pool[T]{}
	if factory != nil {
		pl.p.New = func() any { return factory() }
	}
	return pl
}

// Get retrieves a T from the pool, or calls the factory function if the pool
// is empty. Returns the zero value of T if no factory was provided and the
// pool is empty.
func (p *Pool[T]) Get() T {
	v := p.p.Get()
	if v == nil {
		var zero T
		return zero
	}
	return v.(T)
}

// Put returns x to the pool.
func (p *Pool[T]) Put(x T) {
	p.p.Put(x)
}
