// SPDX-License-Identifier: Apache-2.0

package concur_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/concur"
)

// T#C7 Go: all goroutines complete.
func TestGroup_Go_AllComplete(t *testing.T) {
	t.Parallel()
	g := concur.NewGroup()
	ctx := context.Background()
	const n = 50
	results := make(chan struct{}, n)
	for range n {
		g.Go(ctx, func() error {
			results <- struct{}{}
			return nil
		})
	}
	err := g.WaitDone(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(results), n)
}

// T#C8 Go: all errors collected (not first-wins).
func TestGroup_Go_AllErrorsCollected(t *testing.T) {
	t.Parallel()
	g := concur.NewGroup()
	ctx := context.Background()
	const n = 5
	errA := errors.New("err-a")
	errB := errors.New("err-b")
	for i := range n {
		i := i
		g.Go(ctx, func() error {
			if i%2 == 0 {
				return errA
			}
			return errB
		})
	}
	err := g.WaitDone(ctx)
	assert.True(t, err != nil, "expected non-nil error")
	// All errors should be present in the joined error.
	assert.True(t, errors.Is(err, errA), "expected errA in joined error")
	assert.True(t, errors.Is(err, errB), "expected errB in joined error")
}

// T#C9 Go: panic recovered as PanicError.
func TestGroup_Go_PanicRecoveredAsPanicError(t *testing.T) {
	t.Parallel()
	g := concur.NewGroup()
	ctx := context.Background()
	g.Go(ctx, func() error {
		panic("boom")
	})
	err := g.WaitDone(ctx)
	assert.True(t, err != nil, "expected non-nil error from panic")
	var pe *concur.PanicError
	assert.True(t, errors.As(err, &pe), "expected PanicError in chain")
	assert.Equal(t, pe.Value, any("boom"))
}

// T#C10 TryGo: returns false when at limit.
func TestGroup_TryGo_ReturnsFalseAtLimit(t *testing.T) {
	t.Parallel()
	// Create group with limit=1.
	g := concur.NewGroup(concur.WithLimit(1))
	block := make(chan struct{})
	// Schedule one fn that blocks so the slot is occupied.
	ok := g.TryGo(func() error {
		<-block
		return nil
	})
	assert.True(t, ok, "first TryGo should succeed")

	// Now the second should fail.
	ok2 := g.TryGo(func() error { return nil })
	assert.True(t, !ok2, "second TryGo should return false when at limit")

	close(block)
	ctx := context.Background()
	assert.NoError(t, g.WaitDone(ctx))
}

// T#C11 WaitDone: returns nil when all fns succeed.
func TestGroup_WaitDone_ReturnsNilOnSuccess(t *testing.T) {
	t.Parallel()
	g := concur.NewGroup()
	ctx := context.Background()
	for range 10 {
		g.Go(ctx, func() error { return nil })
	}
	err := g.WaitDone(ctx)
	assert.NoError(t, err)
}

// T#C12 WaitDone: returns joined error when fns fail.
func TestGroup_WaitDone_ReturnsJoinedErrors(t *testing.T) {
	t.Parallel()
	g := concur.NewGroup()
	ctx := context.Background()
	sentinel := errors.New("sentinel")
	g.Go(ctx, func() error { return sentinel })
	err := g.WaitDone(ctx)
	assert.ErrorIs(t, err, sentinel)
}
