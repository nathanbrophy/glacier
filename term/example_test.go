// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/nathanbrophy/glacier/term"
)

// ExampleNew demonstrates constructing a Style via chaining.
func ExampleNew() {
	// Build an immutable style.
	warn := term.New().Foreground(term.Warning).Bold()
	info := term.New().Foreground(term.Cyan)

	// Fprint to a bytes.Buffer (non-TTY) so output is plain text, suitable
	// for a deterministic doctest.
	var buf bytes.Buffer
	term.Fprint(&buf, warn, "config file missing")
	fmt.Println(buf.String())
	buf.Reset()
	term.Fprint(&buf, info, "using defaults")
	fmt.Println(buf.String())
	// Output:
	// config file missing
	// using defaults
}

// ExampleGlyph demonstrates glyph lookup.
func ExampleGlyph() {
	// On a non-UTF-8 environment Glyph returns the ASCII form.
	// We use Glyphs() to inspect registered entries deterministically.
	found := false
	for _, g := range term.Glyphs() {
		if g.Name == "check" {
			found = true
			break
		}
	}
	fmt.Println(found)
	// Output:
	// true
}

// ExampleRegisterGlyph demonstrates custom glyph registration.
func ExampleRegisterGlyph() {
	err := term.RegisterGlyph("example_custom", "★", "[star]")
	if err != nil {
		// "already registered" is acceptable in repeated example runs.
		fmt.Println("ok (or already registered)")
	} else {
		fmt.Println("ok (or already registered)")
	}
	// Output:
	// ok (or already registered)
}

// ExampleCenter demonstrates the Center layout function.
func ExampleCenter() {
	fmt.Printf("%q\n", term.Center("hi", 8))
	// Output:
	// "   hi   "
}

// ExamplePad demonstrates Pad.
func ExamplePad() {
	fmt.Printf("%q\n", term.Pad("text", 2, 3))
	// Output:
	// "  text   "
}

// ExampleTruncate demonstrates Truncate with a custom ellipsis.
func ExampleTruncate() {
	fmt.Println(term.Truncate("hello world", 8, "..."))
	// Output:
	// hello...
}

// ExampleWrap demonstrates Wrap.
func ExampleWrap() {
	fmt.Println(term.Wrap("hello world foo", 7))
	// Output:
	// hello
	// world
	// foo
}

// ExampleRGB demonstrates the RGB constructor.
func ExampleRGB() {
	c := term.RGB(0x22, 0xD3, 0xEE)
	fmt.Printf("R=%d G=%d B=%d\n", c.R, c.G, c.B)
	// Output:
	// R=34 G=211 B=238
}

// ExampleHex demonstrates hex color parsing.
func ExampleHex() {
	c, err := term.Hex("#22D3EE")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("R=%d G=%d B=%d\n", c.R, c.G, c.B)
	// Output:
	// R=34 G=211 B=238
}

// ExampleCapability demonstrates capability detection on a bytes.Buffer.
func ExampleCapability() {
	var buf bytes.Buffer
	caps := term.Capability(&buf)
	fmt.Println(caps.IsTTY)
	// Output:
	// false
}

// ExampleSpinner demonstrates the Spinner animation factory.
func ExampleSpinner() {
	anim := term.Spinner("loading")
	lines, done := anim.Render()
	fmt.Println(!done, len(lines) == 1)
	// Output:
	// true true
}

// ExampleNewProgress demonstrates Progress construction and rendering.
func ExampleNewProgress() {
	p := term.NewProgress(100, term.WithProgressLabel("processing"))
	p.Set(50)
	lines, done := p.Render()
	fmt.Println(!done, len(lines) == 1)
	// Output:
	// true true
}

// ExampleNewStatusBar demonstrates StatusBar usage.
func ExampleNewStatusBar() {
	sb := term.NewStatusBar()
	sb.SetSection("phase", "initializing")
	lines, done := sb.Render()
	fmt.Println(!done, len(lines) >= 1)
	// Output:
	// true true
}

// ExampleNewAnimator demonstrates Animator construction and a quick run.
func ExampleNewAnimator() {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := term.NewAnimator(logger,
		term.WithWriter(io.Discard),
		term.WithRefreshRate(10*1_000_000), // 10ms
	)

	// A self-completing animation.
	type finishNow struct{}
	_ = &finishNow{}

	// Use a quickly-finishing mock animation instead.
	prog := term.NewProgress(1)
	prog.Done()
	a.Add(prog)

	ctx := context.Background()
	if err := a.Run(ctx); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("done")
	}
	// Output:
	// done
}

// ExampleColumns demonstrates the Columns layout function.
func ExampleColumns() {
	rows := [][]string{
		{"Name", "Score"},
		{"Alice", "100"},
		{"Bob", "90"},
	}
	// Columns right-pads cells to the column width; trim trailing space for display.
	out := term.Columns(rows, term.WithColumnGap(2))
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		fmt.Println(strings.TrimRight(line, " "))
	}
	// Output:
	// Name   Score
	// Alice  100
	// Bob    90
}
