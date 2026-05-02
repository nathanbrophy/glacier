// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/assert/require"
	"github.com/nathanbrophy/glacier/term"
)

func TestProgressRenderBasic(t *testing.T) {
	t.Parallel()
	p := term.NewProgress(100)
	p.Set(50)
	lines, done := p.Render()
	assert.False(t, done, "Progress.Render() done=true before Done() called")
	require.Equal(t, len(lines), 1)
	assert.Contains(t, lines[0], "%")
}

func TestProgressDone(t *testing.T) {
	t.Parallel()
	p := term.NewProgress(100)
	p.Done()
	_, done := p.Render()
	assert.True(t, done, "Progress.Render() done=false after Done() called")
}

func TestProgressIncrement(t *testing.T) {
	t.Parallel()
	p := term.NewProgress(100)
	p.Increment(25)
	p.Increment(25)
	lines, _ := p.Render()
	assert.Contains(t, lines[0], "50%")
}

func TestProgressIndeterminate(t *testing.T) {
	t.Parallel()
	p := term.NewProgress(-1)
	lines, done := p.Render()
	assert.False(t, done, "indeterminate Progress.Render() done=true before Done()")
	require.Equal(t, len(lines), 1)
}

func TestProgressSetNegative(t *testing.T) {
	t.Parallel()
	// Negative Set values are accepted (L-add-8).
	p := term.NewProgress(100)
	p.Set(-50)
	lines, done := p.Render()
	assert.False(t, done, "Progress.Render() done=true unexpectedly")
	_ = lines // just verify no panic
}

func TestProgressWithOptions(t *testing.T) {
	t.Parallel()
	p := term.NewProgress(1024*1024,
		term.WithProgressLabel("Downloading"),
		term.WithProgressShowSpeed(),
		term.WithProgressShowETA(),
		term.WithProgressShowBytes(),
		term.WithProgressGlyph("=", "-"),
	)
	p.Set(512 * 1024)
	lines, _ := p.Render()
	require.Equal(t, len(lines), 1)
	assert.True(t, strings.Contains(lines[0], "Downloading"), "Progress.Render() missing label: "+lines[0])
}

func TestProgressZeroTotal(t *testing.T) {
	t.Parallel()
	p := term.NewProgress(0)
	lines, _ := p.Render()
	// total=0 means 100% immediately.
	assert.Contains(t, lines[0], "100%")
}

func TestNewProgressDeterminate(t *testing.T) {
	t.Parallel()
	p := term.NewProgress(200)
	p.Set(100)
	lines, _ := p.Render()
	assert.Contains(t, lines[0], "50%")
}
