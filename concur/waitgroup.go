// SPDX-License-Identifier: Apache-2.0

package concur

import (
	"context"
	"sync"
)

// WaitGroup embeds sync.WaitGroup and adds WaitCtx for ctx-aware waiting.
type WaitGroup struct {
	sync.WaitGroup
}

// WaitCtx waits for the WaitGroup counter to reach zero or for ctx to be
// cancelled. Returns ErrCancelled if ctx is cancelled before the counter reaches zero.
func (wg *WaitGroup) WaitCtx(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		wg.WaitGroup.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ErrCancelled
	}
}
