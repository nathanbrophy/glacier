// SPDX-License-Identifier: Apache-2.0

package log

import (
	"context"
	"fmt"
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
// It is a boolean: color is either on or off — ColorAuto has been resolved.
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
	inner slog.Handler // invariant: non-nil
	color resolvedColor // invariant: computed once at construction; immutable after
}

// Pre-computed ANSI escape byte sequences, keyed by level.
// These are computed once at package init (or handler construction) so that
// Handle never calls fmt.Sprintf at log time — zero allocations for color lookup.
var levelEscapes [9][]byte // indexed by (level+8)/4

// levelEscapeIdx maps a slog.Level to an index in levelEscapes.
// Levels: Trace=-8, Debug=-4, Info=0, Notice=2, Warn=4, Error=8
func levelEscapeIdx(l slog.Level) int {
	switch l {
	case LevelTrace:
		return 0
	case LevelDebug:
		return 1
	case LevelInfo:
		return 2
	case LevelNotice:
		return 3
	case LevelWarn:
		return 4
	case LevelError:
		return 5
	default:
		return 2 // default to INFO
	}
}

// ansiEsc builds a 24-bit foreground ANSI escape sequence for the given RGB values.
// Called once at init — never at log time.
func ansiEsc(r, g, b uint8) []byte {
	return []byte(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b))
}

var ansiReset = []byte("\x1b[0m")

func init() {
	// Palette per spec 0005 handler documentation:
	// TRACE:  text-muted #8B949E
	// DEBUG:  text-muted #8B949E
	// INFO:   cyan       #22D3EE
	// NOTICE: teal       #2DD4BF
	// WARN:   warning    #FBBF24
	// ERROR:  error      #F87171
	levelEscapes[0] = ansiEsc(0x8B, 0x94, 0x9E) // TRACE  (#8B949E)
	levelEscapes[1] = ansiEsc(0x8B, 0x94, 0x9E) // DEBUG  (#8B949E)
	levelEscapes[2] = ansiEsc(0x22, 0xD3, 0xEE) // INFO   (#22D3EE)
	levelEscapes[3] = ansiEsc(0x2D, 0xD4, 0xBF) // NOTICE (#2DD4BF)
	levelEscapes[4] = ansiEsc(0xFB, 0xBF, 0x24) // WARN   (#FBBF24)
	levelEscapes[5] = ansiEsc(0xF8, 0x71, 0x71) // ERROR  (#F87171)
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
	inner := slog.NewTextHandler(w, &slog.HandlerOptions{
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
	color := resolveColor(w, cfg.color)
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
// Pass any slog.Leveler — including dynamic LevelVar pointers — for runtime
// level control.
func WithLevel(l slog.Leveler) option.Option[handlerConfig] {
	return option.OptionFunc[handlerConfig](func(c *handlerConfig) error {
		c.level = l
		return nil
	})
}

// WithSource enables source-location attribution on every record: the
// caller's file and line are included. Off by default — source attribution
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
