// SPDX-License-Identifier: Apache-2.0

package term

import (
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
	filledGlyph string // default "█"
	emptyGlyph  string // default "░"
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
	total   int64
	current atomic.Int64
	done    atomic.Bool
	cfg     progressConfig

	mu      sync.Mutex
	samples []speedSample // sliding 5-second window
	started time.Time
}

// NewProgress constructs a Progress with the given total byte/item count.
// total == -1 for indeterminate (unknown total) mode.
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
	return &Progress{total: total, cfg: cfg, started: time.Now()}
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
