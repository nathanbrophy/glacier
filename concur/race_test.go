//go:build race

// SPDX-License-Identifier: Apache-2.0

package concur_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/concur"
)

// T#C30 Multiple goroutines: concurrent Lock/Unlock no race.
func TestRace_Mutex_ConcurrentLockUnlock(t *testing.T) {
	t.Parallel()
	var mu concur.Mutex
	var counter int
	const n = 200
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
}

// T#C31 Multiple goroutines: concurrent Acquire/Release no race.
func TestRace_Semaphore_ConcurrentAcquireRelease(t *testing.T) {
	t.Parallel()
	s, _ := concur.NewSemaphore(10)
	ctx := context.Background()
	const n = 200
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			_ = s.Acquire(ctx, 1)
			_ = s.Release(1)
		}()
	}
	wg.Wait()
}

// T#C32 Once.Do: concurrent calls — only one fn executes.
func TestRace_Once_ConcurrentDo(t *testing.T) {
	t.Parallel()
	var o concur.Once[int]
	var calls atomic.Int64
	const n = 200
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
	assert.True(t, calls.Load() == int64(1), fmt.Sprintf("Once.Do fn called %d times; want 1", calls.Load()))
}
