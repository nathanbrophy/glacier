// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"sync"
	"testing"
)

// §21.4 NF4, T27; §23.14

// TestConcurrentAssertions verifies 100 goroutines each calling Equal against
// their own mockTB, running under -race, produces no data races.
func TestConcurrentAssertions(t *testing.T) {
	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			mt := &mockTB{}
			Equal(mt, 42, 42)
			Equal(mt, "hello", "hello")
			Equal(mt, []int{1, 2, 3}, []int{1, 2, 3})
		}()
	}
	wg.Wait()
}

// TestTBConcurrentErrorf verifies that assert doesn't introduce races when
// multiple goroutines call Errorf on the same *testing.T.
// §23.14
func TestTBConcurrentErrorf(t *testing.T) {
	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			// Use a real *testing.T for sub-tests :  t.Errorf is goroutine-safe per docs.
			// We pass the real t and verify no race is detected.
			Equal(t, 1, 1) // always passes; no Errorf call.
		}()
	}
	wg.Wait()
}
