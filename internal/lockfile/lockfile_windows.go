// SPDX-License-Identifier: Apache-2.0

//go:build windows

package lockfile

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

// acquireLock implements Lock/RLock for Windows using LockFileEx.
func acquireLock(ctx context.Context, path string, exclusive bool) (Unlock, error) {
	f, err := openLockFile(path)
	if err != nil {
		return nil, err
	}

	flags := uint32(windows.LOCKFILE_FAIL_IMMEDIATELY)
	if exclusive {
		flags |= windows.LOCKFILE_EXCLUSIVE_LOCK
	}

	// Lock entire file (offset 0, max length).
	const lockLen = ^uint32(0)
	for {
		var ol windows.Overlapped
		err := windows.LockFileEx(windows.Handle(f.Fd()), flags, 0, lockLen, lockLen, &ol)
		if err == nil {
			return makeUnlock(f, lockLen), nil
		}
		// Contended; retry until ctx done.
		select {
		case <-ctx.Done():
			_ = f.Close()
			return nil, ErrTimeout
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// makeUnlock returns an idempotent Unlock that releases the LockFileEx region
// and closes the file handle.
func makeUnlock(f interface {
	Fd() uintptr
	Close() error
}, lockLen uint32) Unlock {
	var once sync.Once
	return func() error {
		var err error
		once.Do(func() {
			var ol windows.Overlapped
			_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, lockLen, lockLen, &ol)
			err = f.Close()
		})
		return err
	}
}
