// SPDX-License-Identifier: Apache-2.0

package lockfile

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// ErrTimeout is returned when ctx is cancelled before a lock can be acquired.
var ErrTimeout = errors.New("lockfile: lock acquisition cancelled")

// Unlock releases the lock and closes the underlying file handle.
// Safe to call multiple times; subsequent calls are no-ops.
type Unlock func() error

// Lock acquires an exclusive advisory lock on path. The path is opened (or
// created with 0o600 permissions if absent) and held by the returned closer
// until Unlock is called. The lock is advisory: cooperating callers see
// mutual exclusion; non-cooperating callers can still bypass it.
//
// If ctx is cancelled before the lock is acquired, Lock returns ErrTimeout.
func Lock(ctx context.Context, path string) (Unlock, error) {
	return acquireLock(ctx, path, true)
}

// RLock acquires a shared advisory lock on path. Multiple shared locks may
// coexist; an exclusive lock blocks them and is blocked by them. See Lock
// for path semantics.
func RLock(ctx context.Context, path string) (Unlock, error) {
	return acquireLock(ctx, path, false)
}

// openLockFile opens or creates path for read+write at 0o600.
func openLockFile(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("lockfile: open %s: %w", path, err)
	}
	return f, nil
}
