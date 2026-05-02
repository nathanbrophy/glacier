// SPDX-License-Identifier: Apache-2.0

// Package assert provides Glacier's test assertions and runtime Must
// helpers. Test assertions report failures via t.Errorf and return bool,
// allowing callers to stack assertions and see all failures in one run.
// For halt-on-failure semantics use the sister package
// github.com/nathanbrophy/glacier/assert/require.
//
// Equal goes beyond reflect.DeepEqual: pointer-aware, map-order-insensitive,
// optionally slice-order-insensitive via IgnoreOrder, case-insensitive via
// IgnoreCase, float-tolerant via WithDelta, and field-selective via
// IgnoreFields. Custom types that implement Equal(any) bool participate via
// that method. A primitive fast path bypasses reflection entirely when T is
// comparable and got == want by ==.
//
// Arguments that may contain secrets (tokens, passwords, private keys)
// should be wrapped with log.RedactValue before passing to assertions —
// the diff renderer calls LogValue() on slog.LogValuer values, rendering
// them as [REDACTED] instead of the raw secret.
//
// Runtime Must helpers (Must, Must2, Mustf) panic on violation and are
// intended only for initialization-time invariants in package init,
// main() setup, and test harness setup. Never use Must in library hot
// paths — library code must not panic (see CLAUDE.md directive 4).
package assert
