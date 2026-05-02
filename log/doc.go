// SPDX-License-Identifier: Apache-2.0

// Package log provides a thin, opinionated layer over Go's log/slog. It
// delivers Glacier's default text and JSON handlers with stable attribute
// ordering and TTY-aware color from the spec-0001 palette, context-based
// attribute attachment (the "ctx carries attrs, never handlers" rule),
// explicit logger injection and retrieval, opt-in source-location attribution,
// two extra levels (Trace and Notice) for richer instrumentation, and a
// Redact helper for explicit secret marking. Every Glacier package emits
// structured logs through this surface so the framework's diagnostic output
// is consistent.
package log
