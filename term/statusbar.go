// SPDX-License-Identifier: Apache-2.0

package term

import (
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
	layout StatusBarLayout
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

// StatusBar is a multi-section status display. Sections are keyed by name.
// Thread-safe.
type StatusBar struct {
	cfg      statusBarConfig
	mu       sync.RWMutex
	sections map[string]string
	order    []string // insertion-order tracking
	closed   atomic.Bool
}

// NewStatusBar constructs an empty StatusBar.
func NewStatusBar(opts ...StatusBarOption) *StatusBar {
	cfg := statusBarConfig{}
	for _, o := range opts {
		if o == nil {
			continue
		}
		_ = o.applyStatusBar(&cfg)
	}
	return &StatusBar{
		cfg:      cfg,
		sections: make(map[string]string),
	}
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

// Close stops the StatusBar's render participation and releases resources.
// Idempotent. Implements io.Closer.
// Concurrency: goroutine-safe.
func (s *StatusBar) Close() error {
	s.closed.Store(true)
	return nil
}
