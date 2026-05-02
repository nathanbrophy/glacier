// SPDX-License-Identifier: Apache-2.0

package term

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// progressConfig holds options for a Progress animation.
type progressConfig struct {
	label       string
	showSpeed   bool
	showETA     bool
	showBytes   bool
	barStyle    Style
	filledGlyph string    // default "█"
	emptyGlyph  string    // default "░"
	animator    *Animator // nil unless WithProgressAnimator was applied
}

// ProgressOption configures a Progress animation.
type ProgressOption interface{ applyProgress(*progressConfig) error }

type progressOptionFunc func(*progressConfig) error

func (f progressOptionFunc) applyProgress(c *progressConfig) error { return f(c) }

// WithProgressLabel sets the descriptive label rendered after the bar.
func WithProgressLabel(s string) ProgressOption {
	return progressOptionFunc(func(c *progressConfig) error { c.label = s; return nil })
}

// WithProgressShowSpeed enables bytes/sec or items/sec display.
// Speed is computed from a sliding 5-second window of increment events.
func WithProgressShowSpeed() ProgressOption {
	return progressOptionFunc(func(c *progressConfig) error { c.showSpeed = true; return nil })
}

// WithProgressShowETA enables estimated-time-remaining display.
// ETA is computed from current speed and remaining bytes.
// Displays "ETA --" when speed is zero.
func WithProgressShowETA() ProgressOption {
	return progressOptionFunc(func(c *progressConfig) error { c.showETA = true; return nil })
}

// WithProgressShowBytes enables "3.2 MB / 4.8 MB" byte-count display.
// Uses 1024-based SI prefixes (KiB, MiB, GiB).
func WithProgressShowBytes() ProgressOption {
	return progressOptionFunc(func(c *progressConfig) error { c.showBytes = true; return nil })
}

// WithProgressBarStyle applies s to the filled portion of the progress bar.
func WithProgressBarStyle(s Style) ProgressOption {
	return progressOptionFunc(func(c *progressConfig) error { c.barStyle = s; return nil })
}

// WithProgressGlyph sets the filled (default "█") and empty (default "░") glyphs.
func WithProgressGlyph(filled, empty string) ProgressOption {
	return progressOptionFunc(func(c *progressConfig) error {
		c.filledGlyph = filled
		c.emptyGlyph = empty
		return nil
	})
}

// WithProgressAnimator injects a shared Animator into the Progress.
// When set, Run is a no-op — the consumer is expected to drive the injected
// Animator themselves. The Progress is registered into the shared Animator at
// construction time; Done or Close marks it complete.
func WithProgressAnimator(a *Animator) ProgressOption {
	return progressOptionFunc(func(c *progressConfig) error {
		c.animator = a
		return nil
	})
}

// speedSample records a data point for the sliding-window speed calculation.
type speedSample struct {
	t time.Time
	n int64
}

// Progress is a progress-bar animation. Thread-safe.
//
// Invariant: current is clamped to [0, total] for percentage display;
// values outside the range are accepted and render >100% or <0% visually.
// Invariant: total == -1 signals indeterminate mode (spinner-style bar + byte counter).
type Progress struct {
	total    int64
	current  atomic.Int64
	done     atomic.Bool
	cfg      progressConfig
	injected bool      // true when animator was provided via WithProgressAnimator
	owned    bool      // true when Run constructed the internal Animator
	animator *Animator // shared (injected) or internal (owned); nil before Run if not injected

	mu      sync.Mutex
	samples []speedSample // sliding 5-second window
	started time.Time
}

// NewProgress constructs a Progress with the given total byte/item count.
// total == -1 for indeterminate (unknown total) mode.
//
// When WithProgressAnimator is provided, the Progress is registered into the
// shared Animator immediately and Run becomes a no-op.
func NewProgress(total int64, opts ...ProgressOption) *Progress {
	cfg := progressConfig{
		filledGlyph: "█",
		emptyGlyph:  "░",
	}
	for _, o := range opts {
		if o == nil {
			continue
		}
		_ = o.applyProgress(&cfg)
	}
	injected := cfg.animator != nil
	p := &Progress{total: total, cfg: cfg, started: time.Now(), injected: injected}
	if injected {
		p.animator = cfg.animator
		cfg.animator.Add(p)
	}
	return p
}

// Set sets the current progress to n. Goroutine-safe (atomic).
// Negative values are accepted and render as-is.
func (p *Progress) Set(n int64) {
	p.current.Store(n)
	p.recordSample(n)
}

// Increment adds n to the current progress. Goroutine-safe (atomic).
func (p *Progress) Increment(n int64) {
	cur := p.current.Add(n)
	p.recordSample(cur)
}

// Done marks the progress as complete. Subsequent Render calls return done=true.
// Goroutine-safe. Idempotent.
func (p *Progress) Done() {
	p.done.Store(true)
}

// Run starts the progress animation. If no Animator was injected via
// WithProgressAnimator, an internal Animator is constructed (preferring
// slog.Default() when available, falling back to a fresh os.Stderr logger) and
// this Progress is added to it. Run blocks until ctx is cancelled or Done is
// called (which causes Render to return done=true and the Animator to exit).
//
// If an Animator was injected, Run is a no-op and returns nil immediately; the
// caller is responsible for running that Animator.
//
// Concurrency: goroutine-safe.
func (p *Progress) Run(ctx context.Context) error {
	if p.injected {
		return nil
	}
	a := NewAnimator(newInternalLogger())
	p.animator = a
	p.owned = true
	a.Add(p)
	return a.Run(ctx)
}

// Close stops the Progress animation. If this Progress owns an internal Animator,
// Close stops it. If using an injected Animator, Close marks the animation done
// so its next Render returns done=true.
//
// Idempotent; returns nil on every call.
// Concurrency: goroutine-safe.
func (p *Progress) Close() error {
	p.done.Store(true)
	if p.owned && p.animator != nil {
		return p.animator.Close()
	}
	return nil
}

// recordSample records a data point for the speed window.
func (p *Progress) recordSample(n int64) {
	if !p.cfg.showSpeed && !p.cfg.showETA {
		return
	}
	now := time.Now()
	p.mu.Lock()
	p.samples = append(p.samples, speedSample{t: now, n: n})
	// Trim samples older than 5 seconds.
	cutoff := now.Add(-5 * time.Second)
	i := 0
	for i < len(p.samples) && p.samples[i].t.Before(cutoff) {
		i++
	}
	p.samples = p.samples[i:]
	p.mu.Unlock()
}

// speed returns bytes/sec from the sliding window. Returns 0 if insufficient data.
func (p *Progress) speed() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.samples) < 2 {
		return 0
	}
	first := p.samples[0]
	last := p.samples[len(p.samples)-1]
	dt := last.t.Sub(first.t).Seconds()
	if dt <= 0 {
		return 0
	}
	dn := float64(last.n - first.n)
	return dn / dt
}

// Render implements Animation. Returns a single line.
// Performance target: ≤ 2 µs/op per §23.13.
func (p *Progress) Render() ([]string, bool) {
	if p.done.Load() {
		return nil, true
	}
	cur := p.current.Load()
	total := p.total

	caps := Capability(os.Stderr)
	termW := caps.Width
	if termW <= 0 {
		termW = 80
	}

	var sb strings.Builder

	// Build suffix parts first to know how wide the bar can be.
	var parts []string
	if total >= 0 {
		pct := int(math.Round(float64(cur) / float64(total) * 100))
		if total == 0 {
			pct = 100
		}
		parts = append(parts, fmt.Sprintf("%d%%", pct))
	}
	if p.cfg.showBytes {
		if total >= 0 {
			parts = append(parts, fmt.Sprintf("%s/%s", formatBytes(cur), formatBytes(total)))
		} else {
			parts = append(parts, formatBytes(cur))
		}
	}
	if p.cfg.showSpeed {
		sp := p.speed()
		if sp > 0 {
			parts = append(parts, formatBytes(int64(sp))+"/s")
		}
	}
	if p.cfg.showETA && total >= 0 {
		sp := p.speed()
		rem := total - cur
		if sp > 0 && rem > 0 {
			eta := time.Duration(float64(rem)/sp) * time.Second
			parts = append(parts, "ETA "+fmtDuration(eta))
		} else {
			parts = append(parts, "ETA --")
		}
	}
	if p.cfg.label != "" {
		parts = append(parts, p.cfg.label)
	}

	suffix := strings.Join(parts, "  ")
	suffixW := utf8.RuneCountInString(suffix)

	// Bar width: termW - 2 (brackets) - 1 (space) - suffixW
	barW := termW - 3 - suffixW
	if barW < 4 {
		barW = 4
	}

	if total >= 0 {
		filled := int(math.Round(float64(cur) / float64(total) * float64(barW)))
		if total == 0 {
			filled = barW
		}
		if filled > barW {
			filled = barW
		}
		if filled < 0 {
			filled = 0
		}
		filledStr := strings.Repeat(p.cfg.filledGlyph, filled)
		emptyStr := strings.Repeat(p.cfg.emptyGlyph, barW-filled)
		styledFilled := renderTo(p.cfg.barStyle, filledStr, os.Stderr)
		sb.WriteString("[")
		sb.WriteString(styledFilled)
		sb.WriteString(emptyStr)
		sb.WriteString("]")
	} else {
		// Indeterminate: animated spinner-style bar.
		frame := int(time.Since(p.started).Milliseconds()/100) % barW
		inner := strings.Repeat(p.cfg.emptyGlyph, barW)
		runes := []rune(inner)
		if frame < len(runes) {
			runes[frame] = []rune(p.cfg.filledGlyph)[0]
		}
		sb.WriteString("[")
		sb.WriteString(string(runes))
		sb.WriteString("]")
	}

	if suffix != "" {
		sb.WriteString(" ")
		sb.WriteString(suffix)
	}

	return []string{sb.String()}, false
}

// formatBytes formats n as a human-readable 1024-based size string.
func formatBytes(n int64) string {
	const (
		kib = 1024
		mib = 1024 * kib
		gib = 1024 * mib
	)
	switch {
	case n >= gib:
		return fmt.Sprintf("%.1f GiB", float64(n)/float64(gib))
	case n >= mib:
		return fmt.Sprintf("%.1f MiB", float64(n)/float64(mib))
	case n >= kib:
		return fmt.Sprintf("%.1f KiB", float64(n)/float64(kib))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// fmtDuration formats a duration as a human-readable string.
func fmtDuration(d time.Duration) string {
	s := int(d.Seconds())
	if s < 60 {
		return fmt.Sprintf("%ds", s)
	}
	m := s / 60
	s = s % 60
	if m < 60 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	h := m / 60
	m = m % 60
	return fmt.Sprintf("%dh%dm", h, m)
}
