// SPDX-License-Identifier: Apache-2.0

package log

import "log/slog"

// LevelTrace is below Debug :  for very-verbose tracing that is stripped in
// production builds. Glacier handlers render this as "TRACE". Stdlib handlers
// render it as "DEBUG-4".
const LevelTrace slog.Level = -8

// LevelDebug mirrors slog.LevelDebug (-4). Provided for symmetry.
const LevelDebug slog.Level = slog.LevelDebug

// LevelInfo mirrors slog.LevelInfo (0). Provided for symmetry.
const LevelInfo slog.Level = slog.LevelInfo

// LevelNotice is between Info and Warn :  for important non-warning events
// (config reloaded, connection established). Glacier handlers render this as
// "NOTICE". Stdlib handlers render it as "INFO+2".
const LevelNotice slog.Level = 2

// LevelWarn mirrors slog.LevelWarn (4). Provided for symmetry.
const LevelWarn slog.Level = slog.LevelWarn

// LevelError mirrors slog.LevelError (8). Provided for symmetry.
const LevelError slog.Level = slog.LevelError

// levelLabel returns the canonical Glacier label for l.
func levelLabel(l slog.Level) string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelNotice:
		return "NOTICE"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return l.String()
	}
}
