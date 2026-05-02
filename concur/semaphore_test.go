// SPDX-License-Identifier: Apache-2.0

package concur_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/concur"
)

// T#C15 NewSemaphore(0) returns ErrInvalidPermits.
func TestNewSemaphore_ZeroCapacity(t *testing.T) {
	t.Parallel()
	_, err := concur.NewSemaphore(0)
	assert.ErrorIs(t, err, concur.ErrInvalidPermits)
}

// T#C16 NewSemaphore(-1) returns ErrInvalidPermits.
func TestNewSemaphore_NegativeCapacity(t *testing.T) {
	t.Parallel()
	_, err := concur.NewSemaphore(-1)
	assert.ErrorIs(t, err, concur.ErrInvalidPermits)
}

// T#C17 Acquire + Release: basic round-trip.
func TestSemaphore_AcquireRelease_RoundTrip(t *testing.T) {
	t.Parallel()
	s, err := concur.NewSemaphore(3)
	assert.NoError(t, err)

	ctx := context.Background()
	assert.NoError(t, s.Acquire(ctx, 2))
	assert.NoError(t, s.Release(2))
}

// T#C18 Acquire: blocks when at capacity; succeeds after Release.
func TestSemaphore_Acquire_BlocksUntilRelease(t *testing.T) {
	t.Parallel()
	s, err := concur.NewSemaphore(1)
	assert.NoError(t, err)

	ctx := context.Background()
	assert.NoError(t, s.Acquire(ctx, 1))

	acquired := make(chan error, 1)
	go func() {
		acquired <- s.Acquire(ctx, 1)
	}()

	// Should be blocked.
	select {
	case <-acquired:
		require.True(t, false, "Acquire should be blocked when at capacity")
	case <-time.After(30 * time.Millisecond):
	}

	assert.NoError(t, s.Release(1))

	select {
	case err := <-acquired:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		require.True(t, false, "Acquire did not unblock after Release")
	}
}

// T#C19 TryAcquire: returns false when at capacity.
func TestSemaphore_TryAcquire_ReturnsFalseAtCapacity(t *testing.T) {
	t.Parallel()
	s, err := concur.NewSemaphore(1)
	assert.NoError(t, err)

	ctx := context.Background()
	assert.NoError(t, s.Acquire(ctx, 1))

	ok, err := s.TryAcquire(1)
	assert.NoError(t, err)
	assert.True(t, !ok, "TryAcquire should return false when at capacity")

	assert.NoError(t, s.Release(1))
}

// T#C20 Release: returns ErrInvalidPermits for n > capacity.
func TestSemaphore_Release_ExceedsCapacity(t *testing.T) {
	t.Parallel()
	s, err := concur.NewSemaphore(2)
	assert.NoError(t, err)

	err = s.Release(3)
	assert.ErrorIs(t, err, concur.ErrInvalidPermits)
}

// T#C21 Acquire: cancelled ctx returns ErrCancelled.
func TestSemaphore_Acquire_CancelledCtxReturnsErrCancelled(t *testing.T) {
	t.Parallel()
	s, err := concur.NewSemaphore(1)
	assert.NoError(t, err)

	ctx := context.Background()
	// Fill the semaphore.
	assert.NoError(t, s.Acquire(ctx, 1))

	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err = s.Acquire(cancelCtx, 1)
	assert.ErrorIs(t, err, concur.ErrCancelled)

	// Release so cleanup is clean.
	assert.NoError(t, s.Release(1))

	// Concurrent cancel: fill semaphore and cancel while blocked.
	s2, _ := concur.NewSemaphore(1)
	assert.NoError(t, s2.Acquire(ctx, 1))
	cancelCtx2, cancel2 := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	var acquireErr error
	go func() {
		defer wg.Done()
		acquireErr = s2.Acquire(cancelCtx2, 1)
	}()
	time.Sleep(10 * time.Millisecond)
	cancel2()
	wg.Wait()
	assert.ErrorIs(t, acquireErr, concur.ErrCancelled)
}
