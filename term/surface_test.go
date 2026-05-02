// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/term"
)

// TestSurfaceClosed_TermPackage verifies that all exported symbols exist in the
// package surface as specified by the API section of spec 0016.
// This is a compile-time check: if the symbol doesn't exist, the file won't compile.
func TestSurfaceClosed_TermPackage(t *testing.T) {
	t.Parallel()

	// ── Capability detection ──────────────────────────────────────────────
	var _ term.Capabilities
	var _ term.ColorSupport
	_ = term.ColorNone
	_ = term.Color16
	_ = term.Color256
	_ = term.Color24Bit
	var _ func(io.Writer) term.Capabilities = term.Capability

	// ── Color ─────────────────────────────────────────────────────────────
	var _ term.Color
	var _ func(uint8, uint8, uint8) term.Color = term.RGB
	var _ func(string) (term.Color, error) = term.Hex
	_ = term.Cyan
	_ = term.Teal
	_ = term.Bg
	_ = term.Surface
	_ = term.Surface2
	_ = term.Text
	_ = term.TextMuted
	_ = term.TextFaint
	_ = term.Success
	_ = term.Warning
	_ = term.Error
	_ = term.Info
	_ = term.Border
	_ = term.Cyan100
	_ = term.Cyan300
	_ = term.Cyan500
	_ = term.Cyan700
	_ = term.Teal500
	_ = term.Teal700

	// ── Style ─────────────────────────────────────────────────────────────
	var _ term.Style
	var _ func() term.Style = term.New
	var _ func(term.Style, string) string = term.Sprint
	var _ func(io.Writer, term.Style, string) = term.Fprint
	s := term.New()
	var _ term.Style = s.Foreground(term.Cyan)
	var _ term.Style = s.Background(term.Bg)
	var _ term.Style = s.Bold()
	var _ term.Style = s.Italic()
	var _ term.Style = s.Underline()
	var _ term.Style = s.Dim()
	var _ term.Style = s.Strike()
	var _ string = s.Render("x")

	// ── Glyphs ────────────────────────────────────────────────────────────
	var _ term.GlyphInfo
	var _ func(string) string = term.Glyph
	var _ func(string, string, string) error = term.RegisterGlyph
	var _ func() []term.GlyphInfo = term.Glyphs

	// ── Beauty writer ─────────────────────────────────────────────────────
	var _ func(string, ...term.BoxOption) string = term.Box
	var _ func() term.BoxOption = term.WithRoundedCorners
	var _ func() term.BoxOption = term.WithSharpCorners
	var _ func() term.BoxOption = term.WithDoubleBorders
	var _ func(term.Style) term.BoxOption = term.WithBorderStyle
	var _ func(int, int, int, int) term.BoxOption = term.WithPadding
	var _ func(string) term.BoxOption = term.WithTitle
	var _ func(term.Style) term.BoxOption = term.WithTitleStyle
	var _ func(string, int) string = term.Center
	var _ func(string, int) string = term.Justify
	var _ func(string, int, int) string = term.Pad
	var _ func(string, int, string) string = term.Truncate
	var _ func(string, int) string = term.Wrap
	var _ term.ColumnOption
	var _ func(int) term.ColumnOption = term.WithColumnGap
	var _ func(int, term.Alignment) term.ColumnOption = term.WithColumnAlignment
	var _ func([][]string, ...term.ColumnOption) string = term.Columns
	var _ func(term.Style, ...string) string = term.Banner
	var _ term.Alignment
	_ = term.AlignLeft
	_ = term.AlignCenter
	_ = term.AlignRight

	// ── Prompts ───────────────────────────────────────────────────────────
	var _ func(string) term.PromptOption = term.WithDefault
	var _ func(func(string) error) term.PromptOption = term.WithValidator
	var _ func(int) term.PromptOption = term.WithMaxAttempts
	var _ func(time.Duration) term.PromptOption = term.WithTimeout
	var _ func() term.ConfirmOption = term.WithDefaultYes
	var _ func(context.Context, string, ...term.PromptOption) (string, error) = term.Prompt
	var _ func(context.Context, string) (string, error) = term.Password
	var _ func(context.Context, string, ...term.ConfirmOption) (bool, error) = term.Confirm

	// ── Animator ─────────────────────────────────────────────────────────
	var _ term.AnimatorOption
	var _ term.Animation
	var _ term.Handle
	var _ *term.Animator

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := term.NewAnimator(logger)
	if a == nil {
		t.Fatal("NewAnimator returned nil")
	}
	h := a.Add(&neverDoneAnimation{})
	h.Cancel()
	_ = a.Pause
	_ = a.Resume
	_ = a.Close

	// Animator option constructors.
	var _ func(time.Duration) term.AnimatorOption = term.WithRefreshRate
	var _ func(io.Writer) term.AnimatorOption = term.WithWriter
	var _ func(int) term.AnimatorOption = term.WithMaxBufferedRecords

	// ── Built-in animations ───────────────────────────────────────────────
	var _ term.SpinnerOption
	var _ func(term.Style) term.SpinnerOption = term.WithSpinnerStyle
	var _ func([]string) term.SpinnerOption = term.WithSpinnerFrames
	var _ func(string, ...term.SpinnerOption) term.Animation = term.Spinner

	var _ *term.Progress
	var _ func(int64, ...term.ProgressOption) *term.Progress = term.NewProgress
	var _ term.ProgressOption
	var _ func(string) term.ProgressOption = term.WithProgressLabel
	var _ func() term.ProgressOption = term.WithProgressShowSpeed
	var _ func() term.ProgressOption = term.WithProgressShowETA
	var _ func() term.ProgressOption = term.WithProgressShowBytes
	var _ func(term.Style) term.ProgressOption = term.WithProgressBarStyle
	var _ func(string, string) term.ProgressOption = term.WithProgressGlyph

	var _ *term.StatusBar
	var _ func(...term.StatusBarOption) *term.StatusBar = term.NewStatusBar
	var _ term.StatusBarOption
	var _ func(term.StatusBarLayout) term.StatusBarOption = term.WithStatusBarLayout
	_ = term.StatusBarLines
	_ = term.StatusBarColumns

	var _ *term.DownloadProgress
	var _ func(io.Reader, int64, string, ...term.ProgressOption) *term.DownloadProgress = term.NewDownloadProgress

	// SelectOption.
	var _ term.SelectOption

	// ── Errors ────────────────────────────────────────────────────────────
	var _ error = term.ErrCancelled
	var _ error = term.ErrNotInteractive
	var _ error = term.ErrTimeout
	var _ error = term.ErrTooManyAttempts
	var _ *term.HexParseError
	var _ *term.GlyphError
}
