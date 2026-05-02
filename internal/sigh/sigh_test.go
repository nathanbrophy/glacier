// SPDX-License-Identifier: Apache-2.0

package sigh_test

import (
	"context"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/internal/sigh"
)

func TestNotifyCancel(t *testing.T) {
	ctx, stop := sigh.Notify(context.Background())
	defer stop()

	// Cancelling the stop function should cancel the context.
	stop()
	select {
	case <-ctx.Done():
		// expected
	case <-time.After(time.Second):
		t.Fatal("context not cancelled after stop()")
	}
}

func TestNotifyParentCancelPropagate(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())
	ctx, stop := sigh.Notify(parent)
	defer stop()

	parentCancel()
	select {
	case <-ctx.Done():
		// expected: parent cancellation propagates
	case <-time.After(time.Second):
		t.Fatal("context not cancelled after parent cancel")
	}
}

func TestNotifyAlreadyCancelledParent(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())
	parentCancel() // cancel before Notify

	ctx, stop := sigh.Notify(parent)
	defer stop()

	select {
	case <-ctx.Done():
		// expected immediately
	case <-time.After(time.Second):
		t.Fatal("context not cancelled when parent already cancelled")
	}
}
