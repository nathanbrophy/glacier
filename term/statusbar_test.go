// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"strings"
	"testing"

	"github.com/nathanbrophy/glacier/term"
)

func TestStatusBarSetSection(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar()
	sb.SetSection("phase", "running")
	lines, done := sb.Render()
	if done {
		t.Error("StatusBar.Render() done=true before Close()")
	}
	found := false
	for _, l := range lines {
		if strings.Contains(l, "phase") && strings.Contains(l, "running") {
			found = true
		}
	}
	if !found {
		t.Errorf("StatusBar.Render() missing section content: %v", lines)
	}
}

func TestStatusBarRemoveSection(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar()
	sb.SetSection("a", "hello")
	sb.SetSection("b", "world")
	sb.Remove("a")
	lines, _ := sb.Render()
	for _, l := range lines {
		if strings.Contains(l, "a:") {
			t.Errorf("removed section 'a' still appears in render: %v", lines)
		}
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
	if err1 != nil {
		t.Errorf("Close() first = %v, want nil", err1)
	}
	if err2 != nil {
		t.Errorf("Close() second = %v, want nil (idempotent)", err2)
	}
}

func TestStatusBarCloseMakesDone(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar()
	sb.SetSection("x", "v")
	_ = sb.Close()
	_, done := sb.Render()
	if !done {
		t.Error("after Close(), StatusBar.Render() done=false")
	}
}

func TestStatusBarColumnLayout(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar(term.WithStatusBarLayout(term.StatusBarColumns))
	sb.SetSection("cpu", "12%")
	sb.SetSection("mem", "1.2GB")
	lines, done := sb.Render()
	if done {
		t.Error("StatusBar.Render() done=true before Close()")
	}
	if len(lines) != 1 {
		t.Errorf("StatusBarColumns: expected 1 line, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "cpu") || !strings.Contains(lines[0], "mem") {
		t.Errorf("StatusBarColumns line missing sections: %q", lines[0])
	}
}

func TestStatusBarLineLayout(t *testing.T) {
	t.Parallel()
	sb := term.NewStatusBar(term.WithStatusBarLayout(term.StatusBarLines))
	sb.SetSection("s1", "v1")
	sb.SetSection("s2", "v2")
	lines, _ := sb.Render()
	if len(lines) < 2 {
		t.Errorf("StatusBarLines: expected ≥ 2 lines, got %d: %v", len(lines), lines)
	}
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
		if strings.Contains(l, "old") {
			t.Errorf("old value still present after replacement: %v", lines)
		}
	}
	if !found {
		t.Errorf("new value not found in lines: %v", lines)
	}
}
