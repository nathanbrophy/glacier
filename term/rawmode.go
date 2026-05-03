// SPDX-License-Identifier: Apache-2.0

package term

import (
	"context"
	"os"
	"sync"

	xterm "golang.org/x/term"
)

// AcquireRaw puts stdin into raw mode and returns a restore function.
// The restore function must be called (typically via defer) to return stdin to
// cooked mode. The restore is panic-safe: even if the caller panics, calling
// the returned restore func will still restore the terminal.
//
// AcquireRaw returns ErrNotInteractive when stdin is not a TTY.
//
// Concurrency: not goroutine-safe; caller owns the terminal while raw mode is active.
func AcquireRaw(_ context.Context) (restore func(), err error) {
	fd := int(os.Stdin.Fd())
	if !xterm.IsTerminal(fd) {
		return func() {}, ErrNotInteractive
	}
	oldState, err := xterm.MakeRaw(fd)
	if err != nil {
		return func() {}, err
	}
	var once sync.Once
	restore = func() {
		once.Do(func() {
			_ = xterm.Restore(fd, oldState)
		})
	}
	return restore, nil
}
