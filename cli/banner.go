// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	_ "embed"
	"io"
	"sync"

	"github.com/nathanbrophy/glacier/term"
)

//go:embed assets/banner.txt
var bannerRaw []byte

var (
	bannerOnce    sync.Once
	bannerPlain   []byte
	bannerColored []byte
)

func initBanner() {
	bannerOnce.Do(func() {
		bannerPlain = bytes.TrimRight(bannerRaw, "\n")
		// The colored version is rendered at write time via writeBanner.
		bannerColored = bannerPlain
	})
}

// writeBanner writes the banner to w, applying ANSI gradient color when the
// writer is a TTY with color support and NO_COLOR is not set.
func writeBanner(w io.Writer) {
	initBanner()
	caps := term.Capability(w)
	if caps.IsTTY && caps.SupportsColor >= term.Color24Bit && !caps.NoColorEnv {
		// Apply a simple cyan→blue gradient across lines.
		lines := bytes.Split(bannerColored, []byte("\n"))
		for i, line := range lines {
			// Cycle through a simple palette: cyan at top, deeper blue at bottom.
			r, g, b := gradientColor(i, len(lines))
			_, _ = w.Write([]byte("\x1b[38;2;" +
				itoa(r) + ";" + itoa(g) + ";" + itoa(b) + "m"))
			_, _ = w.Write(line)
			_, _ = w.Write([]byte("\x1b[0m\n"))
		}
	} else {
		_, _ = w.Write(bannerPlain)
		_, _ = w.Write([]byte("\n"))
	}
}

// gradientColor returns an RGB color for line i of total lines,
// interpolating from cyan (0,200,200) to blue (0,100,255).
func gradientColor(i, total int) (r, g, b int) {
	if total <= 1 {
		return 0, 200, 200
	}
	t := float64(i) / float64(total-1)
	r = 0
	g = int(200 - t*100)
	b = int(200 + t*55)
	return r, g, b
}

// itoa is a minimal int-to-string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [10]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
