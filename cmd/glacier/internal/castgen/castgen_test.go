// SPDX-License-Identifier: Apache-2.0

package castgen_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/nathanbrophy/glacier/assert"
	"github.com/nathanbrophy/glacier/cmd/glacier/internal/castgen"
)

// fixedCast returns a deterministic Cast for tests. The output uses ANSI
// escapes that exercise every supported SGR family (bold, fg256, fg24bit,
// underline, dim, reset).
func fixedCast() castgen.Cast {
	const out = "" +
		"\x1b[1;38;5;87mʕ•ᴥ•ʔ\x1b[0m \x1b[38;5;87mglacier version\x1b[0m\n" +
		"\x1b[1;38;5;84mʕ⌐■-■ʔ\x1b[0m \x1b[38;5;84mglacier dev\x1b[0m\n" +
		"  \x1b[2mgo:    go1.26.0\x1b[0m\n" +
		"  \x1b[4mos:    windows/amd64\x1b[0m\n"
	return castgen.Cast{
		Scenario: castgen.Scenario{
			Name:  "fixture",
			Title: "test fixture",
			Cols:  60,
			Rows:  6,
		},
		Output:   []byte(out),
		Recorded: time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC),
	}
}

func TestWriteCast_HasV2Header(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := castgen.WriteCast(&buf, fixedCast())
	assert.NoError(t, err)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	assert.True(t, len(lines) == 2,
		"expected 2 lines (header + one event), got %d", len(lines))

	header := lines[0]
	assert.True(t, strings.Contains(header, `"version":2`),
		"header missing version=2: %s", header)
	assert.True(t, strings.Contains(header, `"width":60`),
		"header missing width=60: %s", header)
	assert.True(t, strings.Contains(header, `"timestamp":`),
		"header missing timestamp: %s", header)
}

func TestWriteCast_EventCarriesFullOutput(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := castgen.WriteCast(&buf, fixedCast())
	assert.NoError(t, err)

	out := buf.String()
	// Output is JSON-quoted in the event line; check a recognizable substring.
	assert.True(t, strings.Contains(out, "glacier version"),
		"event line missing captured output: %s", out)
}

func TestWriteSVG_ContainsColoredSpans(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := castgen.WriteSVG(&buf, fixedCast())
	assert.NoError(t, err)

	svg := buf.String()

	// Must be a valid SVG document.
	assert.True(t, strings.HasPrefix(svg, "<?xml"),
		"SVG should start with XML declaration")
	assert.True(t, strings.Contains(svg, "<svg "),
		"SVG should contain <svg> element")
	assert.True(t, strings.Contains(svg, "</svg>"),
		"SVG should be closed")

	// Must contain a colored span for the cyan kaomoji (256 palette index 87).
	// The hex color computed from index 87 should appear in a fill="" attr.
	assert.True(t, strings.Contains(svg, `fill="#`),
		"SVG should contain at least one colored fill attribute")

	// Bold + underline + dim styles should map to font-weight / text-decoration / opacity.
	assert.True(t, strings.Contains(svg, `font-weight="bold"`),
		"SVG missing font-weight=bold for SGR 1: %s", firstFew(svg, 1500))
	assert.True(t, strings.Contains(svg, `text-decoration="underline"`),
		"SVG missing text-decoration=underline for SGR 4")
	assert.True(t, strings.Contains(svg, `opacity="0.6"`),
		"SVG missing opacity=0.6 for SGR 2 (dim)")
}

func TestWriteSVG_NoANSIEscapesLeak(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := castgen.WriteSVG(&buf, fixedCast())
	assert.NoError(t, err)
	// The rendered SVG must not contain raw ANSI escape bytes; everything
	// should be translated into SVG attributes.
	assert.False(t, strings.ContainsRune(buf.String(), '\x1b'),
		"SVG should not contain raw ESC bytes")
}

func TestWriteSVG_HasTitleBarAndTrafficLights(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := castgen.WriteSVG(&buf, fixedCast())
	assert.NoError(t, err)

	svg := buf.String()
	assert.True(t, strings.Contains(svg, "test fixture"),
		"SVG missing scenario title")
	for _, c := range []string{"#ff605c", "#ffbd44", "#00ca4e"} {
		assert.True(t, strings.Contains(svg, c),
			"SVG missing traffic-light color %s", c)
	}
}

// Example demonstrates the canonical use of the package: capture + write
// both artifacts.
func Example() {
	c := castgen.Cast{
		Scenario: castgen.Scenario{Name: "demo", Title: "demo", Cols: 80, Rows: 10},
		Output:   []byte("\x1b[1mhello\x1b[0m world\n"),
	}
	var castBuf, svgBuf bytes.Buffer
	_ = castgen.WriteCast(&castBuf, c)
	_ = castgen.WriteSVG(&svgBuf, c)
	// Output:
}

// firstFew returns the first n bytes of s as a string for diagnostic logging.
func firstFew(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
