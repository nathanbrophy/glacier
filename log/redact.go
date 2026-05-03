// SPDX-License-Identifier: Apache-2.0

package log

import "log/slog"

// Redact wraps v in a slog.LogValuer that always formats as "[REDACTED]",
// regardless of which slog handler receives it. Pass the returned value as a
// slog attribute value:
//
//	slog.Info("auth", "token", log.Redact(token))
//
// The redaction is implemented via the stdlib slog.LogValuer interface, so it
// works with every handler :  including third-party handlers that call
// LogValue() on attribute values before formatting.
//
// Preconditions: none. Redact(nil) returns a value that still formats as
// "[REDACTED]".
// Concurrency: goroutine-safe; the returned value is immutable.
func Redact(v any) slog.LogValuer {
	return redactedValue{v: v}
}

// redactedValue implements slog.LogValuer. The wrapped value is never exposed.
type redactedValue struct{ v any }

// LogValue implements slog.LogValuer. Always returns slog.StringValue("[REDACTED]").
func (r redactedValue) LogValue() slog.Value { return slog.StringValue("[REDACTED]") }
