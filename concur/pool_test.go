// SPDX-License-Identifier: Apache-2.0

package concur_test

import (
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/concur"
)

// T#C22 Get on empty pool with factory: returns factory result.
func TestPool_Get_WithFactory(t *testing.T) {
	t.Parallel()
	p := concur.NewPool(func() int { return 42 })
	got := p.Get()
	assert.Equal(t, got, 42)
}

// T#C23 Get on empty pool without factory: returns zero value.
func TestPool_Get_WithoutFactory(t *testing.T) {
	t.Parallel()
	p := concur.NewPool[string](nil)
	got := p.Get()
	assert.Equal(t, got, "")
}

// T#C24 Put/Get round-trip: Put returns value; Get retrieves it (eventually — GC may collect).
func TestPool_PutGet_RoundTrip(t *testing.T) {
	t.Parallel()
	p := concur.NewPool[int](nil)
	p.Put(99)
	// The pool may or may not return the value (GC can collect), so we accept either 99 or 0.
	got := p.Get()
	assert.True(t, got == 99 || got == 0, "expected 99 or zero value from pool")
}
