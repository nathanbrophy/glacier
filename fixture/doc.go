// SPDX-License-Identifier: Apache-2.0

// Package fixture provides test-resource management — the catch-all for
// everything tests need beyond stdlib testing and Glacier's assert package.
// Distinct from assert (which checks values), fixture manages test inputs,
// outputs, environments, and lifecycle invariants. Three groups of facilities:
//
//   - Persistent test data: Golden for byte-faithful output regression checks,
//     Snapshot[T] for typed structural snapshots, Load/LoadJSON[T] for static
//     testdata reading.
//
//   - Test environment: Clock for deterministic time injection, NewFS for
//     in-memory read-only filesystems, Capture for stdout/stderr capture.
//
//   - Lifecycle invariants: GuardLeaks for end-of-test leak detection of
//     goroutines, temp dirs, env vars, and file descriptors; cleanup
//     assertions enforced via t.Cleanup.
//
// All helpers register their cleanup with t.Cleanup, keeping test bodies
// linear and readable. The env var GLACIER_GOLDEN_UPDATE=1 activates
// auto-write mode for golden and snapshot files.
package fixture
