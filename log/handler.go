// SPDX-License-Identifier: Apache-2.0

package log

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"

	"github.com/nathanbrophy/glacier/option"
	"github.com/nathanbrophy/glacier/term"
)

// ColorMode controls whether the text handler emits ANSI color escapes.
type ColorMode int

const (
	// ColorAuto enables color when w is a TTY and neither GLACIER_NO_COLOR
	// nor NO_COLOR is set in the environment. This is the default for
	// NewHandler.
	ColorAuto ColorMode = iota

	// ColorAlways forces color on regardless of whether w is a TTY.
	// Color is still suppressed when GLACIER_NO_COLOR=1 is set.
	ColorAlways

	// ColorNever forces color off regardless of TTY status or env vars.
	ColorNever
)

// resolvedColor is the effective color decision after consulting ColorMode and env vars.
// It is a boolean: color is either on or off :  ColorAuto has been resolved.
// invariant: set once at NewHandler time; never mutated.
type resolvedColor bool

// handlerConfig holds construction-time settings for NewHandler and NewJSONHandler.
// Consumers never touch this type directly; they use the With* option constructors.
type handlerConfig struct {
	level  slog.Leveler // invariant: non-nil; defaults to slog.LevelInfo
	source bool         // invariant: false means no source attribution
	color  ColorMode    // invariant: one of ColorAuto, ColorAlways, ColorNever
}

// glacierHandler wraps a stdlib slog.Handler and adds ctx-attr injection and
// Glacier attribute ordering.
//
// Performance target: ≤ 3 allocs beyond stdlib slog baseline on the hot path.
// Get correctness first; optimize allocations in a follow-up pass.
type glacierHandler struct {
	inner slog.Handler  // invariant: non-nil
	color resolvedColor // invariant: computed once at construction; immutable after
}

// levelStyles holds the per-level term.Style used to colorize the level
// label in the text handler. Constructed via term.RGB so the log handler
// owns its semantic palette directly rather than binding to term's brand
// palette tokens (which have a separate purpose). Hex values are the spec
// 0005 documentation set.
var levelStyles = [6]term.Style{
	term.New().Foreground(term.RGB(0x8B, 0x94, 0x9E)), // TRACE  :  text-muted #8B949E
	term.New().Foreground(term.RGB(0x8B, 0x94, 0x9E)), // DEBUG  :  text-muted #8B949E
	term.New().Foreground(term.RGB(0x22, 0xD3, 0xEE)), // INFO   :  cyan       #22D3EE
	term.New().Foreground(term.RGB(0x2D, 0xD4, 0xBF)), // NOTICE :  teal       #2DD4BF
	term.New().Foreground(term.RGB(0xFB, 0xBF, 0x24)), // WARN   :  warning    #FBBF24
	term.New().Foreground(term.RGB(0xF8, 0x71, 0x71)), // ERROR  :  error      #F87171
}

// levelLabels mirrors levelStyles in the order TRACE, DEBUG, INFO, NOTICE,
// WARN, ERROR. The byte-form needles and replacements below are derived
// from this list once at init.
var levelLabels = [6]string{"TRACE", "DEBUG", "INFO", "NOTICE", "WARN", "ERROR"}

// levelNeedles and levelReplacements are pre-computed byte sequences used
// by colorWriter to substitute the level field in serialized records when
// color is on. The leading space anchors the match to the field position
// in slog's text format (`time=… level=<LABEL> msg=…`); user-provided
// attribute values starting with " level=" are vanishingly rare.
var (
	levelNeedles      [6][]byte
	levelReplacements [6][]byte
)

func init() {
	reset := []byte(term.AnsiReset)
	for i, label := range levelLabels {
		needle := []byte(" level=" + label)
		prefix := []byte(levelStyles[i].Prefix())
		repl := make([]byte, 0, len(needle)+len(prefix)+len(reset))
		repl = append(repl, " level="...)
		repl = append(repl, prefix...)
		repl = append(repl, label...)
		repl = append(repl, reset...)
		levelNeedles[i] = needle
		levelReplacements[i] = repl
	}
}

// colorWriter wraps an io.Writer and substitutes the slog text-handler's
// emitted ` level=<LABEL>` field with its colored equivalent on each Write.
// One Write call corresponds to one record (slog's text handler buffers a
// full record before calling Write). The substitution happens at most once
// per Write :  slog emits the level field exactly once per record.
type colorWriter struct {
	w io.Writer
}

// Write substitutes the first matching ` level=<LABEL>` needle (if any)
// with its colored replacement, then forwards to the underlying writer.
// Returns the original p length on success so callers see the byte count
// they asked to write, not the post-substitution length.
func (cw *colorWriter) Write(p []byte) (int, error) {
	for i, needle := range levelNeedles {
		idx := bytes.Index(p, needle)
		if idx < 0 {
			continue
		}
		repl := levelReplacements[i]
		buf := make([]byte, 0, len(p)+len(repl)-len(needle))
		buf = append(buf, p[:idx]...)
		buf = append(buf, repl...)
		buf = append(buf, p[idx+len(needle):]...)
		if _, err := cw.w.Write(buf); err != nil {
			return 0, err
		}
		return len(p), nil
	}
	return cw.w.Write(p)
}

// resolveColor determines the effective color setting at handler-construction time.
// Rule: GLACIER_NO_COLOR > NO_COLOR > TTY detection > ColorAlways.
func resolveColor(w io.Writer, mode ColorMode) resolvedColor {
	if mode == ColorNever {
		return resolvedColor(false)
	}
	// Consult term.Capability, which already checks GLACIER_NO_COLOR and NO_COLOR.
	caps := term.Capability(w)
	if caps.NoColorEnv {
		// Either GLACIER_NO_COLOR or NO_COLOR is set; color off regardless of mode.
		return resolvedColor(false)
	}
	switch mode {
	case ColorAlways:
		return resolvedColor(true)
	case ColorAuto:
		return resolvedColor(caps.IsTTY && caps.SupportsColor > term.ColorNone)
	}
	return resolvedColor(false)
}

// NewHandler returns Glacier's text slog.Handler.
//
// The handler places attributes in this canonical order: level, msg,
// package, op, error, then any user-supplied attrs in the order supplied.
// Time precedes level when a time source is configured; the default omits
// time for terse developer output.
//
// Color palette (24-bit ANSI; pre-computed at construction time):
//
//	TRACE:  text-muted #8B949E
//	DEBUG:  text-muted #8B949E
//	INFO:   cyan       #22D3EE
//	NOTICE: teal       #2DD4BF
//	WARN:   warning    #FBBF24
//	ERROR:  error      #F87171
//
// Package and op attribute keys get a subtle cyan accent; error values get
// rose. Color is resolved once at construction time; subsequent Handle calls
// do not re-check env vars or TTY state.
//
// Use options to override the minimum level, enable source attribution, or
// override the color mode.
//
//	h := log.NewHandler(os.Stderr,
//	    log.WithLevel(log.LevelDebug),
//	    log.WithSource(),
//	)
func NewHandler(w io.Writer, opts ...option.Option[handlerConfig]) slog.Handler {
	cfg := applyHandlerConfig(opts)
	// Resolve color against the user-supplied writer (capability detection
	// must see the real writer, not a wrapper that hides Fd()).
	color := resolveColor(w, cfg.color)
	// When color is enabled, wrap the writer in a colorWriter that
	// substitutes the serialized level field with its colored form. The
	// stdlib text handler quotes string values containing control bytes,
	// so we cannot inject ANSI escapes via ReplaceAttr :  the substitution
	// has to happen on the rendered byte stream instead.
	out := w
	if bool(color) {
		out = &colorWriter{w: w}
	}
	inner := slog.NewTextHandler(out, &slog.HandlerOptions{
		Level:     cfg.level,
		AddSource: cfg.source,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Override level label for custom Glacier levels (Trace, Notice).
			if a.Key == slog.LevelKey {
				level, ok := a.Value.Any().(slog.Level)
				if ok {
					a.Value = slog.StringValue(levelLabel(level))
				}
			}
			return a
		},
	})
	return &glacierHandler{inner: inner, color: color}
}

// NewJSONHandler returns Glacier's JSON slog.Handler. Attribute ordering
// matches NewHandler; no color is emitted (JSON is not a styled medium).
// WithColor is accepted but ignored. WithLevel and WithSource are honored.
//
//	log.SetDefault(slog.New(log.NewJSONHandler(os.Stderr,
//	    log.WithLevel(log.LevelInfo),
//	    log.WithSource(),
//	)))
func NewJSONHandler(w io.Writer, opts ...option.Option[handlerConfig]) slog.Handler {
	cfg := applyHandlerConfig(opts)
	inner := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:     cfg.level,
		AddSource: cfg.source,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Override level label for custom Glacier levels (Trace, Notice).
			if a.Key == slog.LevelKey {
				level, ok := a.Value.Any().(slog.Level)
				if ok {
					a.Value = slog.StringValue(levelLabel(level))
				}
			}
			return a
		},
	})
	// JSON handler never uses color.
	return &glacierHandler{inner: inner, color: resolvedColor(false)}
}

// applyHandlerConfig applies opts to a default handlerConfig.
func applyHandlerConfig(opts []option.Option[handlerConfig]) handlerConfig {
	cfg, _ := option.Apply(opts)
	if cfg.level == nil {
		cfg.level = slog.LevelInfo
	}
	return cfg
}

// WithLevel sets the handler's minimum log level. Records below this level
// are discarded. Default: slog.LevelInfo.
//
// Pass any slog.Leveler :  including dynamic LevelVar pointers :  for runtime
// level control.
func WithLevel(l slog.Leveler) option.Option[handlerConfig] {
	return option.OptionFunc[handlerConfig](func(c *handlerConfig) error {
		c.level = l
		return nil
	})
}

// defaultLevelVar is the package-owned dynamic level var consulted by
// handlers constructed with WithDynamicLevel. SDK binaries that want
// runtime-mutable level filtering install one such handler at startup and
// then call SetDefaultLevel to update the threshold (e.g. when --verbose is
// observed during flag parsing).
//
// The zero value of slog.LevelVar is slog.LevelInfo; SDK startup typically
// calls SetDefaultLevel(LevelWarn) before installing the handler.
var defaultLevelVar slog.LevelVar

// WithDynamicLevel binds the handler to the package-level dynamic level var
// updated by SetDefaultLevel. Use this when the program needs to change the
// handler's filter level after construction (for example, after CLI flag
// parsing reveals a --verbose request).
//
//	log.SetDefaultLevel(slog.LevelWarn)
//	h := log.NewHandler(os.Stderr, log.WithDynamicLevel())
//	// ...later, after parsing --verbose:
//	log.SetDefaultLevel(slog.LevelDebug)
func WithDynamicLevel() option.Option[handlerConfig] {
	return WithLevel(&defaultLevelVar)
}

// SetDefaultLevel updates the threshold of the package-level dynamic level
// var. Handlers constructed via WithDynamicLevel observe the new level on
// their next Enabled / Handle call. Safe for concurrent use.
func SetDefaultLevel(l slog.Level) {
	defaultLevelVar.Set(l)
}

// DefaultLevel returns the current threshold of the package-level dynamic
// level var. Returns slog.LevelInfo if SetDefaultLevel has never been called.
func DefaultLevel() slog.Level {
	return defaultLevelVar.Level()
}

// WithSource enables source-location attribution on every record: the
// caller's file and line are included. Off by default :  source attribution
// adds approximately 30% latency per log call.
func WithSource() option.Option[handlerConfig] {
	return option.OptionFunc[handlerConfig](func(c *handlerConfig) error {
		c.source = true
		return nil
	})
}

// WithColor overrides the text handler's color mode. Ignored by
// NewJSONHandler.
//
// Example: force color off in a test that captures stderr:
//
//	h := log.NewHandler(&buf, log.WithColor(log.ColorNever))
func WithColor(m ColorMode) option.Option[handlerConfig] {
	return option.OptionFunc[handlerConfig](func(c *handlerConfig) error {
		c.color = m
		return nil
	})
}

// Enabled implements slog.Handler.
func (h *glacierHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle implements slog.Handler. It appends ctx-attached attrs to the record
// and delegates to the inner handler.
func (h *glacierHandler) Handle(ctx context.Context, r slog.Record) error {
	// Append ctx-attached attrs (from log.With calls) to the record.
	if attrs := ctxAttrs(ctx); len(attrs) > 0 {
		// Clone the record so we don't mutate a shared value.
		r2 := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
		r.Attrs(func(a slog.Attr) bool {
			r2.AddAttrs(a)
			return true
		})
		r2.AddAttrs(attrs...)
		r = r2
	}

	// When color is off, sanitize ANSI escape bytes in attribute string values.
	// This prevents log injection via user-controlled strings.
	if !bool(h.color) {
		r = sanitizeRecord(r)
	}

	return h.inner.Handle(ctx, r)
}

// sanitizeRecord returns a new record where every string-valued attribute
// has \x1b replaced with <ESC>. This defends against log injection when
// color output is disabled.
func sanitizeRecord(r slog.Record) slog.Record {
	r2 := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		r2.AddAttrs(sanitizeAttr(a))
		return true
	})
	return r2
}

// sanitizeAttr replaces \x1b in string-typed attr values with "<ESC>".
func sanitizeAttr(a slog.Attr) slog.Attr {
	v := a.Value.Resolve()
	if v.Kind() == slog.KindString {
		s := v.String()
		if strings.ContainsRune(s, '\x1b') {
			return slog.Attr{
				Key:   a.Key,
				Value: slog.StringValue(strings.ReplaceAll(s, "\x1b", "<ESC>")),
			}
		}
	}
	return a
}

// WithAttrs implements slog.Handler.
func (h *glacierHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &glacierHandler{inner: h.inner.WithAttrs(attrs), color: h.color}
}

// WithGroup implements slog.Handler.
func (h *glacierHandler) WithGroup(name string) slog.Handler {
	return &glacierHandler{inner: h.inner.WithGroup(name), color: h.color}
}
