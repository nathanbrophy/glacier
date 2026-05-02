// SPDX-License-Identifier: Apache-2.0

package term

import (
	"context"
	"fmt"
	"sync/atomic"
)

// spinnerConfig holds resolved options for a Spinner animation.
type spinnerConfig struct {
	style    Style
	frames   []string
	animator *Animator // nil unless WithSpinnerAnimator was applied
}

// SpinnerOption configures a Spinner animation.
type SpinnerOption interface{ applySpinner(*spinnerConfig) error }

type spinnerOptionFunc func(*spinnerConfig) error

func (f spinnerOptionFunc) applySpinner(c *spinnerConfig) error { return f(c) }

// WithSpinnerStyle applies s to the spinning glyph character.
func WithSpinnerStyle(s Style) SpinnerOption {
	return spinnerOptionFunc(func(c *spinnerConfig) error { c.style = s; return nil })
}

// WithSpinnerFrames overrides the default braille-spinner frame sequence.
// Precondition: frames must not be empty.
func WithSpinnerFrames(frames []string) SpinnerOption {
	return spinnerOptionFunc(func(c *spinnerConfig) error {
		if len(frames) == 0 {
			return fmt.Errorf("term: WithSpinnerFrames: frames must not be empty")
		}
		c.frames = frames
		return nil
	})
}

// WithSpinnerAnimator injects a shared Animator into the SpinnerAnimator.
// When set, Run is a no-op — the consumer is expected to drive the injected
// Animator themselves. The SpinnerAnimator is registered into the shared
// Animator at construction time; Close marks it done.
func WithSpinnerAnimator(a *Animator) SpinnerOption {
	return spinnerOptionFunc(func(c *spinnerConfig) error {
		c.animator = a
		return nil
	})
}

// defaultSpinnerFrames is the default braille spinner (8 frames).
var defaultSpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠇"}

// SpinnerAnimator is a self-contained spinner animation. It satisfies the
// Animation interface so it can be passed to Animator.Add for multi-animation
// coordination, and also exposes Run and Close for standalone self-contained use.
//
// Concurrency: all exported methods are goroutine-safe.
type SpinnerAnimator struct {
	label    string
	cfg      spinnerConfig
	frame    atomic.Int64
	done     atomic.Bool
	injected bool      // true when animator was provided via WithSpinnerAnimator
	owned    bool      // true when Run constructed the internal Animator
	animator *Animator // shared (injected) or internal (owned); nil before Run if not injected
}

// Render implements Animation. Returns one line formatted as "<glyph> <label>".
// Performance target: ≤ 500 ns/op per §23.13.
func (s *SpinnerAnimator) Render() ([]string, bool) {
	if s.done.Load() {
		return nil, true
	}
	n := int(s.frame.Add(1) - 1)
	frames := s.cfg.frames
	glyph := frames[n%len(frames)]
	styled := renderTo(s.cfg.style, glyph, nil)
	return []string{styled + " " + s.label}, false
}

// Run starts the spinner. If no Animator was injected via WithSpinnerAnimator,
// an internal Animator is constructed — it uses slog.Default() when the default
// logger's handler is not already intercepted, and falls back to a fresh
// os.Stderr handler otherwise — and this spinner is added to it. Run blocks
// until ctx is cancelled or Close is called.
//
// If an Animator was injected, Run is a no-op and returns nil immediately; the
// caller is responsible for running that Animator.
//
// Concurrency: goroutine-safe.
func (s *SpinnerAnimator) Run(ctx context.Context) error {
	if s.injected {
		return nil
	}
	a := NewAnimator(newInternalLogger())
	s.animator = a
	s.owned = true
	a.Add(s)
	return a.Run(ctx)
}

// Close stops the spinner. If this SpinnerAnimator owns an internal Animator,
// Close stops it. If using an injected Animator, Close marks the animation done
// so its next Render returns done=true.
//
// Idempotent; returns nil on every call.
// Concurrency: goroutine-safe.
func (s *SpinnerAnimator) Close() error {
	s.done.Store(true)
	if s.owned && s.animator != nil {
		return s.animator.Close()
	}
	return nil
}

// Spinner returns a *SpinnerAnimator that cycles through spinner glyphs at each frame.
// Default frames: spinner_braille_0 through spinner_braille_7 (8 frames).
// The label is rendered after the glyph on the same line.
//
// *SpinnerAnimator satisfies the Animation interface; it can be passed to
// Animator.Add for multi-animation coordination, or used standalone via Run/Close.
//
// When WithSpinnerAnimator is provided, the spinner is registered into the shared
// Animator immediately and Run becomes a no-op.
//
// Performance target: ≤ 500 ns/op per frame per §23.13.
func Spinner(label string, opts ...SpinnerOption) *SpinnerAnimator {
	cfg := spinnerConfig{frames: defaultSpinnerFrames}
	for _, o := range opts {
		if o == nil {
			continue
		}
		_ = o.applySpinner(&cfg)
	}
	injected := cfg.animator != nil
	sa := &SpinnerAnimator{label: label, cfg: cfg, injected: injected}
	if injected {
		sa.animator = cfg.animator
		cfg.animator.Add(sa)
	}
	return sa
}
