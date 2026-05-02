// SPDX-License-Identifier: Apache-2.0

// Package log provides Glacier's slog conventions: default text and JSON
// handlers with stable attribute ordering and TTY-aware color, two extra
// levels (Trace and Notice), context-based attribute attachment, explicit
// logger injection and retrieval, and a Redact helper for secret marking.
//
// The "ctx carries attrs, never handlers" rule is enforced here: With
// attaches attributes that the Glacier handlers automatically append to
// any record logged with that context. The handler itself is configured
// at construction and injected when scoping to a request or goroutine is
// needed.
package log
