// SPDX-License-Identifier: Apache-2.0

package term_test

import (
	"bytes"
	"testing"

	"github.com/nathanbrophy/glacier/term"
)

// BenchmarkStyleRender measures the hot path for Style.Render.
// Target: ≤ 100 ns/op + 1 alloc per §23.13.
func BenchmarkStyleRender(b *testing.B) {
	s := term.New().Foreground(term.Cyan).Bold()
	text := "benchmark text"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Render(text)
	}
}

// BenchmarkStyleFprint measures Fprint to a bytes.Buffer (non-TTY).
func BenchmarkStyleFprint(b *testing.B) {
	s := term.New().Foreground(term.Warning)
	var buf bytes.Buffer
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		term.Fprint(&buf, s, "text")
	}
}

// BenchmarkStyleRenderCacheHit measures the second Render (should hit cache).
func BenchmarkStyleRenderCacheHit(b *testing.B) {
	s := term.New().Foreground(term.Teal).Bold().Italic()
	text := "cached"
	// Warm the cache.
	_ = s.Render(text)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Render(text)
	}
}

// BenchmarkCapabilityCacheHit measures the second Capability call (must be zero-alloc).
func BenchmarkCapabilityCacheHit(b *testing.B) {
	w := &bytes.Buffer{}
	_ = term.Capability(w) // warm
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = term.Capability(w)
	}
}

// BenchmarkSpinnerFrame measures Spinner.Render.
// Target: ≤ 500 ns/op per §23.13.
func BenchmarkSpinnerFrame(b *testing.B) {
	anim := term.Spinner("loading")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = anim.Render()
	}
}

// BenchmarkProgressFrame measures Progress.Render.
// Target: ≤ 2 µs/op per §23.13.
func BenchmarkProgressFrame(b *testing.B) {
	p := term.NewProgress(1024*1024,
		term.WithProgressLabel("Downloading"),
		term.WithProgressShowSpeed(),
		term.WithProgressShowETA(),
		term.WithProgressShowBytes(),
	)
	p.Set(512 * 1024)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Render()
	}
}

// BenchmarkGlyphLookup measures the hot path for Glyph().
func BenchmarkGlyphLookup(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = term.Glyph("check")
	}
}

// BenchmarkWrap measures Wrap on a medium-length string.
func BenchmarkWrap(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs."
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = term.Wrap(text, 40)
	}
}

// BenchmarkBox measures Box rendering.
func BenchmarkBox(b *testing.B) {
	text := "Configuration validation failed.\nPort 8080 already in use."
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = term.Box(text, term.WithRoundedCorners(), term.WithTitle("WARNING"))
	}
}
