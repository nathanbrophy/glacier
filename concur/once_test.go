// SPDX-License-Identifier: Apache-2.0

package concur_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/concur"
)

// T#C25 Do: fn called exactly once across 100 goroutines.
func TestOnce_Do_CalledExactlyOnce(t *testing.T) {
	t.Parallel()
	var o concur.Once[int]
	var calls atomic.Int64
	const n = 100

	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			o.Do(func() (int, error) {
				calls.Add(1)
				return 1, nil
			})
		}()
	}
	wg.Wait()
	assert.Equal(t, calls.Load(), int64(1))
}

// T#C26 Do: subsequent calls return memoized (val, err).
func TestOnce_Do_SubsequentCallsMemoized(t *testing.T) {
	t.Parallel()
	var o concur.Once[string]
	sentinel := errors.New("once-err")

	v1, e1 := o.Do(func() (string, error) { return "hello", sentinel })
	v2, e2 := o.Do(func() (string, error) { return "other", nil })

	assert.Equal(t, v1, "hello")
	assert.ErrorIs(t, e1, sentinel)
	assert.Equal(t, v2, "hello")
	assert.ErrorIs(t, e2, sentinel)
}

// T#C27 Do: panic does not memoize; next call re-attempts.
func TestOnce_Do_PanicDoesNotMemoize(t *testing.T) {
	t.Parallel()
	var o concur.Once[int]

	// First call panics.
	func() {
		defer func() { recover() }()
		o.Do(func() (int, error) {
			panic("oops")
		})
	}()

	// Second call should re-attempt and succeed.
	v, err := o.Do(func() (int, error) { return 42, nil })
	assert.NoError(t, err)
	assert.Equal(t, v, 42)

	// Third call should return memoized value.
	v2, err2 := o.Do(func() (int, error) { return 99, nil })
	assert.NoError(t, err2)
	assert.Equal(t, v2, 42)
}
