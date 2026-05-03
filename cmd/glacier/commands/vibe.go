// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nathanbrophy/glacier/cmd/glacier/internal/report"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/shimmer"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/vibetips"
	"github.com/nathanbrophy/glacier/term"
)

// VibeCmd plays the Glacier vibes animation.
//
// +glacier:command name=vibe parent=glacier
type VibeCmd struct {
	// Duration bounds the loop. Default 0s means run until key press or SIGINT.
	//
	// +glacier:default 0s
	Duration time.Duration

	// NoTips suppresses the rotating tip line.
	//
	// +glacier:default false
	NoTips bool

	// Seed seeds the tip-rotation order. 0 means time-based.
	//
	// +glacier:default 0
	Seed int64

	// ASCII forces the kaomoji-only fallback even on capable terminals.
	//
	// +glacier:default false
	ASCII bool
}

// Run plays the glacier vibes animation. On a TTY with no --ascii flag it
// runs an animated box with cycling bear expressions and rotating tips.
// On a non-TTY or with --ascii it prints a static fallback and exits.
func (c *VibeCmd) Run(ctx context.Context) error {
	report.Status(report.Calm, "glacier vibe")

	caps := term.Capability(os.Stderr)
	useFull := caps.IsTTY && !c.ASCII

	if !useFull {
		return c.runStatic()
	}

	// Try to acquire raw mode for keypress detection.
	restore, err := term.AcquireRaw(ctx)
	if err != nil {
		// Fall back to static mode when raw mode is unavailable.
		return c.runStatic()
	}
	defer restore()

	tips := vibetips.Shuffled(c.Seed)

	anim := &vibeAnimation{
		startTime: time.Now(),
		tips:      tips,
		noTips:    c.NoTips,
	}

	animator := term.NewAnimator(slog.Default())
	animator.Add(anim)

	runCtx := ctx
	var runCancel context.CancelFunc

	if c.Duration > 0 {
		runCtx, runCancel = context.WithTimeout(ctx, c.Duration)
		defer runCancel()
	} else {
		// Start a goroutine that cancels the context on any keypress.
		runCtx, runCancel = context.WithCancel(ctx)
		go func() {
			buf := make([]byte, 1)
			_, _ = os.Stdin.Read(buf)
			runCancel()
		}()
		defer runCancel()
	}

	err = animator.Run(runCtx)
	// ErrCancelled means the user pressed a key or context was cancelled — that's fine.
	if err != nil && err.Error() != "term: animator: cancelled" {
		// Unwrap to check for the known cancellation sentinel.
		if !isCancelledErr(err) {
			report.Status(report.Err, err.Error())
			return err
		}
	}

	report.Status(report.Confident, "nice.")
	return nil
}

// runStatic prints a static fallback: kaomoji, wordmark, tagline, and one tip.
func (c *VibeCmd) runStatic() error {
	bear := "ʕ•ᴥ•ʔ"
	fmt.Fprintln(os.Stdout, bear+"  GLACIER")
	fmt.Fprintln(os.Stdout, "Less plumbing. More Go.")
	if !c.NoTips {
		tips := vibetips.Shuffled(c.Seed)
		if len(tips) > 0 {
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, bear+" "+tips[0].Body)
		}
	}
	return nil
}

// isCancelledErr reports whether err is the animator's cancellation sentinel.
// We compare the error string because the sentinel lives in a different package.
func isCancelledErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "cancelled")
}

// vibeAnimation is the animation rendered by glacier vibe.
type vibeAnimation struct {
	startTime time.Time
	tips      []vibetips.Tip
	noTips    bool
	tick      atomic.Int64
}

// Render implements term.Animation. It cycles bear expressions every 3 seconds
// and rotates the tip line every 5 seconds. The wordmark shimmer advances
// one gradient stop per tick (100 ms), completing a full 6-stop aurora cycle
// every 600 ms per spec 0032 D-S58.
func (v *vibeAnimation) Render() ([]string, bool) {
	t := v.tick.Add(1)

	// Bear expression cycles every 30 ticks (3 seconds at 100ms/frame).
	expressionIdx := (t / 30) % 3
	kaomojis := [3]string{"ʕ•ᴥ•ʔ", "ʕ⌐■-■ʔ", "ʕ•_•ʔ"}
	bearKaomoji := kaomojis[expressionIdx]

	tip := ""
	if !v.noTips && len(v.tips) > 0 {
		// Rotate every 50 ticks (5 seconds at 100ms/frame).
		tipIdx := int((t / 50) % int64(len(v.tips)))
		tip = bearKaomoji + " " + v.tips[tipIdx].Body
	}

	// Shimmer phase: one stop per tick, full cycle every 6 ticks (600 ms).
	caps := term.Capability(os.Stderr)
	phase := int(t % 6)
	c24 := term.ShouldColor(os.Stderr) && caps.SupportsColor >= term.Color24Bit
	c256 := term.ShouldColor(os.Stderr) && !c24 && caps.SupportsColor >= term.Color256
	wm := shimmer.Wordmark(phase, c24, c256)

	var content strings.Builder
	content.WriteString(bearKaomoji + "  " + wm + "\n\n")
	content.WriteString("Less plumbing. More Go.\n")
	if tip != "" {
		content.WriteString("\n" + tip + "\n")
	}
	content.WriteString("\nPress any key to exit.")

	box := term.Box(
		content.String(),
		term.WithTitle("glacier vibes"),
		term.WithRoundedCorners(),
		term.WithPadding(1, 2, 1, 2),
	)
	return strings.Split(box, "\n"), false
}
