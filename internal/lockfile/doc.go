// SPDX-License-Identifier: Apache-2.0

// Package lockfile provides cooperative cross-process advisory file locking.
// Lock acquires an exclusive lock; RLock acquires a shared lock. Both return
// an Unlock function that releases the lock and closes the file handle.
//
// Implementations are platform-specific:
//   - Unix: golang.org/x/sys/unix.Flock with LOCK_EX or LOCK_SH
//   - Windows: golang.org/x/sys/windows.LockFileEx with LOCKFILE_EXCLUSIVE_LOCK
//
// Usage:
//
//	unlock, err := lockfile.Lock(ctx, path+".lock")
//	if err != nil { return err }
//	defer unlock()
//	// ... critical section over path ...
//
// The lock is advisory: cooperating processes that all use this package will
// serialize correctly. Misbehaving processes can still read or write the
// underlying file directly. The lock file is created with 0o600 permissions
// to mirror the data file's privacy posture.
package lockfile
