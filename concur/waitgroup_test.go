// SPDX-License-Identifier: Apache-2.0

package concur_test

import (
	"context"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/concur"
)

// T#C28 WaitCtx: returns nil when all Done calls complete.
func TestWaitGroup_WaitCtx_ReturnsNilOnCompletion(t *testing.T) {
	t.Parallel()
	var wg concur.WaitGroup
	const n = 10
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond)
		}()
	}
	ctx := context.Background()
	err := wg.WaitCtx(ctx)
	assert.NoError(t, err)
}

// T#C29 WaitCtx: returns ErrCancelled when ctx cancelled before done.
func TestWaitGroup_WaitCtx_ReturnsCancelledWhenCtxCancelled(t *testing.T) {
	t.Parallel()
	var wg concur.WaitGroup
	wg.Add(1) // never Done'd — keeps wg from completing

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := wg.WaitCtx(ctx)
	assert.ErrorIs(t, err, concur.ErrCancelled)

	// Clean up: mark done so GC can collect the goroutine.
	wg.Done()
}
