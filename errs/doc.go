// SPDX-License-Identifier: Apache-2.0

// Package errs provides the helpers underlying Glacier's error story:
// chain-preserving wrapping with optional stack capture, nil-aware joining,
// tree-walking iteration, library-register-validating sentinels, multi-target
// matching, retry markers, and stable error codes.
//
// errs declares no sentinels of its own; per-package sentinels (Err<Cause>)
// live in their own packages and use Sentinel for construction.
package errs
