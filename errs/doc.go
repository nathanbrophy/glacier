// SPDX-License-Identifier: Apache-2.0

// Package errs provides helpers for the Glacier error story. The package
// declares no sentinels of its own — those live in each consuming package.
// What errs provides: chain-preserving Wrap (with optional fluent stack-trace
// capture), drop-nils-and-collapse-singletons Join, tree-walking Chain
// iterator over the full error tree, library-register-enforcing Sentinel
// factory, the multi-target IsAny check, retry-classification markers
// (MarkRetryable / Retryable), and a stable-error-code interface (Coded /
// Code). Every package's error story rolls up through these helpers without
// ever importing each other's sentinels.
package errs
