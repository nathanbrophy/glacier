// SPDX-License-Identifier: Apache-2.0

// Package cache is Glacier's generic key-value cache.
//
// One interface, three implementations:
//
//   - [New] returns an in-memory [Cache] backed by a map+RWMutex.
//   - [NewDisk] returns a disk-backed [Cache] that persists each key as a
//     separate JSON file with advisory file locking for cross-process safety.
//   - [NewLayered] composes a primary and a backing [Cache] with write-through
//     semantics so a process can survive a restart without losing entries.
//
// All implementations are generic over a value type V, support per-key TTL
// with a hybrid (per-instance default + per-Set override) policy, and
// collapse concurrent misses on the same key onto one loader call via
// [Cache.GetOrLoad] singleflight.
//
// The [Cache] interface carries the +glacier:mock marker so dependent
// packages can swap in a mock for tests with mock.Of[Cache[V]]().
//
// Zero new direct dependencies: stdlib + internal/safefile + internal/safejson
// + internal/lockfile + obs (optional, no-op when no OTEL endpoint configured).
//
// Hot-path Get on the in-memory implementation allocates zero bytes when the
// key is present and unexpired (verified by BenchmarkMemHit).
package cache
