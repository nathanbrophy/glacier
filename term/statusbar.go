// SPDX-License-Identifier: Apache-2.0

package term

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
)

// StatusBarLayout controls how sections are arranged.
type StatusBarLayout int

const (
	// StatusBarLines renders one section per line (default).
	StatusBarLines StatusBarLayout = iota
	// StatusBarColumns renders sections side-by-side in columns.
	StatusBarColumns
)

// statusBarConfig holds options for a StatusBar animation.
type statusBarConfig struct {
	layout   StatusBarLayout
	animator *Animator // nil unless WithStatusBarAnimator was applied
}

// StatusBarOption configures a StatusBar animation.
type StatusBarOption interface{ applyStatusBar(*statusBarConfig) error }

type statusBarOptionFunc func(*statusBarConfig) error

func (f statusBarOptionFunc) applyStatusBar(c *statusBarConfig) error { return f(c) }

// WithStatusBarLayout controls whether sections render on separate lines
// (StatusBarLines, default) or side-by-side in columns (StatusBarColumns).
func WithStatusBarLayout(l StatusBarLayout) StatusBarOption {
	return statusBarOptionFunc(func(c *statusBarConfig) error { c.layout = l; return nil })
}

// WithStatusBarAnimator injects a shared Animator into the StatusBar.
// When set, Run is a no-op — the consumer is expected to drive the injected
// Animator themselves. The StatusBar is registered into the shared Animator at
// construction time; Close marks it done.
func WithStatusBarAnimator(a *Animator) StatusBarOption {
	return statusBarOptionFunc(func(c *statusBarConfig) error { c.animator = a; return nil })
}

// StatusBar is a multi-section status display. Sections are keyed by name.
// Thread-safe.
type StatusBar struct {
	cfg      statusBarConfig
	mu       sync.RWMutex
	sections map[string]string
	order    []string // insertion-order tracking
	closed   atomic.Bool
	injected bool      // true when animator was provided via WithStatusBarAnimator
	owned    bool      // true when Run constructed the internal Animator
	animator *Animator // shared (injected) or internal (owned); nil before Run if not injected
}

// NewStatusBar constructs an empty StatusBar.
//
// When WithStatusBarAnimator is provided, the StatusBar is registered into the
// shared Animator immediately and Run becomes a no-op.
func NewStatusBar(opts ...StatusBarOption) *StatusBar {
	cfg := statusBarConfig{}
	for _, o := range opts {
		if o == nil {
			continue
		}
		_ = o.applyStatusBar(&cfg)
	}
	injected := cfg.animator != nil
	sb := &StatusBar{
		cfg:      cfg,
		sections: make(map[string]string),
		injected: injected,
	}
	if injected {
		sb.animator = cfg.animator
		cfg.animator.Add(sb)
	}
	return sb
}

// SetSection adds or replaces section name with content.
// Goroutine-safe.
func (s *StatusBar) SetSection(name, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.sections[name]; !exists {
		s.order = append(s.order, name)
	}
	s.sections[name] = content
}

// Remove deletes section name. No-op if name is not present.
// Goroutine-safe.
func (s *StatusBar) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.sections[name]; !exists {
		return
	}
	delete(s.sections, name)
	for i, k := range s.order {
		if k == name {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
}

// Render implements Animation. Returns one line per section (StatusBarLines)
// or a single row of columns (StatusBarColumns). Returns done=false always;
// StatusBar is closed by the caller via its Handle.
func (s *StatusBar) Render() ([]string, bool) {
	if s.closed.Load() {
		return nil, true
	}
	s.mu.RLock()
	order := make([]string, len(s.order))
	copy(order, s.order)
	sects := make(map[string]string, len(s.sections))
	for k, v := range s.sections {
		sects[k] = v
	}
	s.mu.RUnlock()

	switch s.cfg.layout {
	case StatusBarColumns:
		var row []string
		for _, k := range order {
			row = append(row, k+": "+sects[k])
		}
		return []string{strings.Join(row, "  |  ")}, false
	default: // StatusBarLines
		var lines []string
		for _, k := range order {
			lines = append(lines, k+": "+sects[k])
		}
		if len(lines) == 0 {
			return []string{""}, false
		}
		return lines, false
	}
}

// Run starts the StatusBar animation. If no Animator was injected via
// WithStatusBarAnimator, an internal Animator is constructed (preferring
// slog.Default() when available, falling back to a fresh os.Stderr logger) and
// this StatusBar is added to it. Run blocks until ctx is cancelled or Close
// is called (which causes Render to return done=true and the Animator to exit).
//
// If an Animator was injected, Run is a no-op and returns nil immediately; the
// caller is responsible for running that Animator.
//
// Concurrency: goroutine-safe.
func (s *StatusBar) Run(ctx context.Context) error {
	if s.injected {
		return nil
	}
	a := NewAnimator(newInternalLogger())
	s.animator = a
	s.owned = true
	a.Add(s)
	return a.Run(ctx)
}

// Close stops the StatusBar's render participation and releases resources.
// If this StatusBar owns an internal Animator, Close stops it. If using an
// injected Animator, Close marks the animation done so its next Render returns
// done=true.
//
// Idempotent. Implements io.Closer.
// Concurrency: goroutine-safe.
func (s *StatusBar) Close() error {
	s.closed.Store(true)
	if s.owned && s.animator != nil {
		return s.animator.Close()
	}
	return nil
}
