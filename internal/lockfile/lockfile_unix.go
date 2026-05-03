// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package lockfile

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// acquireLock implements Lock/RLock for Unix using flock(2).
func acquireLock(ctx context.Context, path string, exclusive bool) (Unlock, error) {
	f, err := openLockFile(path)
	if err != nil {
		return nil, err
	}

	op := unix.LOCK_SH
	if exclusive {
		op = unix.LOCK_EX
	}

	// Try non-blocking first; if contended, retry with backoff until ctx done.
	for {
		err := unix.Flock(int(f.Fd()), op|unix.LOCK_NB)
		if err == nil {
			return makeUnlock(f), nil
		}
		// Lock held by another process; wait and retry until ctx fires.
		select {
		case <-ctx.Done():
			_ = f.Close()
			return nil, ErrTimeout
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// makeUnlock returns an Unlock closure that idempotently releases the flock
// and closes the file handle.
func makeUnlock(f interface {
	Fd() uintptr
	Close() error
}) Unlock {
	var once sync.Once
	return func() error {
		var err error
		once.Do(func() {
			_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
			err = f.Close()
		})
		return err
	}
}
