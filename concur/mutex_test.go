// SPDX-License-Identifier: Apache-2.0

package concur_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/concur"
)

// T#C1 Lock/Unlock: mutual exclusion — two goroutines race for lock; only one at a time proceeds.
func TestMutex_LockUnlock_MutualExclusion(t *testing.T) {
	t.Parallel()
	var mu concur.Mutex
	var counter int
	const n = 100
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
	assert.Equal(t, counter, n)
}

// T#C2 LockCtx: cancelled ctx returns ErrCancelled.
func TestMutex_LockCtx_CancelledReturnsErrCancelled(t *testing.T) {
	t.Parallel()
	var mu concur.Mutex
	mu.Lock() // hold the lock so LockCtx must wait
	defer mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	err := mu.LockCtx(ctx)
	assert.ErrorIs(t, err, concur.ErrCancelled)
}

// T#C3 LockCtx: non-cancelled ctx eventually acquires.
func TestMutex_LockCtx_EventuallyAcquires(t *testing.T) {
	t.Parallel()
	var mu concur.Mutex
	mu.Lock()

	ctx := context.Background()
	acquired := make(chan error, 1)
	go func() {
		acquired <- mu.LockCtx(ctx)
	}()

	// Release the lock after a brief pause.
	time.Sleep(10 * time.Millisecond)
	mu.Unlock()

	err := <-acquired
	assert.NoError(t, err)
	// Ensure we can unlock the lock that LockCtx acquired.
	mu.Unlock()
}

// T#C4 RWMutex: multiple readers can hold simultaneously.
func TestRWMutex_MultipleReadersSimultaneous(t *testing.T) {
	t.Parallel()
	var mu concur.RWMutex
	const readers = 10
	ready := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(readers)
	held := make(chan struct{}, readers)

	for range readers {
		go func() {
			defer wg.Done()
			<-ready
			mu.RLock()
			held <- struct{}{}
			time.Sleep(20 * time.Millisecond)
			mu.RUnlock()
		}()
	}

	close(ready)
	// All readers should be able to acquire the read lock concurrently.
	for range readers {
		select {
		case <-held:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for reader to acquire RLock")
		}
	}
	wg.Wait()
}

// T#C5 RWMutex: writer blocks while reader holds.
func TestRWMutex_WriterBlocksWhileReaderHolds(t *testing.T) {
	t.Parallel()
	var mu concur.RWMutex
	mu.RLock()

	writerAcquired := make(chan struct{})
	go func() {
		mu.Lock()
		close(writerAcquired)
		mu.Unlock()
	}()

	// Writer should not have acquired yet.
	select {
	case <-writerAcquired:
		t.Fatal("writer acquired lock while reader holds it")
	case <-time.After(30 * time.Millisecond):
	}

	mu.RUnlock()

	// Now writer should acquire.
	select {
	case <-writerAcquired:
	case <-time.After(2 * time.Second):
		t.Fatal("writer never acquired lock after reader released")
	}
}

// T#C6 RLockCtx: cancelled ctx returns ErrCancelled.
func TestRWMutex_RLockCtx_CancelledReturnsErrCancelled(t *testing.T) {
	t.Parallel()
	var mu concur.RWMutex
	// Acquire write lock so RLockCtx must wait.
	mu.Lock()
	defer mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	err := mu.RLockCtx(ctx)
	assert.ErrorIs(t, err, concur.ErrCancelled)
}
