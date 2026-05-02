// SPDX-License-Identifier: Apache-2.0

// Package fixture provides test-resource management — the catch-all for
// everything tests need beyond stdlib testing and Glacier's assert package.
// Distinct from assert (which checks values), fixture manages test inputs,
// outputs, environments, and lifecycle invariants. Three groups of facilities:
// persistent test data (Golden for byte-faithful output regression checks,
// Snapshot[T] for typed structural snapshots, Load/LoadJSON[T] for static
// testdata reading); test environment (Clock for deterministic time, MockFS for
// in-memory filesystem, Capture for stdout/stderr capture); and lifecycle
// invariants (GuardLeaks for end-to-end leak detection of temp dirs,
// goroutines, env vars, and file descriptors, with cleanup-time assertions
// enforced via t.Cleanup).
package fixture
