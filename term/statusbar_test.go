// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/term"
)

func TestStatusBarSetSection(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar()
	sb.SetSection("phase", "running")
	lines, done := sb.Render()
	assert.False(t, done, "StatusBar.Render() done=true before Close()")
	found := false
	for _, l := range lines {
		if strings.Contains(l, "phase") && strings.Contains(l, "running") {
			found = true
		}
	}
	assert.True(t, found, fmt.Sprintf("StatusBar.Render() missing section content: %v", lines))
}

func TestStatusBarRemoveSection(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar()
	sb.SetSection("a", "hello")
	sb.SetSection("b", "world")
	sb.Remove("a")
	lines, _ := sb.Render()
	for _, l := range lines {
		assert.False(t, strings.Contains(l, "a:"),
			fmt.Sprintf("removed section 'a' still appears in render: %v", lines))
	}
}

func TestStatusBarRemoveNonExistent(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar()
	// Remove on non-existent key is a no-op (no panic).
	sb.Remove("ghost")
}

func TestStatusBarCloseIdempotent(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar()
	err1 := sb.Close()
	err2 := sb.Close()
	assert.NoError(t, err1, "Close() first = %v, want nil", err1)
	assert.NoError(t, err2, "Close() second = %v, want nil (idempotent)", err2)
}

func TestStatusBarCloseMakesDone(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar()
	sb.SetSection("x", "v")
	_ = sb.Close()
	_, done := sb.Render()
	assert.True(t, done, "after Close(), StatusBar.Render() done=false")
}

func TestStatusBarColumnLayout(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar(term.WithStatusBarLayout(term.StatusBarColumns))
	sb.SetSection("cpu", "12%")
	sb.SetSection("mem", "1.2GB")
	lines, done := sb.Render()
	assert.False(t, done, "StatusBar.Render() done=true before Close()")
	assert.True(t, len(lines) == 1, fmt.Sprintf("StatusBarColumns: expected 1 line, got %d: %v", len(lines), lines))
	assert.True(t, strings.Contains(lines[0], "cpu") && strings.Contains(lines[0], "mem"),
		fmt.Sprintf("StatusBarColumns line missing sections: %q", lines[0]))
}

func TestStatusBarLineLayout(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar(term.WithStatusBarLayout(term.StatusBarLines))
	sb.SetSection("s1", "v1")
	sb.SetSection("s2", "v2")
	lines, _ := sb.Render()
	assert.True(t, len(lines) >= 2,
		fmt.Sprintf("StatusBarLines: expected >= 2 lines, got %d: %v", len(lines), lines))
}

func TestStatusBarReplacesSection(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar()
	sb.SetSection("key", "old")
	sb.SetSection("key", "new")
	lines, _ := sb.Render()
	found := false
	for _, l := range lines {
		if strings.Contains(l, "new") {
			found = true
		}
		assert.False(t, strings.Contains(l, "old"),
			fmt.Sprintf("old value still present after replacement: %v", lines))
	}
	assert.True(t, found, fmt.Sprintf("new value not found in lines: %v", lines))
}
