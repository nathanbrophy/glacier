// SPDX-License-Identifier: Apache-2.0

package term

import (
	"fmt"
	"sync/atomic"
)

// spinnerConfig holds resolved options for a Spinner animation.
type spinnerConfig struct {
	style  Style
	frames []string
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

// defaultSpinnerFrames is the default braille spinner (8 frames).
var defaultSpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠇"}

// spinnerAnimation implements Animation for a Spinner.
type spinnerAnimation struct {
	label string
	cfg   spinnerConfig
	frame atomic.Int64
}

// Render implements Animation. Returns one line formatted as "<glyph> <label>".
// Always reports done=false: a spinner has no natural completion. Removal from
// the active set is the caller's responsibility via Handle.Cancel on the
// Handle returned by Animator.Add.
// Performance target: ≤ 500 ns/op per §23.13.
func (s *spinnerAnimation) Render() ([]string, bool) {
	n := int(s.frame.Add(1) - 1)
	frames := s.cfg.frames
	glyph := frames[n%len(frames)]
	styled := renderTo(s.cfg.style, glyph, nil)
	return []string{styled + " " + s.label}, false
}

// Spinner returns an Animation that cycles through spinner glyphs at each frame.
// Default frames: spinner_braille_0 through spinner_braille_7 (8 frames).
// The label is rendered after the glyph on the same line.
//
// A Spinner has no natural completion — unlike Progress, there is no Done
// method. Stop a running spinner via Handle.Cancel on the Handle returned by
// Animator.Add; this removes it from the active set on the next frame.
//
// Performance target: ≤ 500 ns/op per frame per §23.13.
func Spinner(label string, opts ...SpinnerOption) Animation {
	cfg := spinnerConfig{frames: defaultSpinnerFrames}
	for _, o := range opts {
		if o == nil {
			continue
		}
		_ = o.applySpinner(&cfg)
	}
	return &spinnerAnimation{label: label, cfg: cfg}
}
