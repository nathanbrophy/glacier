// SPDX-License-Identifier: Apache-2.0

package concur_test

import (
	"context"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/concur"
)

// T#C13 Go after WaitDone panics.
func TestGroup_Go_AfterWaitDonePanics(t *testing.T) {
	t.Parallel()
	g := concur.NewGroup()
	ctx := context.Background()
	_ = g.WaitDone(ctx)

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		g.Go(ctx, func() error { return nil })
	}()
	assert.True(t, panicked, "Go after WaitDone should panic")
}

// T#C14 TryGo after WaitDone panics.
func TestGroup_TryGo_AfterWaitDonePanics(t *testing.T) {
	t.Parallel()
	g := concur.NewGroup()
	ctx := context.Background()
	_ = g.WaitDone(ctx)

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		g.TryGo(func() error { return nil })
	}()
	assert.True(t, panicked, "TryGo after WaitDone should panic")
}
